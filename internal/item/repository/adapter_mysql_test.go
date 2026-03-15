package repository

import (
	"testing"
)

func TestBuildMySQLFTSQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"buy", "+buy*"},
		{"buy milk", "+buy* +milk*"},
		{"Buy Milk", "+Buy* +Milk*"},
		{"", ""},
		{"buy+milk", "+buymilk*"},
		{"buy-milk", "+buymilk*"},
		{"buy*milk", "+buymilk*"},
		{"+", ""},
		{"+ -", ""},
	}
	for _, tt := range tests {
		got := buildMySQLFTSQuery(tt.input)
		if got != tt.expected {
			t.Errorf("buildMySQLFTSQuery(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMySQLAdapterSearchCondition(t *testing.T) {
	a := &mysqlAdapter{}

	t.Run("FTSMode", func(t *testing.T) {
		cond, arg := a.SearchCondition("buy", true)
		wantCond := "MATCH(name) AGAINST(? IN BOOLEAN MODE)"
		wantArg := "+buy*"
		if cond != wantCond {
			t.Errorf("condition = %q, want %q", cond, wantCond)
		}
		if arg != wantArg {
			t.Errorf("arg = %q, want %q", arg, wantArg)
		}
	})

	t.Run("LIKEFallback", func(t *testing.T) {
		cond, arg := a.SearchCondition("buy", false)
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
