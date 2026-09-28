import pytest

from ai.app.agent.agent_router import AgentRouter
from ai.app.agent.knowledge_agent import KnowledgeAgent
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

    def __init__(self):
        self.get_lead_called = False

    async def get_lead(self, lead_id):
        self.get_lead_called = True

        return {
            "id": lead_id,
            "name": "Lead Teste",
        }


class FakeStateRepository:

    def __init__(self):
        self.saved_state = None

    async def get(self, conversation_id):
        return None

    async def save(self, state):
        self.saved_state = state


class FakeRAGService:

    def __init__(self):
        self.called = False
        self.question = None

    async def retrieve(
        self,
        session,
        question,
        limit=3,
    ):
        self.called = True
        self.question = question

        return [
            {
                "title": "Financiamento",
                "content": (
                    "Os imóveis podem ser adquiridos através "
                    "de financiamento bancário."
                ),
            }
        ]


class FakeLLMProvider:

    def __init__(self):
        self.called = False
        self.prompt = None

    async def generate_response(self, prompt):
        self.called = True
        self.prompt = prompt

        return """
        {
            "message": "Os imóveis podem ser adquiridos através de financiamento bancário."
        }
        """


class FakeSessionContext:

    async def __aenter__(self):
        return None

    async def __aexit__(self, exc_type, exc, traceback):
        return False


class FakeAsyncSessionLocal:

    def __call__(self):
        return FakeSessionContext()


@pytest.mark.asyncio
async def test_knowledge_flow_uses_knowledge_agent(monkeypatch):

    # ---------------------------------------------------------
    # Conversation
    # ---------------------------------------------------------

    conversation = Conversation(
        id="conversation-1",
        lead_id="lead-1",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content=(
                    "Como funciona o financiamento dos imóveis?"
                ),
            )
        ],
    )

    # ---------------------------------------------------------
    # Dependencies
    # ---------------------------------------------------------

    api_client = FakeAPIClient()
    state_repository = FakeStateRepository()
    rag_service = FakeRAGService()
    llm_provider = FakeLLMProvider()

    knowledge_agent = KnowledgeAgent(
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
    )

    # ---------------------------------------------------------
    # AI Service
    # ---------------------------------------------------------

    service = AIService(
        api_client=api_client,
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=state_repository,
        state_machine=ConversationStateMachine(),
        rag_service=rag_service,
        agent_router=AgentRouter(),
        knowledge_agent=knowledge_agent,
    )

    # ---------------------------------------------------------
    # Evita acesso ao PostgreSQL durante o teste
    # ---------------------------------------------------------

    monkeypatch.setattr(
        "ai.app.infrastructure.database.AsyncSessionLocal",
        FakeAsyncSessionLocal(),
    )

    # ---------------------------------------------------------
    # Execute
    # ---------------------------------------------------------

    result = await service.generate_response(
        conversation
    )

    # ---------------------------------------------------------
    # Response
    # ---------------------------------------------------------

    assert result["message"] == (
        "Os imóveis podem ser adquiridos através de financiamento bancário."
    )

    assert result["intent"] == "general_inquiry"

    assert result["matched_properties"] == []

    assert result["selected_property_id"] is None

    # ---------------------------------------------------------
    # API Client
    # ---------------------------------------------------------

    assert api_client.get_lead_called is True

    # ---------------------------------------------------------
    # RAG
    # ---------------------------------------------------------

    assert rag_service.called is True

    assert rag_service.question == (
        "Como funciona o financiamento dos imóveis?"
    )

    # ---------------------------------------------------------
    # LLM
    # ---------------------------------------------------------

    assert llm_provider.called is True

    assert llm_provider.prompt is not None

    assert (
        "Como funciona o financiamento dos imóveis?"
        in llm_provider.prompt
    )

    assert "Financiamento" in llm_provider.prompt

    # ---------------------------------------------------------
    # State
    # ---------------------------------------------------------

    assert state_repository.saved_state is not None

@pytest.mark.asyncio
async def test_knowledge_flow_for_visit_question(monkeypatch):

    conversation = Conversation(
        id="conversation-visit",
        lead_id="lead-1",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content="Como faço para visitar um imóvel?",
            )
        ],
    )

    api_client = FakeAPIClient()
    state_repository = FakeStateRepository()
    rag_service = FakeRAGService()
    llm_provider = FakeLLMProvider()

    knowledge_agent = KnowledgeAgent(
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
    )

    service = AIService(
        api_client=api_client,
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=state_repository,
        state_machine=ConversationStateMachine(),
        rag_service=rag_service,
        agent_router=AgentRouter(),
        knowledge_agent=knowledge_agent,
    )

    monkeypatch.setattr(
        "ai.app.infrastructure.database.AsyncSessionLocal",
        FakeAsyncSessionLocal(),
    )

    result = await service.generate_response(
        conversation
    )

    assert result["intent"] == "general_inquiry"

    assert result["matched_properties"] == []

    assert result["selected_property_id"] is None

    assert rag_service.called is True

    assert rag_service.question == (
        "Como faço para visitar um imóvel?"
    )

    assert llm_provider.called is True

@pytest.mark.asyncio
async def test_knowledge_flow_for_regions_question(monkeypatch):

    conversation = Conversation(
        id="conversation-regions",
        lead_id="lead-1",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content="Quais regiões vocês atendem?",
            )
        ],
    )

    api_client = FakeAPIClient()
    state_repository = FakeStateRepository()
    rag_service = FakeRAGService()
    llm_provider = FakeLLMProvider()

    knowledge_agent = KnowledgeAgent(
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
    )

    service = AIService(
        api_client=api_client,
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=state_repository,
        state_machine=ConversationStateMachine(),
        rag_service=rag_service,
        agent_router=AgentRouter(),
        knowledge_agent=knowledge_agent,
    )

    monkeypatch.setattr(
        "ai.app.infrastructure.database.AsyncSessionLocal",
        FakeAsyncSessionLocal(),
    )

    result = await service.generate_response(
        conversation
    )

    assert result["intent"] == "general_inquiry"

    assert result["matched_properties"] == []

    assert result["selected_property_id"] is None

    assert rag_service.called is True

    assert rag_service.question == (
        "Quais regiões vocês atendem?"
    )

    assert llm_provider.called is True

