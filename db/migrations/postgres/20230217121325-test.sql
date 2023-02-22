
-- +migrate Up

CREATE TABLE IF NOT EXISTS enum_executions(
    id INT PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS assets(
    id INT PRIMARY KEY,
    type VARCHAR(255),
    content JSONB,
    CONSTRAINT fk_enum_executions
        FOREIGN KEY (enum_execution_id)
        REFERENCES enum_executions(id)
        ON DELETE SET NULL);

CREATE TABLE IF NOT EXISTS relations(
    id INT PRIMARY KEY,
    type VARCHAR(255),
    CONSTRAINT fk_from_asset
        FOREIGN KEY (from_asset_id)
        REFERENCES assets(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_to_asset
        FOREIGN KEY (to_asset_id)
        REFERENCES assets(id)
        ON DELETE CASCADE);

-- +migrate Down
