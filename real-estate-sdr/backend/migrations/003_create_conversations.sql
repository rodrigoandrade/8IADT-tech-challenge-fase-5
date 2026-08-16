CREATE TABLE conversations (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                               lead_id UUID NOT NULL,

                               status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

                               started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               ended_at TIMESTAMPTZ NULL,

                               CONSTRAINT fk_conversations_lead
                                   FOREIGN KEY (lead_id)
                                       REFERENCES leads(id)
                                       ON DELETE CASCADE,

                               CONSTRAINT ck_conversations_status
                                   CHECK (
                                       status IN (
                                                  'ACTIVE',
                                                  'CLOSED'
                                           )
                                       ),

                               CONSTRAINT ck_conversations_dates
                                   CHECK (
                                       (status = 'ACTIVE' AND ended_at IS NULL)
                                           OR
                                       (status = 'CLOSED' AND ended_at IS NOT NULL)
                                       )
);

CREATE INDEX idx_conversations_lead_id
    ON conversations(lead_id);

CREATE INDEX idx_conversations_status
    ON conversations(status);

CREATE INDEX idx_conversations_started_at
    ON conversations(started_at);