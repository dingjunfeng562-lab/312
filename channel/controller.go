package channel

import (
	"chat/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SyncChargeForm struct {
	Overwrite bool           `json:"overwrite"`
	Data      ChargeSequence `json:"data"`
}

func GetInfo(c *gin.Context) {
	c.JSON(http.StatusOK, SystemInstance.AsInfo())
}

func AttachmentService(c *gin.Context) {
	// /attachments/:hash -> ~/storage/attachments/:hash
	hash := c.Param("hash")
	c.File(fmt.Sprintf("storage/attachments/%s", hash))
}

func DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.DeleteChannel(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func ActivateChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.ActivateChannel(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func DeactivateChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.DeactivateChannel(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetChannelList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   ConduitInstance.Sequence,
	})
}

func GetChannel(c *gin.Context) {
	id := c.Param("id")
	channel := ConduitInstance.Sequence.GetChannelById(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": channel != nil,
		"data":   channel,
	})
}

func CreateChannel(c *gin.Context) {
	var channel Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	channel.Id = ConduitInstance.GetMaxId() + 1
	state := prepareChannel(&channel)
	var charge ChargeSequence
	if state == nil {
		charge, state = syncChannelCharge(&channel)
	}
	if state == nil {
		state = ConduitInstance.CreateChannel(&channel)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":        state == nil,
		"error":         utils.GetError(state),
		"synced_prices": len(charge),
	})
}

func UpdateChannel(c *gin.Context) {
	var channel Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	id := c.Param("id")
	channel.Id = utils.ParseInt(id)

	state := prepareChannel(&channel)
	var charge ChargeSequence
	if state == nil {
		charge, state = syncChannelCharge(&channel)
	}
	if state == nil {
		state = ConduitInstance.UpdateChannel(channel.Id, &channel)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":        state == nil,
		"error":         utils.GetError(state),
		"synced_prices": len(charge),
	})
}

func prepareChannel(channel *Channel) error {
	if !channel.AutoModels {
		return nil
	}
	models, err := channel.FetchModels()
	if err != nil {
		return err
	}
	channel.Models = models
	return nil
}

func syncChannelCharge(channel *Channel) (ChargeSequence, error) {
	if !channel.AutoPrice {
		return nil, nil
	}

	charge, err := channel.FetchCharge()
	if err != nil {
		return nil, err
	}
	if err := ChargeInstance.SyncRules(charge, channel.PriceOverwrite); err != nil {
		return nil, err
	}
	return charge, nil
}

func SetCharge(c *gin.Context) {
	var charge Charge
	if err := c.ShouldBindJSON(&charge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := ChargeInstance.SetRule(charge)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetChargeList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   ChargeInstance.ListActiveRules(ConduitInstance),
	})
}

func FetchChannelCharge(c *gin.Context) {
	instance := ConduitInstance.Sequence.GetChannelById(utils.ParseInt(c.Param("id")))
	if instance == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": false,
			"error":  "channel not found",
			"data":   ChargeSequence{},
		})
		return
	}

	charge, err := instance.FetchCharge()
	c.JSON(http.StatusOK, gin.H{
		"status": err == nil,
		"error":  utils.GetError(err),
		"data":   charge,
	})
}

func DeleteCharge(c *gin.Context) {
	id := c.Param("id")
	state := ChargeInstance.DeleteRule(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func SyncCharge(c *gin.Context) {
	var form SyncChargeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
	}

	state := ChargeInstance.SyncRules(form.Data, form.Overwrite)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   SystemInstance,
	})
}

func UpdateConfig(c *gin.Context) {
	var config SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := SystemInstance.UpdateConfig(&config)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}
