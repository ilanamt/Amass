-- +migrate Up

-- see https://www.sqlite.org/foreignkeys.html#fk_enable about enabling foreign keys
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS enum_executions(
    id INTEGER PRIMARY KEY,
    created_at TEXT DEFAULT (unixepoch()))
    STRICT;

CREATE TABLE IF NOT EXISTS assets(
    id INTEGER PRIMARY KEY,
    created_at TEXT DEFAULT (unixepoch()),
    enum_execution_id INTEGER,
    type TEXT,
    content TEXT,
    FOREIGN KEY(enum_execution_id) REFERENCES enum_executions(id) ON DELETE SET NULL)
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
DROP TABLE assets;
DROP TABLE enum_executions;
