BEGIN;

CREATE TABLE IF NOT EXISTS student_pre_registration (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email VARCHAR(255) NOT NULL UNIQUE,
    token UUID NOT NULL UNIQUE,

    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,

    resend_count INTEGER NOT NULL DEFAULT 0,
    resend_window_start TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP
);
COMMIT;
