from ai.app.agent.agent_type import AgentType
from ai.app.agent.base_agent import BaseAgent


class QualificationAgent(BaseAgent):

    @property
    def agent_type(self) -> AgentType:
        return AgentType.QUALIFICATION

    async def execute(self, **kwargs):
        raise NotImplementedError(
            "QualificationAgent ainda será integrado ao fluxo existente."
        )