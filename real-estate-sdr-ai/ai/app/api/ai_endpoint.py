from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel

from ai.app.models.ai_response import AIResponse
from ai.app.services.ai_service import AIService


router = APIRouter(
    prefix="/ai",
    tags=["AI"],
)


class GenerateResponseRequest(BaseModel):
    conversation_id: str


def get_ai_service() -> AIService:
    """
    Dependency responsible for providing the AIService.

    The concrete implementation will be configured during
    application bootstrap.
    """
    raise NotImplementedError(
        "AIService dependency is not configured"
    )


@router.post(
    "/conversations/generate",
    response_model=AIResponse,
)
def generate_ai_response(
        request: GenerateResponseRequest,
        ai_service: AIService = Depends(get_ai_service),
) -> AIResponse:
    """
    Generates an AI response for an existing conversation.

    The conversation_id is used as the single source of context.
    The AI service is responsible for retrieving the lead,
    conversation history and available properties.
    """

    try:
        return ai_service.generate_response(
            conversation_id=request.conversation_id,
        )

    except ValueError as exc:
        raise HTTPException(
            status_code=400,
            detail=str(exc),
        ) from exc

    except Exception as exc:
        raise HTTPException(
            status_code=500,
            detail="Failed to generate AI response",
        ) from exc
