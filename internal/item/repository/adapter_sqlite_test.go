package repository

import (
	"testing"
)

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"buy", `"buy"*`},
		{"buy milk", `"buy"* "milk"*`},
		{"Buy Milk", `"Buy"* "Milk"*`},
		{"", ""},
	}
	for _, tt := range tests {
		got := buildFTSQuery(tt.input)
		if got != tt.expected {
			t.Errorf("buildFTSQuery(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSQLiteAdapterSearchCondition(t *testing.T) {
	t.Run("FTSMode", func(t *testing.T) {
		a := &sqliteAdapter{ftsEnabled: true}
		cond, arg := a.searchCondition("buy")
		wantCond := "item.rowid IN (SELECT rowid FROM item_fts WHERE item_fts MATCH ?)"
		wantArg := `"buy"*`
		if cond != wantCond {
			t.Errorf("condition = %q, want %q", cond, wantCond)
		}
		if arg != wantArg {
			t.Errorf("arg = %q, want %q", arg, wantArg)
		}
	})

	t.Run("LIKEFallback", func(t *testing.T) {
		a := &sqliteAdapter{ftsEnabled: false}
		cond, arg := a.searchCondition("buy")
		wantCond := "LOWER(name) LIKE LOWER(?)"
		wantArg := "%buy%"
		if cond != wantCond {
			t.Errorf("condition = %q, want %q", cond, wantCond)
		}
		if arg != wantArg {
			t.Errorf("arg = %q, want %q", arg, wantArg)
		}
	})
}
