import logging
from datetime import datetime, timedelta, timezone
from typing import Any

from ai.app.agent.state import ConversationStateType
from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.repositories.conversation_state_repository import (
    ConversationStateRepository,
)


logger = logging.getLogger(__name__)


class FollowUpService:
    """
    Serviço responsável por identificar conversas inativas
    e enviar automaticamente um follow-up.

    Regra:

        última mensagem = USER
        +
        tempo de inatividade >= threshold
        +
        follow_up_count < max_count
        +
        conversa não está em estado final
        =>
        envia follow-up
    """

    def __init__(
        self,
        api_client: RealEstateAPIClient,
        state_repository: ConversationStateRepository,
        inactivity_minutes: int = 1,
        max_follow_ups: int = 1,
    ):
        self.api_client = api_client
        self.state_repository = state_repository
        self.inactivity_minutes = inactivity_minutes
        self.max_follow_ups = max_follow_ups

    async def process(self) -> int:
        """
        Verifica todas as conversas e envia follow-up
        quando necessário.

        Retorna a quantidade de follow-ups enviados.
        """

        conversations = await self.api_client.get_conversations(
            limit=100,
            offset=0,
        )

        sent = 0

        for conversation in conversations:
            conversation_id = self._get_conversation_id(
                conversation
            )

            if not conversation_id:
                continue

            try:
                was_sent = await self.process_conversation(
                    conversation_id
                )

                if was_sent:
                    sent += 1

            except Exception:
                logger.exception(
                    "Error processing follow-up "
                    "for conversation=%s",
                    conversation_id,
                )

        return sent

    async def process_conversation(self, conversation_id: str) -> bool:
        messages = await self.api_client.get_messages(conversation_id)

        if not messages:
            return False

        last_user_message = None

        for message in reversed(messages):
            if self._get_role(message) == "USER":
                last_user_message = message
                break

        if last_user_message is None:
            return False

        state = await self.state_repository.get(conversation_id)

        if state is None:
            return False

        if state.follow_up_count >= self.max_follow_ups:
            return False

        if state.state in (
                ConversationStateType.SCHEDULED,
                ConversationStateType.COMPLETED,
                ConversationStateType.LOST,
        ):
            return False

        created_at = self._get_created_at(last_user_message)

        if created_at is None:
            logger.warning(
                "Mensagem do lead sem created_at. conversation_id=%s",
                conversation_id,
            )
            return False

        if not self._is_inactive(created_at):
            return False

        state.previous_state = state.state
        state.previous_next_action = state.next_action

        content = self._build_follow_up_message(state)

        await self.api_client.create_message(
            conversation_id=conversation_id,
            role="ASSISTANT",
            content=content,
        )

        state.follow_up_count += 1
        state.waiting_for_customer = True
        state.state = ConversationStateType.WAITING_LEAD
        state.next_action = None

        await self.state_repository.save(state)

        logger.info(
            "Follow-up enviado. conversation_id=%s follow_up_count=%s",
            conversation_id,
            state.follow_up_count,
        )

        return True

    def _is_inactive(
        self,
        created_at: datetime,
    ) -> bool:

        now = datetime.now(timezone.utc)

        if created_at.tzinfo is None:
            created_at = created_at.replace(
                tzinfo=timezone.utc
            )

        threshold = timedelta(
            minutes=self.inactivity_minutes
        )

        return now - created_at >= threshold

    @staticmethod
    def _get_conversation_id(
        conversation: dict[str, Any],
    ) -> str | None:

        value = (
            conversation.get("id")
            or conversation.get("ID")
        )

        if value is None:
            return None

        return str(value)

    @staticmethod
    def _get_role(
        message: dict[str, Any],
    ) -> str:

        value = (
            message.get("role")
            or message.get("Role")
            or ""
        )

        return str(value).strip().upper()

    @staticmethod
    def _get_created_at(
        message: dict[str, Any],
    ) -> datetime | None:

        value = (
            message.get("created_at")
            or message.get("CreatedAt")
        )

        if value is None:
            return None

        if isinstance(value, datetime):
            return value

        try:
            return datetime.fromisoformat(
                str(value).replace(
                    "Z",
                    "+00:00",
                )
            )

        except ValueError:
            return None

    @staticmethod
    def _build_follow_up_message(
        state,
    ) -> str:

        if state.next_action:
            return (
                "Oi! Tudo bem? 😊 "
                "Vi que nossa conversa ficou em aberto. "
                "Posso continuar te ajudando "
                "com a busca do imóvel?"
            )

        return (
            "Oi! Tudo bem? 😊 "
            "Vi que nossa conversa ficou em aberto. "
            "Posso continuar te ajudando?"
        )