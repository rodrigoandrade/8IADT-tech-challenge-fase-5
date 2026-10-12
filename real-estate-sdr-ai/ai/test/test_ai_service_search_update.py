import pytest

from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.models.ai_response import AIResponse
from ai.app.services.ai_service import AIService


@pytest.fixture
def service():
    # O método testado não depende das dependências do __init__.
    return AIService.__new__(AIService)


@pytest.fixture
def existing_state():
    return ConversationState(
        conversation_id="test-search-update",
        flow=ConversationFlow.BUY,
        state=ConversationStateType.RECOMMEND_PROPERTIES,
        next_action=NextAction.RECOMMEND_PROPERTY,
        transaction_type="buy",
        property_type="apartamento",
        region="Moema",
        budget=400000,
        bedrooms=2,
        urgency="3 meses",
        selected_property_id="old-property-id",
    )


def make_response(**qualification):
    return AIResponse(
        message="Entendi os critérios da sua busca.",
        qualification=qualification,
        transaction_type="buy",
    )


def test_region_change_preserves_other_criteria(
    service,
    existing_state,
):
    response = make_response(region="Pinheiros")

    updated = service._update_state_from_llm(
        state=existing_state,
        response=response,
        lead_message="E na região de Pinheiros?",
    )

    assert updated.region == "Pinheiros"
    assert updated.budget == 400000
    assert updated.bedrooms == 2
    assert updated.property_type == "apartamento"
    assert updated.selected_property_id is None


def test_region_change_to_morumbi(
    service,
    existing_state,
):
    response = make_response(region="Morumbi")

    updated = service._update_state_from_llm(
        state=existing_state,
        response=response,
        lead_message="E no Morumbi?",
    )

    assert updated.region == "Morumbi"
    assert updated.budget == 400000
    assert updated.bedrooms == 2


def test_new_apartment_search_in_sao_paulo(
    service,
    existing_state,
):
    response = make_response(
        property_type="apartamento",
        region="São Paulo",
    )

    updated = service._update_state_from_llm(
        state=existing_state,
        response=response,
        lead_message="Quero apartamentos em São Paulo",
    )

    assert updated.property_type == "apartamento"
    assert updated.region == "São Paulo"
    assert updated.selected_property_id is None


def test_budget_change_preserves_other_criteria(
    service,
    existing_state,
):
    response = make_response(budget=500000)

    updated = service._update_state_from_llm(
        state=existing_state,
        response=response,
        lead_message="Meu orçamento agora é de R$ 500 mil",
    )

    assert updated.budget == 500000
    assert updated.region == "Moema"
    assert updated.bedrooms == 2
    assert updated.property_type == "apartamento"


def test_explicit_bedroom_count_overrides_previous_value(
    service,
    existing_state,
):
    response = make_response(bedrooms=3)

    updated = service._update_state_from_llm(
        state=existing_state,
        response=response,
        lead_message="Agora quero pelo menos 3 quartos",
    )

    assert updated.bedrooms == 3