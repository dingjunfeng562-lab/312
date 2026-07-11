package admin

import (
	"chat/admin/analysis"
	"chat/channel"
	"chat/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GenerateInvitationForm struct {
	Type   string  `json:"type"`
	Quota  float32 `json:"quota"`
	Number int     `json:"number"`
}

type DeleteInvitationForm struct {
	Code string `json:"code"`
}

type GenerateRedeemForm struct {
	Quota  float32 `json:"quota"`
	Number int     `json:"number"`
}

type PasswordMigrationForm struct {
	Id       int64  `json:"id"`
	Password string `json:"password"`
}

type EmailMigrationForm struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
}

type UserProfileForm struct {
	Id         int64    `json:"id" binding:"required"`
	Username   string   `json:"username" binding:"required"`
	Email      string   `json:"email"`
	UsedQuota  *float32 `json:"used_quota" binding:"required"`
	TotalMonth *int64   `json:"total_month" binding:"required"`
	Enterprise *bool    `json:"enterprise" binding:"required"`
}

type DeleteUserForm struct {
	Id int64 `json:"id" binding:"required"`
}

type SetAdminForm struct {
	Id    int64 `json:"id"`
	Admin bool  `json:"admin"`
}

type BanForm struct {
	Id  int64 `json:"id"`
	Ban bool  `json:"ban"`
}

type QuotaOperationForm struct {
	Id       int64    `json:"id" binding:"required"`
	Quota    *float32 `json:"quota" binding:"required"`
	Override bool     `json:"override"`
}

type SubscriptionOperationForm struct {
	Id      int64  `json:"id" binding:"required"`
	Expired string `json:"expired" binding:"required"`
}

type SubscriptionLevelForm struct {
	Id    int64  `json:"id" binding:"required"`
	Level *int64 `json:"level" binding:"required"`
}

type ReleaseUsageForm struct {
	Id int64 `json:"id" binding:"required"`
}

type UpdateRootPasswordForm struct {
	Password string `json:"password" binding:"required"`
}

type SetInvitationCodeForm struct {
	Id             int64  `json:"id" binding:"required"`
	InvitationCode string `json:"invitation_code"`
}

func UpdateMarketAPI(c *gin.Context) {
	var form MarketModelList
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	err := MarketInstance.SetActiveModels(form, channel.ConduitInstance)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func GetMarketAPI(c *gin.Context) {
	expand := c.Query("expand") == "true" // 新增：是否展开为渠道+模型组合

	if !expand {
		models := MarketInstance.GetAllModels().ActiveChannelModels(channel.ConduitInstance, false)
		for index := range models {
			models[index].Channels = nil
			if models[index].ChannelId != nil {
				if instance := channel.ConduitInstance.GetSequence().GetChannelById(*models[index].ChannelId); instance != nil {
					models[index].ChannelName = instance.GetName()
					models[index].Channels = []string{instance.GetName()}
				}
			} else {
				for _, instance := range channel.ConduitInstance.GetActiveSequence() {
					if instance.IsHit(models[index].Id) {
						models[index].Channels = append(models[index].Channels, instance.GetName())
					}
				}
			}
		}
		c.JSON(http.StatusOK, models)
		return
	}

	// 新逻辑：展开为渠道+模型组合（用于前台展示）
	expandedModels := MarketModelList{}
	configModels := MarketInstance.GetAllModels()

	// 遍历所有激活的渠道
	for _, ch := range channel.ConduitInstance.GetActiveSequence() {
		// 遍历渠道支持的模型
		for _, modelId := range ch.GetModels() {
			// 查找该模型的市场配置
			var baseModel *MarketModel
			for i := range configModels {
				if configModels[i].Id == modelId {
					baseModel = &configModels[i]
					break
				}
			}

			// 如果没有配置，跳过（不在市场中显示）
			if baseModel == nil {
				continue
			}

			// 检查是否已经有针对该渠道的特定配置
			hasSpecificConfig := false
			for i := range configModels {
				if configModels[i].Id == modelId &&
					configModels[i].ChannelId != nil &&
					*configModels[i].ChannelId == ch.Id {
					// 使用特定配置
					model := configModels[i]
					model.ChannelName = ch.GetName()
					model.Channels = nil
					expandedModels = append(expandedModels, model)
					hasSpecificConfig = true
					break
				}
			}

			// 如果没有特定配置，使用基础配置
			if !hasSpecificConfig {
				model := *baseModel
				model.ChannelId = &ch.Id
				model.ChannelName = ch.GetName()
				model.Channels = nil
				expandedModels = append(expandedModels, model)
			}
		}
	}

	c.JSON(http.StatusOK, expandedModels)
}

func InfoAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	c.JSON(http.StatusOK, InfoForm{
		OnlineChats:       utils.GetConns(),
		SubscriptionCount: analysis.GetSubscriptionUsers(db),
		BillingToday:      analysis.GetBillingToday(cache),
		BillingMonth:      analysis.GetBillingMonth(cache),
		BillingYesterday:  analysis.GetBillingYesterday(cache),
		BillingLastMonth:  analysis.GetBillingLastMonth(cache),
	})
}

