import pytest

from ai.app.agent.property_agent import PropertyAgent
from ai.app.agent.response import AgentResponse


class FakePropertyMatcher:

    def __init__(self, properties):
        self.properties = properties
        self.called = False
        self.received_state = None

    async def match(self, state):
        self.called = True
        self.received_state = state
        return self.properties


@pytest.mark.asyncio
async def test_property_agent_returns_matched_properties():

    properties = [
        {
            "id": "property-1",
            "title": "Apartamento Moema",
            "price": 475000,
        },
        {
            "id": "property-2",
            "title": "Apartamento Vila Mariana",
            "price": 490000,
        },
    ]

    matcher = FakePropertyMatcher(properties)
    agent = PropertyAgent(property_matcher=matcher)

    state = object()

    result = await agent.execute(state=state)

    assert isinstance(result, AgentResponse)

    assert matcher.called is True
    assert matcher.received_state is state

    assert result.message == (
        "Encontrei imóveis que correspondem "
        "aos critérios informados."
    )

    assert result.matched_properties == properties
    assert result.selected_property_id == "property-1"


@pytest.mark.asyncio
async def test_property_agent_when_no_properties_are_found():

    matcher = FakePropertyMatcher([])
    agent = PropertyAgent(property_matcher=matcher)

    result = await agent.execute(state=object())

    assert isinstance(result, AgentResponse)

    assert result.message == (
        "Não encontrei imóveis que correspondam "
        "aos critérios informados."
    )

    assert result.matched_properties == []
    assert result.selected_property_id is None