//go:build ignore

package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "modernc.org/sqlite"
)

func sha2Encrypt(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func main() {
	db, err := sql.Open("sqlite", "./db/chatnio.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Ensure administrator account is baishuwan
	newPassword := sha2Encrypt("baishuwan0825")

	result, err := db.Exec(`
		UPDATE auth 
		SET username = ?, password = ?, email = ?, is_admin = ?
		WHERE username IN (?, ?) OR is_admin = ?
	`, "baishuwan", newPassword, "baishuwan@example.com", true, "root", "baishuwan", true)

	if err != nil {
		panic(err)
	}

	affected, _ := result.RowsAffected()
	fmt.Printf("✓ Successfully updated %d user(s)\n", affected)
	fmt.Println("✓ Admin username changed: root → baishuwan")
	fmt.Println("✓ Password reset to: baishuwan0825")
	fmt.Printf("✓ Password hash: %s\n", newPassword)

	// Verify the update
	var username, password string
	err = db.QueryRow("SELECT username, password FROM auth WHERE username = ?", "baishuwan").Scan(&username, &password)
	if err != nil {
		fmt.Println("✗ Verification failed:", err)
		return
	}

	fmt.Println("\n=== Verification ===")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password Hash: %s\n", password)
	fmt.Println("\n✓ Admin account is now: baishuwan / baishuwan0825")
}
