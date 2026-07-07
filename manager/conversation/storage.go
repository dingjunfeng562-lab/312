package conversation

import (
	"chat/auth"
	"chat/globals"
	"chat/utils"
	"database/sql"
	"fmt"
	"time"
)

const (
	defaultHistoryRetentionDays = 7
	videoHistoryRetentionDays   = 3
)

func getConversationRetentionDays(model string) int {
	if globals.GetModelType(model) == globals.ModelTypeVideo {
		return videoHistoryRetentionDays
	}

	return defaultHistoryRetentionDays
}

func GetConversationRetentionDays(model string) int {
	return getConversationRetentionDays(model)
}

func parseDatabaseTime(value interface{}) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case []byte:
		return parseDatabaseTimeString(string(v))
	case string:
		return parseDatabaseTimeString(v)
	default:
		return time.Time{}, false
	}
}

func parseDatabaseTimeString(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func IsConversationExpired(model string, updatedAt time.Time) bool {
	if updatedAt.IsZero() {
		return false
	}

	retention := time.Duration(getConversationRetentionDays(model)) * 24 * time.Hour
	return time.Now().After(updatedAt.Add(retention))
}

func IsConversationExpiredValue(model string, updatedAt interface{}) bool {
	t, ok := parseDatabaseTime(updatedAt)
	return ok && IsConversationExpired(model, t)
}

func normalizeModelValue(value interface{}) string {
	switch v := value.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	default:
		return globals.GPT3Turbo
	}
}

func deleteConversationByID(db *sql.DB, userId int64, conversationId int64) error {
	if _, err := globals.ExecDb(db, "DELETE FROM sharing WHERE user_id = ? AND conversation_id = ?", userId, conversationId); err != nil {
		return err
	}

	_, err := globals.ExecDb(db, "DELETE FROM conversation WHERE user_id = ? AND conversation_id = ?", userId, conversationId)
	return err
}

func DeleteExpiredConversations(db *sql.DB, userId int64) {
	rows, err := globals.QueryDb(db, `
		SELECT conversation_id, model, updated_at
		FROM conversation
		WHERE user_id = ?
	`, userId)
	if err != nil {
		return
	}

	var expired []int64
	for rows.Next() {
		var (
			conversationId int64
			modelValue     interface{}
			updatedAt      interface{}
		)
		if err := rows.Scan(&conversationId, &modelValue, &updatedAt); err != nil {
			continue
		}

		if IsConversationExpiredValue(normalizeModelValue(modelValue), updatedAt) {
			expired = append(expired, conversationId)
		}
	}

	if err := rows.Close(); err != nil {
		globals.Warn(err)
	}

	for _, conversationId := range expired {
		if err := deleteConversationByID(db, userId, conversationId); err != nil {
			globals.Warn(fmt.Sprintf("failed to delete expired conversation: %s", err.Error()))
		}
	}
}

func (c *Conversation) SaveConversation(db *sql.DB) bool {
	if c.UserID == -1 {
		// anonymous request
		return true
	}

	data := utils.ToJson(c.GetMessage())
	query := `
		INSERT INTO conversation (user_id, conversation_id, conversation_name, data, model, task_id) VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE conversation_name = VALUES(conversation_name), data = VALUES(data), model = VALUES(model), task_id = VALUES(task_id), updated_at = CURRENT_TIMESTAMP
	`
	if globals.SqliteEngine {
		query = `
			INSERT INTO conversation (user_id, conversation_id, conversation_name, data, model, task_id) VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(user_id, conversation_id) DO UPDATE SET conversation_name = excluded.conversation_name, data = excluded.data, model = excluded.model, task_id = excluded.task_id, updated_at = CURRENT_TIMESTAMP
		`
	}

	stmt, err := globals.PrepareDb(db, query)
	if err != nil {
		globals.Info(fmt.Sprintf("prepare error during save conversation: %s", err.Error()))
		return false
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			globals.Warn(err)
		}
	}(stmt)

	var taskID sql.NullString
	if c.TaskID != "" {
		taskID = sql.NullString{String: c.TaskID, Valid: true}
	}

	_, err = stmt.Exec(c.UserID, c.Id, c.Name, data, c.Model, taskID)
	if err != nil {
		globals.Info(fmt.Sprintf("execute error during save conversation: %s", err.Error()))
		return false
	}
	return true
}
func GetConversationLengthByUserID(db *sql.DB, userId int64) int64 {
	var length int64
	err := globals.QueryRowDb(db, "SELECT MAX(conversation_id) FROM conversation WHERE user_id = ?", userId).Scan(&length)
	if err != nil || length < 0 {
		return 0
	}
	return length
}

func LoadConversation(db *sql.DB, userId int64, conversationId int64) *Conversation {
	conversation := Conversation{
		UserID: userId,
		Id:     conversationId,
	}

	var (
		data      string
		model     interface{}
		taskID    sql.NullString
		updatedAt interface{}
	)
	err := globals.QueryRowDb(db, `
		SELECT conversation_name, model, data, task_id, updated_at FROM conversation
		WHERE user_id = ? AND conversation_id = ?
		`, userId, conversationId).Scan(&conversation.Name, &model, &data, &taskID, &updatedAt)
	if err != nil {
		return nil
	}

	conversation.Model = normalizeModelValue(model)
	if taskID.Valid {
		conversation.TaskID = taskID.String
	}
	if t, ok := parseDatabaseTime(updatedAt); ok {
		if IsConversationExpired(conversation.Model, t) {
			if err := deleteConversationByID(db, userId, conversationId); err != nil {
				globals.Warn(fmt.Sprintf("failed to delete expired conversation: %s", err.Error()))
			}
			return nil
		}
		conversation.UpdatedAt = utils.ConvertSqlTime(t)
	}

	conversation.Message, err = utils.Unmarshal[[]globals.Message]([]byte(data))
	if err != nil {
		return nil
	}

	return &conversation
}

func LoadConversationList(db *sql.DB, userId int64) []Conversation {
	DeleteExpiredConversations(db, userId)

	var conversationList []Conversation
	rows, err := globals.QueryDb(db, `
			SELECT conversation_id, conversation_name, model, updated_at FROM conversation WHERE user_id = ?
			ORDER BY updated_at DESC, conversation_id DESC LIMIT 100
	`, userId)
	if err != nil {
		return conversationList
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	for rows.Next() {
		var (
			conversation Conversation
			modelValue   interface{}
			updatedAt    interface{}
		)
		err := rows.Scan(&conversation.Id, &conversation.Name, &modelValue, &updatedAt)
		if err != nil {
			continue
		}
		conversation.Model = normalizeModelValue(modelValue)
		if t, ok := parseDatabaseTime(updatedAt); ok {
			if IsConversationExpired(conversation.Model, t) {
				continue
			}
			conversation.UpdatedAt = utils.ConvertSqlTime(t)
		}
		conversationList = append(conversationList, conversation)
	}

	return conversationList
}

func (c *Conversation) DeleteConversation(db *sql.DB) bool {
	if err := deleteConversationByID(db, c.UserID, c.Id); err != nil {
		return false
	}
	return true
}

func (c *Conversation) RenameConversation(db *sql.DB, name string) bool {
	_, err := globals.ExecDb(db, "UPDATE conversation SET conversation_name = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND conversation_id = ?", name, c.UserID, c.Id)
	if err != nil {
		return false
	}
	return true
}

func DeleteAllConversations(db *sql.DB, user auth.User) error {
	_, err := globals.ExecDb(db, "DELETE FROM conversation WHERE user_id = ?", user.GetID(db))
	return err
}
