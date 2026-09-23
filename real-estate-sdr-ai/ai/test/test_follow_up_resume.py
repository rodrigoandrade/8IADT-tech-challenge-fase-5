from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state_machine import ConversationStateMachine


machine = ConversationStateMachine()


# ============================================================
# TESTE 1
# Entrar em follow-up preservando estado anterior
# ============================================================

state = ConversationState(
    conversation_id="test-resume-001",
    flow=ConversationFlow.BUY,
    state=ConversationStateType.QUALIFY_BEDROOMS,
    next_action=NextAction.ASK_BEDROOMS,
    region="zona sul",
    budget=800000,
    property_type="apartamento",
    waiting_for_customer=False,
)

print("\n" + "=" * 60)
print("ESTADO ORIGINAL")
print("=" * 60)

print("state:", state.state)
print("next_action:", state.next_action)
print("previous_state:", state.previous_state)
print("region:", state.region)
print("budget:", state.budget)


# Entra em follow-up
state = machine.enter_follow_up(state)

print("\n" + "=" * 60)
print("APÓS enter_follow_up()")
print("=" * 60)

print("state:", state.state)
print("next_action:", state.next_action)
print("previous_state:", state.previous_state)
print("waiting:", state.waiting_for_customer)
print("region:", state.region)
print("budget:", state.budget)


assert state.state == ConversationStateType.WAITING_LEAD
assert state.next_action is None
assert state.previous_state == ConversationStateType.QUALIFY_BEDROOMS
assert state.waiting_for_customer is True

# Dados da qualificação devem continuar preservados
assert state.region == "zona sul"
assert state.budget == 800000
assert state.property_type == "apartamento"

print("\n✅ TESTE 1 PASSOU")


# ============================================================
# TESTE 2
# Retomar conversa após follow-up
# ============================================================

state = machine.resume_after_follow_up(state)

print("\n" + "=" * 60)
print("APÓS resume_after_follow_up()")
print("=" * 60)

print("state:", state.state)
print("next_action:", state.next_action)
print("previous_state:", state.previous_state)
print("waiting:", state.waiting_for_customer)
print("region:", state.region)
print("budget:", state.budget)


assert state.state == ConversationStateType.QUALIFY_BEDROOMS
assert state.next_action == NextAction.ASK_BEDROOMS
assert state.waiting_for_customer is False

# Dados devem continuar preservados
assert state.region == "zona sul"
assert state.budget == 800000
assert state.property_type == "apartamento"

print("\n✅ TESTE 2 PASSOU")


# ============================================================
# RESULTADO
# ============================================================

print("\n" + "=" * 60)
print("🎉 TODOS OS TESTES DE RESUME PASSARAM")
print("=" * 60)