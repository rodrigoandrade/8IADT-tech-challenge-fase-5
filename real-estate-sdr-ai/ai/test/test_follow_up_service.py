from datetime import datetime, timedelta, timezone
from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

from ai.app.agent.state import ConversationStateType, NextAction
from ai.app.services.follow_up_service import FollowUpService


CONVERSATION_ID = "conversation-1"


def build_state(
    *,
    state=ConversationStateType.QUALIFY_BUDGET,
    next_action=NextAction.ASK_BUDGET,
    follow_up_count=0,
    waiting_for_customer=False,
):
    return SimpleNamespace(
        state=state,
        next_action=next_action,
        previous_state=None,
        previous_next_action=None,
        follow_up_count=follow_up_count,
        waiting_for_customer=waiting_for_customer,
    )


def build_message(
    *,
    role="USER",
    created_at=None,
    content="Quero continuar procurando um imóvel.",
):
    if created_at is None:
        created_at = (
            datetime.now(timezone.utc)
            - timedelta(minutes=2)
        )

    return {
        "role": role,
        "content": content,
        "created_at": created_at.isoformat(),
    }


def build_service(
    *,
    conversations=None,
    messages=None,
    state=None,
    inactivity_minutes=1,
    max_follow_ups=1,
):
    api_client = SimpleNamespace(
        get_conversations=AsyncMock(
            return_value=conversations or []
        ),
        get_messages=AsyncMock(
            return_value=messages or []
        ),
        create_message=AsyncMock(),
    )

    state_repository = SimpleNamespace(
        get=AsyncMock(return_value=state),
        save=AsyncMock(side_effect=lambda value: value),
    )

    service = FollowUpService(
        api_client=api_client,
        state_repository=state_repository,
        inactivity_minutes=inactivity_minutes,
        max_follow_ups=max_follow_ups,
    )

    return service, api_client, state_repository


@pytest.mark.asyncio
async def test_process_conversation_sends_follow_up_when_lead_is_inactive():
    state = build_state()

    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(minutes=2)
            ),
        )
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=state,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is True

    api_client.create_message.assert_awaited_once()

    call = api_client.create_message.await_args

    assert call.kwargs["conversation_id"] == CONVERSATION_ID
    assert call.kwargs["role"] == "ASSISTANT"
    assert "conversa ficou em aberto" in call.kwargs["content"]

    assert state.follow_up_count == 1
    assert state.waiting_for_customer is True
    assert state.state == ConversationStateType.WAITING_LEAD
    assert state.next_action is None

    state_repository.save.assert_awaited_once_with(state)


@pytest.mark.asyncio
async def test_process_conversation_preserves_previous_state_before_follow_up():
    state = build_state(
        state=ConversationStateType.QUALIFY_BEDROOMS,
        next_action=NextAction.ASK_BEDROOMS,
    )

    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(minutes=2)
            ),
        )
    ]

    service, _, _ = build_service(
        messages=messages,
        state=state,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is True

    assert (
        state.previous_state
        == ConversationStateType.QUALIFY_BEDROOMS
    )

    assert (
        state.previous_next_action
        == NextAction.ASK_BEDROOMS
    )

    assert state.state == ConversationStateType.WAITING_LEAD
    assert state.next_action is None


@pytest.mark.asyncio
async def test_process_conversation_sends_follow_up_when_assistant_replied_after_user():
    state = build_state()

    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(minutes=2)
            ),
            content="Quero continuar procurando um imóvel.",
        ),
        build_message(
            role="ASSISTANT",
            created_at=datetime.now(timezone.utc),
            content="Claro! Quantos quartos você procura?",
        ),
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=state,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is True

    api_client.create_message.assert_awaited_once()

    call = api_client.create_message.await_args

    assert call.kwargs["conversation_id"] == CONVERSATION_ID
    assert call.kwargs["role"] == "ASSISTANT"
    assert "conversa ficou em aberto" in call.kwargs["content"]

    assert state.follow_up_count == 1
    assert state.waiting_for_customer is True
    assert state.state == ConversationStateType.WAITING_LEAD
    assert state.next_action is None

    state_repository.save.assert_awaited_once_with(state)


@pytest.mark.asyncio
async def test_process_conversation_does_not_send_when_follow_up_limit_was_reached():
    state = build_state(
        follow_up_count=1,
    )

    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(minutes=2)
            ),
        )
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=state,
        max_follow_ups=1,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is False

    api_client.create_message.assert_not_awaited()
    state_repository.save.assert_not_awaited()


