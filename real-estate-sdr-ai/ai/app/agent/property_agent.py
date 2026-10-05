from typing import Any

from ai.app.agent.agent_type import AgentType
from ai.app.agent.base_agent import BaseAgent
from ai.app.agent.response import AgentResponse


class PropertyAgent(BaseAgent):

    def __init__(self, property_matcher: Any):
        self.property_matcher = property_matcher

    @property
    def agent_type(self) -> AgentType:
        return AgentType.PROPERTY

    async def execute(
        self,
        *,
        state: Any,
    ) -> AgentResponse:

        matched_properties = await self.property_matcher.match(
            state
        )

        if not matched_properties:
            return AgentResponse(
                message=(
                    "Não encontrei imóveis que correspondam "
                    "aos critérios informados."
                )
            )

        selected_property_id = matched_properties[0].get("id")

        return AgentResponse(
            message=(
                "Encontrei imóveis que correspondem "
                "aos critérios informados."
            ),
            matched_properties=matched_properties,
            selected_property_id=selected_property_id,
        )