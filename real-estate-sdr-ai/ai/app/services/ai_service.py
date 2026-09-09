import json
import unicodedata

from dataclasses import replace
from typing import Any

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
from ai.app.models.conversation import (
    Conversation,
    ConversationMessage,
    MessageRole,
)
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.repositories.conversation_state_repository import (
    ConversationStateRepository,
)


class AIService:

    SOUTH_ZONE_NEIGHBORHOODS = {
        "moema",
        "vila mariana",
        "saude",
        "campo belo",
        "campo grande",
        "santo amaro",
        "ipiranga",
        "brooklin",
        "brooklin paulista",
        "morumbi",
        "vila andrade",
        "chacara flora",
        "jardim paulista",
        "jardim europa",
        "itaim bibi",
    }

    def __init__(
        self,
        api_client: RealEstateAPIClient,
        llm_provider: LLMProvider,
        prompt_builder: PromptBuilder,
        response_parser: ResponseParser,
        state_repository: ConversationStateRepository,
        state_machine: ConversationStateMachine,
    ):
        self.api_client = api_client
        self.llm_provider = llm_provider
        self.prompt_builder = prompt_builder
        self.response_parser = response_parser
        self.state_repository = state_repository
        self.state_machine = state_machine

    # =========================================================
    # PUBLIC
    # =========================================================

    async def generate_response(
        self,
        conversation: Conversation,
    ) -> dict[str, Any]:

        conversation_id = str(
            conversation.id
        )

        # ---------------------------------------------------------
        # 1. Validação
        # ---------------------------------------------------------

        if not conversation_id.strip():
            raise ValueError(
                "conversation id cannot be empty"
            )

        if not conversation.lead_id.strip():
            raise ValueError(
                "lead id cannot be empty"
            )

        current_lead_message = (
            self._get_current_lead_message(
                conversation
            )
        )

        if current_lead_message is None:
            raise ValueError(
                "conversation must contain a lead message"
            )

        # ---------------------------------------------------------
        # 2. Buscar contexto do backend
        # ---------------------------------------------------------

        lead = await self.api_client.get_lead(
            conversation.lead_id
        )

        persisted_messages = (
            await self.api_client.get_messages(
                conversation_id
            )
        )

        # ---------------------------------------------------------
        # 3. Montar histórico completo
        # ---------------------------------------------------------

        history = self._build_conversation_history(
            conversation=conversation,
            persisted_messages=persisted_messages,
        )

        conversation_with_history = replace(
            conversation,
            messages=history,
        )

        # ---------------------------------------------------------
        # 4. Persistir mensagem do lead caso necessário
        # ---------------------------------------------------------

        await self._persist_lead_message_if_needed(
            conversation_id=conversation_id,
            message=current_lead_message,
            persisted_messages=persisted_messages,
        )

        # ---------------------------------------------------------
        # 5. Recuperar estado persistido
        # ---------------------------------------------------------

        state = await self.state_repository.get(
            conversation_id
        )

        if state is None:
            state = ConversationState(
                conversation_id=conversation_id,
                state=ConversationStateType.NEW,
            )

        # ---------------------------------------------------------
        # 6. Buscar imóveis
        #
        # Neste momento enviamos os imóveis ao LLM somente como
        # contexto. A escolha definitiva será feita pelo matching
        # determinístico abaixo.
        # ---------------------------------------------------------

        properties = (
            await self.api_client.get_properties()
        )

        # ---------------------------------------------------------
        # 7. Construir prompt
        # ---------------------------------------------------------

        prompt = self.prompt_builder.build(
            lead=lead,
            conversation=conversation_with_history,
            properties=properties,
        )

        # ---------------------------------------------------------
        # 8. LLM interpreta a conversa
        # ---------------------------------------------------------

        raw_response = (
            await self.llm_provider.generate_response(
                prompt
            )
        )

        # ---------------------------------------------------------
        # 9. Parse da resposta
        # ---------------------------------------------------------

        response = self.response_parser.parse(
            raw_response
        )

        # ---------------------------------------------------------
        # 10. Atualizar estado
        # ---------------------------------------------------------

        state = self._update_state_from_llm(
            state=state,
            response=response,
        )

        # ---------------------------------------------------------
        # 11. State Machine
        # ---------------------------------------------------------

        state = self.state_machine.transition(
            state
        )

        # ---------------------------------------------------------
        # 12. Matching determinístico
        # ---------------------------------------------------------

        matched_properties = []

        if state.next_action == NextAction.MATCH_PROPERTY:

            matched_properties = self._match_properties(
                state=state,
                properties=properties,
            )

            if matched_properties:

                selected_property_id = str(
                    matched_properties[0]["id"]
                )

                state = replace(
                    state,
                    state=ConversationStateType.RECOMMEND_PROPERTIES,
                    next_action=NextAction.RECOMMEND_PROPERTY,
                    selected_property_id=selected_property_id,
                )

            else:

                state = replace(
                    state,
                    state=ConversationStateType.RECOMMEND_PROPERTIES,
                    next_action=NextAction.RECOMMEND_PROPERTY,
                    selected_property_id=None,
                )

        # ---------------------------------------------------------
        # 13. Persistir estado
        # ---------------------------------------------------------

        state = await self.state_repository.save(
            state
        )

        # ---------------------------------------------------------
        # 14. Gerar resposta final
        #
        # Se houver imóveis encontrados, fazemos uma segunda
        # chamada ao LLM contendo somente os imóveis compatíveis.
        # ---------------------------------------------------------

        if state.next_action == NextAction.RECOMMEND_PROPERTY:

            recommendation_prompt = (
                self.prompt_builder.build(
                    lead=lead,
                    conversation=conversation_with_history,
                    properties=matched_properties,
                )
            )

            recommendation_raw_response = (
                await self.llm_provider.generate_response(
                    recommendation_prompt
                )
            )

            response = self.response_parser.parse(
                recommendation_raw_response
            )

        # ---------------------------------------------------------
        # 15. Persistir resposta do assistente
        # ---------------------------------------------------------

        assistant_message = response.message

        if assistant_message:
            await self.api_client.create_message(
                conversation_id=conversation_id,
                role="ASSISTANT",
                content=assistant_message,
            )

        # ---------------------------------------------------------
        # 16. Retornar resposta
        # ---------------------------------------------------------

        return self._build_response(
            response=response,
            state=state,
            matched_properties=matched_properties,
        )

    # =========================================================
    # CONVERSATION HISTORY
    # =========================================================

    def _build_conversation_history(
        self,
        conversation: Conversation,
        persisted_messages: list[dict[str, Any]],
    ) -> list[ConversationMessage]:

        history: list[ConversationMessage] = []

        # ---------------------------------------------------------
        # Mensagens persistidas no backend
        # ---------------------------------------------------------

        for message in persisted_messages:

            converted = self._convert_api_message(
                message
            )

            if converted is not None:
                history.append(converted)

        # ---------------------------------------------------------
        # Mensagens recebidas nesta requisição
        # ---------------------------------------------------------

        for message in conversation.messages:

            if not self._message_exists(
                history,
                message,
            ):
                history.append(message)

        return history

    def _convert_api_message(
        self,
        message: dict[str, Any],
    ) -> ConversationMessage | None:

        role = message.get("role")
        content = message.get("content")

        if not content:
            return None

        role_mapping = {
            "USER": MessageRole.LEAD,
            "user": MessageRole.LEAD,
            "LEAD": MessageRole.LEAD,
            "lead": MessageRole.LEAD,

            "ASSISTANT": MessageRole.ASSISTANT,
            "assistant": MessageRole.ASSISTANT,

            "SYSTEM": MessageRole.SYSTEM,
            "system": MessageRole.SYSTEM,
        }

        mapped_role = role_mapping.get(
            str(role)
        )

        if mapped_role is None:
            return None

        return ConversationMessage(
            role=mapped_role,
            content=str(content),
        )

    @staticmethod
    def _message_exists(
        history: list[ConversationMessage],
        message: ConversationMessage,
    ) -> bool:

        return any(
            existing.role == message.role
            and existing.content.strip()
            == message.content.strip()
            for existing in history
        )

    # =========================================================
    # CURRENT MESSAGE
    # =========================================================

    @staticmethod
    def _get_current_lead_message(
        conversation: Conversation,
    ) -> ConversationMessage | None:

        return conversation.last_lead_message

    # =========================================================
    # MESSAGE PERSISTENCE
    # =========================================================

    async def _persist_lead_message_if_needed(
        self,
        conversation_id: str,
        message: ConversationMessage,
        persisted_messages: list[dict[str, Any]],
    ) -> None:

        if self._api_contains_message(
            persisted_messages,
            role="USER",
            content=message.content,
        ):
            return

        await self.api_client.create_message(
            conversation_id=conversation_id,
            role="USER",
            content=message.content,
        )

    @staticmethod
    def _api_contains_message(
        messages: list[dict[str, Any]],
        role: str,
        content: str,
    ) -> bool:

        expected_content = content.strip()

        for message in messages:

            message_role = str(
                message.get("role", "")
            ).upper()

            message_content = str(
                message.get("content", "")
            ).strip()

            if (
                message_role == role
                and message_content == expected_content
            ):
                return True

        return False

    # =========================================================
    # STATE UPDATE
    # =========================================================

    def _update_state_from_llm(
        self,
        state: ConversationState,
        response,
    ) -> ConversationState:

        metadata = response.metadata or {}

        # ---------------------------------------------------------
        # O PromptBuilder atual retorna qualification dentro do
        # JSON, mas o ResponseParser atual não preserva esse objeto
        # no AIResponse.
        #
        # Por isso, o qualification deve ser recuperado do contexto
        # quando necessário.
        #
        # Os campos que já existirem no state são preservados.
        # ---------------------------------------------------------

        state = replace(
            state,

            intent=self._enum_value(
                response.intent
            ) or state.intent,

            qualification_score=(
                response.qualification_score
                if response.qualification_score is not None
                else state.qualification_score
            ),
        )

        # ---------------------------------------------------------
        # transaction_type
        #
        # O ResponseParser não possui esse campo no AIResponse.
        # Pode existir no metadata caso venha do parser/prompt.
        # ---------------------------------------------------------

        transaction_type = metadata.get(
            "transaction_type"
        )

        if transaction_type:
            state = replace(
                state,
                transaction_type=transaction_type,
            )

        # ---------------------------------------------------------
        # Property ID
        # ---------------------------------------------------------

        property_id = metadata.get(
            "property_id"
        )

        if property_id:
            state = replace(
                state,
                selected_property_id=property_id,
            )

        # ---------------------------------------------------------
        # Resolver fluxo
        # ---------------------------------------------------------

        flow = self._resolve_flow(
            state
        )

        if flow is not None:
            state = replace(
                state,
                flow=flow,
            )

        return state

    # =========================================================
    # FLOW
    # =========================================================

    @staticmethod
    def _resolve_flow(
        state: ConversationState,
    ) -> ConversationFlow | None:

        if state.flow is not None:
            return state.flow

        transaction_type = (
            state.transaction_type or ""
        ).strip().lower()

        intent = (
            state.intent or ""
        ).strip().lower()

        if transaction_type in {
            "buy",
            "purchase",
            "compra",
            "comprar",
        }:
            return ConversationFlow.BUY

        if transaction_type in {
            "investment",
            "investimento",
            "invest",
        }:
            return ConversationFlow.INVESTMENT

        if intent in {
            "property_interest",
            "purchase",
            "buy",
            "compra",
        }:
            return ConversationFlow.BUY

        if intent in {
            "investment",
            "investimento",
        }:
            return ConversationFlow.INVESTMENT

        return None

    # =========================================================
    # PROPERTY MATCHING
    # =========================================================

    def _match_properties(
        self,
        state: ConversationState,
        properties: list[dict[str, Any]],
    ) -> list[dict[str, Any]]:

        matched = []

        for property_data in properties:

            if not self._is_available(
                property_data
            ):
                continue

            if not self._matches_property_type(
                state,
                property_data,
            ):
                continue

            if not self._matches_transaction_type(
                state,
                property_data,
            ):
                continue

            if not self._matches_budget(
                state,
                property_data,
            ):
                continue

            if not self._matches_bedrooms(
                state,
                property_data,
            ):
                continue

            if not self._matches_region(
                state,
                property_data,
            ):
                continue

            matched.append(
                property_data
            )

        matched.sort(
            key=self._property_price
        )

        return matched[:3]

    @staticmethod
    def _is_available(
        property_data: dict[str, Any],
    ) -> bool:

        available = property_data.get(
            "available"
        )

        if available is not None:
            return bool(available)

        status = str(
            property_data.get(
                "status",
                "",
            )
        ).upper()

        return status in {
            "AVAILABLE",
            "ACTIVE",
        }

    def _matches_property_type(
        self,
        state: ConversationState,
        property_data: dict[str, Any],
    ) -> bool:

        if not state.property_type:
            return True

        expected = self._normalize(
            state.property_type
        )

        actual = self._normalize(
            property_data.get(
                "property_type",
                "",
            )
        )

        if not actual:
            return False

        return (
            expected in actual
            or actual in expected
        )

    def _matches_transaction_type(
        self,
        state: ConversationState,
        property_data: dict[str, Any],
    ) -> bool:

        if not state.transaction_type:
            return True

        transaction = self._normalize(
            state.transaction_type
        )

        actual = self._normalize(
            property_data.get(
                "transaction_type",
                "",
            )
        )

        if not actual:
            return True

        if transaction in {
            "buy",
            "purchase",
            "compra",
            "comprar",
        }:
            return actual in {
                "sale",
                "venda",
                "buy",
                "compra",
            }

        return True

    @staticmethod
    def _matches_budget(
        state: ConversationState,
        property_data: dict[str, Any],
    ) -> bool:

        if state.budget is None:
            return True

        price = AIService._property_price(
            property_data
        )

        return price <= state.budget

    @staticmethod
    def _matches_bedrooms(
        state: ConversationState,
        property_data: dict[str, Any],
    ) -> bool:

        if state.bedrooms is None:
            return True

        bedrooms = property_data.get(
            "bedrooms"
        )

        if bedrooms is None:
            return False

        try:
            return (
                int(bedrooms)
                >= state.bedrooms
            )
        except (
            TypeError,
            ValueError,
        ):
            return False

    def _matches_region(
        self,
        state: ConversationState,
        property_data: dict[str, Any],
    ) -> bool:

        if not state.region:
            return True

        requested_region = self._normalize(
            state.region
        )

        neighborhood = self._normalize(
            property_data.get(
                "neighborhood",
                "",
            )
        )

        city = self._normalize(
            property_data.get(
                "city",
                "",
            )
        )

        # ---------------------------------------------------------
        # Zona Sul
        # ---------------------------------------------------------

        if "zona sul" in requested_region:

            normalized_neighborhoods = {
                self._normalize(
                    item
                )
                for item
                in self.SOUTH_ZONE_NEIGHBORHOODS
            }

            return (
                neighborhood
                in normalized_neighborhoods
            )

        # ---------------------------------------------------------
        # Bairro / cidade
        # ---------------------------------------------------------

        return (
            requested_region in neighborhood
            or requested_region in city
            or neighborhood in requested_region
        )

    # =========================================================
    # RESPONSE
    # =========================================================

    def _build_response(
        self,
        response,
        state: ConversationState,
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
                state.qualification_score
            ),

            "flow": (
                state.flow.value
                if state.flow
                else None
            ),

            "state": (
                state.state.value
                if state.state
                else None
            ),

            "next_action": (
                state.next_action.value
                if state.next_action
                else None
            ),

            "qualification": {
                "region": state.region,
                "budget": state.budget,
                "bedrooms": state.bedrooms,
                "urgency": state.urgency,
                "property_type": state.property_type,
                "investor_profile": (
                    state.investor_profile
                ),
                "investment_ticket": (
                    state.investment_ticket
                ),
                "expected_return": (
                    state.expected_return
                ),
            },

            "selected_property_id": (
                state.selected_property_id
            ),

            "matched_properties": (
                matched_properties
            ),
        }

    # =========================================================
    # HELPERS
    # =========================================================

    @staticmethod
    def _enum_value(
        value,
    ) -> str | None:

        if value is None:
            return None

        if hasattr(value, "value"):
            return value.value

        return str(value)

    @staticmethod
    def _property_price(
        property_data: dict[str, Any],
    ) -> float:

        value = property_data.get(
            "price"
        )

        if value is None:
            return float("inf")

        try:
            return float(value)
        except (
            TypeError,
            ValueError,
        ):
            return float("inf")

    @staticmethod
    def _normalize(
        value: Any,
    ) -> str:

        if value is None:
            return ""

        text = str(value).strip().lower()

        return "".join(
            char
            for char in unicodedata.normalize(
                "NFD",
                text,
            )
            if unicodedata.category(char)
            != "Mn"
        )
