import asyncio

from ai.app.agent.state import (
    ConversationFlow,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state_machine import ConversationStateMachine
from ai.app.llm.response_parser import ResponseParser
from ai.app.models.conversation import (
    Conversation,
    ConversationMessage,
    MessageRole,
)
from ai.app.prompt.prompt_builder import PromptBuilder
from ai.app.services.ai_service import AIService


# ============================================================
# FAKE STATE REPOSITORY
# ============================================================

class FakeStateRepository:

    def __init__(self):
        self.states = {}

    async def get(self, conversation_id):
        return self.states.get(conversation_id)

    async def save(self, state):
        self.states[state.conversation_id] = state


# ============================================================
# FAKE REAL ESTATE API
# ============================================================

class FakeRealEstateAPIClient:

    async def get_lead(self, lead_id):

        return {
            "id": lead_id,
            "name": "Rodrigo",
        }

    async def get_messages(self, conversation_id):

        return []

    async def create_message(
        self,
        conversation_id,
        role,
        content,
    ):

        pass

    async def get_properties(self):

        return [
            {
                "id": "property-001",
                "title": "Apartamento Vila Mariana",
                "property_type": "apartamento",
                "price": 750000,
                "bedrooms": 2,
                "neighborhood": "Vila Mariana",
                "city": "São Paulo",
                "status": "AVAILABLE",
            }
        ]


# ============================================================
# FAKE LLM
# ============================================================

class FakeLLMProvider:

    def __init__(self, responses):

        self.responses = responses
        self.index = 0

    async def generate_response(self, prompt):

        if self.index >= len(self.responses):

            raise RuntimeError(
                "FakeLLMProvider: no more responses configured"
            )

        response = self.responses[self.index]

        self.index += 1

        return response


# ============================================================
# CREATE SERVICE
# ============================================================

def create_service(
    repository,
    llm_provider,
):

    return AIService(
        api_client=FakeRealEstateAPIClient(),
        llm_provider=llm_provider,
        prompt_builder=PromptBuilder(),
        response_parser=ResponseParser(),
        state_repository=repository,
        state_machine=ConversationStateMachine(),
    )


# ============================================================
# TESTE 3.1
# PRIMEIRA MENSAGEM
# ============================================================

async def test_first_message(
    service,
    repository,
):

    conversation_id = "scenario-3-integration-001"

    conversation = Conversation(
        id=conversation_id,
        lead_id="lead-001",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content=(
                    "Estou procurando apartamento "
                    "na zona sul."
                ),
            )
        ],
    )

    result = await service.generate_response(
        conversation
    )

    print("\n" + "=" * 60)
    print("CENÁRIO 3 - TESTE 3.1")
    print("PRIMEIRA MENSAGEM")
    print("=" * 60)

    print("message:", result["message"])
    print("flow:", result["flow"])
    print("state:", result["state"])
    print("next_action:", result["next_action"])
    print("qualification:", result["qualification"])

    assert result["flow"] == ConversationFlow.BUY.value

    assert result["state"] == (
        ConversationStateType.QUALIFY_BUDGET.value
    )

    assert result["next_action"] == (
        NextAction.ASK_BUDGET.value
    )

    assert result["qualification"]["region"] == (
        "zona sul"
    )

    assert result["qualification"]["property_type"] == (
        "apartamento"
    )

    state = await repository.get(
        conversation_id
    )

    assert state is not None

    assert state.flow == ConversationFlow.BUY

    assert state.state == (
        ConversationStateType.QUALIFY_BUDGET
    )

    assert state.next_action == (
        NextAction.ASK_BUDGET
    )

    assert state.region == "zona sul"

    assert state.property_type == "apartamento"

    print("\n✅ TESTE 3.1 PASSOU")


# ============================================================
# TESTE 3.2
# LEAD INFORMA ORÇAMENTO
# ============================================================

