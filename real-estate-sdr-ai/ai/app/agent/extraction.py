from dataclasses import dataclass
from typing import Optional

@dataclass
class ExtractedConversationData:
    """
    Informações extraídas pelo LLM a partir da conversa.
    """

    intent: Optional[str] = None
    transaction_type: Optional[str] = None

    property_type: Optional[str] = None
    region: Optional[str] = None
    budget: Optional[float] = None
    bedrooms: Optional[int] = None
    urgency: Optional[str] = None

    investor_profile: Optional[str] = None
    investment_ticket: Optional[float] = None
    expected_return: Optional[float] = None