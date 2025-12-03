CREATE ROLE scorehub WITH LOGIN PASSWORD '1';
CREATE DATABASE scorehub OWNER scorehub;

GRANT ALL PRIVILEGES ON DATABASE scorehub TO scorehub;

-- From here on, run against the "scorehub" database (\c scorehub in psql)
-- Ensure schema usage and object privileges in the public schema
GRANT ALL ON SCHEMA public TO scorehub;
GRANT USAGE ON SCHEMA public TO scorehub;

-- Grant on existing objects in public schema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO scorehub;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO scorehub;

-- Ensure future objects in public schema also get privileges
ALTER DEFAULT PRIVILEGES FOR ROLE scorehub IN SCHEMA public
    GRANT ALL ON TABLES TO scorehub;
ALTER DEFAULT PRIVILEGES FOR ROLE scorehub IN SCHEMA public
    GRANT ALL ON SEQUENCES TO scorehub;

-- scorehub database initialization
-- Tables and indexes for users and notifications
-- Safe to run multiple times due to IF NOT EXISTS usage

-- Users table
CREATE TABLE IF NOT EXISTS public.users
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    score      BIGINT       NOT NULL DEFAULT 0,
    deleted    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Helpful index if querying by score (e.g., leaderboards)
CREATE INDEX IF NOT EXISTS idx_users_score ON public.users (score);

-- Enforce unique emails only for non-deleted users
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active ON public.users (email) WHERE deleted = FALSE;

-- Notifications table
CREATE TABLE IF NOT EXISTS public.notifications
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    message    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_notifications_user
        FOREIGN KEY (user_id)
            REFERENCES public.users (id)
            ON DELETE CASCADE
);

-- Indexes to accelerate common queries
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON public.notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON public.notifications (created_at);
-- Efficient retrieval of latest notifications per user
CREATE INDEX IF NOT EXISTS idx_notifications_user_created_at ON public.notifications (user_id, created_at DESC);
