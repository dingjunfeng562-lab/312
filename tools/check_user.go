//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "./db/chatnio.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, username, password, email, is_admin, is_banned FROM auth LIMIT 10")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println("=== Users in Database ===")
	for rows.Next() {
		var id int
		var username, password, email string
		var isAdmin, isBanned bool
		if err := rows.Scan(&id, &username, &password, &email, &isAdmin, &isBanned); err != nil {
			panic(err)
		}
		fmt.Printf("ID: %d\nUsername: %s\nPassword Hash: %s\nEmail: %s\nIs Admin: %t\nIs Banned: %t\n\n",
			id, username, password, email, isAdmin, isBanned)
	}
}
