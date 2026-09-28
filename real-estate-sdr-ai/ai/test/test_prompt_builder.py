from ai.app.prompt.prompt_builder import PromptBuilder


def test_build_knowledge_prompt():
    builder = PromptBuilder()

    prompt = builder.build_knowledge_prompt(
        question="Como funciona o financiamento?",
        knowledge=[
            {
                "title": "Financiamento",
                "content": (
                    "Os imóveis podem ser adquiridos através de "
                    "financiamento bancário, sujeito à análise de "
                    "crédito e às condições estabelecidas pela "
                    "instituição financeira."
                ),
                "similarity": 0.69,
            }
        ],
    )

    assert "Como funciona o financiamento?" in prompt
    assert "financiamento bancário" in prompt
    assert "sujeito à análise de crédito" in prompt

    # similarity não precisa ser enviada para o LLM
    assert "0.69" not in prompt

    # O prompt de conhecimento não deve carregar
    # regras do fluxo comercial.
    assert "next_action" not in prompt
    assert "qualification" not in prompt