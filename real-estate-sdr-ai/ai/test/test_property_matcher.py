import pytest

from ai.app.agent.property_matcher import PropertyMatcher


class FakeAIService:

    def __init__(self):
        self.called = False
        self.received_state = None

    async def _match_properties(self, state):
        self.called = True
        self.received_state = state

        return [
            {
                "id": "property-1",
                "title": "Apartamento Teste",
                "price": 475000,
            }
        ]


@pytest.mark.asyncio
async def test_property_matcher_delegates_to_ai_service():

    ai_service = FakeAIService()

    matcher = PropertyMatcher(
        ai_service=ai_service
    )

    state = object()

    result = await matcher.match(state)

    assert ai_service.called is True
    assert ai_service.received_state is state

    assert len(result) == 1
    assert result[0]["id"] == "property-1"