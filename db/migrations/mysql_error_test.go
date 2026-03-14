package migrations

// Internal test package so we can access the unexported isMySQLError helper.

import (
	"errors"
	"fmt"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestIsMySQLError_MatchingCode(t *testing.T) {
	err := &mysql.MySQLError{Number: 1061, Message: "Duplicate key name"}
	assert.True(t, isMySQLError(err, 1061))
}

func TestIsMySQLError_WrongCode(t *testing.T) {
	err := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}
	assert.False(t, isMySQLError(err, 1061))
}

func TestIsMySQLError_NonMySQLError(t *testing.T) {
	err := errors.New("some other error")
	assert.False(t, isMySQLError(err, 1061))
}

func TestIsMySQLError_WrappedMySQLError(t *testing.T) {
	// errors.As unwraps, so a wrapped *MySQLError should still match.
	inner := &mysql.MySQLError{Number: 1061, Message: "Duplicate key name"}
	wrapped := fmt.Errorf("exec failed: %w", inner)
	assert.True(t, isMySQLError(wrapped, 1061))
}

func TestIsMySQLError_Nil(t *testing.T) {
	// Passing nil should never panic and should return false.
	assert.False(t, isMySQLError(nil, 1061))
}
