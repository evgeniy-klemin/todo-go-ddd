-- +goose Up
ALTER TABLE item ADD COLUMN description TEXT DEFAULT '';

CREATE VIRTUAL TABLE IF NOT EXISTS item_fts USING fts5(item_id UNINDEXED, name, description);

-- Populate FTS index from existing data
INSERT INTO item_fts(item_id, name, description) SELECT id, name, COALESCE(description, '') FROM item;

-- Triggers to keep FTS in sync
CREATE TRIGGER item_fts_ai AFTER INSERT ON item BEGIN
    INSERT INTO item_fts(item_id, name, description) VALUES (new.id, new.name, COALESCE(new.description, ''));
END;

CREATE TRIGGER item_fts_ad AFTER DELETE ON item BEGIN
    DELETE FROM item_fts WHERE item_id = old.id;
END;

CREATE TRIGGER item_fts_au AFTER UPDATE ON item BEGIN
    DELETE FROM item_fts WHERE item_id = old.id;
    INSERT INTO item_fts(item_id, name, description) VALUES (new.id, new.name, COALESCE(new.description, ''));
END;

-- +goose Down
DROP TRIGGER IF EXISTS item_fts_au;
DROP TRIGGER IF EXISTS item_fts_ad;
DROP TRIGGER IF EXISTS item_fts_ai;
DROP TABLE IF EXISTS item_fts;
-- SQLite does not support DROP COLUMN, so we leave the column in place for down migration
