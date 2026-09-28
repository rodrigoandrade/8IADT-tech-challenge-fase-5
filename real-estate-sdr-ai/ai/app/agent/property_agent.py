from ai.app.agent.agent_type import AgentType
from ai.app.agent.base_agent import BaseAgent


class PropertyAgent(BaseAgent):

    @property
    def agent_type(self) -> AgentType:
        return AgentType.PROPERTY

    async def execute(self, **kwargs):
        raise NotImplementedError(
            "PropertyAgent ainda será integrado ao fluxo existente."
        )