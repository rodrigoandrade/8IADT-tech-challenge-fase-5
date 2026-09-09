import os
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI

from ai.app.api.ai_routes import router as ai_router


# =========================================================
# ENVIRONMENT
# =========================================================

BASE_DIR = Path(__file__).resolve().parents[2]

load_dotenv(BASE_DIR / ".env")


# =========================================================
# APPLICATION
# =========================================================

def create_app() -> FastAPI:

    app = FastAPI(
        title="Real Estate SDR AI",
        description="AI service for the Real Estate SDR platform",
        version="1.0.0",
    )

    # =====================================================
    # HEALTH
    # =====================================================

    @app.get("/health")
    async def health():
        return {
            "status": "ok",
        }

    # =====================================================
    # ROUTES
    # =====================================================

    app.include_router(ai_router)

    return app


# =========================================================
# APPLICATION INSTANCE
# =========================================================

app = create_app()