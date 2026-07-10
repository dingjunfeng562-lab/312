package admin

import (
	"chat/globals"
	"chat/utils"
	"database/sql"
	"time"
)

type InvitationUsageDetail struct {
	Code        string  `json:"code"`
	Quota       float32 `json:"quota"`
	Type        string  `json:"type"`
	Used        bool    `json:"used"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   *string `json:"expires_at"`
	IsExpired   bool    `json:"is_expired"`
	CreatorName string  `json:"creator_name"`
	UsedByUser  string  `json:"used_by_user"`
	UsedAt      *string `json:"used_at"`
	UsedIP      string  `json:"used_ip"`
	Notes       string  `json:"notes"`
}

func GetInvitationUsageDetail(db *sql.DB, code string) (*InvitationUsageDetail, error) {
	row := globals.QueryRowDb(db, `
		SELECT
			i.code, i.quota, i.type, i.used, i.created_at,
			i.expires_at, i.creator_name, i.used_at, i.used_ip, i.notes,
			COALESCE(a.username, '-') AS used_by_user
		FROM invitation i
		LEFT JOIN auth a ON a.id = i.used_id
		WHERE i.code = ?
	`, code)

	var detail InvitationUsageDetail
	var createdAt []uint8
	var expiresAt sql.NullTime
	var creatorName sql.NullString
	var usedAt sql.NullTime
	var usedIP sql.NullString
	var notes sql.NullString

	err := row.Scan(
		&detail.Code,
		&detail.Quota,
		&detail.Type,
		&detail.Used,
		&createdAt,
		&expiresAt,
		&creatorName,
		&usedAt,
		&usedIP,
		&notes,
		&detail.UsedByUser,
	)
	if err != nil {
		return nil, err
	}

	if stamp := utils.ConvertTime(createdAt); stamp != nil {
		detail.CreatedAt = stamp.Format("2006-01-02 15:04:05")
	}

	if expiresAt.Valid {
		formatted := expiresAt.Time.Format("2006-01-02 15:04:05")
		detail.ExpiresAt = &formatted
		detail.IsExpired = time.Now().After(expiresAt.Time)
	}

	if creatorName.Valid {
		detail.CreatorName = creatorName.String
	} else {
		detail.CreatorName = "system"
	}

	if usedAt.Valid {
		formatted := usedAt.Time.Format("2006-01-02 15:04:05")
		detail.UsedAt = &formatted
	}

	if usedIP.Valid {
		detail.UsedIP = usedIP.String
	}

	if notes.Valid {
		detail.Notes = notes.String
	}

	return &detail, nil
}

type ExpiredInvitationData struct {
	Code        string `json:"code"`
	ExpiresAt   string `json:"expires_at"`
	CreatorName string `json:"creator_name"`
	Used        bool   `json:"used"`
}

func GetExpiredInvitations(db *sql.DB) ([]ExpiredInvitationData, error) {
	rows, err := globals.QueryDb(db, `
		SELECT code, expires_at, creator_name, used
		FROM invitation
		WHERE expires_at IS NOT NULL AND expires_at < ?
		ORDER BY expires_at DESC
	`, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ExpiredInvitationData
	for rows.Next() {
		var item ExpiredInvitationData
		var expiresAt []uint8
		var creatorName sql.NullString
		if err := rows.Scan(&item.Code, &expiresAt, &creatorName, &item.Used); err != nil {
			return nil, err
		}
		if stamp := utils.ConvertTime(expiresAt); stamp != nil {
			item.ExpiresAt = stamp.Format("2006-01-02 15:04:05")
		}
		if creatorName.Valid {
			item.CreatorName = creatorName.String
		} else {
			item.CreatorName = "system"
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
