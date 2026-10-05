import json
from typing import Any

from ai.app.agent.agent_type import AgentType
from ai.app.agent.base_agent import BaseAgent
from ai.app.agent.response import AgentResponse
from ai.app.agent.llm_provider import LLMProvider
from ai.app.prompt.prompt_builder import PromptBuilder


class KnowledgeAgent(BaseAgent):

    def __init__(
        self,
        llm_provider: LLMProvider,
        prompt_builder: PromptBuilder,
    ):
        self.llm_provider = llm_provider
        self.prompt_builder = prompt_builder

    @property
    def agent_type(self) -> AgentType:
        return AgentType.KNOWLEDGE

    async def execute(
        self,
        *,
        rag_service: Any,
        session: Any,
        question: str,
    ) -> AgentResponse:

        knowledge = await rag_service.retrieve(
            session=session,
            question=question,
            limit=3,
        )

        prompt = self.prompt_builder.build_knowledge_prompt(
            question=question,
            knowledge=knowledge,
        )

        raw_response = await self.llm_provider.generate_response(
            prompt
        )

        try:
            data = json.loads(raw_response)

        except json.JSONDecodeError as exc:
            raise ValueError(
                "KnowledgeAgent recebeu uma resposta inválida do LLM."
            ) from exc

        message = data.get("message")

        if not message:
            raise ValueError(
                "KnowledgeAgent recebeu JSON sem o campo 'message'."
            )

        return AgentResponse(
            message=message,
        )