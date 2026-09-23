# ai/app/main.py

import logging
import os
from contextlib import asynccontextmanager
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI

from ai.app.api.ai_routes import router as ai_router
from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.workers.follow_up_worker import FollowUpWorker


# =========================================================
# LOGGING
# =========================================================

logging.basicConfig(
    level=logging.INFO,
)

logger = logging.getLogger(__name__)


# =========================================================
# ENVIRONMENT
# =========================================================

BASE_DIR = Path(__file__).resolve().parents[2]

load_dotenv(
    BASE_DIR / ".env"
)


# =========================================================
# CONFIGURATION
# =========================================================

REAL_ESTATE_API_URL = os.getenv(
    "REAL_ESTATE_API_URL",
)

if not REAL_ESTATE_API_URL:
    raise RuntimeError(
        "REAL_ESTATE_API_URL is not configured"
    )


FOLLOW_UP_AFTER_MINUTES = int(
    os.getenv(
        "FOLLOW_UP_AFTER_MINUTES",
        "1",
    )
)

FOLLOW_UP_INTERVAL_SECONDS = int(
    os.getenv(
        "FOLLOW_UP_INTERVAL_SECONDS",
        "10",
    )
)

FOLLOW_UP_MAX_COUNT = int(
    os.getenv(
        "FOLLOW_UP_MAX_COUNT",
        "1",
    )
)


# =========================================================
# CLIENT
# =========================================================

real_estate_api_client = RealEstateAPIClient(
    base_url=REAL_ESTATE_API_URL,
)


# =========================================================
# FOLLOW-UP WORKER
# =========================================================

follow_up_worker = FollowUpWorker(
    api_client=real_estate_api_client,
    inactivity_minutes=FOLLOW_UP_AFTER_MINUTES,
    max_follow_ups=FOLLOW_UP_MAX_COUNT,
    interval_seconds=FOLLOW_UP_INTERVAL_SECONDS,
)


# =========================================================
# LIFESPAN
# =========================================================

@asynccontextmanager
async def lifespan(app: FastAPI):

    logger.info(
        "Starting Real Estate SDR AI..."
    )

    await follow_up_worker.start()

    logger.info(
        "Real Estate SDR AI started."
    )

    try:
        yield

    finally:

        logger.info(
            "Stopping Real Estate SDR AI..."
        )

        await follow_up_worker.stop()

        logger.info(
            "Real Estate SDR AI stopped."
        )


# =========================================================
# APPLICATION
# =========================================================

app = FastAPI(
    title="Real Estate SDR AI",
    description="AI SDR service for real estate leads",
    version="1.0.0",
    lifespan=lifespan,
)


# =========================================================
# ROUTES
# =========================================================

app.include_router(
    ai_router,
)


# =========================================================
# HEALTH
# =========================================================

@app.get("/health")
async def health():
    return {
        "status": "ok",
        "service": "real-estate-sdr-ai",
    }