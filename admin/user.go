package admin

import (
	"chat/globals"
	"chat/utils"
	"context"
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// AuthLike is to solve the problem of import cycle
type AuthLike struct {
	ID int64 `json:"id"`
}

func (a *AuthLike) GetID(_ *sql.DB) int64 {
	return a.ID
}

func (a *AuthLike) HitID() int64 {
	return a.ID
}

func getUsersForm(db *sql.DB, page int64, search string, filter UserFilter) PaginationForm {
	if page < 0 {
		page = 0
	}
	var users []interface{}
	var total int64

	where := []string{"(auth.username LIKE ? OR COALESCE(auth.email, '') LIKE ?)"}
	args := []interface{}{"%" + search + "%", "%" + search + "%"}
	if filter.Admin == "yes" {
		where = append(where, "auth.is_admin = TRUE")
	} else if filter.Admin == "no" {
		where = append(where, "auth.is_admin = FALSE")
	}
	if filter.Ban == "yes" {
		where = append(where, "auth.is_banned = TRUE")
	} else if filter.Ban == "no" {
		where = append(where, "auth.is_banned = FALSE")
	}
	if filter.Plan == "yes" {
		where = append(where, "COALESCE(subscription.level, 0) > 0 AND subscription.expired_at > CURRENT_TIMESTAMP")
	} else if filter.Plan == "no" {
		where = append(where, "(COALESCE(subscription.level, 0) = 0 OR subscription.expired_at IS NULL OR subscription.expired_at <= CURRENT_TIMESTAMP)")
	}

	fromWhere := ` FROM auth
		LEFT JOIN quota ON quota.user_id = auth.id
		LEFT JOIN subscription ON subscription.user_id = auth.id
		WHERE ` + strings.Join(where, " AND ")
	if err := globals.QueryRowDb(db, "SELECT COUNT(*)"+fromWhere, args...).Scan(&total); err != nil {
		return PaginationForm{
			Status:  false,
			Message: err.Error(),
		}
	}

	sorts := map[string]string{
		"id-asc": "auth.id ASC", "id-desc": "auth.id DESC",
		"quota-asc": "COALESCE(quota.quota, 0) ASC", "quota-desc": "COALESCE(quota.quota, 0) DESC",
		"used-quota-asc": "COALESCE(quota.used, 0) ASC", "used-quota-desc": "COALESCE(quota.used, 0) DESC",
		"plan-asc": "COALESCE(subscription.level, 0) ASC", "plan-desc": "COALESCE(subscription.level, 0) DESC",
	}
	order, ok := sorts[filter.Sort]
	if !ok {
		order = sorts["id-asc"]
	}
	queryArgs := append(append([]interface{}{}, args...), pagination, page*pagination)
	rows, err := globals.QueryDb(db, `
		SELECT 
		    auth.id, auth.username, auth.email, auth.is_admin,
		    quota.quota, quota.used,
		    subscription.expired_at, subscription.total_month, subscription.enterprise, subscription.level,
		    auth.is_banned,
		    COALESCE((
		        SELECT invitation.code
		        FROM invitation
		        WHERE invitation.used_id = auth.id AND invitation.used = TRUE
		        ORDER BY invitation.updated_at DESC, invitation.id DESC
		        LIMIT 1
		    ), '') as invitation_code
	`+fromWhere+` ORDER BY `+order+` LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return PaginationForm{
			Status:  false,
			Message: err.Error(),
		}
	}

	for rows.Next() {
		var user UserData
		var (
			email             sql.NullString
			expired           []uint8
			quota             sql.NullFloat64
			usedQuota         sql.NullFloat64
			totalMonth        sql.NullInt64
			isEnterprise      sql.NullBool
			subscriptionLevel sql.NullInt64
			isBanned          sql.NullBool
		)
		if err := rows.Scan(&user.Id, &user.Username, &email, &user.IsAdmin, &quota, &usedQuota, &expired, &totalMonth, &isEnterprise, &subscriptionLevel, &isBanned, &user.InvitationCode); err != nil {
			return PaginationForm{
				Status:  false,
				Message: err.Error(),
			}
		}
		if email.Valid {
			user.Email = email.String
		}
		if quota.Valid {
			user.Quota = float32(quota.Float64)
		}
		if usedQuota.Valid {
			user.UsedQuota = float32(usedQuota.Float64)
		}
		if totalMonth.Valid {
			user.TotalMonth = totalMonth.Int64
		}
		if subscriptionLevel.Valid {
			user.Level = int(subscriptionLevel.Int64)
		}
		stamp := utils.ConvertTime(expired)
		if stamp != nil {
			user.IsSubscribed = stamp.After(time.Now())
			user.ExpiredAt = stamp.Format("2006-01-02 15:04:05")
		}
		user.Enterprise = isEnterprise.Valid && isEnterprise.Bool
		user.IsBanned = isBanned.Valid && isBanned.Bool

		users = append(users, user)
	}

	return PaginationForm{
		Status: true,
		Total:  int(math.Ceil(float64(total) / float64(pagination))),
		Data:   users,
	}
}

func userExists(db *sql.DB, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid user id")
	}
	var count int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE id = ?", id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func requireUser(db *sql.DB, id int64) error {
	exists, err := userExists(db, id)
	if err != nil {
		return fmt.Errorf("failed to query user: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}
	return nil
}

func configuredRootUsername() string {
	username := strings.TrimSpace(viper.GetString("root.username"))
	if username == "" {
		return "baishuwan"
	}
	return username
}

func isRootUser(db *sql.DB, id int64) (bool, error) {
	var username string
	if err := globals.QueryRowDb(db, "SELECT username FROM auth WHERE id = ?", id).Scan(&username); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("user not found")
		}
		return false, err
	}
	return username == configuredRootUsername(), nil
}

func validateUsername(username string) bool {
	username = strings.TrimSpace(username)
	return len(username) >= 2 && len(username) <= 24
}

func validateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 1 || len(email) > 255 {
		return false
	}
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

// clearUserCache clears all cache keys starting with nio:user:
func clearUserCache(cache *redis.Client) error {
	if cache == nil {
		return nil
	}
	ctx := context.Background()
	iter := cache.Scan(ctx, 0, "nio:user:*", 100).Iterator()
	for iter.Next(ctx) {
		if err := cache.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete cache key %s: %v", iter.Val(), err)
		}
	}
	return iter.Err()
}

func passwordMigration(db *sql.DB, cache *redis.Client, id int64, password string) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	password = strings.TrimSpace(password)
	if len(password) < 6 || len(password) > 36 {
		return fmt.Errorf("password length must be between 6 and 36")
	}
	hash_passwd := utils.Sha2Encrypt(password)

	// Update password in database
	result, err := globals.ExecDb(db, `
		UPDATE auth SET password = ? WHERE id = ?
	`, hash_passwd, id)

	if err != nil {
		return err
	}

	if rows, rowsErr := result.RowsAffected(); rowsErr == nil && rows == 0 {
		return fmt.Errorf("user not found")
	}

	if err := clearUserCache(cache); err != nil {
		return fmt.Errorf("failed to clear user cache: %v", err)
	}

	return nil
}

func emailMigration(db *sql.DB, cache *redis.Client, id int64, email string) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	email = strings.TrimSpace(email)
	if !validateEmail(email) {
		return fmt.Errorf("invalid email format")
	}
	var duplicate int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE email = ? AND id <> ?", email, id).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate > 0 {
		return fmt.Errorf("email is already in use")
	}
	_, err := globals.ExecDb(db, `
		UPDATE auth SET email = ? WHERE id = ?
	`, email, id)
	if err != nil {
		return err
	}
	return clearUserCache(cache)
}

func setAdmin(db *sql.DB, cache *redis.Client, id int64, isAdmin bool) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	root, err := isRootUser(db, id)
	if err != nil {
		return err
	}
	if root && !isAdmin {
		return fmt.Errorf("configured root administrator cannot be demoted")
	}
	if !isAdmin {
		var current bool
		var admins int
		if err := globals.QueryRowDb(db, "SELECT is_admin FROM auth WHERE id = ?", id).Scan(&current); err != nil {
			return err
		}
		if current {
			if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE is_admin = TRUE").Scan(&admins); err != nil {
				return err
			}
			if admins <= 1 {
				return fmt.Errorf("cannot remove the final administrator")
			}
		}
	}
	_, err = globals.ExecDb(db, `
		UPDATE auth SET is_admin = ? WHERE id = ?
	`, isAdmin, id)
	if err != nil {
		return err
	}
	return clearUserCache(cache)
}

func banUser(db *sql.DB, cache *redis.Client, id int64, isBanned bool) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	root, err := isRootUser(db, id)
	if err != nil {
		return err
	}
	if root && isBanned {
		return fmt.Errorf("configured root administrator cannot be banned")
	}
	_, err = globals.ExecDb(db, `
		UPDATE auth SET is_banned = ? WHERE id = ?
	`, isBanned, id)
	if err != nil {
		return err
	}
	return clearUserCache(cache)
}

func quotaMigration(db *sql.DB, id int64, quota float32, override bool) (float32, error) {
	if id <= 0 {
		return 0, fmt.Errorf("invalid user id")
	}
	if math.IsNaN(float64(quota)) || math.IsInf(float64(quota), 0) {
		return 0, fmt.Errorf("invalid points value")
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start points transaction: %w", err)
	}
	defer tx.Rollback()

	var userExists int
	if err := tx.QueryRow("SELECT COUNT(*) FROM auth WHERE id = ?", id).Scan(&userExists); err != nil {
		return 0, fmt.Errorf("failed to query user: %w", err)
	}
	if userExists == 0 {
		return 0, fmt.Errorf("user not found")
	}

	var current float32
	err = tx.QueryRow("SELECT quota FROM quota WHERE user_id = ?", id).Scan(&current)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query current points: %w", err)
	}
	if err == sql.ErrNoRows {
		current = 0
	}

	updated := quota
	if !override {
		updated = current + quota
	}
	if math.IsNaN(float64(updated)) || math.IsInf(float64(updated), 0) {
		return 0, fmt.Errorf("invalid resulting points balance")
	}
	if updated < 0 {
		return 0, fmt.Errorf("points balance cannot be negative")
	}

	if err == sql.ErrNoRows {
		if _, err := tx.Exec("INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?)", id, updated, 0.); err != nil {
			return 0, fmt.Errorf("failed to create points balance: %w", err)
		}
	} else if _, err := tx.Exec("UPDATE quota SET quota = ? WHERE user_id = ?", updated, id); err != nil {
		return 0, fmt.Errorf("failed to update points balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit points transaction: %w", err)
	}
	return updated, nil
}

func subscriptionMigration(db *sql.DB, id int64, expired string) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	var count int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM subscription WHERE user_id = ?", id).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err := globals.ExecDb(db, "INSERT INTO subscription (user_id, expired_at, level) VALUES (?, ?, ?)", id, expired, 0)
		return err
	}
	_, err := globals.ExecDb(db, "UPDATE subscription SET expired_at = ? WHERE user_id = ?", expired, id)
	return err
}

func subscriptionLevelMigration(db *sql.DB, id int64, level int64) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	if level < 0 || level > 3 {
		return fmt.Errorf("invalid subscription level")
	}

	var count int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM subscription WHERE user_id = ?", id).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err := globals.ExecDb(db, "INSERT INTO subscription (user_id, level) VALUES (?, ?)", id, level)
		return err
	}
	_, err := globals.ExecDb(db, "UPDATE subscription SET level = ? WHERE user_id = ?", level, id)
	return err
}

func releaseUsage(db *sql.DB, cache *redis.Client, id int64) error {
	if err := requireUser(db, id); err != nil {
		return err
	}

	// 订阅功能已移除，直接返回成功
	return nil
}

func updateUserProfile(db *sql.DB, cache *redis.Client, id int64, profile UserProfileUpdate) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	profile.Username = strings.TrimSpace(profile.Username)
	profile.Email = strings.TrimSpace(profile.Email)
	if !validateUsername(profile.Username) {
		return fmt.Errorf("username length must be between 2 and 24")
	}
	if profile.Email != "" && !validateEmail(profile.Email) {
		return fmt.Errorf("invalid email format")
	}
	if math.IsNaN(float64(profile.UsedQuota)) || math.IsInf(float64(profile.UsedQuota), 0) || profile.UsedQuota < 0 {
		return fmt.Errorf("used points cannot be negative or invalid")
	}
	if profile.TotalMonth < 0 {
		return fmt.Errorf("total subscription months cannot be negative")
	}

	root, err := isRootUser(db, id)
	if err != nil {
		return err
	}
	if root && profile.Username != configuredRootUsername() {
		return fmt.Errorf("configured root username cannot be changed")
	}
	var duplicate int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE username = ? AND id <> ?", profile.Username, id).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate > 0 {
		return fmt.Errorf("username is already in use")
	}
	if profile.Email != "" {
		if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE email = ? AND id <> ?", profile.Email, id).Scan(&duplicate); err != nil {
			return err
		}
		if duplicate > 0 {
			return fmt.Errorf("email is already in use")
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start user update transaction: %w", err)
	}
	defer tx.Rollback()
	var email interface{}
	if profile.Email != "" {
		email = profile.Email
	}
	if _, err := tx.Exec(globals.PreflightSql("UPDATE auth SET username = ?, email = ? WHERE id = ?"), profile.Username, email, id); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	var quotaCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM quota WHERE user_id = ?", id).Scan(&quotaCount); err != nil {
		return fmt.Errorf("failed to query points: %w", err)
	}
	if quotaCount == 0 {
		if _, err := tx.Exec("INSERT INTO quota (user_id, quota, used) VALUES (?, ?, ?)", id, 0, profile.UsedQuota); err != nil {
			return fmt.Errorf("failed to create points record: %w", err)
		}
	} else if _, err := tx.Exec("UPDATE quota SET used = ? WHERE user_id = ?", profile.UsedQuota, id); err != nil {
		return fmt.Errorf("failed to update used points: %w", err)
	}
	var subscriptionCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM subscription WHERE user_id = ?", id).Scan(&subscriptionCount); err != nil {
		return fmt.Errorf("failed to query subscription: %w", err)
	}
	if subscriptionCount == 0 {
		if _, err := tx.Exec("INSERT INTO subscription (user_id, level, total_month, enterprise) VALUES (?, ?, ?, ?)", id, 0, profile.TotalMonth, profile.Enterprise); err != nil {
			return fmt.Errorf("failed to create subscription record: %w", err)
		}
	} else if _, err := tx.Exec("UPDATE subscription SET total_month = ?, enterprise = ? WHERE user_id = ?", profile.TotalMonth, profile.Enterprise, id); err != nil {
		return fmt.Errorf("failed to update subscription details: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit user update: %w", err)
	}
	return clearUserCache(cache)
}

func deleteUser(db *sql.DB, cache *redis.Client, id int64, currentUsername string) error {
	if err := requireUser(db, id); err != nil {
		return err
	}
	var username string
	var admin bool
	if err := globals.QueryRowDb(db, "SELECT username, is_admin FROM auth WHERE id = ?", id).Scan(&username, &admin); err != nil {
		return err
	}
	if username == configuredRootUsername() {
		return fmt.Errorf("configured root administrator cannot be deleted")
	}
	if username == currentUsername {
		return fmt.Errorf("you cannot delete your own account")
	}
	if admin {
		var admins int
		if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE is_admin = TRUE").Scan(&admins); err != nil {
			return err
		}
		if admins <= 1 {
			return fmt.Errorf("cannot delete the final administrator")
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start deletion transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec("UPDATE invitation SET used = FALSE, used_id = NULL WHERE used_id = ?", id); err != nil {
		return fmt.Errorf("failed to release invitation codes: %w", err)
	}
	for _, table := range []string{"broadcast", "sharing", "mask", "conversation", "apikey", "subscription", "quota", "package"} {
		column := "user_id"
		if table == "broadcast" {
			column = "poster_id"
		}
		if _, err := tx.Exec("DELETE FROM "+table+" WHERE "+column+" = ?", id); err != nil {
			return fmt.Errorf("failed to delete user data from %s: %w", table, err)
		}
	}
	result, err := tx.Exec("DELETE FROM auth WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if rows, rowsErr := result.RowsAffected(); rowsErr == nil && rows == 0 {
		return fmt.Errorf("user not found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit user deletion: %w", err)
	}
	return clearUserCache(cache)
}

func UpdateRootPassword(db *sql.DB, cache *redis.Client, password string) error {
	password = strings.TrimSpace(password)
	if len(password) < 6 || len(password) > 36 {
		return fmt.Errorf("password length must be between 6 and 36")
	}

	username := strings.TrimSpace(viper.GetString("root.username"))
	if username == "" {
		username = "baishuwan"
	}

	result, err := globals.ExecDb(db, `
		UPDATE auth SET password = ?, is_admin = ?, is_banned = ? WHERE username = ?
	`, utils.Sha2Encrypt(password), true, false, username)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return fmt.Errorf("configured root user %s not found", username)
	}

	// Clear all user related cache
	if err := clearUserCache(cache); err != nil {
		return fmt.Errorf("failed to clear user cache: %v", err)
	}

	return nil
}

// setUserInvitationCode sets or updates the invitation code for a user
func setUserInvitationCode(db *sql.DB, userId int64, invitationCode string) error {
	invitationCode = strings.TrimSpace(invitationCode)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var userExists int
	if err := tx.QueryRow("SELECT COUNT(*) FROM auth WHERE id = ?", userId).Scan(&userExists); err != nil {
		return fmt.Errorf("failed to query user: %w", err)
	}
	if userExists == 0 {
		return fmt.Errorf("user not found")
	}

	// An empty code explicitly removes every invitation association from the user.
	if invitationCode == "" {
		if _, err := tx.Exec(`
			UPDATE invitation SET used = FALSE, used_id = NULL
			WHERE used_id = ? AND used = TRUE
		`, userId); err != nil {
			return fmt.Errorf("failed to clear invitation code: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit invitation update: %w", err)
		}
		return nil
	}

	var invitationId int64
	var used bool
	var currentUsedId sql.NullInt64
	if err := tx.QueryRow(`
		SELECT id, used, used_id FROM invitation WHERE code = ?
	`, invitationCode).Scan(&invitationId, &used, &currentUsedId); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invitation code not found")
		}
		return fmt.Errorf("failed to query invitation code: %w", err)
	}

	if used && (!currentUsedId.Valid || currentUsedId.Int64 != userId) {
		return fmt.Errorf("invitation code is already used by another user")
	}

	if _, err := tx.Exec(`
		UPDATE invitation SET used = FALSE, used_id = NULL
		WHERE used_id = ? AND id <> ?
	`, userId, invitationId); err != nil {
		return fmt.Errorf("failed to clear existing invitation code: %w", err)
	}

	if _, err := tx.Exec(`
		UPDATE invitation SET used = TRUE, used_id = ?
		WHERE id = ? AND (used = FALSE OR used_id = ?)
	`, userId, invitationId, userId); err != nil {
		return fmt.Errorf("failed to set invitation code: %w", err)
	}

	var assignedUserId sql.NullInt64
	var assigned bool
	if err := tx.QueryRow(`
		SELECT used, used_id FROM invitation WHERE id = ?
	`, invitationId).Scan(&assigned, &assignedUserId); err != nil {
		return fmt.Errorf("failed to verify invitation update: %w", err)
	}
	if !assigned || !assignedUserId.Valid || assignedUserId.Int64 != userId {
		return fmt.Errorf("invitation code is already used by another user")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit invitation update: %w", err)
	}
	return nil
}
