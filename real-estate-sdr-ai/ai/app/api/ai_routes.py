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
from ai.app.models.conversation import Conversation
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.rag.embeddings import EmbeddingService
from ai.app.rag.rag_service import RAGService
from ai.app.repositories.knowledge_repository import KnowledgeRepository
from ai.app.services.ai_service import AIService
from ai.app.agent.agent_router import AgentRouter

from ai.app.agent.property_agent import PropertyAgent
from ai.app.services.property_matching_service import PropertyMatchingService

router = APIRouter(
    prefix="/ai",
    tags=["AI"],
)


# =========================================================
# DEPENDENCIES
# =========================================================

def get_ai_service(
    session: AsyncSession = Depends(get_session),
) -> AIService:

    # -----------------------------------------------------
    # Environment / configuration
    # -----------------------------------------------------

    import os

    real_estate_api_url = os.getenv(
        "REAL_ESTATE_API_URL"
    )

    if not real_estate_api_url:
        raise RuntimeError(
            "REAL_ESTATE_API_URL is not configured"
        )

    # -----------------------------------------------------
    # Infrastructure
    # -----------------------------------------------------

    real_estate_api_client = RealEstateAPIClient(
        base_url=real_estate_api_url
    )

    state_repository = PostgresConversationStateRepository(
        session=session
    )

    # -----------------------------------------------------
    # LLM
    # -----------------------------------------------------

    llm_provider = OpenAIProvider()

    # -----------------------------------------------------
    # Prompt
    # -----------------------------------------------------

    prompt_builder = PromptBuilder()

    # -----------------------------------------------------
    # Response parser
    # -----------------------------------------------------

    response_parser = ResponseParser()

    # -----------------------------------------------------
    # State machine
    # -----------------------------------------------------

    state_machine = ConversationStateMachine()

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
        api_client=real_estate_api_client,
        llm_provider=llm_provider,
        prompt_builder=prompt_builder,
        response_parser=response_parser,
        state_repository=state_repository,
        state_machine=state_machine,
        rag_service=rag_service,
        agent_router=AgentRouter(),
        knowledge_agent=KnowledgeAgent(
            llm_provider=llm_provider,
            prompt_builder=prompt_builder,
        ),
        property_agent=PropertyAgent(
            property_matcher=PropertyMatchingService(
                api_client=real_estate_api_client,
            ),
        ),
    )


# =========================================================
# ROUTES
# =========================================================

@router.post("/generate-response")
async def generate_response(
    conversation: Conversation,
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
        ) from exc

    except Exception as exc:
        import logging

        logging.exception(
            "Erro ao gerar resposta da IA"
        )

        raise HTTPException(
            status_code=500,
            detail=str(exc),
        ) from exc