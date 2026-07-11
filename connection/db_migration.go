package connection

import (
	"chat/globals"
	"database/sql"
	"strings"
)

func validSqlError(err error) bool {
	if err == nil {
		return false
	}

	content := strings.ToLower(err.Error())

	// Migrations are intentionally rerun on startup. Treat duplicate schema
	// objects as already applied for both MySQL and SQLite.
	ignored := []string{
		"error 1060",            // MySQL: duplicate column name
		"error 1050",            // MySQL: table already exists
		"duplicate column name", // SQLite
		"already exists",        // SQLite/MySQL schema object
	}
	for _, marker := range ignored {
		if strings.Contains(content, marker) {
			return false
		}
	}

	return true
}

func checkSqlError(_ sql.Result, err error) error {
	if validSqlError(err) {
		return err
	}

	return nil
}

func execSql(db *sql.DB, sql string, args ...interface{}) error {
	return checkSqlError(globals.ExecDb(db, sql, args...))
}

func doMigration(db *sql.DB) error {
	if globals.SqliteEngine {
		return doSqliteMigration(db)
	}

	// v3.10 migration

	// update `quota`, `used` field in `quota` table
	// migrate `DECIMAL(16, 4)` to `DECIMAL(24, 6)`

	if err := execSql(db, `
		ALTER TABLE quota
		MODIFY COLUMN quota DECIMAL(24, 6),
		MODIFY COLUMN used DECIMAL(24, 6);
	`); err != nil {
		return err
	}

	// add new field `is_banned` in `auth` table
	if err := execSql(db, `
		ALTER TABLE auth
		ADD COLUMN is_banned BOOLEAN DEFAULT FALSE;
	`); err != nil {
		return err
	}

	// add new field `task_id` in `conversation` table to store task id (e.g., video job id)
	if err := execSql(db, `
		ALTER TABLE conversation
		ADD COLUMN task_id VARCHAR(255) NULL;
	`); err != nil {
		return err
	}

	return nil
}

func doSqliteMigration(db *sql.DB) error {
	// v3.10 added sqlite support, no migration needed before this version

	// v4 migration
	// add new field `task_id` in `conversation` table to store task id (e.g., video job id)
	if err := execSql(db, `
		ALTER TABLE conversation
		ADD COLUMN task_id VARCHAR(255) NULL;
	`); err != nil {
		return err
	}

	return nil
}
