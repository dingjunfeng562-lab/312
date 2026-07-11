package utils

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// GetClientIP 获取客户端真实 IP 地址
// 优先级：X-Real-IP > X-Forwarded-For > RemoteAddr
func GetClientIP(c *gin.Context) string {
	// 尝试从 X-Real-IP 获取
	ip := strings.TrimSpace(c.GetHeader("X-Real-IP"))
	if ip != "" && ip != "unknown" {
		return ip
	}

	// 尝试从 X-Forwarded-For 获取（取第一个 IP）
	forwarded := strings.TrimSpace(c.GetHeader("X-Forwarded-For"))
	if forwarded != "" && forwarded != "unknown" {
		// X-Forwarded-For 可能包含多个 IP，格式：client, proxy1, proxy2
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
			if ip != "" && ip != "unknown" {
				return ip
			}
		}
	}

	// 回退到 RemoteAddr
	ip = c.ClientIP()
	return ip
}
