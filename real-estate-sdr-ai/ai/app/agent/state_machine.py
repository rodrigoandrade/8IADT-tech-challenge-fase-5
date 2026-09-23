from dataclasses import replace

from .state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)


class ConversationStateMachine:
    """
    Máquina de estados do agente SDR.

    Responsabilidades:

    - determinar o fluxo comercial;
    - controlar a etapa de qualificação;
    - determinar a próxima ação;
    - controlar o estado de follow-up.

    O LLM interpreta a linguagem natural e extrai informações.

    A máquina de estados determina a próxima etapa comercial.
    """

    def transition(
        self,
        state: ConversationState,
    ) -> ConversationState:

        # --------------------------------------------------------
        # ESTADOS FINAIS
        # --------------------------------------------------------

        if self._is_final_state(state):
            return state

        # --------------------------------------------------------
        # FOLLOW-UP
        # --------------------------------------------------------

        if state.state in (
            ConversationStateType.WAITING_LEAD,
            ConversationStateType.FOLLOW_UP_PENDING,
            ConversationStateType.SEND_FOLLOW_UP,
        ):
            return self._transition_follow_up(state)

        # --------------------------------------------------------
        # FLUXO AINDA NÃO IDENTIFICADO
        # --------------------------------------------------------

        if state.flow is None:
            return self._identify_intent(state)

        # --------------------------------------------------------
        # BUY
        # --------------------------------------------------------

        if state.flow == ConversationFlow.BUY:
            return self._transition_buy(state)

        # --------------------------------------------------------
        # INVESTMENT
        # --------------------------------------------------------

        if state.flow == ConversationFlow.INVESTMENT:
            return self._transition_investment(state)

        return self._identify_intent(state)

    # ============================================================
    # COMMON
    # ============================================================

    def _identify_intent(
        self,
        state: ConversationState,
    ) -> ConversationState:

        return replace(
            state,
            state=ConversationStateType.IDENTIFY_INTENT,
            next_action=NextAction.IDENTIFY_INTENT,
        )

    # ============================================================
    # BUY FLOW
    # ============================================================

    def _transition_buy(
        self,
        state: ConversationState,
    ) -> ConversationState:
        """
        Fluxo de compra:

        1. orçamento
        2. quartos
        3. região
        4. urgência
        5. matching
        6. recomendação
        7. agendamento
        """

        # --------------------------------------------------------
        # ORÇAMENTO
        # --------------------------------------------------------

        if state.budget is None:
            return replace(
                state,
                state=ConversationStateType.QUALIFY_BUDGET,
                next_action=NextAction.ASK_BUDGET,
            )

        # --------------------------------------------------------
        # QUARTOS
        # --------------------------------------------------------

        if state.bedrooms is None:
            return replace(
                state,
                state=ConversationStateType.QUALIFY_BEDROOMS,
                next_action=NextAction.ASK_BEDROOMS,
            )

        # --------------------------------------------------------
        # REGIÃO
        # --------------------------------------------------------

        if not self._has_value(state.region):
            return replace(
                state,
                state=ConversationStateType.QUALIFY_REGION,
                next_action=NextAction.ASK_LOCATION,
            )

        # --------------------------------------------------------
        # URGÊNCIA
        # --------------------------------------------------------

        if not self._has_value(state.urgency):
            return replace(
                state,
                state=ConversationStateType.QUALIFY_URGENCY,
                next_action=NextAction.ASK_URGENCY,
            )

        # --------------------------------------------------------
        # MATCHING
        #
        # Se ainda estamos na etapa de qualificação, inicia
        # o matching.
        # --------------------------------------------------------

        if state.state != ConversationStateType.RECOMMEND_PROPERTIES:
            return replace(
                state,
                state=ConversationStateType.RECOMMEND_PROPERTIES,
                next_action=NextAction.MATCH_PROPERTY,
            )

        # --------------------------------------------------------
        # RECOMENDAÇÃO
        #
        # O AIService executou o matching e agora podemos
        # avançar para a recomendação.
        # --------------------------------------------------------

        if (
            state.state
            == ConversationStateType.RECOMMEND_PROPERTIES
            and state.selected_property_id is not None
        ):
            return replace(
                state,
                state=ConversationStateType.RECOMMEND_PROPERTIES,
                next_action=NextAction.RECOMMEND_PROPERTY,
            )

        # --------------------------------------------------------
        # RECOMENDAÇÃO SEM IMÓVEL SELECIONADO
        #
        # Mantém a conversa na etapa de matching.
        # --------------------------------------------------------

        return replace(
            state,
            state=ConversationStateType.RECOMMEND_PROPERTIES,
            next_action=NextAction.MATCH_PROPERTY,
        )

    # ============================================================
    # INVESTMENT FLOW
    # ============================================================

    def _transition_investment(
        self,
        state: ConversationState,
    ) -> ConversationState:
        """
        Fluxo de investimento:

        1. perfil
        2. ticket
        3. retorno esperado
        4. especialista
        """

        # --------------------------------------------------------
        # PERFIL
        # --------------------------------------------------------

        if not self._has_value(
            state.investor_profile
        ):
            return replace(
                state,
                state=ConversationStateType.INVESTOR_PROFILE,
                next_action=NextAction.ASK_INVESTOR_PROFILE,
            )

        # --------------------------------------------------------
        # TICKET
        # --------------------------------------------------------

        if state.investment_ticket is None:
            return replace(
                state,
                state=ConversationStateType.INVESTMENT_TICKET,
                next_action=NextAction.ASK_INVESTMENT_TICKET,
            )

        # --------------------------------------------------------
        # RETORNO
        # --------------------------------------------------------

        if state.expected_return is None:
            return replace(
                state,
                state=ConversationStateType.EXPECTED_RETURN,
                next_action=NextAction.ASK_EXPECTED_RETURN,
            )

        # --------------------------------------------------------
        # ESPECIALISTA
        # --------------------------------------------------------

        return replace(
            state,
            state=ConversationStateType.SPECIALIST_HANDOFF,
            next_action=NextAction.HANDOFF_SPECIALIST,
        )

    # ============================================================
    # FOLLOW-UP
    # ============================================================

    def _transition_follow_up(
        self,
        state: ConversationState,
    ) -> ConversationState:

        # --------------------------------------------------------
        # AGUARDANDO CLIENTE
        # --------------------------------------------------------

        if state.state == ConversationStateType.WAITING_LEAD:

            return replace(
                state,
                state=ConversationStateType.FOLLOW_UP_PENDING,
                next_action=None,
            )

        # --------------------------------------------------------
        # FOLLOW-UP PENDENTE
        # --------------------------------------------------------

        if state.state == ConversationStateType.FOLLOW_UP_PENDING:

            return replace(
                state,
                state=ConversationStateType.SEND_FOLLOW_UP,
                next_action=NextAction.FOLLOW_UP,
                follow_up_count=state.follow_up_count + 1,
            )

        # --------------------------------------------------------
        # FOLLOW-UP ENVIADO
        # --------------------------------------------------------

        if state.state == ConversationStateType.SEND_FOLLOW_UP:

            return replace(
                state,
                state=ConversationStateType.WAITING_LEAD,
                next_action=None,
                waiting_for_customer=True,
            )

        return state

    # ============================================================
    # FOLLOW-UP HELPERS
    # ============================================================

    def enter_follow_up(
            self,
            state: ConversationState,
    ) -> ConversationState:

        return replace(
            state,
            previous_state=state.state,
            previous_next_action=state.next_action,
            state=ConversationStateType.WAITING_LEAD,
            next_action=None,
            waiting_for_customer=True,
        )

    def resume_after_follow_up(
        self,
        state: ConversationState,
    ) -> ConversationState:

        previous_state = state.previous_state

        if previous_state is None:
            return self.transition(state)

        return replace(
            state,
            state=previous_state,
            previous_state=None,
            next_action=state.previous_next_action,
            previous_next_action=None,
            waiting_for_customer=False,
        )

    # ============================================================
    # HELPERS
    # ============================================================

    @staticmethod
    def _has_value(
        value: object,
    ) -> bool:

        if value is None:
            return False

        if isinstance(value, str):
            return bool(value.strip())

        return True

    @staticmethod
    def _is_final_state(
        state: ConversationState,
    ) -> bool:

        return state.state in (
            ConversationStateType.SCHEDULED,
            ConversationStateType.COMPLETED,
            ConversationStateType.LOST,
        )
