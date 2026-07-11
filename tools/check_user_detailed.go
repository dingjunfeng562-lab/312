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

	// 查询用户详细信息
	query := `
		SELECT 
			auth.id, 
			auth.username, 
			auth.email, 
			auth.is_admin, 
			auth.is_banned,
			COALESCE(quota.quota, 0) as total_quota,
			COALESCE(quota.used, 0) as used_quota
		FROM auth
		LEFT JOIN quota ON auth.id = quota.user_id
		ORDER BY auth.id
	`

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println("╔════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Ai idea 用户权限详细报告                               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	adminCount := 0
	userCount := 0
	bannedCount := 0

	for rows.Next() {
		var id int
		var username, email string
		var isAdmin, isBanned bool
		var totalQuota, usedQuota float64

		if err := rows.Scan(&id, &username, &email, &isAdmin, &isBanned, &totalQuota, &usedQuota); err != nil {
			panic(err)
		}

		if isAdmin {
			adminCount++
		} else {
			userCount++
		}
		if isBanned {
			bannedCount++
		}

		// 计算剩余配额
		remainingQuota := totalQuota - usedQuota
		usagePercent := 0.0
		if totalQuota > 0 {
			usagePercent = (usedQuota / totalQuota) * 100
		}

		// 权限标识
		role := "👤 普通用户"
		if isAdmin {
			role = "👑 管理员"
		}
		if isBanned {
			role = "🚫 已封禁"
		}

		fmt.Println("─────────────────────────────────────────────────────────────────")
		fmt.Printf("ID:            %d\n", id)
		fmt.Printf("用户名:        %s\n", username)
		fmt.Printf("邮箱:          %s\n", email)
		fmt.Printf("角色:          %s\n", role)
		fmt.Printf("总配额:        %.2f\n", totalQuota)
		fmt.Printf("已使用:        %.2f\n", usedQuota)
		fmt.Printf("剩余配额:      %.2f\n", remainingQuota)
		if totalQuota > 0 {
			fmt.Printf("使用率:        %.2f%%\n", usagePercent)
		}
		fmt.Printf("状态:          %s\n", getStatus(isAdmin, isBanned))
		fmt.Println()
	}

	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println("                            统计信息                              ")
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Printf("👑 管理员数量:     %d\n", adminCount)
	fmt.Printf("👤 普通用户数量:   %d\n", userCount)
	fmt.Printf("🚫 封禁用户数量:   %d\n", bannedCount)
	fmt.Printf("📊 总用户数:       %d\n", adminCount+userCount)
	fmt.Println("═════════════════════════════════════════════════════════════════")
}

func getStatus(isAdmin, isBanned bool) string {
	if isBanned {
		return "🚫 已封禁"
	}
	if isAdmin {
		return "✅ 正常（管理员权限）"
	}
	return "✅ 正常"
}
