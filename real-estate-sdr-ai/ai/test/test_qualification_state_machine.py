import pytest

from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state_machine import ConversationStateMachine


@pytest.fixture
def machine():
    return ConversationStateMachine()


def test_identifies_intent_when_flow_is_missing(machine):
    state = ConversationState(
        conversation_id="test-intent",
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.IDENTIFY_INTENT
    assert result.next_action == NextAction.IDENTIFY_INTENT


def test_buy_flow_asks_budget_first(machine):
    state = ConversationState(
        conversation_id="test-buy-budget",
        flow=ConversationFlow.BUY,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.QUALIFY_BUDGET
    assert result.next_action == NextAction.ASK_BUDGET


def test_buy_flow_asks_bedrooms_after_budget(machine):
    state = ConversationState(
        conversation_id="test-buy-bedrooms",
        flow=ConversationFlow.BUY,
        budget=500000,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.QUALIFY_BEDROOMS
    assert result.next_action == NextAction.ASK_BEDROOMS


def test_buy_flow_asks_region_after_budget_and_bedrooms(machine):
    state = ConversationState(
        conversation_id="test-buy-region",
        flow=ConversationFlow.BUY,
        budget=500000,
        bedrooms=3,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.QUALIFY_REGION
    assert result.next_action == NextAction.ASK_LOCATION


def test_buy_flow_asks_urgency_after_region(machine):
    state = ConversationState(
        conversation_id="test-buy-urgency",
        flow=ConversationFlow.BUY,
        budget=500000,
        bedrooms=3,
        region="São Paulo",
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.QUALIFY_URGENCY
    assert result.next_action == NextAction.ASK_URGENCY


def test_buy_flow_starts_matching_after_qualification(machine):
    state = ConversationState(
        conversation_id="test-buy-matching",
        flow=ConversationFlow.BUY,
        budget=500000,
        bedrooms=3,
        region="São Paulo",
        urgency="within_3_months",
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.RECOMMEND_PROPERTIES
    assert result.next_action == NextAction.MATCH_PROPERTY


def test_buy_flow_recommends_selected_property(machine):
    state = ConversationState(
        conversation_id="test-buy-recommend",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.RECOMMEND_PROPERTIES,
        budget=500000,
        bedrooms=3,
        region="São Paulo",
        urgency="within_3_months",
        selected_property_id="property-test-001",
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.RECOMMEND_PROPERTIES
    assert result.next_action == NextAction.RECOMMEND_PROPERTY


def test_investment_flow_asks_investor_profile_first(machine):
    state = ConversationState(
        conversation_id="test-investment-profile",
        flow=ConversationFlow.INVESTMENT,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.INVESTOR_PROFILE
    assert result.next_action == NextAction.ASK_INVESTOR_PROFILE
