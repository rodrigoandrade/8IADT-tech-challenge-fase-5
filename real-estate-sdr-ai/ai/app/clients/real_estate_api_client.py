from typing import Any

import httpx


class RealEstateAPIClient:
    """
    HTTP client responsible for communicating with the
    real-estate-sdr-api service.

    All operations are asynchronous because the AI service
    is running on FastAPI.
    """

    def __init__(
        self,
        base_url: str,
        timeout: float = 10.0,
    ):
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout

    async def create_message(
        self,
        conversation_id: str,
        role: str,
        content: str,
    ) -> dict[str, Any]:

        payload = {
            "conversation_id": conversation_id,
            "role": role,
            "content": content,
        }

        url = f"{self._base_url}/messages"

        try:
            async with httpx.AsyncClient(
                timeout=self._timeout
            ) as client:

                response = await client.post(
                    url,
                    json=payload,
                )

        except httpx.RequestError as exc:
            raise RuntimeError(
                "Failed to communicate with "
                f"real-estate-sdr-api: {exc}"
            ) from exc

        if response.status_code >= 400:
            raise RuntimeError(
                "real-estate-sdr-api returned "
                f"HTTP {response.status_code}: "
                f"{response.text}"
            )

        try:
            return response.json()

        except ValueError as exc:
            raise RuntimeError(
                "Invalid JSON response from "
                f"real-estate-sdr-api: {response.text}"
            ) from exc

    async def get_conversation(
        self,
        conversation_id: str,
    ) -> dict[str, Any]:

        return await self._get(
            f"/conversations/{conversation_id}"
        )

    async def get_conversations(
            self,
            limit: int = 100,
            offset: int = 0,
    ) -> list[dict[str, Any]]:

        response = await self._get(
            "/conversations",
            params={
                "limit": limit,
                "offset": offset,
            },
        )

        if isinstance(response, list):
            return response

        if isinstance(response, dict):
            items = response.get("items")

            if isinstance(items, list):
                return items

            value = response.get("value")

            if isinstance(value, list):
                return value

        return []

    async def get_lead(
        self,
        lead_id: str,
    ) -> dict[str, Any]:

        return await self._get(
            f"/leads/{lead_id}"
        )

    async def get_properties(
        self,
        params: dict[str, Any] | None = None,
    ) -> list[dict[str, Any]]:

        response = await self._get(
            "/properties",
            params=params,
        )

        if isinstance(response, list):
            return response

        if isinstance(response, dict):
            items = response.get("items", [])

            if isinstance(items, list):
                return items

        return []

    async def get_messages(
        self,
        conversation_id: str,
    ) -> list[dict[str, Any]]:

        response = await self._get(
            f"/conversations/{conversation_id}/messages"
        )

        if isinstance(response, list):
            return response

        if isinstance(response, dict):
            items = response.get("items", [])

            if isinstance(items, list):
                return items

        return []

    async def _get(
        self,
        path: str,
        params: dict[str, Any] | None = None,
    ) -> Any:

        url = f"{self._base_url}{path}"

        try:
            async with httpx.AsyncClient(
                timeout=self._timeout
            ) as client:

                response = await client.get(
                    url,
                    params=params,
                )

        except httpx.RequestError as exc:
            raise RuntimeError(
                f"Failed to communicate with real-estate-sdr-api: {exc}"
            ) from exc

        if response.status_code == 404:
            raise ValueError(
                f"Resource not found: {path}"
            )

        if response.status_code >= 400:
            raise RuntimeError(
                "real-estate-sdr-api returned "
                f"HTTP {response.status_code}: "
                f"{response.text}"
            )

        try:
            return response.json()

        except ValueError as exc:
            raise RuntimeError(
                f"Invalid JSON response from real-estate-sdr-api: {path}"
            ) from exc