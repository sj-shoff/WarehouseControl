-- +goose Up
CREATE TABLE IF NOT EXISTS items_history (
    id SERIAL PRIMARY KEY,
    item_id INT NOT NULL,
    action TEXT NOT NULL,
    old_data JSONB,
    new_data JSONB,
    changed_by TEXT NOT NULL,
    changed_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS items_history;