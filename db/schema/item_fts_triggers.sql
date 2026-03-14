CREATE TRIGGER IF NOT EXISTS item_ai AFTER INSERT ON item BEGIN
	INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
END;
CREATE TRIGGER IF NOT EXISTS item_ad AFTER DELETE ON item BEGIN
	INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
END;
CREATE TRIGGER IF NOT EXISTS item_au AFTER UPDATE ON item BEGIN
	INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
	INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
END;
