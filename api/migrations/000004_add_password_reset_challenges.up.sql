CREATE TABLE password_reset_challenges (
    password_reset_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    code_hash BYTEA NOT NULL,
    password_reset_token_hash BYTEA,
    failed_attempts SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    code_expires_at TIMESTAMPTZ NOT NULL,
    password_reset_token_expires_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    consumed_at TIMESTAMPTZ,
    invalidated_at TIMESTAMPTZ,

    CONSTRAINT password_reset_challenges_user_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT password_reset_challenges_code_hash_unique
        UNIQUE (code_hash),
    CONSTRAINT password_reset_challenges_token_hash_unique
        UNIQUE (password_reset_token_hash),
    CONSTRAINT password_reset_challenges_failed_attempts_range
        CHECK (failed_attempts BETWEEN 0 AND 5),
    CONSTRAINT password_reset_challenges_updated_after_creation
        CHECK (updated_at >= created_at),
    CONSTRAINT password_reset_challenges_code_expiration_after_creation
        CHECK (code_expires_at > created_at),
    CONSTRAINT password_reset_challenges_confirmation_after_creation
        CHECK (confirmed_at IS NULL OR confirmed_at >= created_at),
    CONSTRAINT password_reset_challenges_consumption_after_confirmation
        CHECK (
            consumed_at IS NULL
            OR (confirmed_at IS NOT NULL AND consumed_at >= confirmed_at)
        ),
    CONSTRAINT password_reset_challenges_invalidation_after_creation
        CHECK (invalidated_at IS NULL OR invalidated_at >= created_at),
    CONSTRAINT password_reset_challenges_token_after_confirmation
        CHECK (
            (confirmed_at IS NULL
                AND password_reset_token_hash IS NULL
                AND password_reset_token_expires_at IS NULL)
            OR
            (confirmed_at IS NOT NULL
                AND password_reset_token_hash IS NOT NULL
                AND password_reset_token_expires_at IS NOT NULL
                AND password_reset_token_expires_at > confirmed_at)
        ),
    CONSTRAINT password_reset_challenges_single_terminal_state
        CHECK (consumed_at IS NULL OR invalidated_at IS NULL)
);

CREATE INDEX password_reset_challenges_user_id_idx
    ON password_reset_challenges (user_id);

CREATE UNIQUE INDEX password_reset_challenges_active_user_idx
    ON password_reset_challenges (user_id)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE INDEX password_reset_challenges_code_expires_at_idx
    ON password_reset_challenges (code_expires_at);

CREATE INDEX password_reset_challenges_token_expires_at_idx
    ON password_reset_challenges (password_reset_token_expires_at)
    WHERE password_reset_token_expires_at IS NOT NULL;

CREATE INDEX password_reset_challenges_consumed_at_idx
    ON password_reset_challenges (consumed_at)
    WHERE consumed_at IS NOT NULL;

CREATE INDEX password_reset_challenges_invalidated_at_idx
    ON password_reset_challenges (invalidated_at)
    WHERE invalidated_at IS NOT NULL;
