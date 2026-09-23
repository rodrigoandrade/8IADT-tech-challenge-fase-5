from dataclasses import dataclass, field
from enum import Enum
from typing import Any


class LeadIntent(str, Enum):
    PROPERTY_INTEREST = "property_interest"
    GENERAL_INQUIRY = "general_inquiry"
    PRICE_INQUIRY = "price_inquiry"
    SCHEDULING = "scheduling"
    NEGOTIATION = "negotiation"
    UNKNOWN = "unknown"


class LeadStatus(str, Enum):
    NEW = "new"
    CONTACTED = "contacted"
    QUALIFIED = "qualified"
    CONVERTED = "converted"
    LOST = "lost"
    UNKNOWN = "unknown"


@dataclass(frozen=True)
class AIResponse:
    """
    Resposta estruturada produzida pelo serviço de IA.

    Contém tanto a resposta textual para o lead quanto
    os dados estruturados utilizados para atualizar o
    estado da conversa.
    """

    message: str

    intent: LeadIntent = LeadIntent.UNKNOWN

    lead_status: LeadStatus = LeadStatus.UNKNOWN

    qualification_score: float | None = None

    next_action: str | None = None

    reasoning: str | None = None

    metadata: dict[str, Any] = field(default_factory=dict)

    # Tipo da negociação identificado pelo LLM.
    # Exemplos: "buy", "investment".
    transaction_type: str | None = None

    # Dados de qualificação extraídos da conversa.
    #
    # Exemplo:
    # {
    #     "region": "zona sul",
    #     "budget": 800000,
    #     "bedrooms": 2,
    #     "urgency": "3 meses",
    #     "property_type": "apartamento",
    #     "investor_profile": None,
    #     "investment_ticket": None,
    #     "expected_return": None
    # }
    qualification: dict[str, Any] = field(default_factory=dict)

    def __post_init__(self):
        if not self.message or not self.message.strip():
            raise ValueError("message cannot be empty")

        if self.qualification is None:
            object.__setattr__(self, "qualification", {})

        if self.metadata is None:
            object.__setattr__(self, "metadata", {})

    @property
    def property_id(self) -> str | None:
        """
        Retorna o ID do imóvel quando informado nos metadados.
        """
        value = self.metadata.get("property_id")

        if value is None:
            return None

        return str(value)

    @property
    def region(self) -> str | None:
        """
        Retorna a região identificada na qualificação.
        """
        value = self.qualification.get("region")

        if value is None:
            return None

        return str(value)

    @property
    def budget(self) -> float | None:
        """
        Retorna o orçamento identificado na qualificação.
        """
        value = self.qualification.get("budget")

        if value is None:
            return None

        try:
            return float(value)
        except (TypeError, ValueError):
            return None

    @property
    def bedrooms(self) -> int | None:
        """
        Retorna a quantidade de quartos identificada.
        """
        value = self.qualification.get("bedrooms")

        if value is None:
            return None

        try:
            return int(value)
        except (TypeError, ValueError):
            return None

    @property
    def urgency(self) -> str | None:
        """
        Retorna a urgência identificada.
        """
        value = self.qualification.get("urgency")

        if value is None:
            return None

        return str(value)

    @property
    def property_type(self) -> str | None:
        """
        Retorna o tipo de imóvel identificado.
        """
        value = self.qualification.get("property_type")

        if value is None:
            return None

        return str(value)

    @property
    def investor_profile(self) -> str | None:
        """
        Retorna o perfil do investidor identificado.
        """
        value = self.qualification.get("investor_profile")

        if value is None:
            return None

        return str(value)

    @property
    def investment_ticket(self) -> float | None:
        """
        Retorna o ticket disponível para investimento.
        """
        value = self.qualification.get("investment_ticket")

        if value is None:
            return None

        try:
            return float(value)
        except (TypeError, ValueError):
            return None

    @property
    def expected_return(self) -> float | None:
        """
        Retorna o retorno esperado pelo investidor.
        """
        value = self.qualification.get("expected_return")

        if value is None:
            return None

        try:
            return float(value)
        except (TypeError, ValueError):
            return None