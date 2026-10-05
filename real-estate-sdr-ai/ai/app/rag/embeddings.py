import os

from openai import AsyncOpenAI


class EmbeddingService:

    def __init__(self):
        api_key = os.getenv("OPENAI_API_KEY")

        if not api_key:
            raise RuntimeError("OPENAI_API_KEY não configurada")

        self.client = AsyncOpenAI(api_key=api_key)
        self.model = os.getenv(
            "OPENAI_EMBEDDING_MODEL",
            "text-embedding-3-small",
        )

    async def generate(self, text: str) -> list[float]:
        response = await self.client.embeddings.create(
            model=self.model,
            input=text,
        )

        return response.data[0].embedding