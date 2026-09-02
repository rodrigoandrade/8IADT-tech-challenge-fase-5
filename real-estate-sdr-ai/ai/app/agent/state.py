from dataclasses import dataclass
from typing import Optional


@dataclass
class ConversationState:
    conversation_id: str

    intent: Optional[str] = None
    transaction_type: Optional[str] = None

    region: Optional[str] = None
    budget: Optional[float] = None
    bedrooms: Optional[int] = None
    urgency: Optional[str] = None

    investor_profile: Optional[str] = None
    expected_return: Optional[float] = None

    qualification_score: Optional[float] = None
    temperature: Optional[str] = None

    next_action: Optional[str] = None