@pytest.mark.parametrize(
    "final_state",
    [
        ConversationStateType.SCHEDULED,
        ConversationStateType.COMPLETED,
        ConversationStateType.LOST,
    ],
)
@pytest.mark.asyncio
async def test_process_conversation_does_not_send_for_final_states(
    final_state,
):
    state = build_state(
        state=final_state,
    )

    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(minutes=2)
            ),
        )
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=state,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is False

    api_client.create_message.assert_not_awaited()
    state_repository.save.assert_not_awaited()


@pytest.mark.asyncio
async def test_process_conversation_does_not_send_when_conversation_is_not_inactive():
    state = build_state()

    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(seconds=10)
            ),
        )
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=state,
        inactivity_minutes=1,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is False

    api_client.create_message.assert_not_awaited()
    state_repository.save.assert_not_awaited()


@pytest.mark.asyncio
async def test_process_conversation_does_not_send_when_state_does_not_exist():
    messages = [
        build_message(
            role="USER",
            created_at=(
                datetime.now(timezone.utc)
                - timedelta(minutes=2)
            ),
        )
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=None,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is False

    state_repository.get.assert_awaited_once_with(
        CONVERSATION_ID
    )

    api_client.create_message.assert_not_awaited()
    state_repository.save.assert_not_awaited()


@pytest.mark.asyncio
async def test_process_conversation_does_not_send_when_there_are_no_messages():
    state = build_state()

    service, api_client, state_repository = build_service(
        messages=[],
        state=state,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is False

    api_client.create_message.assert_not_awaited()
    state_repository.save.assert_not_awaited()


@pytest.mark.asyncio
async def test_process_conversation_does_not_send_when_created_at_is_missing():
    state = build_state()

    messages = [
        {
            "role": "USER",
            "content": "Quero continuar.",
        }
    ]

    service, api_client, state_repository = build_service(
        messages=messages,
        state=state,
    )

    result = await service.process_conversation(
        CONVERSATION_ID
    )

    assert result is False

    api_client.create_message.assert_not_awaited()
    state_repository.save.assert_not_awaited()


@pytest.mark.asyncio
async def test_process_sends_follow_up_only_for_eligible_conversations():
    now = datetime.now(timezone.utc)

    conversations = [
        {"id": "conversation-eligible"},
        {"id": "conversation-recent"},
        {"id": "conversation-final"},
    ]

    eligible_messages = [
        build_message(
            role="USER",
            created_at=now - timedelta(minutes=2),
        )
    ]

    recent_messages = [
        build_message(
            role="USER",
            created_at=now - timedelta(seconds=10),
        )
    ]

    final_messages = [
        build_message(
            role="USER",
            created_at=now - timedelta(minutes=2),
        )
    ]

    eligible_state = build_state()

    recent_state = build_state()

    final_state = build_state(
        state=ConversationStateType.COMPLETED,
    )

    async def get_messages(conversation_id):
        messages_by_conversation = {
            "conversation-eligible": eligible_messages,
            "conversation-recent": recent_messages,
            "conversation-final": final_messages,
        }

        return messages_by_conversation[conversation_id]

    async def get_state(conversation_id):
        states_by_conversation = {
            "conversation-eligible": eligible_state,
            "conversation-recent": recent_state,
            "conversation-final": final_state,
        }

        return states_by_conversation[conversation_id]

    api_client = SimpleNamespace(
        get_conversations=AsyncMock(
            return_value=conversations
        ),
        get_messages=AsyncMock(
            side_effect=get_messages
        ),
        create_message=AsyncMock(),
    )

    state_repository = SimpleNamespace(
        get=AsyncMock(
            side_effect=get_state
        ),
        save=AsyncMock(),
    )

    service = FollowUpService(
        api_client=api_client,
        state_repository=state_repository,
        inactivity_minutes=1,
        max_follow_ups=1,
    )

    result = await service.process()

    assert result == 1

    assert api_client.create_message.await_count == 1

    call = api_client.create_message.await_args

    assert (
        call.kwargs["conversation_id"]
        == "conversation-eligible"
    )

    assert eligible_state.follow_up_count == 1
    assert (
        eligible_state.state
        == ConversationStateType.WAITING_LEAD
    )

    assert recent_state.follow_up_count == 0
    assert final_state.follow_up_count == 0