from abc import ABC, abstractmethod

from ai.app.agent.agent_type import AgentType


class BaseAgent(ABC):

    @property
    @abstractmethod
    def agent_type(self) -> AgentType:
        pass

    @abstractmethod
    async def execute(self, **kwargs):
        pass