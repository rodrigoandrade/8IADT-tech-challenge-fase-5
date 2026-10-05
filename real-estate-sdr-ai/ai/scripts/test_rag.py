import asyncio
from pathlib import Path

from dotenv import load_dotenv

from ai.app.infrastructure.database import AsyncSessionLocal
from ai.app.rag.embeddings import EmbeddingService
from ai.app.rag.rag_service import RAGService
from ai.app.repositories.knowledge_repository import KnowledgeRepository


BASE_DIR = Path(__file__).resolve().parents[2]
load_dotenv(BASE_DIR / ".env")


async def main():
    rag_service = RAGService(
        embedding_service=EmbeddingService(),
        knowledge_repository=KnowledgeRepository(),
    )

    questions = [
        "Quais regiões de São Paulo vocês atendem?",
        "Posso financiar o imóvel?",
        "Como faço para visitar um imóvel?",
    ]

    async with AsyncSessionLocal() as session:
        for question in questions:
            print("\n" + "=" * 80)
            print(f"PERGUNTA: {question}")
            print("=" * 80)

            results = await rag_service.retrieve(
                session=session,
                question=question,
                limit=2,
            )

            for result in results:
                print(
                    f"\nTítulo: {result['title']}"
                    f"\nSimilaridade: {result['similarity']:.4f}"
                    f"\nConteúdo: {result['content']}"
                )


if __name__ == "__main__":
    asyncio.run(main())