from ai.app.agent.agent_router import AgentRouter
from ai.app.agent.agent_type import AgentType


def test_financing_question_goes_to_knowledge():
    router = AgentRouter()

    result = router.route(
        "Como funciona o financiamento dos imóveis?"
    )

    assert result == AgentType.KNOWLEDGE


def test_visit_question_goes_to_knowledge():
    router = AgentRouter()

    result = router.route(
        "Como faço para visitar um imóvel?"
    )

    assert result == AgentType.KNOWLEDGE


def test_buy_apartment_goes_to_qualification():
    router = AgentRouter()

    result = router.route(
        "Quero comprar um apartamento"
    )

    assert result == AgentType.QUALIFICATION


def test_property_search_goes_to_property():
    router = AgentRouter()

    result = router.route(
        "Me mostre apartamentos em Moema"
    )

    assert result == AgentType.PROPERTY