from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from ai.app.agent.knowledge_agent import KnowledgeAgent

from ai.app.agent.state_machine import ConversationStateMachine
from ai.app.clients.real_estate_api_client import RealEstateAPIClient
from ai.app.infrastructure.database import get_session
from ai.app.infrastructure.repositories.postgres_conversation_state_repository import (
    PostgresConversationStateRepository,
)
from ai.app.llm.openai_provider import OpenAIProvider
from ai.app.llm.response_parser import ResponseParser
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.rag.embeddings import EmbeddingService
from ai.app.rag.rag_service import RAGService
from ai.app.repositories.knowledge_repository import KnowledgeRepository
from ai.app.services.ai_service import AIService
from ai.app.agent.agent_router import AgentRouter

router = APIRouter(
    prefix="/ai",
    tags=["AI"],
)


def get_ai_service(
    session: AsyncSession = Depends(get_session),
) -> AIService:

    # -----------------------------------------------------
    # State repository
    # -----------------------------------------------------

    state_repository = PostgresConversationStateRepository(
        session
    )

    # -----------------------------------------------------
    # RAG
    # -----------------------------------------------------

    embedding_service = EmbeddingService()

    knowledge_repository = KnowledgeRepository()

    rag_service = RAGService(
        embedding_service=embedding_service,
        knowledge_repository=knowledge_repository,
    )

    # -----------------------------------------------------
    # AI Service
    # -----------------------------------------------------

    return AIService(
        api_client=RealEstateAPIClient(),
        llm_provider=OpenAIProvider(),
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=state_repository,
        state_machine=ConversationStateMachine(),
        rag_service=rag_service,
        agent_router=AgentRouter(),
        knowledge_agent=KnowledgeAgent(),
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