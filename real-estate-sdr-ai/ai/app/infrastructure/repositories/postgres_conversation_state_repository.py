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


class PostgresConversationStateRepository(
    ConversationStateRepository
):

    def __init__(self, session: AsyncSession):
        self.session = session

    async def get(
        self,
        conversation_id: str,
    ) -> ConversationState | None:

        conversation_uuid = uuid.UUID(conversation_id)

        result = await self.session.execute(
            select(ConversationStateModel).where(
                ConversationStateModel.conversation_id
                == conversation_uuid
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

        conversation_uuid = uuid.UUID(
            state.conversation_id
        )

        result = await self.session.execute(
            select(ConversationStateModel).where(
                ConversationStateModel.conversation_id
                == conversation_uuid
            )
        )

        model = result.scalar_one_or_none()

        if model is None:
            model = ConversationStateModel(
                conversation_id=conversation_uuid
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

        model.previous_state = (
            state.previous_state.value
            if state.previous_state
            else None
        )

        model.intent = state.intent
        model.transaction_type = state.transaction_type
        model.property_type = state.property_type
        model.region = state.region
        model.budget = state.budget
        model.bedrooms = state.bedrooms
        model.urgency = state.urgency
        model.investor_profile = state.investor_profile
        model.investment_ticket = state.investment_ticket
        model.expected_return = state.expected_return
        model.qualification_score = (
            state.qualification_score
        )
        model.temperature = state.temperature

        model.selected_property_id = (
            uuid.UUID(state.selected_property_id)
            if state.selected_property_id
            else None
        )

        model.follow_up_count = state.follow_up_count
        model.waiting_for_customer = (
            state.waiting_for_customer
        )

        await self.session.commit()
        await self.session.refresh(model)

        return self._to_domain(model)

    async def delete(
        self,
        conversation_id: str,
    ) -> None:

        conversation_uuid = uuid.UUID(
            conversation_id
        )

        await self.session.execute(
            delete(ConversationStateModel).where(
                ConversationStateModel.conversation_id
                == conversation_uuid
            )
        )

        await self.session.commit()

    @staticmethod
    def _to_domain(
        model: ConversationStateModel,
    ) -> ConversationState:

        return ConversationState(
            conversation_id=str(
                model.conversation_id
            ),

            flow=(
                ConversationFlow(model.flow)
                if model.flow
                else None
            ),

            state=ConversationStateType(
                model.state
            ),

            next_action=(
                NextAction(model.next_action)
                if model.next_action
                else None
            ),

            previous_state=(
                ConversationStateType(
                    model.previous_state
                )
                if model.previous_state
                else None
            ),

            intent=model.intent,
            transaction_type=model.transaction_type,
            property_type=model.property_type,
            region=model.region,

            budget=(
                float(model.budget)
                if model.budget is not None
                else None
            ),

            bedrooms=model.bedrooms,
            urgency=model.urgency,
            investor_profile=model.investor_profile,

            investment_ticket=(
                float(model.investment_ticket)
                if model.investment_ticket is not None
                else None
            ),

            expected_return=(
                float(model.expected_return)
                if model.expected_return is not None
                else None
            ),

            qualification_score=(
                float(model.qualification_score)
                if model.qualification_score is not None
                else None
            ),

            temperature=model.temperature,

            selected_property_id=(
                str(model.selected_property_id)
                if model.selected_property_id
                else None
            ),

            follow_up_count=model.follow_up_count,
            waiting_for_customer=(
                model.waiting_for_customer
            ),
        )