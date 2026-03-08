-- +goose Up
ALTER TABLE item ADD FULLTEXT INDEX idx_item_name_fulltext (name);

-- +goose Down
ALTER TABLE item DROP INDEX idx_item_name_fulltext;
