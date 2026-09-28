from ai.app.agent.agent_type import AgentType
from ai.app.agent.knowledge_agent import KnowledgeAgent
from ai.app.agent.property_agent import PropertyAgent
from ai.app.agent.qualification_agent import QualificationAgent
from ai.app.agent.response import AgentResponse
from ai.app.prompt.prompt_builder import PromptBuilder


class FakeLLMProvider:

    async def generate_response(self, prompt):
        return AgentResponse(message="Resposta fake")


def create_knowledge_agent():
    return KnowledgeAgent(
        llm_provider=FakeLLMProvider(),
        prompt_builder=PromptBuilder(),
    )


def test_knowledge_agent_type():

    agent = create_knowledge_agent()

    assert agent.agent_type == AgentType.KNOWLEDGE


def test_qualification_agent_type():

    agent = QualificationAgent()

    assert agent.agent_type == AgentType.QUALIFICATION


def test_property_agent_type():

    agent = PropertyAgent()

    assert agent.agent_type == AgentType.PROPERTY