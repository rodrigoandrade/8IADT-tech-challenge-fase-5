from typing import Any

from ai.app.llm.provider import LLMProvider


class LLMService:
    """
    Serviço responsável por intermediar as chamadas ao modelo de linguagem.

    A aplicação conversa com este serviço, e não diretamente com
    o provedor (OpenAI, Azure OpenAI, etc.).
    """

    def __init__(self, provider: LLMProvider):
        self._provider = provider

    def generate(
            self,
            prompt: str,
            **kwargs: Any,
    ) -> str:
        """
        Envia um prompt ao LLM e retorna a resposta textual.
        """
        if not prompt or not prompt.strip():
            raise ValueError("prompt cannot be empty")

        return self._provider.generate(
            prompt=prompt,
            **kwargs,
        )