from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum


class MessageRole(str, Enum):
    """
    Identifica quem enviou uma mensagem na conversa.
    """

    LEAD = "lead"
    ASSISTANT = "assistant"
    SYSTEM = "system"


@dataclass(frozen=True)
class ConversationMessage:
    """
    Representa uma mensagem individual da conversa.
    """

    role: MessageRole
    content: str
    timestamp: datetime | None = None

    def __post_init__(self) -> None:
        if not self.content.strip():
            raise ValueError("message content cannot be empty")


@dataclass(frozen=True)
class Conversation:
    """
    Representa o histórico de uma conversa entre o lead e o SDR.
    """

    id: str
    lead_id: str

    messages: list[ConversationMessage] = field(
        default_factory=list
    )

    created_at: datetime | None = None
    updated_at: datetime | None = None

    def __post_init__(self) -> None:
        if not self.id.strip():
            raise ValueError("conversation id cannot be empty")

        if not self.lead_id.strip():
            raise ValueError("lead id cannot be empty")

    def add_message(
            self,
            message: ConversationMessage,
    ) -> "Conversation":
        """
        Retorna uma nova conversa contendo a nova mensagem.

        A classe é imutável (frozen=True), portanto não altera
        diretamente a lista existente.
        """

        return Conversation(
            id=self.id,
            lead_id=self.lead_id,
            messages=[
                *self.messages,
                message,
            ],
            created_at=self.created_at,
            updated_at=message.timestamp,
        )

    @property
    def last_message(self) -> ConversationMessage | None:
        """
        Retorna a última mensagem da conversa.
        """

        if not self.messages:
            return None

        return self.messages[-1]

    @property
    def last_lead_message(self) -> ConversationMessage | None:
        """
        Retorna a última mensagem enviada pelo lead.
        """

        for message in reversed(self.messages):
            if message.role == MessageRole.LEAD:
                return message

        return None

    def format_for_prompt(self) -> str:
        """
        Formata o histórico para ser utilizado pelo PromptBuilder.
        """

        if not self.messages:
            return "Nenhuma mensagem anterior."

        role_names = {
            MessageRole.LEAD: "Lead",
            MessageRole.ASSISTANT: "SDR",
            MessageRole.SYSTEM: "Sistema",
        }

        return "\n".join(
            f"{role_names[message.role]}: {message.content}"
            for message in self.messages
        )
