# ai/app/main.py

import logging
import os
import time
import uuid
from contextlib import asynccontextmanager
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI, Request

from ai.app.api.ai_routes import router as ai_router
from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.workers.follow_up_worker import FollowUpWorker

from ai.app.observability.context import (
    set_correlation_id,
)


# =========================================================
# LOGGING
# =========================================================

class ObservabilityFormatter(
    logging.Formatter
):
    def format(
        self,
        record: logging.LogRecord,
    ) -> str:

        timestamp = self.formatTime(
            record,
            self.datefmt,
        )

        correlation_id = getattr(
            record,
            "correlation_id",
            "-",
        )

        method = getattr(
            record,
            "method",
            "-",
        )

        path = getattr(
            record,
            "path",
            "-",
        )

        status_code = getattr(
            record,
            "status_code",
            "-",
        )

        duration_ms = getattr(
            record,
            "duration_ms",
            "-",
        )

        message = record.getMessage()

        return (
            f"{timestamp} "
            f"{record.levelname} "
            f"{record.name} "
            f"event={message} "
            f"correlation_id={correlation_id} "
            f"method={method} "
            f"path={path} "
            f"status_code={status_code} "
            f"duration_ms={duration_ms}"
        )


handler = logging.StreamHandler()

handler.setFormatter(
    ObservabilityFormatter(
        "%(asctime)s"
    )
)

logging.basicConfig(
    level=logging.INFO,
    handlers=[handler],
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
# OBSERVABILITY
# =========================================================

@app.middleware("http")
async def observability_middleware(
    request: Request,
    call_next,
):
    correlation_id = request.headers.get(
        "X-Correlation-ID"
    )

    correlation_id = request.headers.get(
        "X-Correlation-ID"
    )

    if not correlation_id:
        correlation_id = str(uuid.uuid4())

    set_correlation_id(
        correlation_id
    )

    start_time = time.perf_counter()

    logger.info(
        "request_started",
        extra={
            "correlation_id": correlation_id,
            "method": request.method,
            "path": request.url.path,
        },
    )

    try:
        response = await call_next(request)

        duration_ms = (
            time.perf_counter() - start_time
        ) * 1000

        response.headers["X-Correlation-ID"] = correlation_id

        logger.info(
            "request_completed",
            extra={
                "correlation_id": correlation_id,
                "method": request.method,
                "path": request.url.path,
                "status_code": response.status_code,
                "duration_ms": round(duration_ms, 2),
            },
        )

        return response

    except Exception:
        duration_ms = (
            time.perf_counter() - start_time
        ) * 1000

        logger.exception(
            "request_failed",
            extra={
                "correlation_id": correlation_id,
                "method": request.method,
                "path": request.url.path,
                "duration_ms": round(duration_ms, 2),
            },
        )

        raise


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