async def test_budget_message(
    service,
    repository,
):

    conversation_id = "scenario-3-integration-001"

    conversation = Conversation(
        id=conversation_id,
        lead_id="lead-001",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content=(
                    "Estou procurando apartamento "
                    "na zona sul."
                ),
            ),
            ConversationMessage(
                role=MessageRole.ASSISTANT,
                content=(
                    "Claro! Qual é o seu orçamento?"
                ),
            ),
            ConversationMessage(
                role=MessageRole.LEAD,
                content="Até 800 mil.",
            ),
        ],
    )

    result = await service.generate_response(
        conversation
    )

    print("\n" + "=" * 60)
    print("CENÁRIO 3 - TESTE 3.2")
    print("LEAD INFORMA ORÇAMENTO")
    print("=" * 60)

    print("message:", result["message"])
    print("flow:", result["flow"])
    print("state:", result["state"])
    print("next_action:", result["next_action"])
    print("qualification:", result["qualification"])

    assert result["flow"] == ConversationFlow.BUY.value

    assert result["state"] == (
        ConversationStateType.QUALIFY_BEDROOMS.value
    )

    assert result["next_action"] == (
        NextAction.ASK_BEDROOMS.value
    )

    assert result["qualification"]["region"] == (
        "zona sul"
    )

    assert result["qualification"]["property_type"] == (
        "apartamento"
    )

    assert result["qualification"]["budget"] == 800000

    state = await repository.get(
        conversation_id
    )

    assert state is not None

    assert state.flow == ConversationFlow.BUY

    assert state.state == (
        ConversationStateType.QUALIFY_BEDROOMS
    )

    assert state.next_action == (
        NextAction.ASK_BEDROOMS
    )

    assert state.region == "zona sul"

    assert state.property_type == "apartamento"

    assert state.budget == 800000

    print("\n✅ TESTE 3.2 PASSOU")


# ============================================================
# TESTE 3.3
# LEAD PARA DE RESPONDER
# ============================================================

async def test_follow_up_cycle(
    repository,
):

    conversation_id = "scenario-3-integration-001"

    state = await repository.get(
        conversation_id
    )

    assert state is not None

    print("\n" + "=" * 60)
    print("CENÁRIO 3 - TESTE 3.3")
    print("LEAD PARA DE RESPONDER")
    print("=" * 60)

    print("ESTADO ANTES DO FOLLOW-UP")
    print("flow:", state.flow)
    print("state:", state.state)
    print("next_action:", state.next_action)
    print("region:", state.region)
    print("budget:", state.budget)
    print("property_type:", state.property_type)

    assert state.flow == ConversationFlow.BUY

    assert state.state == (
        ConversationStateType.QUALIFY_BEDROOMS
    )

    assert state.next_action == (
        NextAction.ASK_BEDROOMS
    )

    assert state.region == "zona sul"

    assert state.budget == 800000

    assert state.property_type == "apartamento"

    machine = ConversationStateMachine()

    # --------------------------------------------------------
    # ENTER FOLLOW-UP
    # --------------------------------------------------------

    state = machine.enter_follow_up(
        state
    )

    print("\nESTADO APÓS enter_follow_up()")
    print("flow:", state.flow)
    print("state:", state.state)
    print("next_action:", state.next_action)
    print("previous_state:", state.previous_state)
    print(
        "previous_next_action:",
        state.previous_next_action,
    )
    print("waiting:", state.waiting_for_customer)

    assert state.state == (
        ConversationStateType.WAITING_LEAD
    )

    assert state.next_action is None

    assert state.previous_state == (
        ConversationStateType.QUALIFY_BEDROOMS
    )

    assert state.previous_next_action == (
        NextAction.ASK_BEDROOMS
    )

    assert state.waiting_for_customer is True

    assert state.region == "zona sul"

    assert state.budget == 800000

    assert state.property_type == "apartamento"

    await repository.save(state)

    # --------------------------------------------------------
    # FOLLOW_UP_PENDING
    # --------------------------------------------------------

    state = machine.transition(
        state
    )

    print("\nESTADO APÓS transition()")
    print("state:", state.state)
    print("next_action:", state.next_action)
    print("waiting:", state.waiting_for_customer)

    assert state.state == (
        ConversationStateType.FOLLOW_UP_PENDING
    )

    assert state.next_action is None

    # --------------------------------------------------------
    # SEND_FOLLOW_UP
    # --------------------------------------------------------

    state = machine.transition(
        state
    )

    print("\nESTADO APÓS SEGUNDA transition()")
    print("state:", state.state)
    print("next_action:", state.next_action)
    print("follow_up_count:", state.follow_up_count)

    assert state.state == (
        ConversationStateType.SEND_FOLLOW_UP
    )

    assert state.next_action == (
        NextAction.FOLLOW_UP
    )

    assert state.follow_up_count == 1

    # --------------------------------------------------------
    # AFTER SEND
    # --------------------------------------------------------

    state = machine.transition(
        state
    )

    print("\nESTADO APÓS ENVIO DO FOLLOW-UP")
    print("state:", state.state)
    print("next_action:", state.next_action)
    print("waiting:", state.waiting_for_customer)

    assert state.state == (
        ConversationStateType.WAITING_LEAD
    )

    assert state.next_action is None

    assert state.waiting_for_customer is True

    assert state.flow == ConversationFlow.BUY

    assert state.region == "zona sul"

    assert state.budget == 800000

    assert state.property_type == "apartamento"

    assert state.previous_state == (
        ConversationStateType.QUALIFY_BEDROOMS
    )

    assert state.previous_next_action == (
        NextAction.ASK_BEDROOMS
    )

    await repository.save(state)

    print("\n✅ TESTE 3.3 PASSOU")


