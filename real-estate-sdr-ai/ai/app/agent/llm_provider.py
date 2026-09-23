from abc import ABC, abstractmethod

from .response import AgentResponse


class LLMProvider(ABC):

    @abstractmethod
    async def generate_response(
        self,
        prompt: str,
    ) -> AgentResponse:
        """
        Executa o LLM e retorna os dados estruturados.
        """
        raise NotImplementedError