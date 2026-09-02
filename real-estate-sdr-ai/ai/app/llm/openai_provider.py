import os

from openai import AsyncOpenAI

from ai.app.llm.provider import LLMProvider


class OpenAIProvider(LLMProvider):

    def __init__(self):
        api_key = os.getenv("OPENAI_API_KEY")

        model = os.getenv(
            "OPENAI_MODEL",
            "gpt-4o-mini",
        )

        if not api_key:
            raise RuntimeError(
                "OPENAI_API_KEY is not configured"
            )

        self.client = AsyncOpenAI(
            api_key=api_key,
        )

        self.model = model

    async def generate_response(
        self,
        prompt: str,
    ) -> str:

        response = await self.client.chat.completions.create(
            model=self.model,
            messages=[
                {
                    "role": "user",
                    "content": prompt,
                }
            ],
            temperature=0.2,
        )

        content = response.choices[0].message.content

        if not content:
            raise RuntimeError(
                "OpenAI returned an empty response"
            )

        return content