# ============================================================
# TESTE 3.4
# LEAD RESPONDE APÓS FOLLOW-UP
# ============================================================

async def test_lead_returns_after_follow_up(
    service,
    repository,
    llm_provider,
):

    conversation_id = "scenario-3-integration-001"

    state = await repository.get(
        conversation_id
    )

    assert state is not None

    print("\n" + "=" * 60)
    print("CENÁRIO 3 - TESTE 3.4")
    print("LEAD RESPONDE APÓS FOLLOW-UP")
    print("=" * 60)

    print("ESTADO ANTES DA RESPOSTA")
    print("flow:", state.flow)
    print("state:", state.state)
    print("next_action:", state.next_action)
    print("previous_state:", state.previous_state)
    print(
        "previous_next_action:",
        state.previous_next_action,
    )
    print("region:", state.region)
    print("budget:", state.budget)

    # --------------------------------------------------------
    # Deve estar aguardando o lead
    # --------------------------------------------------------

    assert state.state == (
        ConversationStateType.WAITING_LEAD
    )

    assert state.waiting_for_customer is True

    assert state.previous_state == (
        ConversationStateType.QUALIFY_BEDROOMS
    )

    assert state.previous_next_action == (
        NextAction.ASK_BEDROOMS
    )

    # --------------------------------------------------------
    # NOVA MENSAGEM
    # --------------------------------------------------------

    conversation = Conversation(
        id=conversation_id,
        lead_id="lead-001",
        messages=[
            ConversationMessage(
                role=MessageRole.LEAD,
                content=(
                    "Estou procurando apartamento "
                    "na zona sul."
                ),
            ),
            ConversationMessage(
                role=MessageRole.ASSISTANT,
                content=(
                    "Claro! Qual é o seu orçamento?"
                ),
            ),
            ConversationMessage(
                role=MessageRole.LEAD,
                content="Até 800 mil.",
            ),
            ConversationMessage(
                role=MessageRole.ASSISTANT,
                content=(
                    "Perfeito! Quantos quartos "
                    "você procura?"
                ),
            ),
            ConversationMessage(
                role=MessageRole.LEAD,
                content="Quero 2 quartos.",
            ),
        ],
    )

    # --------------------------------------------------------
    # RESPOSTA DO LLM
    # --------------------------------------------------------

    llm_provider.responses.append(
        """
        {
            "message": "Perfeito! Vou considerar apartamentos de 2 quartos.",
            "intent": "purchase",
            "transaction_type": "buy",
            "qualification": {
                "bedrooms": 2
            },
            "next_action": "ask_location"
        }
        """
    )

    # --------------------------------------------------------
    # EXECUTAR AI SERVICE
    # --------------------------------------------------------

    result = await service.generate_response(
        conversation
    )

    print("\nRESULTADO APÓS RESPOSTA DO LEAD")
    print("message:", result["message"])
    print("flow:", result["flow"])
    print("state:", result["state"])
    print("next_action:", result["next_action"])
    print("qualification:", result["qualification"])

    # --------------------------------------------------------
    # ASSERTIONS
    # --------------------------------------------------------

    assert result["flow"] == (
        ConversationFlow.BUY.value
    )

    assert result["state"] != (
        ConversationStateType.WAITING_LEAD.value
    )

    assert result["qualification"]["bedrooms"] == 2

    assert result["qualification"]["region"] == (
        "zona sul"
    )

    assert result["qualification"]["budget"] == 800000

    assert result["qualification"]["property_type"] == (
        "apartamento"
    )

    # --------------------------------------------------------
    # ESTADO PERSISTIDO
    # --------------------------------------------------------

    persisted_state = await repository.get(
        conversation_id
    )

    assert persisted_state is not None

    assert persisted_state.flow == (
        ConversationFlow.BUY
    )

    assert persisted_state.region == "zona sul"

    assert persisted_state.budget == 800000

    assert persisted_state.property_type == (
        "apartamento"
    )

    assert persisted_state.bedrooms == 2

    assert persisted_state.waiting_for_customer is False

    assert persisted_state.previous_state is None

    assert persisted_state.previous_next_action is None

    print("\n✅ TESTE 3.4 PASSOU")
    print("   lead respondeu após follow-up")
    print("   bedrooms = 2")
    print("   região preservada")
    print("   orçamento preservado")
    print("   tipo de imóvel preservado")
    print("   waiting_for_customer = False")
    print("   contexto de follow-up limpo")


