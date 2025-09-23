CREATE TABLE IF NOT EXISTS github_users (
    id             INTEGER PRIMARY KEY,
    login          TEXT    NOT NULL,
    node_id        TEXT,
    avatar_url     TEXT,
    url            TEXT,
    html_url       TEXT,
    type           TEXT,
    user_view_type TEXT,
    site_admin     BOOLEAN DEFAULT FALSE,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_github_users_login ON github_users (login);
