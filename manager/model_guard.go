package manager

import (
	"chat/admin"
	"chat/channel"
	"fmt"
)

func checkMarketModelEnabled(model string) error {
	if channel.ConduitInstance.HasChannel(model) &&
		admin.MarketInstance.IsModelEnabledForActiveChannels(model, channel.ConduitInstance) {
		return nil
	}
	return fmt.Errorf("model is disabled by administrator (model: %s)", model)
}
