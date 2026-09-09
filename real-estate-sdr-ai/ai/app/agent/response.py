from typing import Optional
from pydantic import BaseModel

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
    """
    Resposta estruturada produzida pelo LLM.

    O LLM é responsável somente por interpretar a mensagem
    e extrair informações.

    A decisão da próxima ação pertence à StateMachine.
    """

    message: str

    intent: Optional[str] = None

    transaction_type: Optional[str] = None

    qualification: Qualification = Qualification()