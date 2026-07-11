package admin

import (
	"database/sql"
	"math"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func newQuotaTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	for _, statement := range []string{
		"CREATE TABLE auth (id INTEGER PRIMARY KEY, username TEXT)",
		"CREATE TABLE quota (user_id INTEGER PRIMARY KEY, quota REAL, used REAL)",
		"INSERT INTO auth (id, username) VALUES (1, 'alice'), (2, 'bob')",
		"INSERT INTO quota (user_id, quota, used) VALUES (1, 100, 5)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("prepare database with %q: %v", statement, err)
		}
	}

	return db
}

func TestQuotaMigrationChangesAndSetsPoints(t *testing.T) {
	db := newQuotaTestDB(t)

	quota, err := quotaMigration(db, 1, 25, false)
	if err != nil || quota != 125 {
		t.Fatalf("increase points = (%v, %v), want (125, nil)", quota, err)
	}

	quota, err = quotaMigration(db, 1, -20, false)
	if err != nil || quota != 105 {
		t.Fatalf("decrease points = (%v, %v), want (105, nil)", quota, err)
	}

	quota, err = quotaMigration(db, 1, 42.5, true)
	if err != nil || quota != 42.5 {
		t.Fatalf("set points = (%v, %v), want (42.5, nil)", quota, err)
	}

	var used float32
	if err := db.QueryRow("SELECT used FROM quota WHERE user_id = 1").Scan(&used); err != nil {
		t.Fatalf("query used points: %v", err)
	}
	if used != 5 {
		t.Fatalf("used points changed to %v, want 5", used)
	}
}

func TestQuotaMigrationCreatesMissingBalance(t *testing.T) {
	db := newQuotaTestDB(t)

	quota, err := quotaMigration(db, 2, 30, false)
	if err != nil || quota != 30 {
		t.Fatalf("create points = (%v, %v), want (30, nil)", quota, err)
	}
}

func TestQuotaMigrationRejectsInvalidChanges(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		quota    float32
		override bool
		message  string
	}{
		{name: "missing user", id: 99, quota: 1, message: "user not found"},
		{name: "negative result", id: 1, quota: -101, message: "cannot be negative"},
		{name: "negative override", id: 1, quota: -1, override: true, message: "cannot be negative"},
		{name: "not a number", id: 1, quota: float32(math.NaN()), message: "invalid points value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newQuotaTestDB(t)
			_, err := quotaMigration(db, test.id, test.quota, test.override)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("quotaMigration error = %v, want message containing %q", err, test.message)
			}
		})
	}
}
