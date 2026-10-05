import asyncio
from types import SimpleNamespace
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from ai.app.workers.follow_up_worker import FollowUpWorker


@pytest.mark.asyncio
async def test_start_creates_worker_task():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
        inactivity_minutes=1,
        max_follow_ups=1,
        interval_seconds=60,
    )

    with patch.object(
        worker,
        "_run",
        new_callable=AsyncMock,
    ) as mock_run:
        await worker.start()

        assert worker._running is True
        assert worker._task is not None

        await asyncio.sleep(0)

        mock_run.assert_awaited_once()

        await worker.stop()

    assert worker._running is False
    assert worker._task is None


@pytest.mark.asyncio
async def test_start_does_not_create_duplicate_task():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
        interval_seconds=60,
    )

    with patch.object(
        worker,
        "_run",
        new_callable=AsyncMock,
    ) as mock_run:
        await worker.start()

        first_task = worker._task

        await worker.start()

        second_task = worker._task

        assert first_task is second_task

        await asyncio.sleep(0)

        mock_run.assert_awaited_once()

        await worker.stop()


@pytest.mark.asyncio
async def test_stop_cancels_running_task():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
        interval_seconds=60,
    )

    async def blocking_run():
        try:
            await asyncio.Event().wait()
        except asyncio.CancelledError:
            raise

    with patch.object(
        worker,
        "_run",
        side_effect=blocking_run,
    ):
        await worker.start()

        assert worker._running is True
        assert worker._task is not None

        await asyncio.sleep(0)

        await worker.stop()

    assert worker._running is False
    assert worker._task is None


@pytest.mark.asyncio
async def test_stop_is_safe_when_worker_is_not_running():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
    )

    assert worker._running is False
    assert worker._task is None

    await worker.stop()

    assert worker._running is False
    assert worker._task is None


@pytest.mark.asyncio
async def test_process_creates_follow_up_service_and_processes_conversations():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
        inactivity_minutes=5,
        max_follow_ups=2,
        interval_seconds=60,
    )

    mock_follow_up_service = MagicMock()
    mock_follow_up_service.process = AsyncMock(
        return_value=2
    )

    with (
        patch(
            "ai.app.workers.follow_up_worker.AsyncSessionLocal"
        ) as mock_session_local,
        patch(
            "ai.app.workers.follow_up_worker.PostgresConversationStateRepository"
        ) as mock_repository,
        patch(
            "ai.app.workers.follow_up_worker.FollowUpService",
            return_value=mock_follow_up_service,
        ) as mock_service,
    ):
        session = AsyncMock()

        session_context = MagicMock()
        session_context.__aenter__ = AsyncMock(
            return_value=session
        )
        session_context.__aexit__ = AsyncMock(
            return_value=None
        )

        mock_session_local.return_value = session_context

        await worker._process()

    mock_session_local.assert_called_once()

    mock_repository.assert_called_once_with(
        session
    )

    mock_service.assert_called_once_with(
        api_client=api_client,
        state_repository=mock_repository.return_value,
        inactivity_minutes=5,
        max_follow_ups=2,
    )

    mock_follow_up_service.process.assert_awaited_once()


@pytest.mark.asyncio
async def test_run_calls_process_repeatedly():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
        interval_seconds=0,
    )

    process_count = 0

    async def fake_process():
        nonlocal process_count

        process_count += 1

        if process_count >= 2:
            worker._running = False

    with patch.object(
            worker,
            "_process",
            new_callable=AsyncMock,
            side_effect=fake_process,
    ) as mock_process:
        worker._running = True

        await worker._run()

        assert process_count == 2
        assert mock_process.await_count == 2


@pytest.mark.asyncio
async def test_run_continues_when_process_raises_exception():
    api_client = MagicMock()

    worker = FollowUpWorker(
        api_client=api_client,
        interval_seconds=0,
    )

    process_count = 0

    async def fake_process():
        nonlocal process_count

        process_count += 1

        if process_count == 1:
            raise RuntimeError("temporary failure")

        worker._running = False

    with patch.object(
            worker,
            "_process",
            new_callable=AsyncMock,
            side_effect=fake_process,
    ) as mock_process:
        worker._running = True

        await worker._run()

        assert process_count == 2
        assert mock_process.await_count == 2