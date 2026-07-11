package manager

import (
	"chat/admin"
	"chat/channel"
	"chat/globals"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ModelAPI(c *gin.Context) {
	c.JSON(http.StatusOK, globals.V1ListModels)
}

func MarketAPI(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	expandedModels := admin.MarketModelList{}
	for _, model := range admin.MarketInstance.GetAllModels().ActiveChannelModels(channel.ConduitInstance, true) {
		if model.ChannelId == nil {
			expandedModels = append(expandedModels, model)
			continue
		}
		ch := channel.ConduitInstance.GetSequence().GetChannelById(*model.ChannelId)
		model.ChannelName = ch.GetName()
		model.Channels = nil
		expandedModels = append(expandedModels, model)
	}

	c.JSON(http.StatusOK, expandedModels)
}

func ChargeAPI(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.JSON(http.StatusOK, channel.ChargeInstance.ListActiveRules(channel.ConduitInstance))
}

func PlanAPI(c *gin.Context) {
	// 订阅功能已移除，返回空列表
	c.JSON(http.StatusOK, []interface{}{})
}

func sendErrorResponse(c *gin.Context, err error, types ...string) {
	var errType string
	if len(types) > 0 {
		errType = types[0]
	} else {
		errType = "chatnio_api_error"
	}

	c.JSON(http.StatusServiceUnavailable, RelayErrorResponse{
		Error: TranshipmentError{
			Message: err.Error(),
			Type:    errType,
		},
	})
}

func abortWithErrorResponse(c *gin.Context, err error, types ...string) {
	sendErrorResponse(c, err, types...)
	c.Abort()
}
