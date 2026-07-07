package manager

import (
	"chat/admin"
	"fmt"
)

func checkMarketModelEnabled(model string) error {
	if admin.MarketInstance.IsModelEnabled(model) {
		return nil
	}
	return fmt.Errorf("model is disabled by administrator (model: %s)", model)
}