func ModelAnalysisAPI(c *gin.Context) {
	cache := utils.GetCacheFromContext(c)
	c.JSON(http.StatusOK, analysis.GetSortedModelData(cache))
}

func RequestAnalysisAPI(c *gin.Context) {
	cache := utils.GetCacheFromContext(c)
	c.JSON(http.StatusOK, analysis.GetRequestData(cache))
}

func BillingAnalysisAPI(c *gin.Context) {
	cache := utils.GetCacheFromContext(c)
	c.JSON(http.StatusOK, analysis.GetBillingData(cache))
}

func ErrorAnalysisAPI(c *gin.Context) {
	cache := utils.GetCacheFromContext(c)
	c.JSON(http.StatusOK, analysis.GetErrorData(cache))
}

func UserTypeAnalysisAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	if form, err := analysis.GetUserTypeData(db); err != nil {
		c.JSON(http.StatusOK, &analysis.UserTypeForm{})
	} else {
		c.JSON(http.StatusOK, form)
	}
}

func RedeemListAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	page, _ := strconv.Atoi(c.Query("page"))
	c.JSON(http.StatusOK, GetRedeemData(db, int64(page)))
}

func DeleteRedeemAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	var form DeleteInvitationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	err := DeleteRedeemCode(db, form.Code)
	c.JSON(http.StatusOK, gin.H{
		"status": err == nil,
		"error":  err,
	})
}

func InvitationPaginationAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	page, _ := strconv.Atoi(c.Query("page"))
	c.JSON(http.StatusOK, GetInvitationPagination(db, int64(page)))
}

func DeleteInvitationAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	var form DeleteInvitationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	err := DeleteInvitationCode(db, form.Code)
	c.JSON(http.StatusOK, gin.H{
		"status": err == nil,
		"error":  err,
	})
}
func GenerateInvitationAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	var form GenerateInvitationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GenerateInvitations(db, form.Number, form.Quota, form.Type))
}

func GenerateRedeemAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	var form GenerateRedeemForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GenerateRedeemCodes(db, form.Number, form.Quota))
}

func UserPaginationAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	page, _ := strconv.Atoi(c.Query("page"))
	search := strings.TrimSpace(c.Query("search"))
	filter := UserFilter{
		Plan: strings.TrimSpace(c.Query("plan")), Admin: strings.TrimSpace(c.Query("admin")),
		Ban: strings.TrimSpace(c.Query("ban")), Sort: strings.TrimSpace(c.Query("sort")),
	}
	c.JSON(http.StatusOK, getUsersForm(db, int64(page), search, filter))
}

func UpdateUserProfileAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)
	var form UserProfileForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}
	err := updateUserProfile(db, cache, form.Id, UserProfileUpdate{
		Username: form.Username, Email: form.Email, UsedQuota: *form.UsedQuota,
		TotalMonth: *form.TotalMonth, Enterprise: *form.Enterprise,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func UpdatePasswordAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form PasswordMigrationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	err := passwordMigration(db, cache, form.Id, form.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func UpdateEmailAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form EmailMigrationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := emailMigration(db, cache, form.Id, form.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func SetAdminAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form SetAdminForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := setAdmin(db, cache, form.Id, form.Admin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func BanAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form BanForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := banUser(db, cache, form.Id, form.Ban)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func UserQuotaAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form QuotaOperationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	quota, err := quotaMigration(db, form.Id, *form.Quota, form.Override)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	if err := clearUserCache(cache); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"quota":  quota,
	})
}

func UserSubscriptionAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form SubscriptionOperationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	// convert to time
	if _, err := time.Parse("2006-01-02 15:04:05", form.Expired); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	if err := subscriptionMigration(db, form.Id, form.Expired); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	if err := clearUserCache(cache); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func SubscriptionLevelAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form SubscriptionLevelForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := subscriptionLevelMigration(db, form.Id, *form.Level)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	if err := clearUserCache(cache); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func ReleaseUsageAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	var form ReleaseUsageForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := releaseUsage(db, cache, form.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func UpdateRootPasswordAPI(c *gin.Context) {
	var form UpdateRootPasswordForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)
	err := UpdateRootPassword(db, cache, form.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func ListLoggerAPI(c *gin.Context) {
	c.JSON(http.StatusOK, ListLogs())
}

func DownloadLoggerAPI(c *gin.Context) {
	path := c.Query("path")
	getBlobFile(c, path)
}

func DeleteLoggerAPI(c *gin.Context) {
	path := c.Query("path")
	if err := deleteLogFile(path); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func ConsoleLoggerAPI(c *gin.Context) {
	n := utils.ParseInt(c.Query("n"))

	content := getLatestLogs(n)

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"content": content,
	})
}

func SetInvitationCodeAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	var form SetInvitationCodeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := setUserInvitationCode(db, form.Id, form.InvitationCode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

func DeleteUserAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)
	var form DeleteUserForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}
	if err := deleteUser(db, cache, form.Id, c.GetString("user")); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}
