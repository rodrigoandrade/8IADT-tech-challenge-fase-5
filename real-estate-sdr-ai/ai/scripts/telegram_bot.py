"""Telegram bridge for Real Estate SDR AI with persistent update deduplication.

Place at: real-estate-sdr-ai/ai/scripts/telegram_bot.py
Required root .env:
  TELEGRAM_BOT_TOKEN=...
  REAL_ESTATE_API_URL=http://localhost:8080
  API_KEY=...
  TELEGRAM_LEAD_MAPPING=TELEGRAM_USER_ID:LEAD_UUID[,TELEGRAM_USER_ID:LEAD_UUID]

The bot never creates leads or conversations. It picks the newest ACTIVE
conversation for the mapped existing lead.
"""
from __future__ import annotations

import logging
import os
import sqlite3
from contextlib import contextmanager
import time
from pathlib import Path
from typing import Any

import requests
from dotenv import load_dotenv

PROJECT_ROOT = Path(__file__).resolve().parents[2]
load_dotenv(PROJECT_ROOT / ".env")

logger = logging.getLogger("telegram_bot")


class IntegrationError(RuntimeError):
    """A request to Telegram or the Go API failed."""


def parse_mapping(raw: str) -> dict[int, str]:
    mapping: dict[int, str] = {}
    for entry in raw.split(","):
        entry = entry.strip()
        if not entry:
            continue
        try:
            user_text, lead_id = entry.split(":", 1)
            user_id = int(user_text.strip())
            lead_id = lead_id.strip()
        except ValueError as exc:
            raise ValueError(
                "TELEGRAM_LEAD_MAPPING inválido. Use TELEGRAM_USER_ID:LEAD_UUID."
            ) from exc
        if user_id <= 0 or not lead_id:
            raise ValueError("ID do Telegram e LEAD_UUID devem ser válidos.")
        if user_id in mapping and mapping[user_id] != lead_id:
            raise ValueError(f"Usuário Telegram duplicado no mapeamento: {user_id}.")
        mapping[user_id] = lead_id
    return mapping


def as_list(payload: Any) -> list[dict[str, Any]]:
    if isinstance(payload, list):
        return [item for item in payload if isinstance(item, dict)]
    if isinstance(payload, dict):
        for key in ("items", "value", "conversations", "data"):
            value = payload.get(key)
            if isinstance(value, list):
                return [item for item in value if isinstance(item, dict)]
    return []


def newest_active_conversation_from_payload(payload: Any) -> dict[str, Any] | None:
    active = [
        item for item in as_list(payload)
        if str(item.get("status", "")).upper() == "ACTIVE" and item.get("id")
    ]
    return max(active, key=lambda item: str(item.get("started_at") or ""), default=None)


def response_text(payload: Any) -> str:
    if isinstance(payload, str):
        return payload.strip()
    if not isinstance(payload, dict):
        return ""
    for key in ("message", "content", "response", "answer", "text"):
        value = payload.get(key)
        if isinstance(value, str) and value.strip():
            return value.strip()
        if isinstance(value, dict):
            for nested_key in ("content", "text", "message"):
                nested = value.get(nested_key)
                if isinstance(nested, str) and nested.strip():
                    return nested.strip()
    return ""


class UpdateStore:
    """SQLite-backed processed-update ledger.

    A successful update is recorded only after its handler completes. This
    survives restarts and avoids normal duplicate polling. A crash between an
    external POST /messages side effect and recording completion can still
    cause a retry; full end-to-end exactly-once needs idempotency in the Go API.
    """

    def __init__(self, db_path: str | Path):
        self.db_path = str(db_path)
        Path(self.db_path).parent.mkdir(parents=True, exist_ok=True)
        with self._connect() as connection:
            connection.execute("""
                CREATE TABLE IF NOT EXISTS telegram_updates (
                    update_id INTEGER PRIMARY KEY,
                    status TEXT NOT NULL CHECK(status IN ('processing', 'processed')),
                    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
                )
            """)

    @contextmanager
    def _connect(self):
        # sqlite3.Connection as a context manager commits/rolls back but does
        # not close the connection. Explicitly close it for Windows compatibility.
        connection = sqlite3.connect(self.db_path, timeout=10)
        try:
            connection.execute("PRAGMA busy_timeout=10000")
            yield connection
            connection.commit()
        except Exception:
            connection.rollback()
            raise
        finally:
            connection.close()

    def last_processed_update_id(self) -> int | None:
        with self._connect() as connection:
            row = connection.execute(
                "SELECT MAX(update_id) FROM telegram_updates WHERE status='processed'"
            ).fetchone()
        return row[0] if row and row[0] is not None else None

    def is_processed(self, update_id: int) -> bool:
        with self._connect() as connection:
            row = connection.execute(
                "SELECT status FROM telegram_updates WHERE update_id=?", (update_id,)
            ).fetchone()
        return bool(row and row[0] == "processed")

    def mark_processing(self, update_id: int) -> None:
        with self._connect() as connection:
            connection.execute(
                """INSERT INTO telegram_updates(update_id, status)
                   VALUES (?, 'processing')
                   ON CONFLICT(update_id) DO UPDATE SET
                     status='processing', updated_at=CURRENT_TIMESTAMP
                   WHERE telegram_updates.status != 'processed'""",
                (update_id,),
            )

    def mark_processed(self, update_id: int) -> None:
        with self._connect() as connection:
            connection.execute(
                """INSERT INTO telegram_updates(update_id, status)
                   VALUES (?, 'processed')
                   ON CONFLICT(update_id) DO UPDATE SET
                     status='processed', updated_at=CURRENT_TIMESTAMP""",
                (update_id,),
            )


