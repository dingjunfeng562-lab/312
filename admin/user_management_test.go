package admin

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

func newUserManagementTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	statements := []string{
		`CREATE TABLE auth (id INTEGER PRIMARY KEY, username TEXT UNIQUE, email TEXT UNIQUE, password TEXT, token TEXT, is_admin BOOLEAN, is_banned BOOLEAN)`,
		`CREATE TABLE quota (user_id INTEGER PRIMARY KEY, quota REAL, used REAL)`,
		`CREATE TABLE subscription (user_id INTEGER PRIMARY KEY, expired_at TEXT, total_month INTEGER, enterprise BOOLEAN, level INTEGER)`,
		`CREATE TABLE invitation (id INTEGER PRIMARY KEY, code TEXT UNIQUE, used BOOLEAN, used_id INTEGER, updated_at TEXT)`,
		`CREATE TABLE broadcast (poster_id INTEGER)`,
		`CREATE TABLE sharing (user_id INTEGER)`,
		`CREATE TABLE mask (user_id INTEGER)`,
		`CREATE TABLE conversation (user_id INTEGER)`,
		`CREATE TABLE apikey (user_id INTEGER)`,
		`CREATE TABLE package (user_id INTEGER)`,
		`INSERT INTO auth VALUES (1, 'rootadmin', 'root@example.com', 'x', 'x', TRUE, FALSE)`,
		`INSERT INTO auth VALUES (2, 'alice', 'alice@example.com', 'x', 'x', FALSE, FALSE)`,
		`INSERT INTO auth VALUES (3, 'bob', 'bob@example.com', 'x', 'x', TRUE, TRUE)`,
		`INSERT INTO quota VALUES (1, 100, 10), (2, 50, 25), (3, 200, 5)`,
		`INSERT INTO subscription VALUES (1, '2099-01-01 00:00:00', 12, TRUE, 3)`,
		`INSERT INTO subscription VALUES (2, '2000-01-01 00:00:00', 2, FALSE, 1)`,
		`INSERT INTO invitation VALUES (1, 'INV-BOB', TRUE, 3, '2026-01-01 00:00:00')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("prepare database with %q: %v", statement, err)
		}
	}
	oldRoot := viper.GetString("root.username")
	viper.Set("root.username", "rootadmin")
	t.Cleanup(func() { viper.Set("root.username", oldRoot) })
	return db
}

func TestGetUsersFormAppliesFiltersAndSort(t *testing.T) {
	db := newUserManagementTestDB(t)
	form := getUsersForm(db, 0, "", UserFilter{Admin: "yes", Ban: "all", Plan: "all", Sort: "quota-desc"})
	if !form.Status || len(form.Data) != 2 {
		t.Fatalf("filtered list = status %v, data %v, message %q", form.Status, len(form.Data), form.Message)
	}
	first := form.Data[0].(UserData)
	second := form.Data[1].(UserData)
	if first.Username != "bob" || second.Username != "rootadmin" {
		t.Fatalf("quota-desc order = %s, %s", first.Username, second.Username)
	}
	if first.InvitationCode != "INV-BOB" {
		t.Fatalf("invitation code = %q, want INV-BOB", first.InvitationCode)
	}

	form = getUsersForm(db, 0, "alice@example", UserFilter{Plan: "no", Sort: "id-asc"})
	if !form.Status || len(form.Data) != 1 || form.Data[0].(UserData).Username != "alice" {
		t.Fatalf("email/search/plan filter returned %#v (%s)", form.Data, form.Message)
	}
}

func TestUpdateUserProfileUpdatesAllEditableProfileFields(t *testing.T) {
	db := newUserManagementTestDB(t)
	err := updateUserProfile(db, nil, 2, UserProfileUpdate{
		Username: "alice-new", Email: "alice-new@example.com", UsedQuota: 8.5,
		TotalMonth: 7, Enterprise: true,
	})
	if err != nil {
		t.Fatalf("updateUserProfile: %v", err)
	}
	var username, email string
	var used float64
	var months int64
	var enterprise bool
	if err := db.QueryRow(`SELECT auth.username, auth.email, quota.used, subscription.total_month, subscription.enterprise
		FROM auth JOIN quota ON quota.user_id = auth.id JOIN subscription ON subscription.user_id = auth.id WHERE auth.id = 2`).
		Scan(&username, &email, &used, &months, &enterprise); err != nil {
		t.Fatalf("query updated profile: %v", err)
	}
	if username != "alice-new" || email != "alice-new@example.com" || used != 8.5 || months != 7 || !enterprise {
		t.Fatalf("updated profile = %q %q %v %v %v", username, email, used, months, enterprise)
	}
}

func TestUpdateUserProfileAllowsLegacyUserWithoutEmail(t *testing.T) {
	db := newUserManagementTestDB(t)
	if _, err := db.Exec("UPDATE auth SET email = NULL WHERE id = 2"); err != nil {
		t.Fatalf("clear email: %v", err)
	}
	err := updateUserProfile(db, nil, 2, UserProfileUpdate{
		Username: "alice", Email: "", UsedQuota: 11, TotalMonth: 4, Enterprise: true,
	})
	if err != nil {
		t.Fatalf("update legacy profile: %v", err)
	}
	var email sql.NullString
	var used float64
	if err := db.QueryRow(`SELECT auth.email, quota.used FROM auth JOIN quota ON quota.user_id = auth.id WHERE auth.id = 2`).Scan(&email, &used); err != nil {
		t.Fatalf("query legacy profile: %v", err)
	}
	if email.Valid || used != 11 {
		t.Fatalf("legacy profile email=%v used=%v", email, used)
	}
}

func TestAdminSafetyRules(t *testing.T) {
	db := newUserManagementTestDB(t)
	if err := setAdmin(db, nil, 1, false); err == nil || !strings.Contains(err.Error(), "root") {
		t.Fatalf("demote root error = %v", err)
	}
	if err := banUser(db, nil, 1, true); err == nil || !strings.Contains(err.Error(), "root") {
		t.Fatalf("ban root error = %v", err)
	}
	if err := deleteUser(db, nil, 1, "bob"); err == nil || !strings.Contains(err.Error(), "root") {
		t.Fatalf("delete root error = %v", err)
	}
	if err := deleteUser(db, nil, 3, "bob"); err == nil || !strings.Contains(err.Error(), "own account") {
		t.Fatalf("delete current account error = %v", err)
	}
}

func TestDeleteUserCleansDependentDataAndReleasesInvitation(t *testing.T) {
	db := newUserManagementTestDB(t)
	for _, table := range []string{"broadcast", "sharing", "mask", "conversation", "apikey", "package"} {
		column := "user_id"
		if table == "broadcast" {
			column = "poster_id"
		}
		if _, err := db.Exec("INSERT INTO " + table + " (" + column + ") VALUES (3)"); err != nil {
			t.Fatalf("seed %s: %v", table, err)
		}
	}
	if err := deleteUser(db, nil, 3, "rootadmin"); err != nil {
		t.Fatalf("deleteUser: %v", err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM auth WHERE id = 3").Scan(&count); err != nil || count != 0 {
		t.Fatalf("deleted auth count = %d, err = %v", count, err)
	}
	var used bool
	var usedID sql.NullInt64
	if err := db.QueryRow("SELECT used, used_id FROM invitation WHERE code = 'INV-BOB'").Scan(&used, &usedID); err != nil {
		t.Fatalf("query invitation: %v", err)
	}
	if used || usedID.Valid {
		t.Fatalf("invitation remains assigned: used=%v used_id=%v", used, usedID)
	}
}
