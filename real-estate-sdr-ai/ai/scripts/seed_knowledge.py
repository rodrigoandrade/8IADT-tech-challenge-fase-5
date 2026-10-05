import asyncio
import os
from pathlib import Path
from uuid import uuid4

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

    documents = [
        {
            "title": "Regiões atendidas",
            "content": (
                "A imobiliária atua principalmente na Zona Sul "
                "de São Paulo, incluindo Moema, Vila Mariana, "
                "Campo Belo, Brooklin, Morumbi e Santo Amaro."
            ),
            "source": "commercial-policy",
        },
        {
            "title": "Financiamento",
            "content": (
                "Os imóveis podem ser adquiridos através de "
                "financiamento bancário, sujeito à análise de "
                "crédito e às condições estabelecidas pela "
                "instituição financeira."
            ),
            "source": "commercial-policy",
        },
        {
            "title": "Visitas",
            "content": (
                "As visitas aos imóveis devem ser agendadas "
                "previamente com um consultor imobiliário."
            ),
            "source": "commercial-policy",
        },
    ]

    async with AsyncSessionLocal() as session:

        for document in documents:

            await rag_service.index_document(
                session=session,
                document_id=uuid4(),
                title=document["title"],
                content=document["content"],
                source=document["source"],
            )

    print("Knowledge base criada com sucesso.")


if __name__ == "__main__":
    asyncio.run(main())