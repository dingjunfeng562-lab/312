package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func sha2Encrypt(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func main() {
	// 打开数据库连接
	db, err := sql.Open("sqlite3", "./db/chatnio.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	username := "baishuwan"
	newPassword := "baishuwan0825"

	// 加密密码
	hashedPassword := sha2Encrypt(newPassword)

	// 检查用户是否存在
	var userID int64
	var currentUsername string
	err = db.QueryRow("SELECT id, username FROM auth WHERE username = ?", username).Scan(&userID, &currentUsername)
	if err == sql.ErrNoRows {
		fmt.Printf("用户 '%s' 不存在\n", username)
		return
	} else if err != nil {
		log.Fatal("查询用户失败:", err)
	}

	fmt.Printf("找到用户: %s (ID: %d)\n", currentUsername, userID)

	// 更新密码
	result, err := db.Exec("UPDATE auth SET password = ? WHERE username = ?", hashedPassword, username)
	if err != nil {
		log.Fatal("更新密码失败:", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal("获取影响行数失败:", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("✓ 成功更新用户 '%s' 的密码为: %s\n", username, newPassword)
		fmt.Printf("  密码哈希: %s\n", hashedPassword)
	} else {
		fmt.Println("× 密码更新失败，没有行被修改")
	}
}
