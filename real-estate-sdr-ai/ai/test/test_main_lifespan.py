from unittest.mock import AsyncMock, patch

import pytest

from ai.app.main import app, lifespan


@pytest.mark.asyncio
async def test_lifespan_starts_follow_up_worker():

    with patch(
        "ai.app.main.follow_up_worker"
    ) as mock_worker:

        mock_worker.start = AsyncMock()
        mock_worker.stop = AsyncMock()

        async with lifespan(app):

            mock_worker.start.assert_awaited_once()

            mock_worker.stop.assert_not_awaited()

        mock_worker.stop.assert_awaited_once()


@pytest.mark.asyncio
async def test_lifespan_stops_worker_when_application_shuts_down():

    with patch(
        "ai.app.main.follow_up_worker"
    ) as mock_worker:

        mock_worker.start = AsyncMock()
        mock_worker.stop = AsyncMock()

        async with lifespan(app):

            assert mock_worker.start.await_count == 1

        assert mock_worker.stop.await_count == 1


@pytest.mark.asyncio
async def test_lifespan_stops_worker_even_when_application_raises():

    with patch(
        "ai.app.main.follow_up_worker"
    ) as mock_worker:

        mock_worker.start = AsyncMock()
        mock_worker.stop = AsyncMock()

        with pytest.raises(RuntimeError):

            async with lifespan(app):

                mock_worker.start.assert_awaited_once()

                raise RuntimeError(
                    "simulated application error"
                )

        mock_worker.stop.assert_awaited_once()