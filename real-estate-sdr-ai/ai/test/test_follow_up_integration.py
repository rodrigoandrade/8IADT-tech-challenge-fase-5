from datetime import datetime, timedelta, timezone

import pytest

from ai.app.agent.state import (
    ConversationFlow,
    ConversationStateType,
    NextAction,
)
from ai.app.agent.state import ConversationState
from ai.app.services.follow_up_service import FollowUpService


# ============================================================
# FAKE REAL ESTATE API
# ============================================================


class FakeRealEstateAPIClient:

    def __init__(self):
        self.messages = {
            "conversation-follow-up-001": [
                {
                    "id": "message-001",
                    "role": "USER",
                    "content": "Quero um apartamento em Moema.",
                    "created_at": (
                        datetime.now(timezone.utc)
                        - timedelta(minutes=2)
                    ).isoformat(),
                }
            ]
        }

        self.created_messages = []

    async def get_conversations(
        self,
        limit=100,
        offset=0,
    ):
        return [
            {
                "id": "conversation-follow-up-001",
            }
        ]

    async def get_messages(
        self,
        conversation_id,
    ):
        return self.messages.get(
            conversation_id,
            [],
        )

    async def create_message(
        self,
        conversation_id,
        role,
        content,
    ):
        message = {
            "id": f"message-{len(self.created_messages) + 2}",
            "role": role,
            "content": content,
            "created_at": datetime.now(
                timezone.utc
            ).isoformat(),
        }

        self.created_messages.append(message)

        self.messages.setdefault(
            conversation_id,
            [],
        ).append(message)

        return message


# ============================================================
# FAKE STATE REPOSITORY
# ============================================================


class FakeStateRepository:

    def __init__(self):
        self.states = {}

    async def get(
        self,
        conversation_id,
    ):
        return self.states.get(
            conversation_id
        )

    async def save(
        self,
        state,
    ):
        self.states[
            state.conversation_id
        ] = state

        return state

    async def delete(
        self,
        conversation_id,
    ):
        self.states.pop(
            conversation_id,
            None,
        )


# ============================================================
# HELPERS
# ============================================================


def build_state():
    return ConversationState(
        conversation_id="conversation-follow-up-001",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.QUALIFY_BEDROOMS,
        next_action=NextAction.ASK_BEDROOMS,
        previous_state=None,
        previous_next_action=None,
        intent="purchase",
        transaction_type="buy",
        property_type="apartamento",
        region="Moema",
        budget=800000,
        bedrooms=None,
        urgency=None,
        investor_profile=None,
        investment_ticket=None,
        expected_return=None,
        qualification_score=0.6,
        temperature="WARM",
        selected_property_id=None,
        follow_up_count=0,
        waiting_for_customer=False,
    )


# ============================================================
# TESTE 2.1
# FOLLOW-UP AUTOMÁTICO
# ============================================================


