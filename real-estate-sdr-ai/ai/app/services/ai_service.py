from typing import Any
from ai.app.agent.knowledge_agent import KnowledgeAgent
from ai.app.agent.agent_type import AgentType
from ai.app.agent.agent_router import AgentRouter
from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)

from ai.app.agent.state_machine import ConversationStateMachine
from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.llm.provider import LLMProvider
from ai.app.llm.response_parser import ResponseParser
from ai.app.models.conversation import Conversation
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.rag.rag_service import RAGService
from ai.app.repositories.conversation_state_repository import (
    ConversationStateRepository,
)
from ai.app.models.ai_response import (
    AIResponse,
    LeadIntent,
    LeadStatus,
)


class AIService:
    """
    Serviço responsável por orquestrar o atendimento do SDR.

    Responsabilidades:

    - recuperar o lead;
    - recuperar o estado da conversa;
    - interpretar a mensagem através do LLM;
    - consultar a base de conhecimento via RAG;
    - atualizar a qualificação;
    - executar a máquina de estados;
    - buscar imóveis quando necessário;
    - gerar recomendações;
    - persistir o estado;
    - persistir mensagens;
    - retornar o resultado estruturado para a API.
    """

    def __init__(
            self,
            api_client: RealEstateAPIClient,
            llm_provider: LLMProvider,
            prompt_builder: PromptBuilder,
            response_parser: ResponseParser,
            state_repository: ConversationStateRepository,
            state_machine: ConversationStateMachine,
            rag_service: RAGService,
            agent_router: AgentRouter,
            knowledge_agent: KnowledgeAgent,
    ):
        self.api_client = api_client
        self.llm_provider = llm_provider
        self.prompt_builder = prompt_builder
        self.response_parser = response_parser
        self.state_repository = state_repository
        self.state_machine = state_machine
        self.rag_service = rag_service
        self.agent_router = agent_router
        self.knowledge_agent = knowledge_agent

    async def generate_response(
        self,
        conversation: Conversation,
    ) -> dict[str, Any]:

        conversation_id = conversation.id
        lead_id = conversation.lead_id

        if not conversation_id:
            raise ValueError(
                "conversation id cannot be empty"
            )

        if not lead_id:
            raise ValueError(
                "lead id cannot be empty"
            )

        lead_message = conversation.last_lead_message

        if lead_message is None:
            raise ValueError(
                "conversation must contain at least one lead message"
            )

        # ==============================================================
        # 1.5. Identificar o agente responsável
        # ==============================================================

        agent_type = self.agent_router.route(
            lead_message.content
        )

        print(
            f"[AGENT ROUTER] "
            f"message='{lead_message.content}' "
            f"agent='{agent_type.value}'"
        )

        # ==============================================================
        # 1.6. Buscar lead
        # ==============================================================

        lead = await self.api_client.get_lead(
            lead_id
        )

        if lead is None:
            raise ValueError(
                f"Lead not found: {lead_id}"
            )

        # ==============================================================
        # 2. Recuperar ou criar estado
        # ==============================================================

        state = await self._get_or_create_state(
            conversation.id
        )

        if (
            state.state == ConversationStateType.WAITING_LEAD
            and state.previous_state is not None
        ):
            state = self.state_machine.resume_after_follow_up(
                state
            )

        # ==============================================================
        # 3. Persistir mensagem do lead
        # ==============================================================

        # await self._persist_lead_message_if_needed(
        #     conversation_id=conversation_id,
        #     lead_message=lead_message,
        # )

        # ==============================================================
        # 4. Buscar propriedades somente quando necessário
        # ==============================================================

        properties: list[dict[str, Any]] = []

        # Neste primeiro LLM ainda não temos o novo estado calculado.
        # Portanto não fazemos matching aqui.
        #
        # O matching acontece depois que o LLM extrair os dados
        # da mensagem atual e a State Machine determinar MATCH_PROPERTY.

        # ==============================================================
        # 5. Knowledge Agent
        # ==============================================================

        knowledge: list[dict[str, Any]] = []

        if agent_type == AgentType.KNOWLEDGE:
            from ai.app.infrastructure.database import AsyncSessionLocal

            async with AsyncSessionLocal() as session:
                knowledge_response = await self.knowledge_agent.execute(
                    rag_service=self.rag_service,
                    session=session,
                    question=lead_message.content,
                )

            print("\n========== KNOWLEDGE AGENT ==========")
            print(knowledge_response)
            print("=====================================\n")

            response = AIResponse(
                message=knowledge_response.message,
                intent=LeadIntent.GENERAL_INQUIRY,
                lead_status=LeadStatus.CONTACTED,
                qualification_score=state.qualification_score,
                next_action=None,
                reasoning=None,
                metadata={},
                transaction_type=knowledge_response.transaction_type,
                qualification={},
            )

            await self.state_repository.save(state)

            return self._build_response(
                state=state,
                response=response,
                matched_properties=[],
            )

        # ==============================================================
        # 6. Construir prompt
        # ==============================================================

        prompt = self.prompt_builder.build(
            lead=lead,
            conversation=conversation,
            properties=properties,
            matched_properties=[],
            state=state,
            knowledge=knowledge,
        )

        # ==============================================================
        # 7. Executar LLM
        # ==============================================================

        llm_response = await self.llm_provider.generate_response(
            prompt
        )

        print("\n========== LLM RESPONSE ==========")
        print(llm_response)
        print("==================================\n")

        # ==============================================================
        # 8. Parsear resposta
        # ==============================================================

        response = self.response_parser.parse(
            llm_response
        )

        # ==============================================================
        # 9. Atualizar estado com dados extraídos
        # ==============================================================

        state = self._update_state_from_llm(
            state=state,
            response=response,
        )

        # ==============================================================
        # 10. State Machine
        # ==============================================================

        state = self.state_machine.transition(
            state
        )

        response = self._enforce_next_action_message(
            response=response,
            state=state,
        )

        # ==============================================================
        # 11. Matching de imóveis
        # ==============================================================

        matched_properties: list[dict[str, Any]] = []

        if state.next_action == NextAction.MATCH_PROPERTY:

            matched_properties = await self._match_properties(
                state
            )

            if matched_properties:

                state = self._apply_property_match(
                    state=state,
                    matched_properties=matched_properties,
                )

                state = self.state_machine.transition(
                    state
                )

        # ==============================================================
        # 12. Recomendação
        # ==============================================================

        if state.next_action == NextAction.RECOMMEND_PROPERTY:

            # A State Machine pode chegar diretamente em
            # RECOMMEND_PROPERTY sem passar por MATCH_PROPERTY.
            #
            # Nesse caso, matched_properties ainda estará vazio.
            # Como a aplicação é responsável pelo matching,
            # garantimos que os imóveis sejam buscados aqui.

            if not matched_properties:

                matched_properties = await self._match_properties(
                    state
                )

            if matched_properties:

                # Garante que o primeiro imóvel encontrado
                # seja registrado como imóvel selecionado.
                state = self._apply_property_match(
                    state=state,
                    matched_properties=matched_properties,
                )

                response = await self._generate_property_recommendation(
                    lead=lead,
                    conversation=conversation,
                    state=state,
                    properties=matched_properties,
                )

        # ==============================================================
        # 13. Caso não tenha encontrado imóveis
        # ==============================================================

        if (
            state.next_action == NextAction.MATCH_PROPERTY
            and not matched_properties
        ):
            response = self._build_no_property_response(
                response=response,
                state=state,
            )

        # ==============================================================
        # 14. Persistir estado
        # ==============================================================

        await self.state_repository.save(
            state
        )

        # ==============================================================
        # 15. Persistir resposta do assistant
        # ==============================================================

        # await self._persist_assistant_message(
        #     conversation_id=conversation_id,
        #     response=response,
        # )

        # ==============================================================
        # 16. Retornar
        # ==============================================================

        return self._build_response(
            state=state,
            response=response,
            matched_properties=matched_properties,
        )

    # ==================================================================
    # STATE
    # ==================================================================

    async def _get_or_create_state(
        self,
        conversation_id: str,
    ) -> ConversationState:

        state = await self.state_repository.get(
            conversation_id
        )

        if state is not None:
            return state

        return ConversationState(
            conversation_id=conversation_id
        )

    # ==================================================================
    # LLM -> STATE
    # ==================================================================

    def _update_state_from_llm(
        self,
        state: ConversationState,
        response: AIResponse,
    ) -> ConversationState:

        qualification = response.qualification or {}

        # --------------------------------------------------------------
        # Intent
        # --------------------------------------------------------------

        intent = state.intent

        if response.intent is not None:
            intent = self._enum_value(
                response.intent
            )

        # --------------------------------------------------------------
        # Transaction type
        # --------------------------------------------------------------

        transaction_type = state.transaction_type

        if response.transaction_type:
            transaction_type = response.transaction_type

        # --------------------------------------------------------------
        # Flow
        # --------------------------------------------------------------

        flow = state.flow

        if transaction_type == "buy":
            flow = ConversationFlow.BUY

        elif transaction_type == "investment":
            flow = ConversationFlow.INVESTMENT

        # --------------------------------------------------------------
        # Qualification
        # --------------------------------------------------------------

        region = self._value_or_current(
            qualification,
            "region",
            state.region,
        )

        budget = self._value_or_current(
            qualification,
            "budget",
            state.budget,
        )

        bedrooms = self._value_or_current(
            qualification,
            "bedrooms",
            state.bedrooms,
        )

        urgency = self._value_or_current(
            qualification,
            "urgency",
            state.urgency,
        )

        property_type = self._value_or_current(
            qualification,
            "property_type",
            state.property_type,
        )

        investor_profile = self._value_or_current(
            qualification,
            "investor_profile",
            state.investor_profile,
        )

        investment_ticket = self._value_or_current(
            qualification,
            "investment_ticket",
            state.investment_ticket,
        )

        expected_return = self._value_or_current(
            qualification,
            "expected_return",
            state.expected_return,
        )

        # --------------------------------------------------------------
        # Consistência entre fluxo e campos financeiros
        # --------------------------------------------------------------

        # "budget" pertence ao fluxo de compra.
        # "investment_ticket" pertence ao fluxo de investimento.
        #
        # O LLM pode interpretar corretamente o valor e ainda assim
        # preencher os dois campos. A aplicação é a autoridade final
        # sobre o estado, portanto removemos a ambiguidade aqui.

        if flow == ConversationFlow.INVESTMENT:
            budget = None

        elif flow == ConversationFlow.BUY:
            investment_ticket = None

        # --------------------------------------------------------------
        # Qualification score
        # --------------------------------------------------------------

        qualification_score = state.qualification_score

        if response.qualification_score is not None:
            qualification_score = (
                response.qualification_score
            )

        # --------------------------------------------------------------
        # Selected property
        # --------------------------------------------------------------

        selected_property_id = (
            state.selected_property_id
        )

        if response.property_id:
            selected_property_id = (
                response.property_id
            )

        return self._replace_state(
            state,
            flow=flow,
            intent=intent,
            transaction_type=transaction_type,
            property_type=property_type,
            region=region,
            budget=budget,
            bedrooms=bedrooms,
            urgency=urgency,
            investor_profile=investor_profile,
            investment_ticket=investment_ticket,
            expected_return=expected_return,
            qualification_score=qualification_score,
            selected_property_id=selected_property_id,
        )

    @staticmethod
    def _value_or_current(
        qualification: dict[str, Any],
        key: str,
        current_value: Any,
    ) -> Any:

        if key not in qualification:
            return current_value

        value = qualification[key]

        if value is None:
            return current_value

        if isinstance(value, str) and not value.strip():
            return current_value

        return value

    def _enforce_next_action_message(
        self,
        response: AIResponse,
        state: ConversationState,
    ) -> AIResponse:

        messages = {
            NextAction.ASK_BUDGET: (
                "Qual é o orçamento que você pretende investir na compra do imóvel?"
            ),
            NextAction.ASK_BEDROOMS: (
                "Quantos quartos você gostaria que o apartamento tivesse?"
            ),
            NextAction.ASK_LOCATION: (
                "Em qual região ou bairro você gostaria de procurar o imóvel?"
            ),
            NextAction.ASK_URGENCY: (
                "Quando você pretende realizar a compra?"
            ),
        }

        message = messages.get(
            state.next_action
        )

        if not message:
            return response

        return AIResponse(
            message=message,
            intent=response.intent,
            lead_status=response.lead_status,
            qualification_score=response.qualification_score,
            next_action=response.next_action,
            reasoning=response.reasoning,
            metadata=response.metadata,
            transaction_type=response.transaction_type,
            qualification=response.qualification,
        )

    # ==================================================================
    # PROPERTY MATCHING
    # ==================================================================

    async def _match_properties(
        self,
        state: ConversationState,
    ) -> list[dict[str, Any]]:

        properties = await self.api_client.get_properties()

        print("\n========== PROPERTY MATCH ==========")
        print(f"Total properties: {len(properties)}")
        print(f"property_type: {state.property_type}")
        print(f"budget: {state.budget}")
        print(f"bedrooms: {state.bedrooms}")
        print(f"region: {state.region}")

        matched: list[dict[str, Any]] = []

        for property_data in properties:

            property_id = property_data.get("id")
            title = property_data.get("title")

            available = self._is_available(
                property_data
            )

            property_type = self._matches_property_type(
                property_data,
                state.property_type,
            )

            budget = self._matches_budget(
                property_data,
                state.budget,
            )

            bedrooms = self._matches_bedrooms(
                property_data,
                state.bedrooms,
            )

            region = self._matches_region(
                property_data,
                state.region,
            )

            print(
                f"\nProperty: {title}"
                f"\n  id: {property_id}"
                f"\n  available: {available}"
                f"\n  property_type: {property_type}"
                f"\n  budget: {budget}"
                f"\n  bedrooms: {bedrooms}"
                f"\n  region: {region}"
            )

            if not available:
                continue

            if not property_type:
                continue

            if not budget:
                continue

            if not bedrooms:
                continue

            if not region:
                continue

            matched.append(
                property_data
            )

        matched.sort(
            key=self._property_price
        )

        result = matched[:3]

        print("\n========== MATCHED ==========")

        for property_data in result:
            print(
                property_data.get("title"),
                property_data.get("price"),
            )

        print("=============================\n")

        return result

    def _apply_property_match(
        self,
        state: ConversationState,
        matched_properties: list[dict[str, Any]],
    ) -> ConversationState:

        if not matched_properties:
            return state

        property_id = self._property_id(
            matched_properties[0]
        )

        if not property_id:
            return state

        return self._replace_state(
            state,
            selected_property_id=property_id,
        )

    # ==================================================================
    # PROPERTY FILTERS
    # ==================================================================

    @staticmethod
    def _is_available(
        property_data: dict[str, Any],
    ) -> bool:

        status = property_data.get("status")

        if status is not None:
            return str(status).strip().upper() == "AVAILABLE"

        available = property_data.get(
            "available",
            True,
        )

        if isinstance(available, str):
            return available.strip().lower() in {
                "true",
                "1",
                "yes",
                "sim",
            }

        return bool(available)

    def _matches_property_type(
        self,
        property_data: dict[str, Any],
        property_type: str | None,
    ) -> bool:

        if not property_type:
            return True

        actual = property_data.get("property_type")

        if actual is None:
            actual = property_data.get("type")

        if not actual:
            return False

        normalized_actual = self._normalize(
            str(actual)
        )

        normalized_requested = self._normalize(
            property_type
        )

        aliases = {
            "apartamento": {
                "apartamento",
                "apartment",
            },
            "casa": {
                "casa",
                "house",
            },
            "terreno": {
                "terreno",
                "land",
            },
            "sala comercial": {
                "sala comercial",
                "commercial room",
            },
        }

        allowed = aliases.get(
            normalized_requested
        )

        if allowed:
            return normalized_actual in allowed

        return normalized_actual == normalized_requested

    @staticmethod
    def _matches_budget(
        property_data: dict[str, Any],
        budget: float | None,
    ) -> bool:

        if budget is None:
            return True

        price = property_data.get(
            "price"
        )

        if price is None:
            return False

        try:
            return float(price) <= float(budget)

        except (TypeError, ValueError):
            return False

    @staticmethod
    def _matches_bedrooms(
        property_data: dict[str, Any],
        bedrooms: int | None,
    ) -> bool:

        if bedrooms is None:
            return True

        value = property_data.get(
            "bedrooms"
        )

        if value is None:
            return False

        try:
            return int(value) >= int(bedrooms)

        except (TypeError, ValueError):
            return False

    def _matches_region(
        self,
        property_data: dict[str, Any],
        region: str | None,
    ) -> bool:

        if not region:
            return True

        normalized_region = self._normalize(
            region
        )

        city = self._normalize(
            str(
                property_data.get(
                    "city",
                    "",
                )
            )
        )

        neighborhood = self._normalize(
            str(
                property_data.get(
                    "neighborhood",
                    "",
                )
            )
        )

        address = self._normalize(
            str(
                property_data.get(
                    "address",
                    "",
                )
            )
        )

        title = self._normalize(
            str(
                property_data.get(
                    "title",
                    "",
                )
            )
        )

        searchable = " ".join(
            [
                city,
                neighborhood,
                address,
                title,
            ]
        )

        if normalized_region in searchable:
            return True

        if normalized_region in {
            "zona sul",
            "sul",
        }:

            south_neighborhoods = {
                "moema",
                "vila mariana",
                "saude",
                "campo belo",
                "campo grande",
                "santo amaro",
                "ipiranga",
                "brooklin",
                "morumbi",
                "vila andrade",
                "chacara flora",
            }

            normalized_neighborhood = (
                self._normalize(
                    neighborhood
                )
            )

            return normalized_neighborhood in {
                self._normalize(
                    neighborhood_name
                )
                for neighborhood_name
                in south_neighborhoods
            }

        return False

    # ==================================================================
    # PROPERTY RECOMMENDATION
    # ==================================================================

    async def _generate_property_recommendation(
        self,
        lead: Any,
        conversation: Conversation,
        state: ConversationState,
        properties: list[dict[str, Any]],
    ) -> AIResponse:

        prompt = self.prompt_builder.build(
            lead=lead,
            conversation=conversation,
            properties=[],
            matched_properties=properties,
            state=state,
        )

        llm_response = await self.llm_provider.generate_response(
            prompt
        )

        print("\n========== LLM RECOMMENDATION ==========")
        print(llm_response)
        print("=========================================\n")

        return self.response_parser.parse(
            llm_response
        )

    # ==================================================================
    # NO PROPERTY RESPONSE
    # ==================================================================

    @staticmethod
    def _build_no_property_response(
        response: AIResponse,
        state: ConversationState,
    ) -> AIResponse:

        message = (
            "Não encontrei uma opção exatamente dentro "
            "dos critérios informados. Podemos ajustar o "
            "orçamento, a região ou a quantidade de quartos."
        )

        return AIResponse(
            message=message,
            intent=response.intent,
            lead_status=response.lead_status,
            qualification_score=response.qualification_score,
            next_action=response.next_action,
            reasoning=response.reasoning,
            metadata=response.metadata,
            transaction_type=response.transaction_type,
            qualification=response.qualification,
        )

    # ==================================================================
    # MESSAGE PERSISTENCE
    # ==================================================================

    async def _persist_lead_message_if_needed(
        self,
        conversation_id: str,
        lead_message: Any,
    ) -> None:

        existing_messages = (
            await self.api_client.get_messages(
                conversation_id
            )
        )

        for message in existing_messages:

            role = str(
                message.get(
                    "role",
                    "",
                )
            ).upper()

            content = str(
                message.get(
                    "content",
                    ""
                )
            )

            if (
                role == "USER"
                and content.strip()
                == lead_message.content.strip()
            ):
                return

        await self.api_client.create_message(
            conversation_id=conversation_id,
            role="USER",
            content=lead_message.content,
        )

    async def _persist_assistant_message(
        self,
        conversation_id: str,
        response: AIResponse,
    ) -> None:

        await self.api_client.create_message(
            conversation_id=conversation_id,
            role="ASSISTANT",
            content=response.message,
        )

    # ==================================================================
    # RESPONSE
    # ==================================================================

    def _build_response(
        self,
        state: ConversationState,
        response: AIResponse,
        matched_properties: list[dict[str, Any]],
    ) -> dict[str, Any]:

        return {
            "message": response.message,

            "intent": self._enum_value(
                response.intent
            ),

            "lead_status": self._enum_value(
                response.lead_status
            ),

            "qualification_score": (
                response.qualification_score
                if response.qualification_score is not None
                else state.qualification_score
            ),

            "flow": self._enum_value(
                state.flow
            ),

            "state": self._enum_value(
                state.state
            ),

            "next_action": self._enum_value(
                state.next_action
            ),

            "qualification": {
                "region": state.region,
                "budget": state.budget,
                "bedrooms": state.bedrooms,
                "urgency": state.urgency,
                "property_type": state.property_type,
                "investor_profile": state.investor_profile,
                "investment_ticket": state.investment_ticket,
                "expected_return": state.expected_return,
            },

            "selected_property_id": (
                state.selected_property_id
            ),

            "matched_properties": matched_properties,
        }

    # ==================================================================
    # HELPERS
    # ==================================================================

    @staticmethod
    def _property_id(
        property_data: dict[str, Any],
    ) -> str | None:

        value = property_data.get(
            "id"
        )

        if value is None:
            return None

        return str(value)

    @staticmethod
    def _property_price(
        property_data: dict[str, Any],
    ) -> float:

        value = property_data.get(
            "price",
            0,
        )

        try:
            return float(value)

        except (TypeError, ValueError):
            return 0.0

    @staticmethod
    def _normalize(
        value: str,
    ) -> str:

        import unicodedata

        normalized = unicodedata.normalize(
            "NFD",
            value.lower().strip(),
        )

        return "".join(
            character
            for character in normalized
            if unicodedata.category(character) != "Mn"
        )

    @staticmethod
    def _enum_value(
        value: Any,
    ) -> Any:

        if value is None:
            return None

        return getattr(
            value,
            "value",
            value,
        )

    @staticmethod
    def _replace_state(
        state: ConversationState,
        **changes: Any,
    ) -> ConversationState:

        from dataclasses import replace

        return replace(
            state,
            **changes,
        )