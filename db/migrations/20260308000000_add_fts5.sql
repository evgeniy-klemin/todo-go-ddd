-- +goose Up
CREATE VIRTUAL TABLE item_fts USING fts5(name, content='item', content_rowid='rowid');

CREATE TRIGGER item_ai AFTER INSERT ON item BEGIN
  INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
END;

CREATE TRIGGER item_ad AFTER DELETE ON item BEGIN
  INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
END;

CREATE TRIGGER item_au AFTER UPDATE ON item BEGIN
  INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
  INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
END;

-- Rebuild FTS index from existing data
INSERT INTO item_fts(item_fts) VALUES('rebuild');

-- +goose Down
DROP TRIGGER IF EXISTS item_au;
DROP TRIGGER IF EXISTS item_ad;
DROP TRIGGER IF EXISTS item_ai;
DROP TABLE IF EXISTS item_fts;
