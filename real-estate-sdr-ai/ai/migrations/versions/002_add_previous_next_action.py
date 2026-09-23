"""add previous next action

Revision ID: 002_add_previous_next_action
Revises: 001_create_conversation_states
Create Date: 2026-09-16
"""

from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


revision: str = "002_add_previous_next_action"
down_revision: Union[str, None] = "001_create_conversation_states"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column(
        "conversation_states",
        sa.Column(
            "previous_next_action",
            sa.String(50),
            nullable=True,
        ),
    )


def downgrade() -> None:
    op.drop_column(
        "conversation_states",
        "previous_next_action",
    )