# ============================================================
# TEST RUNNER
# ============================================================

async def main():

    repository = FakeStateRepository()

    llm_provider = FakeLLMProvider(
        [
            # ------------------------------------------------
            # TESTE 3.1
            # ------------------------------------------------
            """
            {
                "message": "Claro! Qual é o seu orçamento?",
                "intent": "purchase",
                "transaction_type": "buy",
                "qualification": {
                    "region": "zona sul",
                    "property_type": "apartamento"
                },
                "next_action": "ask_budget"
            }
            """,

            # ------------------------------------------------
            # TESTE 3.2
            # ------------------------------------------------
            """
            {
                "message": "Perfeito! Quantos quartos você procura?",
                "intent": "purchase",
                "transaction_type": "buy",
                "qualification": {
                    "budget": 800000
                },
                "next_action": "ask_bedrooms"
            }
            """,
        ]
    )

    service = create_service(
        repository=repository,
        llm_provider=llm_provider,
    )

    await test_first_message(
        service,
        repository,
    )

    await test_budget_message(
        service,
        repository,
    )

    await test_follow_up_cycle(
        repository
    )

    await test_lead_returns_after_follow_up(
        service,
        repository,
        llm_provider,
    )

    print("\n" + "=" * 60)
    print("🎉 TESTES DE INTEGRAÇÃO 3.1, 3.2, 3.3 E 3.4 PASSARAM")
    print("=" * 60)


if __name__ == "__main__":

    asyncio.run(main())