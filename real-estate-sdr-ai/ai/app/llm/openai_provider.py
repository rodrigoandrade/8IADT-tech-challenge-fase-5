import os

from openai import AsyncOpenAI

from ai.app.llm.provider import LLMProvider


class OpenAIProvider(LLMProvider):
    """
    Implementação do LLMProvider utilizando a API da OpenAI.

    Responsabilidades:
        - configurar o cliente OpenAI;
        - enviar o prompt;
        - solicitar resposta em JSON;
        - retornar o conteúdo textual retornado pelo modelo.

    O provider não é responsável por:
        - controlar a máquina de estados;
        - determinar next_action;
        - persistir ConversationState;
        - interpretar regras comerciais.
    """

    def __init__(
        self,
        api_key: str | None = None,
        model: str | None = None,
    ):
        self.api_key = (
            api_key
            or os.getenv("OPENAI_API_KEY")
        )

        if not self.api_key:
            raise RuntimeError(
                "OPENAI_API_KEY não configurada."
            )

        self.model = (
            model
            or os.getenv(
                "OPENAI_MODEL",
                "gpt-4o-mini",
            )
        )

        self.client = AsyncOpenAI(
            api_key=self.api_key,
        )

    async def generate_response(
        self,
        prompt: str,
    ) -> str:
        """
        Envia o prompt para a OpenAI e retorna
        a resposta textual do modelo.
        """

        if not prompt or not prompt.strip():
            raise ValueError(
                "O prompt não pode ser vazio."
            )

        response = await self.client.chat.completions.create(
            model=self.model,
            messages=[
                {
                    "role": "system",
                    "content": (
                        "Você é um assistente especializado "
                        "em atendimento imobiliário."
                    ),
                },
                {
                    "role": "user",
                    "content": prompt,
                },
            ],
            response_format={
                "type": "json_object",
            },
            temperature=0.2,
        )

        content = response.choices[0].message.content

        if not content:
            raise RuntimeError(
                "A OpenAI retornou uma resposta vazia."
            )

        return content