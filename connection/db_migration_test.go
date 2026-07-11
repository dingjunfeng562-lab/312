package connection

import (
	"errors"
	"testing"
)

func TestValidSqlErrorIgnoresAppliedMigrations(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		valid bool
	}{
		{name: "nil", err: nil, valid: false},
		{name: "mysql duplicate column", err: errors.New("Error 1060: Duplicate column name 'task_id'"), valid: false},
		{name: "mysql existing table", err: errors.New("Error 1050: Table already exists"), valid: false},
		{name: "sqlite duplicate column", err: errors.New("SQL logic error: duplicate column name: task_id (1)"), valid: false},
		{name: "sqlite existing index", err: errors.New("SQL logic error: index idx_name already exists (1)"), valid: false},
		{name: "real failure", err: errors.New("database is locked"), valid: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validSqlError(test.err); got != test.valid {
				t.Fatalf("validSqlError(%v) = %v, want %v", test.err, got, test.valid)
			}
		})
	}
}
