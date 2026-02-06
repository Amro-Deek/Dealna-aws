BEGIN;


CREATE TABLE IF NOT EXISTS sessions (
    session_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id        UUID NOT NULL,
    jti            UUID NOT NULL UNIQUE,

    revoked        BOOLEAN NOT NULL DEFAULT FALSE,

    expires_at     TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at     TIMESTAMP,

    CONSTRAINT fk_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES "User"(user_id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id
    ON sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_sessions_jti
    ON sessions(jti);

CREATE INDEX IF NOT EXISTS idx_sessions_expires_at
    ON sessions(expires_at);

COMMIT;
