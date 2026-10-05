from sqlalchemy.ext.asyncio import AsyncSession

from ai.app.rag.embeddings import EmbeddingService
from ai.app.repositories.knowledge_repository import KnowledgeRepository


class RAGService:

    def __init__(
        self,
        embedding_service: EmbeddingService,
        knowledge_repository: KnowledgeRepository,
    ):
        self.embedding_service = embedding_service
        self.knowledge_repository = knowledge_repository

    async def index_document(
        self,
        session: AsyncSession,
        document_id,
        title: str,
        content: str,
        source: str | None = None,
        metadata: dict | None = None,
    ) -> None:

        embedding = await self.embedding_service.generate(
            content
        )

        await self.knowledge_repository.save(
            session=session,
            document_id=document_id,
            title=title,
            content=content,
            embedding=embedding,
            source=source,
            metadata=metadata,
        )

    async def retrieve(
        self,
        session: AsyncSession,
        question: str,
        limit: int = 3,
    ) -> list[dict]:

        embedding = await self.embedding_service.generate(
            question
        )

        return await self.knowledge_repository.search(
            session=session,
            embedding=embedding,
            limit=limit,
        )