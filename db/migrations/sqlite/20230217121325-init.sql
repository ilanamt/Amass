-- +migrate Up

-- see https://www.sqlite.org/foreignkeys.html#fk_enable about enabling foreign keys
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS executions(
    id INTEGER PRIMARY KEY,
    domains TEXT,
    created_at TEXT DEFAULT (unixepoch()))
    STRICT;

CREATE TABLE IF NOT EXISTS execution_logs(
    id INTEGER PRIMARY KEY,
    asset_id INTEGER,
    execution_id INTEGER,
    created_at TEXT DEFAULT (unixepoch()),
    FOREIGN KEY(asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY(execution_id) REFERENCES executions(id) ON DELETE CASCADE)
    STRICT;


CREATE TABLE IF NOT EXISTS assets(
    id INTEGER PRIMARY KEY,
    created_at TEXT DEFAULT (unixepoch()),
    type TEXT,
    content TEXT)
    STRICT;

CREATE TABLE IF NOT EXISTS relations(
    id INTEGER PRIMARY KEY,
    created_at TEXT DEFAULT (unixepoch()),
    type TEXT,
    from_asset_id INTEGER,
    to_asset_id INTEGER,
    FOREIGN KEY(from_asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY(to_asset_id) REFERENCES assets(id) ON DELETE CASCADE)
    STRICT;

-- +migrate Down

DROP TABLE relations;
DROP TABLE execution_logs;
DROP TABLE assets;
DROP TABLE executions;
