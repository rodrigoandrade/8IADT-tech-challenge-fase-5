from typing import Optional

from pydantic import BaseModel


class Qualification(BaseModel):
    region: Optional[str] = None
    budget: Optional[float] = None
    bedrooms: Optional[int] = None
    urgency: Optional[str] = None


class AgentResponse(BaseModel):
    message: str

    intent: Optional[str] = None
    transaction_type: Optional[str] = None

    qualification: Qualification

    next_action: Optional[str] = None