class TelegramBot:
    def __init__(
        self,
        *,
        bot_token: str,
        api_base_url: str,
        api_key: str,
        lead_mapping: dict[int, str],
        store: UpdateStore,
        session: requests.Session | None = None,
        timeout: float = 35,
    ):
        if not bot_token.strip():
            raise ValueError("TELEGRAM_BOT_TOKEN não configurado.")
        if not api_base_url.strip():
            raise ValueError("REAL_ESTATE_API_URL não configurada.")
        if not api_key.strip():
            raise ValueError("API_KEY não configurada.")
        self.telegram_url = f"https://api.telegram.org/bot{bot_token.strip()}"
        self.api_base_url = api_base_url.rstrip("/")
        self.api_headers = {"X-API-Key": api_key.strip()}
        self.lead_mapping = lead_mapping
        self.store = store
        self.session = session or requests.Session()
        self.timeout = timeout

    def _telegram_call(
        self,
        method: str,
        *,
        params: dict[str, Any] | None = None,
        body: dict[str, Any] | None = None,
        timeout: float | None = None,
    ) -> Any:

        try:
            response = self.session.request(
                "POST" if body is not None else "GET",
                f"{self.telegram_url}/{method}",
                params=params,
                json=body,
                timeout=timeout or self.timeout,
            )
            response.raise_for_status()
            payload = response.json()

        except requests.HTTPError as exc:
            status = (
                exc.response.status_code
                if exc.response is not None
                else "?"
            )
            try:
                detail = exc.response.json().get(
                    "description", "sem descrição"
                ) if exc.response is not None else "sem resposta"
            except ValueError:
                detail = "resposta HTTP não JSON"

            logger.error(
                "Telegram %s retornou HTTP %s: %s",
                method,
                status,
                detail,
            )
            raise IntegrationError(
                f"Falha na chamada Telegram {method}: HTTP {status}."
            ) from exc

        except requests.RequestException as exc:
            logger.error(
                "Erro de rede ao chamar Telegram %s: %s",
                method,
                type(exc).__name__,
            )
            raise IntegrationError(
                f"Falha de rede na chamada Telegram {method}."
            ) from exc

        except ValueError as exc:
            logger.error(
                "Resposta JSON inválida do Telegram em %s.",
                method,
            )
            raise IntegrationError(
                f"Resposta inválida do Telegram em {method}."
            ) from exc

        if not isinstance(payload, dict) or not payload.get("ok"):
            description = payload.get("description", "resposta inválida") if isinstance(payload, dict) else "resposta inválida"
            raise IntegrationError(f"Telegram {method}: {description}")
        return payload.get("result")

    def _api_call(
        self, method: str, path: str, *, body: dict[str, Any] | None = None,
        idempotency_key: str | None = None,
    ) -> Any:
        headers = dict(self.api_headers)
        if idempotency_key:
            # This is effective only if the Go API implements this contract.
            headers["Idempotency-Key"] = idempotency_key
        try:
            response = self.session.request(
                method, f"{self.api_base_url}{path}",
                headers=headers, json=body, timeout=self.timeout,
            )
            response.raise_for_status()
        except requests.HTTPError as exc:
            status = exc.response.status_code if exc.response is not None else "?"
            detail = exc.response.text[:400] if exc.response is not None else ""
            raise IntegrationError(
                f"API Go retornou HTTP {status} em {method} {path}: {detail}"
            ) from exc
        except requests.RequestException as exc:
            raise IntegrationError(f"Falha ao acessar API Go ({method} {path}).") from exc
        if not response.content:
            return None
        try:
            return response.json()
        except ValueError as exc:
            raise IntegrationError(f"Resposta JSON inválida da API Go em {path}.") from exc

    def send_message(self, chat_id: int, text: str) -> None:
        text = (text or "").strip() or "Não consegui obter uma resposta textual. Tente novamente mais tarde."
        for start in range(0, len(text), 4000):
            self._telegram_call(
                "sendMessage",
                body={"chat_id": chat_id, "text": text[start:start + 4000]},
            )

    def handle_message(self, message: dict[str, Any], *, update_id: int | None = None) -> None:
        chat = message.get("chat") or {}
        sender = message.get("from") or {}
        chat_id, user_id = chat.get("id"), sender.get("id")
        text = message.get("text")
        if chat_id is None or user_id is None or sender.get("is_bot"):
            return
        if chat.get("type") != "private":
            self.send_message(chat_id, "Por segurança, fale comigo em uma conversa privada com o bot.")
            return
        if not isinstance(text, str) or not text.strip():
            self.send_message(chat_id, "Por enquanto, consigo processar apenas mensagens de texto.")
            return
        text = text.strip()
        command = text.split(maxsplit=1)[0].split("@")[0]
        if command in ("/start", "/help"):
            self.send_message(chat_id, "Olá! Sou o assistente imobiliário. Envie sua dúvida para continuar o atendimento.")
            return

        lead_id = self.lead_mapping.get(int(user_id))
        if not lead_id:
            logger.info("Usuário Telegram sem mapeamento: %s", user_id)
            self.send_message(chat_id, "Seu usuário ainda não está vinculado a um atendimento. Entre em contato com a equipe responsável.")
            return

        conversations = self._api_call("GET", f"/leads/{lead_id}/conversations")
        conversation = newest_active_conversation_from_payload(conversations)
        if conversation is None:
            self.send_message(chat_id, "Não encontrei uma conversa ativa para seu atendimento. Nenhuma conversa nova foi criada; entre em contato com a equipe responsável.")
            return

        conversation_id = str(conversation["id"])
        # API-level idempotency requires server support. update_id is a stable key.
        result = self._api_call(
            "POST", "/messages",
            body={"conversation_id": conversation_id, "role": "USER", "content": text},
            idempotency_key=f"telegram-update-{update_id}" if update_id is not None else None,
        )
        answer = response_text(result)
        if not answer:
            raise IntegrationError("POST /messages não retornou uma resposta textual reconhecível.")
        self.send_message(chat_id, answer)
        logger.info("Mensagem processada para user_id=%s, conversation_id=%s", user_id, conversation_id)

    def process_update(self, update: dict[str, Any]) -> None:
        update_id = update.get("update_id")
        if not isinstance(update_id, int):
            logger.warning("Update sem update_id válido; ignorando.")
            return
        if self.store.is_processed(update_id):
            logger.info("Update duplicado ignorado: update_id=%s", update_id)
            return

        self.store.mark_processing(update_id)
        message = update.get("message")
        if isinstance(message, dict):
            self.handle_message(message, update_id=update_id)
        self.store.mark_processed(update_id)

    def run(self, retry_seconds: float = 3) -> None:
        logger.info(
            "Bot iniciado; API Go: %s; mapeamentos carregados: %d",
            self.api_base_url, len(self.lead_mapping),
        )
        while True:
            last_id = self.store.last_processed_update_id()
            params: dict[str, Any] = {"timeout": 30, "allowed_updates": '["message"]'}
            if last_id is not None:
                params["offset"] = last_id + 1
            try:
                updates = self._telegram_call("getUpdates", params=params, timeout=35)
                for update in updates or []:
                    # On failure, stop this batch. The update remains unprocessed
                    # and Telegram can deliver it again on the next poll.
                    self.process_update(update)
            except KeyboardInterrupt:
                logger.info("Bot encerrado pelo usuário.")
                break
            except IntegrationError as exc:
                logger.error("Falha temporária no processamento/polling: %s", exc)
                time.sleep(retry_seconds)
            except Exception:
                logger.exception("Erro inesperado no loop principal.")
                time.sleep(retry_seconds)


def main() -> None:
    logging.basicConfig(
        level=os.getenv("LOG_LEVEL", "INFO").upper(),
        format="%(asctime)s %(levelname)s %(name)s - %(message)s",
    )
    token = os.getenv("TELEGRAM_BOT_TOKEN", "").strip()
    api_url = os.getenv("REAL_ESTATE_API_URL", "http://localhost:8080").strip()
    api_key = os.getenv("API_KEY", "").strip() or os.getenv("REAL_ESTATE_API_KEY", "").strip()
    mapping = parse_mapping(os.getenv("TELEGRAM_LEAD_MAPPING", ""))
    db_path = os.getenv("TELEGRAM_UPDATE_DB", str(PROJECT_ROOT / "data" / "telegram_updates.sqlite3"))
    retry_seconds = float(os.getenv("TELEGRAM_POLL_RETRY_SECONDS", "3"))
    timeout = float(os.getenv("TELEGRAM_HTTP_TIMEOUT", "35"))

    bot = TelegramBot(
        bot_token=token,
        api_base_url=api_url,
        api_key=api_key,
        lead_mapping=mapping,
        store=UpdateStore(db_path),
        timeout=timeout,
    )
    bot.run(retry_seconds=retry_seconds)


if __name__ == "__main__":
    main()
