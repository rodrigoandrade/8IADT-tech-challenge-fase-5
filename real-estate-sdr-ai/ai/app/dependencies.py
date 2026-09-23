from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from app.infrastructure.database import get_session
from app.infrastructure.repositories.postgres_conversation_state_repository import (
    PostgresConversationStateRepository,
)
from app.services.ai_service import AIService
from app.services.openai_provider import OpenAIProvider
from app.services.prompt_builder import PromptBuilder
from app.services.response_parser import ResponseParser
from app.agent.state_machine import ConversationStateMachine
from app.clients.real_estate_api_client import RealEstateAPIClient


def get_conversation_state_repository(
    session: AsyncSession = Depends(get_session),
):
    return PostgresConversationStateRepository(session)