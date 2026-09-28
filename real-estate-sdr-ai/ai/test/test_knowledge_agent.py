import pytest

from ai.app.agent.knowledge_agent import KnowledgeAgent
from ai.app.agent.response import AgentResponse
from ai.app.prompt.prompt_builder import PromptBuilder


class FakeRAGService:

    async def retrieve(self, session, question, limit=3):
        return [
            {
                "title": "Financiamento",
                "content": (
                    "Os imóveis podem ser adquiridos através "
                    "de financiamento bancário."
                ),
                "similarity": 0.69,
            }
        ]


class FakeLLMProvider:

    def __init__(self):
        self.last_prompt = None

    async def generate_response(self, prompt):
        self.last_prompt = prompt

        return """
        {
            "message": "Os imóveis podem ser adquiridos através de financiamento bancário."
        }
        """


@pytest.mark.asyncio
async def test_knowledge_agent_retrieves_knowledge_and_generates_response():

    llm_provider = FakeLLMProvider()

    agent = KnowledgeAgent(
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
    )

    result = await agent.execute(
        rag_service=FakeRAGService(),
        session=None,
        question="Como funciona o financiamento?",
    )

    assert isinstance(result, AgentResponse)

    assert (
        result.message
        == "Os imóveis podem ser adquiridos através de financiamento bancário."
    )

    assert llm_provider.last_prompt is not None
    assert "Como funciona o financiamento?" in llm_provider.last_prompt
    assert "Financiamento" in llm_provider.last_prompt
    assert "financiamento bancário" in llm_provider.last_prompt

    # O prompt não deve expor detalhes internos do RAG.
    assert "0.69" not in llm_provider.last_prompt