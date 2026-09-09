from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession

from app.infrastructure.database import get_session
from app.infrastructure.repositories.postgres_conversation_state_repository import (
    PostgresConversationStateRepository,
)
from app.services.ai_service import AIService
from app.llm.openai_provider import OpenAIProvider
from app.prompt.prompt_builder import PromptBuilder
from app.llm.response_parser import ResponseParser
from app.agent.state_machine import ConversationStateMachine
from app.clients.real_estate_api_client import RealEstateAPIClient


router = APIRouter(
    prefix="/ai",
    tags=["AI"],
)


def get_ai_service(
    session: AsyncSession = Depends(get_session),
) -> AIService:

    state_repository = PostgresConversationStateRepository(
        session
    )

    return AIService(
        api_client=RealEstateAPIClient(),
        llm_provider=OpenAIProvider(),
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=state_repository,
        state_machine=ConversationStateMachine(),
    )


@router.post("/generate-response")
async def generate_response(
    conversation: dict,
    service: AIService = Depends(get_ai_service),
):
    try:
        return await service.generate_response(
            conversation
        )

    except ValueError as exc:
        raise HTTPException(
            status_code=400,
            detail=str(exc),
        )

    except Exception as exc:
        raise HTTPException(
            status_code=500,
            detail=str(exc),
        )