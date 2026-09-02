from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.llm.provider import LLMProvider
from ai.app.llm.response_parser import ResponseParser
from ai.app.models.ai_response import AIResponse
from ai.app.models.conversation import Conversation
from ai.app.prompt.prompt_builder import PromptBuilder


class AIService:

    def __init__(
        self,
        llm_provider: LLMProvider,
        prompt_builder: PromptBuilder,
        response_parser: ResponseParser,
        real_estate_api_client: RealEstateAPIClient,
    ) -> None:

        self._llm_provider = llm_provider
        self._prompt_builder = prompt_builder
        self._response_parser = response_parser
        self._real_estate_api_client = real_estate_api_client

    async def generate_response(
        self,
        conversation: Conversation,
    ) -> AIResponse:

        # =====================================================
        # LEAD
        # =====================================================

        lead = await self._real_estate_api_client.get_lead(
            conversation.lead_id
        )

        # =====================================================
        # PROPERTIES
        # =====================================================

        properties = await self._real_estate_api_client.get_properties()

        # =====================================================
        # PROMPT
        # =====================================================

        prompt = self._prompt_builder.build(
            conversation=conversation,
            lead=lead,
            properties=properties,
        )

        # =====================================================
        # LLM
        # =====================================================

        llm_response = await self._llm_provider.generate_response(
            prompt
        )

        # =====================================================
        # PARSER
        # =====================================================

        return self._response_parser.parse(
            llm_response
        )