@pytest.mark.asyncio
async def test_follow_up_integration_sends_message_and_updates_state():

    api_client = FakeRealEstateAPIClient()

    repository = FakeStateRepository()

    state = build_state()

    await repository.save(state)

    service = FollowUpService(
        api_client=api_client,
        state_repository=repository,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    # --------------------------------------------------------
    # EXECUTA FOLLOW-UP
    # --------------------------------------------------------

    result = await service.process()

    # --------------------------------------------------------
    # RESULTADO
    # --------------------------------------------------------

    assert result == 1

    # --------------------------------------------------------
    # MENSAGEM
    # --------------------------------------------------------

    assert len(
        api_client.created_messages
    ) == 1

    follow_up_message = (
        api_client.created_messages[0]
    )

    assert (
        follow_up_message["role"]
        == "ASSISTANT"
    )

    assert (
        "conversa ficou em aberto"
        in follow_up_message["content"]
    )

    # --------------------------------------------------------
    # ESTADO PERSISTIDO
    # --------------------------------------------------------

    persisted_state = await repository.get(
        "conversation-follow-up-001"
    )

    assert persisted_state is not None

    assert (
        persisted_state.state
        == ConversationStateType.WAITING_LEAD
    )

    assert (
        persisted_state.next_action
        is None
    )

    assert (
        persisted_state.follow_up_count
        == 1
    )

    assert (
        persisted_state.waiting_for_customer
        is True
    )

    # --------------------------------------------------------
    # CONTEXTO COMERCIAL PRESERVADO
    # --------------------------------------------------------

    assert (
        persisted_state.previous_state
        == ConversationStateType.QUALIFY_BEDROOMS
    )

    assert (
        persisted_state.previous_next_action
        == NextAction.ASK_BEDROOMS
    )

    assert (
        persisted_state.flow
        == ConversationFlow.BUY
    )

    assert (
        persisted_state.region
        == "Moema"
    )

    assert (
        persisted_state.budget
        == 800000
    )

    assert (
        persisted_state.property_type
        == "apartamento"
    )


# ============================================================
# TESTE 2.2
# NÃO ENVIA SEGUNDO FOLLOW-UP
# ============================================================


@pytest.mark.asyncio
async def test_follow_up_integration_respects_max_follow_ups():

    api_client = FakeRealEstateAPIClient()

    repository = FakeStateRepository()

    state = build_state()

    await repository.save(state)

    service = FollowUpService(
        api_client=api_client,
        state_repository=repository,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    # Primeiro follow-up
    first_result = await service.process()

    assert first_result == 1

    assert len(
        api_client.created_messages
    ) == 1

    # --------------------------------------------------------
    # SEGUNDA TENTATIVA
    # --------------------------------------------------------

    second_result = await service.process()

    assert second_result == 0

    assert len(
        api_client.created_messages
    ) == 1

    persisted_state = await repository.get(
        "conversation-follow-up-001"
    )

    assert (
        persisted_state.follow_up_count
        == 1
    )


# ============================================================
# TESTE 2.3
# ENVIA FOLLOW-UP QUANDO ASSISTANT FOI A ÚLTIMA MENSAGEM,
# DESDE QUE O ÚLTIMO USER ESTEJA INATIVO
# ============================================================


@pytest.mark.asyncio
async def test_follow_up_integration_sends_after_assistant_message_when_user_is_inactive():

    api_client = FakeRealEstateAPIClient()

    repository = FakeStateRepository()

    state = build_state()

    await repository.save(state)

    # --------------------------------------------------------
    # O Fake já começa com um USER antigo:
    #
    # USER
    #  ↓
    # 2 minutos de inatividade
    # --------------------------------------------------------

    # --------------------------------------------------------
    # Simula uma resposta do ASSISTANT depois do USER.
    #
    # Agora a última mensagem é ASSISTANT, mas o último USER
    # continua sendo a referência para a inatividade.
    # --------------------------------------------------------

    await api_client.create_message(
        conversation_id="conversation-follow-up-001",
        role="ASSISTANT",
        content="Posso te ajudar com a busca.",
    )

    service = FollowUpService(
        api_client=api_client,
        state_repository=repository,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    result = await service.process()

    assert result == 1

    # --------------------------------------------------------
    # Apenas o FOLLOW-UP foi criado durante o processamento.
    # --------------------------------------------------------

    assert len(
        api_client.created_messages
    ) == 2

    follow_up = (
        api_client.created_messages[-1]
    )

    assert follow_up["role"] == "ASSISTANT"

    assert (
        "conversa ficou em aberto"
        in follow_up["content"]
    )

    # --------------------------------------------------------
    # Estado da conversa após o FOLLOW-UP
    # --------------------------------------------------------

    persisted_state = await repository.get(
        "conversation-follow-up-001"
    )

    assert (
        persisted_state.follow_up_count
        == 1
    )

    assert (
        persisted_state.state
        == ConversationStateType.WAITING_LEAD
    )

    assert (
        persisted_state.next_action
        is None
    )

    assert (
        persisted_state.waiting_for_customer
        is True
    )


# ============================================================
# TESTE 2.4
# CONVERSA FINAL NÃO RECEBE FOLLOW-UP
# ============================================================


@pytest.mark.parametrize(
    "final_state",
    [
        ConversationStateType.SCHEDULED,
        ConversationStateType.COMPLETED,
        ConversationStateType.LOST,
    ],
)
@pytest.mark.asyncio
async def test_follow_up_integration_does_not_send_for_final_state(
    final_state,
):

    api_client = FakeRealEstateAPIClient()

    repository = FakeStateRepository()

    state = build_state()

    state.state = final_state

    await repository.save(state)

    service = FollowUpService(
        api_client=api_client,
        state_repository=repository,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    result = await service.process()

    assert result == 0

    assert len(
        api_client.created_messages
    ) == 0

    persisted_state = await repository.get(
        "conversation-follow-up-001"
    )

    assert (
        persisted_state.state
        == final_state
    )

    assert (
        persisted_state.follow_up_count
        == 0
    )