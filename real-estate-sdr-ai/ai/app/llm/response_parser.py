import json
from typing import Any

from ai.app.models.ai_response import (
    AIResponse,
    LeadIntent,
    LeadStatus,
)


class ResponseParser:
    """
    Converte a resposta textual produzida pelo LLM em AIResponse.

    O parser aceita respostas JSON no formato atual do domínio e também
    alguns aliases utilizados pelo prompt do agente.
    """

    def parse(
        self,
        response: str,
    ) -> AIResponse:
        """
        Faz o parse da resposta do LLM.

        Raises:
            ValueError: quando a resposta não é um JSON válido ou
                        contém dados incompatíveis com o domínio.
        """

        if not response or not response.strip():
            raise ValueError("LLM response cannot be empty")

        data = self._parse_json(response)

        return self._build_response(data)

    # ============================================================
    # JSON
    # ============================================================

    def _parse_json(
        self,
        response: str,
    ) -> dict[str, Any]:
        """
        Converte a resposta textual em um dicionário.
        """

        cleaned_response = self._clean_response(response)

        try:
            data = json.loads(cleaned_response)

        except json.JSONDecodeError as exc:
            raise ValueError(
                "LLM response is not valid JSON"
            ) from exc

        if not isinstance(data, dict):
            raise ValueError(
                "LLM response must be a JSON object"
            )

        return data

    @staticmethod
    def _clean_response(
        response: str,
    ) -> str:
        """
        Remove blocos Markdown eventualmente retornados pelo LLM.

        Aceita:

            ```json
            {...}
            ```

        e:

            ```
            {...}
            ```
        """

        cleaned = response.strip()

        if cleaned.startswith("```json"):
            cleaned = cleaned[len("```json"):]

        elif cleaned.startswith("```"):
            cleaned = cleaned[len("```"):]

        if cleaned.endswith("```"):
            cleaned = cleaned[:-3]

        return cleaned.strip()

    # ============================================================
    # RESPONSE
    # ============================================================

    def _build_response(
        self,
        data: dict[str, Any],
    ) -> AIResponse:
        """
        Constrói o AIResponse a partir do JSON retornado pelo LLM.
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

        # Compatibilidade com o prompt antigo,
        # que utiliza "lead_score" de 0 a 100.
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

        # Caso o LLM tenha retornado property_id,
        # preservamos essa informação nos metadados.
        property_id = data.get("property_id")

        if property_id is not None:
            metadata["property_id"] = str(property_id)

        return AIResponse(
            message=message,
            intent=intent,
            lead_status=lead_status,
            qualification_score=qualification_score,
            next_action=next_action,
            reasoning=reasoning,
            metadata=metadata,
        )

    # ============================================================
    # MESSAGE
    # ============================================================

    @staticmethod
    def _required_string(
        data: dict[str, Any],
        field: str,
    ) -> str:

        value = data.get(field)

        if not isinstance(value, str) or not value.strip():
            raise ValueError(
                f"required field '{field}' is missing or invalid"
            )

        return value.strip()

    @staticmethod
    def _optional_string(
        value: Any,
    ) -> str | None:

        if value is None:
            return None

        if not isinstance(value, str):
            raise ValueError(
                "optional string field has invalid type"
            )

        value = value.strip()

        return value or None

    # ============================================================
    # INTENT
    # ============================================================

    @staticmethod
    def _parse_intent(
        value: Any,
    ) -> LeadIntent:

        if value is None:
            return LeadIntent.UNKNOWN

        if not isinstance(value, str):
            raise ValueError(
                "intent must be a string"
            )

        normalized = value.strip().lower()

        # Valores nativos do domínio
        try:
            return LeadIntent(normalized)

        except ValueError:
            pass

        # Aliases utilizados pelo LLM/prompt
        aliases = {
            "buy": LeadIntent.PROPERTY_INTEREST,
            "purchase": LeadIntent.PROPERTY_INTEREST,
            "buying": LeadIntent.PROPERTY_INTEREST,

            "rent": LeadIntent.PROPERTY_INTEREST,
            "rental": LeadIntent.PROPERTY_INTEREST,
            "renting": LeadIntent.PROPERTY_INTEREST,

            "property_interest": LeadIntent.PROPERTY_INTEREST,
            "property": LeadIntent.PROPERTY_INTEREST,

            "information": LeadIntent.GENERAL_INQUIRY,
            "info": LeadIntent.GENERAL_INQUIRY,
            "general_inquiry": LeadIntent.GENERAL_INQUIRY,

            "price": LeadIntent.PRICE_INQUIRY,
            "price_inquiry": LeadIntent.PRICE_INQUIRY,

            "visit": LeadIntent.SCHEDULING,
            "schedule": LeadIntent.SCHEDULING,
            "scheduling": LeadIntent.SCHEDULING,

            "negotiation": LeadIntent.NEGOTIATION,
        }

        intent = aliases.get(normalized)

        if intent is not None:
            return intent

        raise ValueError(
            f"invalid lead intent: {value}"
        )

    # ============================================================
    # LEAD STATUS
    # ============================================================

    @staticmethod
    def _parse_lead_status(
        value: Any,
    ) -> LeadStatus:

        if value is None:
            return LeadStatus.UNKNOWN

        if not isinstance(value, str):
            raise ValueError(
                "lead_status must be a string"
            )

        normalized = value.strip().lower()

        # Valores nativos do domínio
        try:
            return LeadStatus(normalized)

        except ValueError:
            pass

        # Status provenientes da API Go.
        aliases = {
            "new": LeadStatus.NEW,

            "contacted": LeadStatus.QUALIFIED,

            "qualified": LeadStatus.QUALIFIED,

            "in_negotiation": LeadStatus.IN_NEGOTIATION,
            "negotiation": LeadStatus.IN_NEGOTIATION,

            "scheduled": LeadStatus.SCHEDULED,

            "converted": LeadStatus.CONVERTED,

            "lost": LeadStatus.LOST,
        }

        status = aliases.get(normalized)

        if status is not None:
            return status

        raise ValueError(
            f"invalid lead status: {value}"
        )

    # ============================================================
    # SCORE
    # ============================================================

    @staticmethod
    def _parse_score(
        value: Any,
    ) -> float | None:

        if value is None:
            return None

        if isinstance(value, bool):
            raise ValueError(
                "qualification_score must be a number"
            )

        try:
            score = float(value)

        except (TypeError, ValueError) as exc:
            raise ValueError(
                "qualification_score must be a number"
            ) from exc

        if not 0.0 <= score <= 1.0:
            raise ValueError(
                "qualification_score must be between 0 and 1"
            )

        return score

    @staticmethod
    def _parse_lead_score(
        value: Any,
    ) -> float:

        if isinstance(value, bool):
            raise ValueError(
                "lead_score must be a number"
            )

        try:
            score = float(value)

        except (TypeError, ValueError) as exc:
            raise ValueError(
                "lead_score must be a number"
            ) from exc

        if not 0.0 <= score <= 100.0:
            raise ValueError(
                "lead_score must be between 0 and 100"
            )

        # Converte 0-100 para 0-1,
        # que é o formato utilizado pelo domínio.
        return score / 100.0

    # ============================================================
    # METADATA
    # ============================================================

    @staticmethod
    def _parse_metadata(
        value: Any,
    ) -> dict[str, str]:

        if value is None:
            return {}

        if not isinstance(value, dict):
            raise ValueError(
                "metadata must be an object"
            )

        metadata: dict[str, str] = {}

        for key, item in value.items():

            if not isinstance(key, str):
                raise ValueError(
                    "metadata keys must be strings"
                )

            # Em vez de exigir que o LLM retorne somente strings,
            # convertemos valores simples para string.
            metadata[key] = str(item)

        return metadata
