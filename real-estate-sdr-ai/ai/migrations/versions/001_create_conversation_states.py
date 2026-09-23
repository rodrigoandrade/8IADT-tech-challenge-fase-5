"""create conversation states

Revision ID: 001_create_conversation_states
Revises:
Create Date: 2026-09-05
"""

from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


revision: str = "001_create_conversation_states"

down_revision: Union[str, None] = None

branch_labels: Union[
    str, Sequence[str], None
] = None

depends_on: Union[
    str, Sequence[str], None
] = None


def upgrade() -> None:

    op.create_table(
        "conversation_states",

        sa.Column(
            "conversation_id",
            sa.UUID(),
            nullable=False,
        ),

        sa.Column(
            "flow",
            sa.String(30),
            nullable=True,
        ),

        sa.Column(
            "state",
            sa.String(50),
            nullable=False,
            server_default="new",
        ),

        sa.Column(
            "next_action",
            sa.String(50),
            nullable=True,
        ),

        sa.Column(
            "previous_state",
            sa.String(50),
            nullable=True,
        ),

        sa.Column(
            "intent",
            sa.String(100),
            nullable=True,
        ),

        sa.Column(
            "transaction_type",
            sa.String(50),
            nullable=True,
        ),

        sa.Column(
            "property_type",
            sa.String(50),
            nullable=True,
        ),

        sa.Column(
            "region",
            sa.String(150),
            nullable=True,
        ),

        sa.Column(
            "budget",
            sa.Numeric(14, 2),
            nullable=True,
        ),

        sa.Column(
            "bedrooms",
            sa.Integer(),
            nullable=True,
        ),

        sa.Column(
            "urgency",
            sa.String(50),
            nullable=True,
        ),

        sa.Column(
            "investor_profile",
            sa.String(100),
            nullable=True,
        ),

        sa.Column(
            "investment_ticket",
            sa.Numeric(14, 2),
            nullable=True,
        ),

        sa.Column(
            "expected_return",
            sa.Numeric(10, 2),
            nullable=True,
        ),

        sa.Column(
            "qualification_score",
            sa.Numeric(5, 4),
            nullable=True,
        ),

        sa.Column(
            "temperature",
            sa.String(30),
            nullable=True,
        ),

        sa.Column(
            "selected_property_id",
            sa.UUID(),
            nullable=True,
        ),

        sa.Column(
            "follow_up_count",
            sa.Integer(),
            nullable=False,
            server_default="0",
        ),

        sa.Column(
            "waiting_for_customer",
            sa.Boolean(),
            nullable=False,
            server_default=sa.false(),
        ),

        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),

        sa.PrimaryKeyConstraint(
            "conversation_id"
        ),

        sa.ForeignKeyConstraint(
            ["conversation_id"],
            ["conversations.id"],
            ondelete="CASCADE",
        ),

        sa.CheckConstraint(
            "budget IS NULL OR budget >= 0",
            name="chk_conversation_states_budget",
        ),

        sa.CheckConstraint(
            "bedrooms IS NULL OR bedrooms >= 0",
            name="chk_conversation_states_bedrooms",
        ),

        sa.CheckConstraint(
            """
            investment_ticket IS NULL
            OR investment_ticket >= 0
            """,
            name="chk_conversation_states_investment_ticket",
        ),

        sa.CheckConstraint(
            """
            expected_return IS NULL
            OR expected_return >= 0
            """,
            name="chk_conversation_states_expected_return",
        ),

        sa.CheckConstraint(
            """
            qualification_score IS NULL
            OR (
                qualification_score >= 0
                AND qualification_score <= 1
            )
            """,
            name="chk_conversation_states_qualification_score",
        ),

        sa.CheckConstraint(
            "follow_up_count >= 0",
            name="chk_conversation_states_follow_up_count",
        ),
    )

    op.create_index(
        "idx_conversation_states_state",
        "conversation_states",
        ["state"],
    )

    op.create_index(
        "idx_conversation_states_waiting",
        "conversation_states",
        ["waiting_for_customer"],
    )

    op.create_index(
        "idx_conversation_states_updated_at",
        "conversation_states",
        ["updated_at"],
    )


def downgrade() -> None:

    op.drop_index(
        "idx_conversation_states_updated_at",
        table_name="conversation_states",
    )

    op.drop_index(
        "idx_conversation_states_waiting",
        table_name="conversation_states",
    )

    op.drop_index(
        "idx_conversation_states_state",
        table_name="conversation_states",
    )

    op.drop_table(
        "conversation_states"
    )