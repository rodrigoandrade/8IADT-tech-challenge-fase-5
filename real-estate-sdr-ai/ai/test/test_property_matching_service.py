import pytest

from ai.app.agent.state import ConversationState
from ai.app.services.property_matching_service import (
    PropertyMatchingService,
)


class FakeAPIClient:

    async def get_properties(self):

        return [
            {
                "id": "1",
                "title": "Apartamento Moema",
                "type": "APARTMENT",
                "price": 475000,
                "bedrooms": 4,
                "status": "AVAILABLE",
                "neighborhood": "Moema",
                "city": "São Paulo",
            },
            {
                "id": "2",
                "title": "Apartamento Caro",
                "type": "APARTMENT",
                "price": 700000,
                "bedrooms": 4,
                "status": "AVAILABLE",
                "neighborhood": "Moema",
                "city": "São Paulo",
            },
            {
                "id": "3",
                "title": "Casa Moema",
                "type": "HOUSE",
                "price": 400000,
                "bedrooms": 4,
                "status": "AVAILABLE",
                "neighborhood": "Moema",
                "city": "São Paulo",
            },
        ]


@pytest.mark.asyncio
async def test_matches_properties_by_qualification():

    service = PropertyMatchingService(
        api_client=FakeAPIClient()
    )

    state = ConversationState(
        conversation_id="conversation-1",
        property_type="apartamento",
        budget=500000,
        bedrooms=3,
        region="Moema",
    )

    result = await service.match(state)

    assert len(result) == 1
    assert result[0]["id"] == "1"
    assert result[0]["title"] == "Apartamento Moema"