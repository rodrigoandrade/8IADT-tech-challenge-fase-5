# ai/app/workers/follow_up_worker.py

import asyncio
import logging

from ai.app.infrastructure.database import AsyncSessionLocal
from ai.app.infrastructure.repositories.postgres_conversation_state_repository import (
    PostgresConversationStateRepository,
)
from ai.app.services.follow_up_service import FollowUpService


logger = logging.getLogger(__name__)


class FollowUpWorker:

    def __init__(
        self,
        api_client,
        inactivity_minutes: int = 1,
        max_follow_ups: int = 1,
        interval_seconds: int = 10,
    ):
        self.api_client = api_client
        self.inactivity_minutes = inactivity_minutes
        self.max_follow_ups = max_follow_ups
        self.interval_seconds = interval_seconds

        self._task: asyncio.Task | None = None
        self._running = False

    async def start(self) -> None:
        """
        Inicia o worker em background.
        """

        if self._running:
            logger.warning(
                "Follow-up worker já está em execução."
            )
            return

        self._running = True

        self._task = asyncio.create_task(
            self._run(),
            name="follow-up-worker",
        )

        logger.info(
            "Follow-up worker iniciado. "
            "Intervalo=%s segundos, "
            "inatividade=%s minutos, "
            "max_follow_ups=%s",
            self.interval_seconds,
            self.inactivity_minutes,
            self.max_follow_ups,
        )

    async def stop(self) -> None:
        """
        Para o worker de forma segura.
        """

        if not self._running:
            return

        logger.info(
            "Parando follow-up worker..."
        )

        self._running = False

        if self._task is not None:
            self._task.cancel()

            try:
                await self._task

            except asyncio.CancelledError:
                pass

            finally:
                self._task = None

        logger.info(
            "Follow-up worker finalizado."
        )

    async def _run(self) -> None:
        """
        Loop principal do worker.
        """

        while self._running:

            try:
                await self._process()

            except asyncio.CancelledError:
                raise

            except Exception:
                logger.exception(
                    "Erro durante execução "
                    "do follow-up worker."
                )

            try:
                await asyncio.sleep(
                    self.interval_seconds
                )

            except asyncio.CancelledError:
                raise

    async def _process(self) -> None:
        """
        Executa uma rodada do processamento.

        Uma nova sessão de banco é criada para cada rodada,
        evitando manter uma sessão aberta durante todo o ciclo
        de vida do worker.
        """

        async with AsyncSessionLocal() as session:

            state_repository = (
                PostgresConversationStateRepository(
                    session
                )
            )

            follow_up_service = FollowUpService(
                api_client=self.api_client,
                state_repository=state_repository,
                inactivity_minutes=self.inactivity_minutes,
                max_follow_ups=self.max_follow_ups,
            )

            sent_count = await follow_up_service.process()

            if sent_count > 0:
                logger.info(
                    "Follow-ups enviados nesta rodada: %s",
                    sent_count,
                )