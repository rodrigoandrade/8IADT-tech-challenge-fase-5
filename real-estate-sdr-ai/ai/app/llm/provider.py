from abc import ABC, abstractmethod


class LLMProvider(ABC):

    @abstractmethod
    async def generate_response(
        self,
        prompt: str,
    ) -> str:
        """
        Generates a textual AI response based on the provided prompt.
        """
        raise NotImplementedError