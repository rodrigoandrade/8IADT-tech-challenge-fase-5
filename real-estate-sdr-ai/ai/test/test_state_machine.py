from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state_machine import ConversationStateMachine


def print_state(title, state):
    print(f"\n{'=' * 60}")
    print(title)
    print(f"{'=' * 60}")
    print(f"flow:         {state.flow}")
    print(f"state:        {state.state}")
    print(f"next_action:  {state.next_action}")
    print(f"waiting:      {state.waiting_for_customer}")
    print(f"follow_up:    {state.follow_up_count}")


machine = ConversationStateMachine()


# ============================================================
# TESTE 1
# WAITING_LEAD -> FOLLOW_UP_PENDING
# ============================================================

state = ConversationState(
    conversation_id="test-follow-up-001",
    flow=ConversationFlow.BUY,
    state=ConversationStateType.WAITING_LEAD,
    next_action=NextAction.FOLLOW_UP,
    waiting_for_customer=True,
)

print_state(
    "ESTADO INICIAL - WAITING_LEAD",
    state,
)

state = machine.transition(state)

print_state(
    "APÓS TRANSITION",
    state,
)

assert state.state == ConversationStateType.FOLLOW_UP_PENDING
assert state.next_action is None

print("\n✅ TESTE 1 PASSOU")


# ============================================================
# TESTE 2
# FOLLOW_UP_PENDING -> SEND_FOLLOW_UP
# ============================================================

state = machine.transition(state)

print_state(
    "FOLLOW_UP_PENDING -> APÓS TRANSITION",
    state,
)

assert state.state == ConversationStateType.SEND_FOLLOW_UP
assert state.next_action == NextAction.FOLLOW_UP
assert state.follow_up_count == 1

print("\n✅ TESTE 2 PASSOU")


# ============================================================
# TESTE 3
# SEND_FOLLOW_UP -> WAITING_LEAD
# ============================================================

state = machine.transition(state)

print_state(
    "SEND_FOLLOW_UP -> APÓS TRANSITION",
    state,
)

assert state.state == ConversationStateType.WAITING_LEAD
assert state.next_action is None
assert state.waiting_for_customer is True

print("\n✅ TESTE 3 PASSOU")


# ============================================================
# TESTE 4
# FOLLOW-UP COM CONTAGEM
# ============================================================

state = ConversationState(
    conversation_id="test-follow-up-002",
    flow=ConversationFlow.BUY,
    state=ConversationStateType.WAITING_LEAD,
    next_action=NextAction.FOLLOW_UP,
    waiting_for_customer=True,
    follow_up_count=1,
)

print_state(
    "FOLLOW-UP COM follow_up_count=1",
    state,
)

state = machine.transition(state)

print_state(
    "APÓS TRANSITION",
    state,
)

assert state.state == ConversationStateType.FOLLOW_UP_PENDING
assert state.next_action is None

print("\n✅ TESTE 4 PASSOU")


# ============================================================
# RESULTADO FINAL
# ============================================================

print("\n")
print("=" * 60)
print("🎉 TODOS OS TESTES DO FOLLOW-UP PASSARAM")
print("=" * 60)