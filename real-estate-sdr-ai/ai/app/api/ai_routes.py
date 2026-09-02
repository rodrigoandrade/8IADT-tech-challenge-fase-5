from fastapi import APIRouter, Depends, HTTPException

from ai.app.models.ai_response import AIResponse
from ai.app.models.conversation import Conversation
from ai.app.services.ai_service import AIService


router = APIRouter(
    prefix="/ai",
    tags=["AI"],
)


def get_ai_service() -> AIService:
    from ai.app.main import ai_service

    if ai_service is None:
        raise RuntimeError(
            "AIService is not initialized"
        )

    return ai_service


@router.post(
    "/generate-response",
    response_model=AIResponse,
)
async def generate_response(
    conversation: Conversation,
    service: AIService = Depends(get_ai_service),
) -> AIResponse:

    try:
        return await service.generate_response(
            conversation
        )

    except ValueError as exc:
        raise HTTPException(
            status_code=400,
            detail=str(exc),
        ) from exc

    except Exception as exc:
        import traceback

        traceback.print_exc()

        raise HTTPException(
            status_code=500,
            detail=str(exc),
        ) from exc