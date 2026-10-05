from uuid import UUID

from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession


class KnowledgeRepository:

    async def save(
        self,
        session: AsyncSession,
        document_id: UUID,
        title: str,
        content: str,
        embedding: list[float],
        source: str | None = None,
        metadata: dict | None = None,
    ) -> None:

        embedding_string = "[" + ",".join(
            str(value) for value in embedding
        ) + "]"

        query = text(
            """
            INSERT INTO knowledge_documents (
                id,
                title,
                content,
                source,
                metadata,
                embedding,
                created_at,
                updated_at
            )
            VALUES (
                :id,
                :title,
                :content,
                :source,
                :metadata,
                CAST(:embedding AS vector),
                NOW(),
                NOW()
            )
            """
        )

        await session.execute(
            query,
            {
                "id": document_id,
                "title": title,
                "content": content,
                "source": source,
                "metadata": metadata,
                "embedding": embedding_string,
            },
        )

        await session.commit()

    async def search(
        self,
        session: AsyncSession,
        embedding: list[float],
        limit: int = 3,
    ) -> list[dict]:

        embedding_string = "[" + ",".join(
            str(value) for value in embedding
        ) + "]"

        query = text(
            """
            SELECT
                id,
                title,
                content,
                source,
                metadata,
                1 - (
                    embedding <=> CAST(:embedding AS vector)
                ) AS similarity
            FROM knowledge_documents
            WHERE embedding IS NOT NULL
            ORDER BY embedding <=> CAST(:embedding AS vector)
            LIMIT :limit
            """
        )

        result = await session.execute(
            query,
            {
                "embedding": embedding_string,
                "limit": limit,
            },
        )

        return [
            dict(row)
            for row in result.mappings().all()
        ]