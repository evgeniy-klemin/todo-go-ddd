package fts

import (
	"database/sql"
	_ "embed"
)

//go:embed item_fts.sql
var ftsTable string

//go:embed item_fts_triggers.sql
var ftsTriggers string

//go:embed item_fts_rebuild.sql
var ftsRebuild string

//go:embed item_fts_drop_triggers.sql
var dropTriggers string

// Apply creates the FTS5 virtual table and triggers on db.
// Returns true on success, false if FTS5 is unavailable or setup fails.
// FTS5 unavailability is graceful degradation — the app falls back to LIKE
// search — so none of the errors below are fatal.
func Apply(db *sql.DB) bool {
	if _, err := db.Exec(ftsTable); err != nil {
		// FTS5 not compiled in or table creation failed; clean up any stale
		// triggers. The cleanup error is intentionally ignored: the triggers
		// may not exist yet, which is expected on a fresh database.
		_, _ = db.Exec(dropTriggers)
		return false
	}
	if _, err := db.Exec(ftsTriggers); err != nil {
		return false
	}
	// Rebuild populates the FTS index from existing rows. A failure here is
	// non-fatal: the index will be populated incrementally as rows are
	// inserted/updated via the triggers we just created.
	_, _ = db.Exec(ftsRebuild)
	return true
}
