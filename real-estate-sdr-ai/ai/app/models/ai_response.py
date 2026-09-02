from dataclasses import dataclass, field
from enum import Enum


class LeadIntent(str, Enum):
    """
    Intenção identificada na conversa com o lead.
    """

    PROPERTY_INTEREST = "property_interest"
    PRICE_INQUIRY = "price_inquiry"
    SCHEDULING = "scheduling"
    GENERAL_INQUIRY = "general_inquiry"
    NEGOTIATION = "negotiation"
    UNKNOWN = "unknown"


class LeadStatus(str, Enum):
    """
    Status comercial do lead.
    """

    NEW = "new"
    QUALIFIED = "qualified"
    IN_NEGOTIATION = "in_negotiation"
    SCHEDULED = "scheduled"
    CONVERTED = "converted"
    LOST = "lost"
    UNKNOWN = "unknown"


@dataclass(frozen=True)
class AIResponse:
    """
    Resposta estruturada produzida pelo modelo de IA.
    """

    message: str

    intent: LeadIntent = LeadIntent.UNKNOWN

    lead_status: LeadStatus = LeadStatus.UNKNOWN

    qualification_score: float | None = None

    next_action: str | None = None

    reasoning: str | None = None

    metadata: dict[str, str] = field(
        default_factory=dict
    )

    def __post_init__(self) -> None:
        if not self.message.strip():
            raise ValueError("message cannot be empty")

        if self.qualification_score is not None:
            if not 0.0 <= self.qualification_score <= 1.0:
                raise ValueError(
                    "qualification_score must be between 0 and 1"
                )