-- SmartPost Database Schema
-- Migration 001: Initial schema

BEGIN;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    telegram_id   BIGINT UNIQUE NOT NULL,
    username      VARCHAR(255),
    first_name    VARCHAR(255),
    timezone      VARCHAR(50) DEFAULT 'Asia/Tashkent',
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Channels table
CREATE TABLE IF NOT EXISTS channels (
    id         BIGSERIAL PRIMARY KEY,
    owner_id   BIGINT REFERENCES users(id) ON DELETE CASCADE,
    chat_id    BIGINT UNIQUE NOT NULL,
    title      VARCHAR(255),
    username   VARCHAR(255),
    is_active  BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Posts table
CREATE TABLE IF NOT EXISTS posts (
    id             BIGSERIAL PRIMARY KEY,
    channel_id     BIGINT REFERENCES channels(id) ON DELETE CASCADE,
    user_id        BIGINT REFERENCES users(id) ON DELETE CASCADE,
    media_type     VARCHAR(20) CHECK (media_type IN ('text','photo','video','video_note')),
    file_id        TEXT,
    caption        TEXT,
    status         VARCHAR(20) DEFAULT 'draft'
                   CHECK (status IN ('draft','scheduled','sent','failed')),
    scheduled_at   TIMESTAMPTZ,
    sent_at        TIMESTAMPTZ,
    error_message  TEXT,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- Buttons table
CREATE TABLE IF NOT EXISTS buttons (
    id          BIGSERIAL PRIMARY KEY,
    post_id     BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    text        VARCHAR(255) NOT NULL,
    url         TEXT NOT NULL,
    color_code  VARCHAR(20) DEFAULT 'default',
    row_index   INT DEFAULT 0,
    col_index   INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_posts_channel_id ON posts(channel_id);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_scheduled ON posts(scheduled_at) WHERE status = 'scheduled';
CREATE INDEX IF NOT EXISTS idx_buttons_post_id ON buttons(post_id);
CREATE INDEX IF NOT EXISTS idx_channels_owner_id ON channels(owner_id);

COMMIT;
