# ai/app/infrastructure/repositories/postgres_conversation_state_repository.py

import uuid

from sqlalchemy import delete, select
from sqlalchemy.ext.asyncio import AsyncSession

from ai.app.agent.state import (
    ConversationFlow,
    ConversationState,
    ConversationStateType,
    NextAction,
)
from ai.app.models.conversation_state import ConversationStateModel
from ai.app.repositories.conversation_state_repository import (
    ConversationStateRepository,
)


class PostgresConversationStateRepository(ConversationStateRepository):

    def __init__(self, session: AsyncSession):
        self.session = session

    async def get(
        self,
        conversation_id: uuid.UUID,
    ) -> ConversationState | None:

        result = await self.session.execute(
            select(ConversationStateModel).where(
                ConversationStateModel.conversation_id == conversation_id
            )
        )

        model = result.scalar_one_or_none()

        if model is None:
            return None

        return self._to_domain(model)

    async def save(
        self,
        state: ConversationState,
    ) -> ConversationState:

        model = await self._get_model(state.conversation_id)

        if model is None:
            model = ConversationStateModel(
                conversation_id=state.conversation_id,
            )
            self.session.add(model)

        model.flow = (
            state.flow.value
            if state.flow
            else None
        )

        model.state = state.state.value

        model.next_action = (
            state.next_action.value
            if state.next_action
            else None
        )

        # Estado anterior ao follow-up.
        model.previous_state = (
            state.previous_state.value
            if state.previous_state
            else None
        )

        # Próxima ação anterior ao follow-up.
        #
        # Esse campo é importante porque durante o follow-up
        # o next_action atual fica como None.
        #
        # Exemplo:
        #
        # antes:
        #   state = QUALIFY_BEDROOMS
        #   next_action = ASK_BEDROOMS
        #
        # durante follow-up:
        #   state = WAITING_LEAD
        #   next_action = None
        #   previous_state = QUALIFY_BEDROOMS
        #   previous_next_action = ASK_BEDROOMS
        #
        # quando o lead responder, o state machine consegue
        # restaurar também a próxima ação.
        model.previous_next_action = (
            state.previous_next_action.value
            if state.previous_next_action
            else None
        )

        model.intent = (
            state.intent
            if state.intent
            else None
        )

        model.transaction_type = (
            state.transaction_type
            if state.transaction_type
            else None
        )

        model.transaction_type = state.transaction_type

        model.property_type = state.property_type

        model.region = state.region

        model.budget = state.budget

        model.bedrooms = state.bedrooms

        model.urgency = state.urgency

        model.investor_profile = state.investor_profile

        model.investment_ticket = state.investment_ticket

        model.expected_return = state.expected_return

        model.qualification_score = state.qualification_score

        model.temperature = state.temperature

        model.selected_property_id = state.selected_property_id

        model.follow_up_count = state.follow_up_count

        model.waiting_for_customer = state.waiting_for_customer

        await self.session.commit()

        await self.session.refresh(model)

        return self._to_domain(model)

    async def delete(
        self,
        conversation_id: uuid.UUID,
    ) -> None:

        await self.session.execute(
            delete(ConversationStateModel).where(
                ConversationStateModel.conversation_id == conversation_id
            )
        )

        await self.session.commit()

    async def _get_model(
        self,
        conversation_id: uuid.UUID,
    ) -> ConversationStateModel | None:

        result = await self.session.execute(
            select(ConversationStateModel).where(
                ConversationStateModel.conversation_id == conversation_id
            )
        )

        return result.scalar_one_or_none()

    @staticmethod
    def _to_domain(
        model: ConversationStateModel,
    ) -> ConversationState:

        return ConversationState(
            conversation_id=model.conversation_id,

            flow=(
                ConversationFlow(model.flow)
                if model.flow
                else None
            ),

            state=ConversationStateType(model.state),

            next_action=(
                NextAction(model.next_action)
                if model.next_action
                else None
            ),

            previous_state=(
                ConversationStateType(model.previous_state)
                if model.previous_state
                else None
            ),

            previous_next_action=(
                NextAction(model.previous_next_action)
                if model.previous_next_action
                else None
            ),

            intent=model.intent,

            transaction_type=model.transaction_type,

            property_type=model.property_type,

            region=model.region,

            budget=model.budget,

            bedrooms=model.bedrooms,

            urgency=model.urgency,

            investor_profile=model.investor_profile,

            investment_ticket=model.investment_ticket,

            expected_return=model.expected_return,

            qualification_score=model.qualification_score,

            temperature=model.temperature,

            selected_property_id=model.selected_property_id,

            follow_up_count=model.follow_up_count,

            waiting_for_customer=model.waiting_for_customer,
        )