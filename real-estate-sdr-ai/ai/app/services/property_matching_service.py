from typing import Any

from ai.app.agent.state import ConversationState
from ai.app.clients.real_estate_api_client import RealEstateAPIClient


class PropertyMatchingService:

    def __init__(
        self,
        api_client: RealEstateAPIClient,
    ):
        self.api_client = api_client

    async def match(
        self,
        state: ConversationState,
    ) -> list[dict[str, Any]]:

        properties = await self.api_client.get_properties()

        print("\n========== PROPERTY MATCH ==========")
        print(f"Total properties: {len(properties)}")
        print(f"property_type: {state.property_type}")
        print(f"budget: {state.budget}")
        print(f"bedrooms: {state.bedrooms}")
        print(f"region: {state.region}")

        matched: list[dict[str, Any]] = []

        for property_data in properties:

            property_id = property_data.get("id")
            title = property_data.get("title")

            available = self._is_available(
                property_data
            )

            property_type = self._matches_property_type(
                property_data,
                state.property_type,
            )

            budget = self._matches_budget(
                property_data,
                state.budget,
            )

            bedrooms = self._matches_bedrooms(
                property_data,
                state.bedrooms,
            )

            region = self._matches_region(
                property_data,
                state.region,
            )

            print(
                f"\nProperty: {title}"
                f"\n  id: {property_id}"
                f"\n  available: {available}"
                f"\n  property_type: {property_type}"
                f"\n  budget: {budget}"
                f"\n  bedrooms: {bedrooms}"
                f"\n  region: {region}"
            )

            if not available:
                continue

            if not property_type:
                continue

            if not budget:
                continue

            if not bedrooms:
                continue

            if not region:
                continue

            matched.append(property_data)

        matched.sort(
            key=self._property_price
        )

        result = matched[:3]

        print("\n========== MATCHED ==========")

        for property_data in result:
            print(
                property_data.get("title"),
                property_data.get("price"),
            )

        print("=============================\n")

        return result

    @staticmethod
    def _is_available(
        property_data: dict[str, Any],
    ) -> bool:

        status = property_data.get("status")

        if status is not None:
            return str(status).strip().upper() == "AVAILABLE"

        available = property_data.get(
            "available",
            True,
        )

        if isinstance(available, str):
            return available.strip().lower() in {
                "true",
                "1",
                "yes",
                "sim",
            }

        return bool(available)

    @staticmethod
    def _normalize(value: str) -> str:
        import unicodedata

        normalized = unicodedata.normalize(
            "NFD",
            value.lower().strip(),
        )

        return "".join(
            character
            for character in normalized
            if unicodedata.category(character) != "Mn"
        )

    def _matches_property_type(
        self,
        property_data: dict[str, Any],
        property_type: str | None,
    ) -> bool:

        if not property_type:
            return True

        actual = property_data.get("property_type")

        if actual is None:
            actual = property_data.get("type")

        if not actual:
            return False

        normalized_actual = self._normalize(
            str(actual)
        )

        normalized_requested = self._normalize(
            property_type
        )

        aliases = {
            "apartamento": {
                "apartamento",
                "apartment",
            },
            "casa": {
                "casa",
                "house",
            },
            "terreno": {
                "terreno",
                "land",
            },
            "sala comercial": {
                "sala comercial",
                "commercial room",
            },
        }

        allowed = aliases.get(
            normalized_requested
        )

        if allowed:
            return normalized_actual in allowed

        return normalized_actual == normalized_requested

    @staticmethod
    def _matches_budget(
        property_data: dict[str, Any],
        budget: float | None,
    ) -> bool:

        if budget is None:
            return True

        price = property_data.get("price")

        if price is None:
            return False

        try:
            return float(price) <= float(budget)
        except (TypeError, ValueError):
            return False

    @staticmethod
    def _matches_bedrooms(
        property_data: dict[str, Any],
        bedrooms: int | None,
    ) -> bool:

        if bedrooms is None:
            return True

        value = property_data.get("bedrooms")

        if value is None:
            return False

        try:
            return int(value) >= int(bedrooms)
        except (TypeError, ValueError):
            return False

    def _matches_region(
        self,
        property_data: dict[str, Any],
        region: str | None,
    ) -> bool:

        if not region:
            return True

        normalized_region = self._normalize(region)

        city = self._normalize(
            str(property_data.get("city", ""))
        )

        neighborhood = self._normalize(
            str(property_data.get("neighborhood", ""))
        )

        address = self._normalize(
            str(property_data.get("address", ""))
        )

        title = self._normalize(
            str(property_data.get("title", ""))
        )

        searchable = " ".join(
            [
                city,
                neighborhood,
                address,
                title,
            ]
        )

        if normalized_region in searchable:
            return True

        if normalized_region in {
            "zona sul",
            "sul",
        }:

            south_neighborhoods = {
                "moema",
                "vila mariana",
                "saude",
                "campo belo",
                "campo grande",
                "santo amaro",
                "ipiranga",
                "brooklin",
                "morumbi",
                "vila andrade",
                "chacara flora",
            }

            return neighborhood in {
                self._normalize(name)
                for name in south_neighborhoods
            }

        return False

    @staticmethod
    def _property_price(
        property_data: dict[str, Any],
    ) -> float:

        value = property_data.get("price")

        if value is None:
            return float("inf")

        try:
            return float(value)
        except (TypeError, ValueError):
            return float("inf")