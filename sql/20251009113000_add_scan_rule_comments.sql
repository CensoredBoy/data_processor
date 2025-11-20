++ sql/20251009113000_add_scan_rule_comments.sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    scan_rule_id INTEGER NOT NULL,
    previous_comment_id INTEGER UNIQUE,
    user_id INTEGER NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (scan_rule_id) REFERENCES scan_rules(id) ON DELETE CASCADE,
    FOREIGN KEY (previous_comment_id) REFERENCES comments(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE scan_rules
    ADD COLUMN latest_comment_id INTEGER UNIQUE,
    ADD CONSTRAINT fk_scan_rules_latest_comment
        FOREIGN KEY (latest_comment_id) REFERENCES comments(id) ON DELETE SET NULL;

CREATE INDEX idx_comments_scan_rule_id ON comments(scan_rule_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);
CREATE INDEX idx_comments_user_id ON comments(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE scan_rules DROP CONSTRAINT IF EXISTS fk_scan_rules_latest_comment;
ALTER TABLE scan_rules DROP COLUMN IF EXISTS latest_comment_id;

DROP TABLE IF EXISTS comments;
-- +goose StatementEnd

