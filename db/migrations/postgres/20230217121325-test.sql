-- +migrate Up

CREATE TABLE IF NOT EXISTS enum_executions(
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS assets(
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    enum_execution_id INT,
    type VARCHAR(255),
    content JSONB,
    CONSTRAINT fk_enum_executions
        FOREIGN KEY (enum_execution_id)
        REFERENCES enum_executions(id)
        ON DELETE SET NULL);

CREATE TABLE IF NOT EXISTS relations(
    id SERIAL PRIMARY KEY,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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

DROP TABLE relations;
DROP TABLE assets;
DROP TABLE enum_executions;
