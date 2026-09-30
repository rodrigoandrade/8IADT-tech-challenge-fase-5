from typing import Any


class PropertyMatcher:

    def __init__(self, ai_service: Any):
        self.ai_service = ai_service

    async def match(self, state: Any) -> list[dict[str, Any]]:
        return await self.ai_service._match_properties(state)