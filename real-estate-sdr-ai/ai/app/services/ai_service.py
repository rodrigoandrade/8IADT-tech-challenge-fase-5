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

from ai.app.agent.property_agent import PropertyAgent

import re

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
            property_agent: PropertyAgent,
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
        self.property_agent = property_agent

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
            lead_message=lead_message.content,
        )

        # ==============================================================
        # 10. State Machine
        # ==============================================================

        state = self.state_machine.transition(
            state
        )

        print("\n========== STATE AFTER TRANSITION ==========")
        print("state.state =", state.state)
        print("state.next_action =", state.next_action)
        print("state.budget =", state.budget)
        print("state.bedrooms =", state.bedrooms)
        print("state.region =", state.region)
        print("state.urgency =", state.urgency)
        print("============================================\n")

        response = self._enforce_next_action_message(
            response=response,
            state=state,
        )

        # ==============================================================
        # 11. Matching de imóveis
        # ==============================================================

        matched_properties: list[dict[str, Any]] = []

        if state.next_action == NextAction.MATCH_PROPERTY:

            property_response = await self.property_agent.execute(
                state=state
            )

            matched_properties = property_response.matched_properties

            if matched_properties:
                state = self._apply_property_match(
                    state=state,
                    matched_properties=matched_properties,
                )

                state = self.state_machine.transition(
                    state
                )


        # ==============================================================
        # 12. Matching e recomendação de imóveis
        # ==============================================================

        if state.next_action in {
            NextAction.MATCH_PROPERTY,
            NextAction.RECOMMEND_PROPERTY,
        }:
            property_response = await self.property_agent.execute(
                state=state
            )

            matched_properties = (
                property_response.matched_properties or []
            )

            if matched_properties:
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
            else:
                # A mensagem final depende do resultado real
                # do matching, não do texto produzido pelo LLM.
                response = self._build_no_property_response(
                    response=response,
                    state=state,
                )


        # ==============================================================
        # 13. Caso não tenha encontrado imóveis
        # ==============================================================

        # O resultado real do matching prevalece sobre o texto do LLM.
        # Trata tanto o caminho MATCH_PROPERTY quanto o caminho em que a
        # máquina de estados já chegou diretamente a RECOMMEND_PROPERTY.
        if (
            state.next_action in (
                NextAction.MATCH_PROPERTY,
                NextAction.RECOMMEND_PROPERTY,
            )
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
        lead_message: str | None = None,
    ) -> ConversationState:
        """Atualiza apenas os critérios informados, preservando os demais."""
        qualification = dict(response.qualification or {})
        original_state = state
        raw_message = (lead_message or "").strip()
        message = raw_message.casefold()

        property_patterns = (
            (r"\bapartamentos?\b", "apartamento"),
            (r"\bcasas?\b", "casa"),
            (r"\bstudios?\b", "studio"),
            (r"\bkitnets?\b", "kitnet"),
            (r"\bterrenos?\b", "terreno"),
        )

        # A mensagem do usuário prevalece sobre qualquer valor inferido pelo LLM.
        explicit_property_type = next(
            (value for pattern, value in property_patterns
             if re.search(pattern, message, re.IGNORECASE)),
            None,
        )

        # Só extrai urgência como resposta direta quando a pergunta anterior
        # era sobre prazo; isso evita interpretar qualquer número como urgência.
        if state.next_action == NextAction.ASK_URGENCY:
            urgency_match = re.search(
                r"\b(\d+)\s*(dias?|semanas?|meses?)\b",
                message,
                re.IGNORECASE,
            )
            if urgency_match:
                qualification["urgency"] = urgency_match.group(0).strip()
            elif re.search(
                r"\b(urgente|imediatamente|imediato|o quanto antes|"
                r"assim que possível|agora)\b",
                message,
                re.IGNORECASE,
            ):
                qualification["urgency"] = raw_message

        # Um pedido de mudança de tipo atualiza o tipo e invalida a seleção
        # anterior, mas não apaga orçamento, região ou quartos já informados.
        property_type_changed = bool(
            explicit_property_type
            and explicit_property_type != state.property_type
        )
        if explicit_property_type:
            qualification["property_type"] = explicit_property_type

        intent = state.intent
        if response.intent is not None:
            intent = self._enum_value(response.intent)

        transaction_type = state.transaction_type
        if response.transaction_type:
            transaction_type = response.transaction_type

        flow = state.flow
        if transaction_type == "buy":
            flow = ConversationFlow.BUY
        elif transaction_type == "investment":
            flow = ConversationFlow.INVESTMENT

        region = self._value_or_current(qualification, "region", state.region)
        budget = self._value_or_current(qualification, "budget", state.budget)
        bedrooms = self._value_or_current(qualification, "bedrooms", state.bedrooms)
        urgency = self._value_or_current(qualification, "urgency", state.urgency)
        property_type = self._value_or_current(
            qualification, "property_type", state.property_type
        )
        investor_profile = self._value_or_current(
            qualification, "investor_profile", state.investor_profile
        )
        investment_ticket = self._value_or_current(
            qualification, "investment_ticket", state.investment_ticket
        )
        expected_return = self._value_or_current(
            qualification, "expected_return", state.expected_return
        )

        if explicit_property_type:
            property_type = explicit_property_type

        if raw_message:
            bedroom_match = re.search(
                r"\b(\d+)\s*(?:quartos?|dormit[oó]rios?)\b",
                message,
                re.IGNORECASE,
            )
            if bedroom_match:
                bedrooms = int(bedroom_match.group(1))

            # Remove a expressão idiomática "na verdade" antes de procurar
            # preposições de localização; ela não é um bairro.
            location_text = re.sub(
                r"\bna\s+verdade\b", " ", raw_message,
                flags=re.IGNORECASE,
            )
            region_match = re.search(
                r"\b(?:em|no|na|bairro)\s+"
                r"(?:regi[aã]o\s+de\s+)?"
                r"([a-záàâãéèêíïóôõöúçñ0-9]"
                r"[a-záàâãéèêíïóôõöúçñ0-9\s'-]*?)"
                r"(?=\s*,|\s*\.|\s*;|\s*\?|"
                r"\s+até\b|\s+com\b|\s+por\b|"
                r"\s+e\b|$)",
                location_text,
                re.IGNORECASE,
            )
            if region_match:
                extracted_region = region_match.group(1).strip(" -,'\"?!")
                # Impede palavras conversacionais de serem gravadas como bairro.
                if extracted_region and extracted_region.casefold() not in {
                    "verdade", "realidade", "seguida", "frente", "casa"
                }:
                    region = extracted_region

            budget_match = re.search(
                r"(?:até|no máximo|máximo de|"
                r"orçamento(?:\s+agora)?(?:\s+é)?(?:\s+de)?)"
                r"\s*R?\$?\s*"
                r"([\d.]+(?:,[\d]{1,2})?)\s*(mil)?",
                message,
                re.IGNORECASE,
            )
            if budget_match:
                normalized_budget = (
                    budget_match.group(1).replace(".", "").replace(",", ".")
                )
                budget = float(normalized_budget)
                if budget_match.group(2):
                    budget *= 1000

        if flow == ConversationFlow.INVESTMENT:
            budget = None
        elif flow == ConversationFlow.BUY:
            investment_ticket = None

        criteria_changed = (
            region != original_state.region
            or budget != original_state.budget
            or bedrooms != original_state.bedrooms
            or property_type != original_state.property_type
        )

        qualification_score = state.qualification_score
        if response.qualification_score is not None:
            qualification_score = response.qualification_score

        # Nunca mantém o imóvel selecionado quando a busca mudou.
        selected_property_id = state.selected_property_id
        if criteria_changed or property_type_changed:
            selected_property_id = None
        elif response.property_id:
            selected_property_id = response.property_id

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
    # ==================================================================

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
        property_type = state.property_type or "imóvel"
        region = state.region
        budget = state.budget
        bedrooms = state.bedrooms

        criteria = []

        if property_type:
            criteria.append(f"tipo: {property_type}")

        if region:
            criteria.append(f"região: {region}")

        if bedrooms is not None:
            criteria.append(f"pelo menos {bedrooms} quartos")

        if budget is not None:
            criteria.append(f"até R$ {budget:,.2f}")

        if criteria:
            criteria_text = "; ".join(criteria)
            message = (
                "Não encontrei imóveis disponíveis que atendam "
                f"aos critérios atuais ({criteria_text}). "
                "Podemos ajustar algum dos filtros para ampliar "
                "as opções."
            )
        else:
            message = (
                "Não encontrei imóveis disponíveis para os "
                "critérios informados. Podemos ajustar a busca."
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
            qualification={
                "region": state.region,
                "budget": state.budget,
                "bedrooms": state.bedrooms,
                "urgency": state.urgency,
                "property_type": state.property_type,
                "investor_profile": state.investor_profile,
                "investment_ticket": state.investment_ticket,
                "expected_return": state.expected_return,
            },
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
