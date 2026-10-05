import json
from typing import Any

from ai.app.models.ai_response import AIResponse, LeadIntent, LeadStatus


class ResponseParser:
    """
    Converte a resposta JSON produzida pelo LLM em um AIResponse
    validado e tipado.
    """

    def parse(self, response: str) -> AIResponse:
        if not response or not response.strip():
            raise ValueError("LLM response cannot be empty")

        data = self._parse_json(response)

        return self._build_response(data)

    def _parse_json(self, response: str) -> dict[str, Any]:
        """
        Faz o parse do JSON retornado pelo LLM.

        Aceita também respostas envolvidas em markdown fences:
        ```json
        {...}
        ```
        """

        content = response.strip()

        if content.startswith("```"):
            lines = content.splitlines()

            if lines and lines[0].startswith("```"):
                lines = lines[1:]

            if lines and lines[-1].strip() == "```":
                lines = lines[:-1]

            content = "\n".join(lines).strip()

        try:
            data = json.loads(content)
        except json.JSONDecodeError as exc:
            raise ValueError(
                f"Invalid JSON returned by LLM: {exc}"
            ) from exc

        if not isinstance(data, dict):
            raise ValueError("LLM response must be a JSON object")

        return data

    def _build_response(self, data: dict[str, Any]) -> AIResponse:
        """
        Constrói o AIResponse a partir do JSON normalizado.
        """

        message = self._required_string(
            data,
            "message",
        )

        intent = self._parse_intent(
            data.get("intent")
        )

        lead_status = self._parse_lead_status(
            data.get("lead_status")
        )

        qualification_score = self._parse_score(
            data.get("qualification_score")
        )

        if qualification_score is None:
            lead_score = data.get("lead_score")

            if lead_score is not None:
                qualification_score = self._parse_lead_score(
                    lead_score
                )

        next_action = self._optional_string(
            data.get("next_action")
        )

        reasoning = self._optional_string(
            data.get("reasoning")
        )

        metadata = self._parse_metadata(
            data.get("metadata")
        )

        property_id = data.get("property_id")

        if property_id is not None:
            metadata["property_id"] = str(property_id)

        transaction_type = self._parse_transaction_type(
            data.get("transaction_type")
        )

        qualification = self._parse_qualification(
            data.get("qualification")
        )

        return AIResponse(
            message=message,
            intent=intent,
            lead_status=lead_status,
            qualification_score=qualification_score,
            next_action=next_action,
            reasoning=reasoning,
            metadata=metadata,
            transaction_type=transaction_type,
            qualification=qualification,
        )

    # ------------------------------------------------------------------
    # Intent
    # ------------------------------------------------------------------

    def _parse_intent(
        self,
        value: Any,
    ) -> LeadIntent:

        if value is None:
            return LeadIntent.UNKNOWN

        normalized = str(value).strip().lower()

        aliases = {
            "property": LeadIntent.PROPERTY_INTEREST,
            "property_interest": LeadIntent.PROPERTY_INTEREST,
            "imovel": LeadIntent.PROPERTY_INTEREST,
            "interesse_imovel": LeadIntent.PROPERTY_INTEREST,
            "interesse_em_imovel": LeadIntent.PROPERTY_INTEREST,

            "general": LeadIntent.GENERAL_INQUIRY,
            "general_inquiry": LeadIntent.GENERAL_INQUIRY,
            "duvida": LeadIntent.GENERAL_INQUIRY,
            "informacao": LeadIntent.GENERAL_INQUIRY,

            "price": LeadIntent.PRICE_INQUIRY,
            "price_inquiry": LeadIntent.PRICE_INQUIRY,
            "preco": LeadIntent.PRICE_INQUIRY,
            "valor": LeadIntent.PRICE_INQUIRY,

            "schedule": LeadIntent.SCHEDULING,
            "scheduling": LeadIntent.SCHEDULING,
            "agendamento": LeadIntent.SCHEDULING,
            "meeting": LeadIntent.SCHEDULING,

            "negotiation": LeadIntent.NEGOTIATION,
            "negociacao": LeadIntent.NEGOTIATION,
            "negociação": LeadIntent.NEGOTIATION,

            "unknown": LeadIntent.UNKNOWN,
        }

        if normalized in aliases:
            return aliases[normalized]

        try:
            return LeadIntent(normalized)
        except ValueError:
            return LeadIntent.UNKNOWN

    # ------------------------------------------------------------------
    # Lead status
    # ------------------------------------------------------------------

    def _parse_lead_status(
        self,
        value: Any,
    ) -> LeadStatus:

        if value is None:
            return LeadStatus.UNKNOWN

        normalized = str(value).strip().lower()

        aliases = {
            "novo": LeadStatus.NEW,
            "new": LeadStatus.NEW,

            "contatado": LeadStatus.CONTACTED,
            "contacted": LeadStatus.CONTACTED,

            "qualificado": LeadStatus.QUALIFIED,
            "qualified": LeadStatus.QUALIFIED,

            "convertido": LeadStatus.CONVERTED,
            "converted": LeadStatus.CONVERTED,

            "perdido": LeadStatus.LOST,
            "lost": LeadStatus.LOST,

            "unknown": LeadStatus.UNKNOWN,
        }

        if normalized in aliases:
            return aliases[normalized]

        try:
            return LeadStatus(normalized)
        except ValueError:
            return LeadStatus.UNKNOWN

    # ------------------------------------------------------------------
    # Transaction type
    # ------------------------------------------------------------------

    def _parse_transaction_type(
        self,
        value: Any,
    ) -> str | None:

        if value is None:
            return None

        normalized = str(value).strip().lower()

        if not normalized:
            return None

        aliases = {
            "compra": "buy",
            "comprar": "buy",
            "buy": "buy",

            "investimento": "investment",
            "investir": "investment",
            "investment": "investment",
        }

        return aliases.get(normalized, normalized)

    # ------------------------------------------------------------------
    # Qualification
    # ------------------------------------------------------------------

    def _parse_qualification(
        self,
        value: Any,
    ) -> dict[str, Any]:

        if value is None:
            return {}

        if not isinstance(value, dict):
            raise ValueError(
                "qualification must be a JSON object"
            )

        qualification: dict[str, Any] = {}

        # Região
        region = value.get("region")

        if region is not None:
            region = str(region).strip()

            if region:
                qualification["region"] = region

        # Orçamento
        budget = self._parse_optional_float(
            value.get("budget")
        )

        if budget is not None:
            qualification["budget"] = budget

        # Quartos
        bedrooms = self._parse_optional_int(
            value.get("bedrooms")
        )

        if bedrooms is not None:
            qualification["bedrooms"] = bedrooms

        # Urgência
        urgency = value.get("urgency")

        if urgency is not None:
            urgency = str(urgency).strip()

            if urgency:
                qualification["urgency"] = urgency

        # Tipo do imóvel
        property_type = value.get("property_type")

        if property_type is not None:
            property_type = str(property_type).strip().lower()

            if property_type:
                qualification["property_type"] = property_type

        # Perfil do investidor
        investor_profile = value.get("investor_profile")

        if investor_profile is not None:
            investor_profile = str(
                investor_profile
            ).strip()

            if investor_profile:
                qualification["investor_profile"] = investor_profile

        # Ticket de investimento
        investment_ticket = self._parse_optional_float(
            value.get("investment_ticket")
        )

        if investment_ticket is not None:
            qualification["investment_ticket"] = investment_ticket

        # Retorno esperado
        expected_return = self._parse_optional_float(
            value.get("expected_return")
        )

        if expected_return is not None:
            qualification["expected_return"] = expected_return

        return qualification

    # ------------------------------------------------------------------
    # Score
    # ------------------------------------------------------------------

    def _parse_score(
        self,
        value: Any,
    ) -> float | None:

        if value is None:
            return None

        try:
            score = float(value)
        except (TypeError, ValueError) as exc:
            raise ValueError(
                "qualification_score must be numeric"
            ) from exc

        if score < 0 or score > 1:
            raise ValueError(
                "qualification_score must be between 0 and 1"
            )

        return score

    def _parse_lead_score(
        self,
        value: Any,
    ) -> float | None:

        if value is None:
            return None

        try:
            score = float(value)
        except (TypeError, ValueError) as exc:
            raise ValueError(
                "lead_score must be numeric"
            ) from exc

        # Alguns prompts podem retornar score em escala 0-100.
        if score > 1:
            score = score / 100.0

        if score < 0 or score > 1:
            raise ValueError(
                "lead_score must be between 0 and 100"
            )

        return score

    # ------------------------------------------------------------------
    # Metadata
    # ------------------------------------------------------------------

    def _parse_metadata(
        self,
        value: Any,
    ) -> dict[str, Any]:

        if value is None:
            return {}

        if not isinstance(value, dict):
            raise ValueError(
                "metadata must be a JSON object"
            )

        return dict(value)

    # ------------------------------------------------------------------
    # Helpers
    # ------------------------------------------------------------------

    def _required_string(
        self,
        data: dict[str, Any],
        field_name: str,
    ) -> str:

        value = data.get(field_name)

        if value is None:
            raise ValueError(
                f"Required field '{field_name}' is missing"
            )

        value = str(value).strip()

        if not value:
            raise ValueError(
                f"Required field '{field_name}' cannot be empty"
            )

        return value

    def _optional_string(
        self,
        value: Any,
    ) -> str | None:

        if value is None:
            return None

        value = str(value).strip()

        return value if value else None

    def _parse_optional_float(
            self,
            value: Any,
    ) -> float | None:
        if value is None:
            return None

        if isinstance(value, str):
            value = value.strip()

            if not value:
                return None

            # Aceita percentuais:
            # "8%"   -> 8.0
            # "8 %"  -> 8.0
            # "8,5%" -> 8.5
            is_percentage = value.endswith("%")

            if is_percentage:
                value = value[:-1].strip()

            value = value.replace(",", ".")

        try:
            return float(value)
        except (TypeError, ValueError) as exc:
            raise ValueError(
                f"Expected numeric value, received: {value}"
            ) from exc

    def _parse_optional_int(
        self,
        value: Any,
    ) -> int | None:

        if value is None:
            return None

        if isinstance(value, str):
            value = value.strip()

            if not value:
                return None

        try:
            return int(value)
        except (TypeError, ValueError) as exc:
            raise ValueError(
                f"Expected integer value, received: {value}"
            ) from exc