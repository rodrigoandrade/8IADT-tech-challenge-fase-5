CREATE TABLE messages (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                          conversation_id UUID NOT NULL,

                          role VARCHAR(20) NOT NULL,

                          content TEXT NOT NULL,

                          created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

                          CONSTRAINT fk_messages_conversation
                              FOREIGN KEY (conversation_id)
                                  REFERENCES conversations(id)
                                  ON DELETE CASCADE,

                          CONSTRAINT ck_messages_role
                              CHECK (
                                  role IN (
                                           'USER',
                                           'ASSISTANT',
                                           'SYSTEM'
                                      )
                                  )
);

CREATE INDEX idx_messages_conversation_created_at
    ON messages(conversation_id, created_at);