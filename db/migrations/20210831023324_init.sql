-- +goose Up
CREATE TABLE item (
	id VARCHAR(36) NOT NULL PRIMARY KEY,
   	name VARCHAR(1000) NOT NULL,
	position INTEGER NOT NULL DEFAULT 1,
	done BOOL NOT NULL DEFAULT FALSE,
	created_at DATETIME NOT NULL
);
CREATE INDEX idx_item_position ON item (position);

-- +goose Down
DROP TABLE item;
