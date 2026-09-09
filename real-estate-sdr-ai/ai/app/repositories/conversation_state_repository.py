from abc import ABC, abstractmethod

from ai.app.agent.state import ConversationState


class ConversationStateRepository(ABC):

    @abstractmethod
    async def get(
        self,
        conversation_id: str,
    ) -> ConversationState | None:
        pass

    @abstractmethod
    async def save(
        self,
        state: ConversationState,
    ) -> ConversationState:
        pass

    @abstractmethod
    async def delete(
        self,
        conversation_id: str,
    ) -> None:
        pass