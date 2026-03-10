BEGIN;

-- recreate sessions table
CREATE TABLE sessions (
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

CREATE INDEX idx_sessions_user_id
    ON sessions(user_id);

CREATE INDEX idx_sessions_jti
    ON sessions(jti);

CREATE INDEX idx_sessions_expires_at
    ON sessions(expires_at);

-- restore removed columns
ALTER TABLE public."User"
ADD COLUMN password_hash character varying(255),
ADD COLUMN failed_login_attempts integer DEFAULT 0,
ADD COLUMN last_login_at timestamp;

-- remove keycloak_sub
ALTER TABLE public."User"
DROP COLUMN IF EXISTS keycloak_sub;

COMMIT;