CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE leads (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       name VARCHAR(150) NOT NULL,
                       email VARCHAR(255) NOT NULL,
                       phone VARCHAR(30) NOT NULL,
                       status VARCHAR(30) NOT NULL DEFAULT 'NEW',
                       created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

                       CONSTRAINT ck_leads_status
                           CHECK (
                               status IN (
                                          'NEW',
                                          'CONTACTED',
                                          'QUALIFIED',
                                          'CONVERTED',
                                          'LOST'
                                   )
                               )
);

CREATE INDEX idx_leads_status
    ON leads(status);

CREATE INDEX idx_leads_phone
    ON leads(phone);

CREATE INDEX idx_leads_created_at
    ON leads(created_at);