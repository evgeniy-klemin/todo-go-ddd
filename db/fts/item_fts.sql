CREATE VIRTUAL TABLE IF NOT EXISTS item_fts USING fts5(name, content='item', content_rowid='rowid')
