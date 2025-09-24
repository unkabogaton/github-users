-- +goose Up
CREATE TABLE IF NOT EXISTS github_users (
    id             BIGINT PRIMARY KEY,
    login          VARCHAR(255) NOT NULL,
    node_id        VARCHAR(255),
    avatar_url     VARCHAR(255),
    url            VARCHAR(255),
    html_url       VARCHAR(255),
    type           VARCHAR(50),
    user_view_type VARCHAR(50),
    site_admin     TINYINT(1) NOT NULL DEFAULT 0,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

ALTER TABLE github_users ADD INDEX idx_github_users_login (login);

-- +goose Down
DROP INDEX idx_github_users_login ON github_users;
DROP TABLE IF EXISTS github_users;
