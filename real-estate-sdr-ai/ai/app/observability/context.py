from contextvars import ContextVar


correlation_id_context: ContextVar[str | None] = ContextVar(
    "correlation_id",
    default=None,
)


def set_correlation_id(
    correlation_id: str,
) -> None:
    correlation_id_context.set(
        correlation_id
    )


def get_correlation_id() -> str | None:
    return correlation_id_context.get()