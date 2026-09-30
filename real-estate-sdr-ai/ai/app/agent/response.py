from typing import Any, Optional

from pydantic import BaseModel, Field


class Qualification(BaseModel):
    region: Optional[str] = None
    budget: Optional[float] = None
    bedrooms: Optional[int] = None
    urgency: Optional[str] = None
    property_type: Optional[str] = None
    investor_profile: Optional[str] = None
    investment_ticket: Optional[float] = None
    expected_return: Optional[float] = None


class AgentResponse(BaseModel):
    message: str
    intent: Optional[str] = None
    transaction_type: Optional[str] = None
    qualification: Qualification = Field(
        default_factory=Qualification
    )
    matched_properties: list[dict[str, Any]] = Field(
        default_factory=list
    )
    selected_property_id: Optional[str] = None