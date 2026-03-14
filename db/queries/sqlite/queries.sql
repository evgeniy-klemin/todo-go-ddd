-- name: GetItemByID :one
SELECT id, name, position, done, created_at FROM item WHERE id = ?;

-- name: InsertItem :exec
INSERT INTO item (id, name, position, done, created_at) VALUES (?, ?, ?, ?, ?);

-- name: UpdateItem :exec
UPDATE item SET name = ?, position = ?, done = ? WHERE id = ?;

-- name: MaxPosition :one
SELECT COALESCE(MAX(position), 0) as max_position FROM item;
