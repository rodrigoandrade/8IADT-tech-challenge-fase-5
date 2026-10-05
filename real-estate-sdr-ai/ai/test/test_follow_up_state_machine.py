import pytest

from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state_machine import ConversationStateMachine


def test_enter_follow_up_preserves_previous_state():

    machine = ConversationStateMachine()

    state = ConversationState(
        conversation_id="conversation-follow-up",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.QUALIFY_BUDGET,
        next_action=NextAction.ASK_BUDGET,
        budget=None,
    )

    result = machine.enter_follow_up(state)

    assert result.state == ConversationStateType.WAITING_LEAD

    assert result.next_action is None

    assert result.waiting_for_customer is True

    assert result.previous_state == (
        ConversationStateType.QUALIFY_BUDGET
    )

    assert result.previous_next_action == (
        NextAction.ASK_BUDGET
    )

def test_waiting_lead_transitions_to_follow_up_pending():

    machine = ConversationStateMachine()

    state = ConversationState(
        conversation_id="conversation-follow-up",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.WAITING_LEAD,
        next_action=None,
        previous_state=ConversationStateType.QUALIFY_BUDGET,
        previous_next_action=NextAction.ASK_BUDGET,
        waiting_for_customer=True,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.FOLLOW_UP_PENDING

    assert result.next_action is None

    assert result.waiting_for_customer is True

def test_follow_up_pending_transitions_to_send_follow_up():

    machine = ConversationStateMachine()

    state = ConversationState(
        conversation_id="conversation-follow-up",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.FOLLOW_UP_PENDING,
        next_action=None,
        follow_up_count=0,
        waiting_for_customer=True,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.SEND_FOLLOW_UP

    assert result.next_action == NextAction.FOLLOW_UP

    assert result.follow_up_count == 1

def test_send_follow_up_returns_to_waiting_lead():

    machine = ConversationStateMachine()

    state = ConversationState(
        conversation_id="conversation-follow-up",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.SEND_FOLLOW_UP,
        next_action=NextAction.FOLLOW_UP,
        follow_up_count=1,
    )

    result = machine.transition(state)

    assert result.state == ConversationStateType.WAITING_LEAD

    assert result.next_action is None

    assert result.waiting_for_customer is True

def test_resume_after_follow_up_restores_previous_state():

    machine = ConversationStateMachine()

    state = ConversationState(
        conversation_id="conversation-follow-up",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.WAITING_LEAD,
        next_action=None,
        previous_state=ConversationStateType.QUALIFY_BUDGET,
        previous_next_action=NextAction.ASK_BUDGET,
        waiting_for_customer=True,
    )

    result = machine.resume_after_follow_up(state)

    assert result.state == (
        ConversationStateType.QUALIFY_BUDGET
    )

    assert result.next_action == NextAction.ASK_BUDGET

    assert result.previous_state is None

    assert result.previous_next_action is None

    assert result.waiting_for_customer is False