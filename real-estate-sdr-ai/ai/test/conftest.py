import pytest

from ai.app.agent.state import ConversationState
from ai.app.agent.state_machine import ConversationStateMachine
from ai.app.llm.provider import LLMProvider
from ai.app.llm.response_parser import ResponseParser
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.services.ai_service import AIService
from ai.app.agent.agent_router import AgentRouter
from ai.app.agent.knowledge_agent import KnowledgeAgent

class InMemoryConversationStateRepository:

    def __init__(self):
        self.states = {}

    async def get(self, conversation_id: str):
        return self.states.get(conversation_id)

    async def save(self, state: ConversationState):
        self.states[state.conversation_id] = state
        return state

    async def delete(self, conversation_id: str):
        self.states.pop(conversation_id, None)


class FakeLLMProvider(LLMProvider):

    def __init__(self):
        self.responses = []

    async def generate_response(self, prompt: str) -> str:

        if not self.responses:
            raise RuntimeError(
                "FakeLLMProvider: nenhuma resposta configurada"
            )

        return self.responses.pop(0)


class FakeRealEstateAPIClient:

    async def get_lead(self, lead_id: str):
        return {
            "id": lead_id,
            "name": "João Teste",
            "email": "joao.teste@example.com",
            "phone": "11999999999",
            "status": "NEW",
        }

    async def get_messages(self, conversation_id: str):
        return []

    async def create_message(
        self,
        conversation_id: str,
        role: str,
        content: str,
    ):
        return {
            "id": "fake-message-id",
            "conversation_id": conversation_id,
            "role": role,
            "content": content,
        }


class FakeRAGService:

    async def retrieve(
        self,
        session,
        question: str,
        limit: int = 3,
    ):
        return []


@pytest.fixture(scope="module")
def repository():
    return InMemoryConversationStateRepository()


@pytest.fixture(scope="module")
def llm_provider():
    provider = FakeLLMProvider()

    provider.responses.extend([
        # ---------------------------------------------------------
        # 3.1 - primeira mensagem
        # ---------------------------------------------------------
        """
        {
            "message": "Claro! Qual é o seu orçamento?",
            "intent": "property_interest",
            "transaction_type": "buy",
            "qualification": {
                "region": "zona sul",
                "property_type": "apartamento"
            },
            "next_action": "ask_budget"
        }
        """,

        # ---------------------------------------------------------
        # 3.2 - orçamento
        # ---------------------------------------------------------
        """
        {
            "message": "Perfeito! Quantos quartos você gostaria?",
            "intent": "property_interest",
            "transaction_type": "buy",
            "qualification": {
                "region": "zona sul",
                "property_type": "apartamento",
                "budget": 800000
            },
            "next_action": "ask_bedrooms"
        }
        """
    ])

    return provider


@pytest.fixture
def prompt_builder():
    return PromptBuilder()


@pytest.fixture
def response_parser():
    return ResponseParser()


@pytest.fixture
def state_machine():
    return ConversationStateMachine()


@pytest.fixture
def api_client():
    return FakeRealEstateAPIClient()


@pytest.fixture
def rag_service():
    return FakeRAGService()


@pytest.fixture
def service(
    api_client,
    llm_provider,
    prompt_builder,
    response_parser,
    repository,
    state_machine,
    rag_service,
):
    return AIService(
        api_client=api_client,
        llm_provider=llm_provider,
        prompt_builder=prompt_builder,
        response_parser=response_parser,
        state_repository=repository,
        state_machine=state_machine,
        rag_service=rag_service,
        agent_router=AgentRouter(),
        knowledge_agent=KnowledgeAgent(
            llm_provider=llm_provider,
            prompt_builder=prompt_builder,
        ),
    )