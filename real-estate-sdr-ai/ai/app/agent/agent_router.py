from ai.app.agent.agent_type import AgentType


class AgentRouter:

    def route(self, message: str) -> AgentType:
        normalized = message.lower().strip()

        if self._is_knowledge_question(normalized):
            return AgentType.KNOWLEDGE

        if self._is_property_request(normalized):
            return AgentType.PROPERTY

        return AgentType.QUALIFICATION

    def _is_knowledge_question(self, message: str) -> bool:
        keywords = (
            "financiamento",
            "financiar",
            "visita",
            "visitar",
            "atendem",
            "atende",
            "região",
            "regioes",
            "regiões",
            "como funciona",
        )

        return any(keyword in message for keyword in keywords)

    def _is_property_request(self, message: str) -> bool:
        keywords = (
            "me mostre",
            "mostrar",
            "encontre",
            "encontrar",
            "imóveis",
            "imoveis",
            "apartamentos",
            "casas",
            "opções",
            "opcoes",
        )

        return any(keyword in message for keyword in keywords)