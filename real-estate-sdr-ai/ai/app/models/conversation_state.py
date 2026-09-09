import uuid
from datetime import datetime
from decimal import Decimal

from sqlalchemy import Boolean, DateTime, Integer, Numeric, String, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from .base import Base


class ConversationStateModel(Base):
    __tablename__ = "conversation_states"

    conversation_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        primary_key=True,
    )

    flow: Mapped[str | None] = mapped_column(
        String(30),
        nullable=True,
    )

    state: Mapped[str] = mapped_column(
        String(50),
        nullable=False,
        default="new",
    )

    next_action: Mapped[str | None] = mapped_column(
        String(50),
        nullable=True,
    )

    previous_state: Mapped[str | None] = mapped_column(
        String(50),
        nullable=True,
    )

    intent: Mapped[str | None] = mapped_column(
        String(100),
        nullable=True,
    )

    transaction_type: Mapped[str | None] = mapped_column(
        String(50),
        nullable=True,
    )

    property_type: Mapped[str | None] = mapped_column(
        String(50),
        nullable=True,
    )

    region: Mapped[str | None] = mapped_column(
        String(150),
        nullable=True,
    )

    budget: Mapped[Decimal | None] = mapped_column(
        Numeric(14, 2),
        nullable=True,
    )

    bedrooms: Mapped[int | None] = mapped_column(
        Integer,
        nullable=True,
    )

    urgency: Mapped[str | None] = mapped_column(
        String(50),
        nullable=True,
    )

    investor_profile: Mapped[str | None] = mapped_column(
        String(100),
        nullable=True,
    )

    investment_ticket: Mapped[Decimal | None] = mapped_column(
        Numeric(14, 2),
        nullable=True,
    )

    expected_return: Mapped[Decimal | None] = mapped_column(
        Numeric(10, 2),
        nullable=True,
    )

    qualification_score: Mapped[Decimal | None] = mapped_column(
        Numeric(5, 4),
        nullable=True,
    )

    temperature: Mapped[str | None] = mapped_column(
        String(30),
        nullable=True,
    )

    selected_property_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True),
        nullable=True,
    )

    follow_up_count: Mapped[int] = mapped_column(
        Integer,
        nullable=False,
        default=0,
    )

    waiting_for_customer: Mapped[bool] = mapped_column(
        Boolean,
        nullable=False,
        default=False,
    )

    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        nullable=False,
        server_default=func.now(),
        onupdate=func.now(),
    )