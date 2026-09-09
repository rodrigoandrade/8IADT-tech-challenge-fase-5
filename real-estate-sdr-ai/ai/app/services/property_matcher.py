from typing import Any


class PropertyMatcher:

    SOUTH_ZONE_NEIGHBORHOODS = {
        "moema",
        "vila mariana",
        "saúde",
        "saude",
        "campo belo",
        "campo grande",
        "santo amaro",
        "ipiranga",
        "chácara santo amaro",
        "chacara santo amaro",
        "brooklin",
        "brooklin novo",
        "morumbi",
        "vila andrade",
        "chácara flora",
        "chacara flora",
    }

    def match(
        self,
        properties: list[dict[str, Any]],
        property_type: str | None = None,
        budget: float | None = None,
        bedrooms: int | None = None,
        region: str | None = None,
    ) -> list[dict[str, Any]]:

        candidates = []

        for property_ in properties:

            if not self._is_available(property_):
                continue

            if not self._matches_property_type(
                property_,
                property_type,
            ):
                continue

            if not self._matches_budget(
                property_,
                budget,
            ):
                continue

            if not self._matches_bedrooms(
                property_,
                bedrooms,
            ):
                continue

            if not self._matches_region(
                property_,
                region,
            ):
                continue

            candidates.append(property_)

        return sorted(
            candidates,
            key=lambda item: item.get("price", float("inf")),
        )

    @staticmethod
    def _is_available(
        property_: dict[str, Any],
    ) -> bool:

        return str(
            property_.get("status", "")
        ).upper() == "AVAILABLE"

    @staticmethod
    def _matches_property_type(
        property_: dict[str, Any],
        property_type: str | None,
    ) -> bool:

        if not property_type:
            return True

        requested = property_type.strip().lower()

        actual = str(
            property_.get("type", "")
        ).strip().lower()

        mapping = {
            "apartamento": "apartment",
            "apartamentos": "apartment",
            "apartment": "apartment",
            "casa": "house",
            "casas": "house",
            "house": "house",
            "terreno": "land",
            "terrenos": "land",
            "land": "land",
            "cobertura": "apartment",
        }

        requested_normalized = mapping.get(
            requested,
            requested,
        )

        return actual == requested_normalized

    @staticmethod
    def _matches_budget(
        property_: dict[str, Any],
        budget: float | None,
    ) -> bool:

        if budget is None:
            return True

        price = property_.get("price")

        if price is None:
            return False

        try:
            return float(price) <= budget
        except (TypeError, ValueError):
            return False

    @staticmethod
    def _matches_bedrooms(
        property_: dict[str, Any],
        bedrooms: int | None,
    ) -> bool:

        if bedrooms is None:
            return True

        value = property_.get("bedrooms")

        if value is None:
            return False

        try:
            # Se o cliente pede 2 quartos,
            # aceitamos 2 ou mais.
            return int(value) >= bedrooms
        except (TypeError, ValueError):
            return False

    def _matches_region(
        self,
        property_: dict[str, Any],
        region: str | None,
    ) -> bool:

        if not region:
            return True

        requested = self._normalize(region)

        neighborhood = self._normalize(
            property_.get("neighborhood", "")
        )

        city = self._normalize(
            property_.get("city", "")
        )

        address = self._normalize(
            property_.get("address", "")
        )

        # Busca direta por bairro/cidade/endereço.
        if requested in neighborhood:
            return True

        if requested in city:
            return True

        if requested in address:
            return True

        # Regra específica para "Zona Sul".
        if self._is_south_zone(requested):

            return neighborhood in {
                self._normalize(item)
                for item in self.SOUTH_ZONE_NEIGHBORHOODS
            }

        return False

    def _is_south_zone(
        self,
        region: str,
    ) -> bool:

        south_terms = {
            "zona sul",
            "sul",
            "zona sul de sao paulo",
            "zona sul sao paulo",
        }

        return region in south_terms

    @staticmethod
    def _normalize(value: Any) -> str:

        if value is None:
            return ""

        return (
            str(value)
            .strip()
            .lower()
        )