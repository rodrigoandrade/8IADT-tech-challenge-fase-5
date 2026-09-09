from dataclasses import dataclass
from enum import Enum
from typing import Optional

class ConversationFlow(str, Enum):
    """
    Fluxo comercial principal da conversa.
    """

    BUY = "buy"
    INVESTMENT = "investment"


class ConversationStateType(str, Enum):
    """
    Estado atual da conversa dentro do fluxo comercial.
    """

    # COMMON
    NEW = "new"
    IDENTIFY_INTENT = "identify_intent"

    # BUY FLOW
    QUALIFY_BUDGET = "qualify_budget"
    QUALIFY_BEDROOMS = "qualify_bedrooms"
    QUALIFY_REGION = "qualify_region"
    QUALIFY_URGENCY = "qualify_urgency"

    RECOMMEND_PROPERTIES = "recommend_properties"
    SCHEDULE_MEETING = "schedule_meeting"

    # INVESTMENT FLOW
    INVESTOR_PROFILE = "investor_profile"
    INVESTMENT_TICKET = "investment_ticket"
    EXPECTED_RETURN = "expected_return"

    SPECIALIST_HANDOFF = "specialist_handoff"

    # FOLLOW-UP
    WAITING_LEAD = "waiting_lead"
    FOLLOW_UP_PENDING = "follow_up_pending"
    SEND_FOLLOW_UP = "send_follow_up"

    # FINAL STATES
    SCHEDULED = "scheduled"
    COMPLETED = "completed"
    LOST = "lost"


class NextAction(str, Enum):
    """
    Próxima ação comercial determinada pela máquina de estados.

    O LLM pode identificar informações da conversa, mas não deve
    ser responsável por determinar sozinho a próxima etapa.
    """

    # COMMON
    IDENTIFY_INTENT = "identify_intent"

    # BUY FLOW
    ASK_BUDGET = "ask_budget"
    ASK_BEDROOMS = "ask_bedrooms"
    ASK_LOCATION = "ask_location"
    ASK_PROPERTY_TYPE = "ask_property_type"
    ASK_URGENCY = "ask_urgency"

    MATCH_PROPERTY = "match_property"
    RECOMMEND_PROPERTY = "recommend_property"

    SCHEDULE_MEETING = "schedule_meeting"

    # INVESTMENT FLOW
    ASK_INVESTOR_PROFILE = "ask_investor_profile"
    ASK_INVESTMENT_TICKET = "ask_investment_ticket"
    ASK_EXPECTED_RETURN = "ask_expected_return"

    HANDOFF_SPECIALIST = "handoff_specialist"

    # FOLLOW-UP
    FOLLOW_UP = "follow_up"


@dataclass
class ConversationState:
    """
    Estado persistente da conversa com o lead.

    O flow representa o objetivo comercial principal.
    Follow-up não é um flow separado: ele é uma condição
    temporal que pode acontecer dentro de qualquer flow.
    """

    # IDENTIFICATION
    conversation_id: str

    # STATE MACHINE
    flow: Optional[ConversationFlow] = None

    state: ConversationStateType = ConversationStateType.NEW

    next_action: Optional[NextAction] = None

    # Guarda o estado comercial antes de entrar em follow-up.
    previous_state: Optional[ConversationStateType] = None

    # INTENT
    intent: Optional[str] = None

    transaction_type: Optional[str] = None

    # BUYER QUALIFICATION
    property_type: Optional[str] = None
    region: Optional[str] = None
    budget: Optional[float] = None
    bedrooms: Optional[int] = None
    urgency: Optional[str] = None

    # INVESTOR QUALIFICATION
    investor_profile: Optional[str] = None
    investment_ticket: Optional[float] = None
    expected_return: Optional[float] = None

    # LEAD QUALIFICATION
    qualification_score: Optional[float] = None
    temperature: Optional[str] = None

    # PROPERTY
    selected_property_id: Optional[str] = None

    # FOLLOW-UP
    follow_up_count: int = 0
    waiting_for_customer: bool = False