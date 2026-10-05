import pytest

from ai.app.agent.agent_router import AgentRouter
from ai.app.agent.agent_type import AgentType
from ai.app.agent.knowledge_agent import KnowledgeAgent
from ai.app.agent.response import AgentResponse
from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state_machine import ConversationStateMachine
from ai.app.llm.response_parser import ResponseParser
from ai.app.models.conversation import (
    Conversation,
    ConversationMessage,
    MessageRole,
)
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.services.ai_service import AIService


class FakeAPIClient:

    async def get_lead(self, lead_id):
        return {
            "id": lead_id,
            "name": "Lead Teste",
        }


class FakeStateRepository:

    def __init__(self, state):
        self.state = state
        self.saved_state = None

    async def get(self, conversation_id):
        return self.state

    async def save(self, state):
        self.saved_state = state


class FakeRAGService:

    async def retrieve(self, session, question, limit=3):
        return []


class FakeLLMProvider:

    async def generate_response(self, prompt):
        return """
        {
            "message": "Encontrei imóveis para você.",
            "intent": "property_interest",
            "transaction_type": "sale",
            "qualification": {
                "property_type": "apartamento",
                "budget": 500000,
                "bedrooms": 3,
                "region": "Moema",
                "urgency": "immediate"
            }
        }
        """


class FakePropertyAgent:

    def __init__(self):
        self.called = False
        self.received_state = None

    async def execute(self, *, state):
        self.called = True
        self.received_state = state

        return AgentResponse(
            message="Encontrei imóveis.",
            matched_properties=[
                {
                    "id": "property-001",
                    "title": "Apartamento Moema",
                    "price": 475000,
                    "bedrooms": 4,
                }
            ],
            selected_property_id="property-001",
        )


@pytest.mark.asyncio
async def test_ai_service_delegates_property_matching_to_property_agent():

    state = ConversationState(
        conversation_id="conversation-property-agent",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.QUALIFY_URGENCY,
        next_action=NextAction.MATCH_PROPERTY,
        property_type="apartamento",
        budget=500000,
        bedrooms=3,
        region="Moema",
        urgency="immediate",
    )

    repository = FakeStateRepository(state)
    property_agent = FakePropertyAgent()

    llm_provider = FakeLLMProvider()

    knowledge_agent = KnowledgeAgent(
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
    )

    service = AIService(
        api_client=FakeAPIClient(),
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=repository,
        state_machine=ConversationStateMachine(),
        rag_service=FakeRAGService(),
        agent_router=AgentRouter(),
        knowledge_agent=knowledge_agent,
        property_agent=property_agent,
    )

    conversation = Conversation(
        id="conversation-property-agent",
        lead_id="lead-001",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content="Quero um apartamento em Moema.",
            )
        ],
    )

    result = await service.generate_response(
        conversation
    )

    assert property_agent.called is True

    assert property_agent.called is True

    received_state = property_agent.received_state

    assert received_state is not None
    assert received_state.conversation_id == "conversation-property-agent"
    assert received_state.flow == ConversationFlow.BUY

    assert received_state.property_type == "apartamento"
    assert received_state.budget == 500000
    assert received_state.bedrooms == 3
    assert received_state.region == "Moema"
    assert received_state.urgency == "immediate"

    assert result["selected_property_id"] == "property-001"

    assert len(result["matched_properties"]) == 1

    assert result["matched_properties"][0]["id"] == "property-001"