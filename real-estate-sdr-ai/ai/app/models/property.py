
from dataclasses import dataclass, field
from typing import Any


@dataclass(frozen=True)
class Property:
    """
    Representa um imóvel dentro do contexto da aplicação de IA.
    """

    id: str
    title: str

    description: str | None = None

    property_type: str | None = None

    city: str | None = None
    state: str | None = None
    neighborhood: str | None = None
    address: str | None = None

    price: float | None = None

    bedrooms: int | None = None
    bathrooms: int | None = None
    parking_spaces: int | None = None

    area: float | None = None

    available: bool = True

    metadata: dict[str, Any] = field(
        default_factory=dict
    )

    def __post_init__(self) -> None:
        if not self.id.strip():
            raise ValueError("property id cannot be empty")

        if not self.title.strip():
            raise ValueError("property title cannot be empty")

        if self.price is not None and self.price < 0:
            raise ValueError("property price cannot be negative")

        if self.area is not None and self.area < 0:
            raise ValueError("property area cannot be negative")

        if self.bedrooms is not None and self.bedrooms < 0:
            raise ValueError("bedrooms cannot be negative")

        if self.bathrooms is not None and self.bathrooms < 0:
            raise ValueError("bathrooms cannot be negative")

        if self.parking_spaces is not None and self.parking_spaces < 0:
            raise ValueError("parking spaces cannot be negative")

