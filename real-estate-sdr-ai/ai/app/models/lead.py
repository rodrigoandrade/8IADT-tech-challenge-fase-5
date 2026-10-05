
from dataclasses import dataclass, field
from typing import Any


@dataclass(frozen=True)
class Lead:
    """
    Representa um lead dentro do contexto da aplicação de IA.
    """

    id: str
    name: str

    phone: str | None = None
    email: str | None = None

    source: str | None = None

    preferences: dict[str, Any] = field(
        default_factory=dict
    )

    def __post_init__(self) -> None:
        if not self.id.strip():
            raise ValueError("lead id cannot be empty")

        if not self.name.strip():
            raise ValueError("lead name cannot be empty")

