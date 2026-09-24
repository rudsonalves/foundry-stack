ALTER TABLE users
    ADD COLUMN email_verified_at TIMESTAMPTZ NOT NULL;

CREATE TABLE email_verification_challenges (
    verification_id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code_hash BYTEA NOT NULL,
    verification_token_hash BYTEA,
    failed_attempts SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    code_expires_at TIMESTAMPTZ NOT NULL,
    verification_token_expires_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    consumed_at TIMESTAMPTZ,
    invalidated_at TIMESTAMPTZ,

    CONSTRAINT email_verification_challenges_email_not_blank
        CHECK (BTRIM(email) <> ''),
    CONSTRAINT email_verification_challenges_code_hash_unique
        UNIQUE (code_hash),
    CONSTRAINT email_verification_challenges_token_hash_unique
        UNIQUE (verification_token_hash),
    CONSTRAINT email_verification_challenges_failed_attempts_range
        CHECK (failed_attempts BETWEEN 0 AND 5),
    CONSTRAINT email_verification_challenges_code_expiration_after_creation
        CHECK (code_expires_at > created_at),
    CONSTRAINT email_verification_challenges_confirmation_after_creation
        CHECK (confirmed_at IS NULL OR confirmed_at >= created_at),
    CONSTRAINT email_verification_challenges_consumption_after_confirmation
        CHECK (
            consumed_at IS NULL
            OR (confirmed_at IS NOT NULL AND consumed_at >= confirmed_at)
        ),
    CONSTRAINT email_verification_challenges_invalidation_after_creation
        CHECK (invalidated_at IS NULL OR invalidated_at >= created_at),
    CONSTRAINT email_verification_challenges_token_after_confirmation
        CHECK (
            (confirmed_at IS NULL
                AND verification_token_hash IS NULL
                AND verification_token_expires_at IS NULL)
            OR
            (confirmed_at IS NOT NULL
                AND verification_token_hash IS NOT NULL
                AND verification_token_expires_at IS NOT NULL
                AND verification_token_expires_at > confirmed_at)
        ),
    CONSTRAINT email_verification_challenges_single_terminal_state
        CHECK (consumed_at IS NULL OR invalidated_at IS NULL)
);

CREATE INDEX email_verification_challenges_email_idx
    ON email_verification_challenges (email);

CREATE INDEX email_verification_challenges_code_expires_at_idx
    ON email_verification_challenges (code_expires_at);

CREATE INDEX email_verification_challenges_token_expires_at_idx
    ON email_verification_challenges (verification_token_expires_at)
    WHERE verification_token_expires_at IS NOT NULL;

CREATE UNIQUE INDEX email_verification_challenges_active_email_idx
    ON email_verification_challenges (email)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;
