import os
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI

from ai.app.api.ai_routes import router as ai_router
from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.llm.openai_provider import OpenAIProvider
from ai.app.llm.response_parser import ResponseParser
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.services.ai_service import AIService


# =========================================================
# ENVIRONMENT
# =========================================================

BASE_DIR = Path(__file__).resolve().parents[2]

load_dotenv(BASE_DIR / ".env")


# =========================================================
# DEPENDENCIES
# =========================================================

ai_service: AIService | None = None


# =========================================================
# APPLICATION
# =========================================================

def create_app() -> FastAPI:
    global ai_service

    app = FastAPI(
        title="Real Estate SDR AI",
        description="AI service for the Real Estate SDR platform",
        version="1.0.0",
    )

    # =====================================================
    # CONFIGURATION
    # =====================================================

    real_estate_api_url = os.getenv(
        "REAL_ESTATE_API_URL"
    )

    if not real_estate_api_url:
        raise RuntimeError(
            "REAL_ESTATE_API_URL is not configured"
        )

    # =====================================================
    # CLIENT
    # =====================================================

    real_estate_api_client = RealEstateAPIClient(
        base_url=real_estate_api_url
    )

    # =====================================================
    # LLM
    # =====================================================

    llm_provider = OpenAIProvider()

    # =====================================================
    # PROMPT
    # =====================================================

    prompt_builder = PromptBuilder()

    # =====================================================
    # RESPONSE PARSER
    # =====================================================

    response_parser = ResponseParser()

    # =====================================================
    # SERVICE
    # =====================================================

    ai_service = AIService(
        llm_provider=llm_provider,
        prompt_builder=prompt_builder,
        response_parser=response_parser,
        real_estate_api_client=real_estate_api_client,
    )

    # =====================================================
    # ROUTES
    # =====================================================

    app.include_router(ai_router)

    # =====================================================
    # HEALTH
    # =====================================================

    @app.get("/health")
    async def health():
        return {
            "status": "ok",
        }

    return app


# =========================================================
# APPLICATION INSTANCE
# =========================================================

app = create_app()
