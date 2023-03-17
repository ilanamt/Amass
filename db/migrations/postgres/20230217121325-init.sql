-- +migrate Up

CREATE TABLE IF NOT EXISTS executions(
    id SERIAL PRIMARY KEY,
    domains TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS assets(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    type VARCHAR(255),
    content JSONB);

CREATE TABLE IF NOT EXISTS execution_logs(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    asset_id INT,
    execution_id INT,
    CONSTRAINT fk_asset
        FOREIGN KEY (asset_id)
        REFERENCES assets(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_execution
        FOREIGN KEY (execution_id)
        REFERENCES executions(id)
        ON DELETE CASCADE);

CREATE TABLE IF NOT EXISTS relations(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    type VARCHAR(255),
    from_asset_id INT,
    to_asset_id INT,
    CONSTRAINT fk_from_asset
        FOREIGN KEY (from_asset_id)
        REFERENCES assets(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_to_asset
        FOREIGN KEY (to_asset_id)
        REFERENCES assets(id)
        ON DELETE CASCADE);

-- +migrate Down

DROP TABLE execution_logs;
DROP TABLE relations;
DROP TABLE assets;
DROP TABLE executions;

