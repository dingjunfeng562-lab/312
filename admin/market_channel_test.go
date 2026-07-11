package admin

import (
	"chat/channel"
	"testing"
)

func TestMarketMigratesLegacyModelsByChannel(t *testing.T) {
	manager := &channel.Manager{Sequence: channel.Sequence{
		&channel.Channel{Id: 1, Name: "primary", Models: []string{"shared-model"}, State: true},
		&channel.Channel{Id: 2, Name: "backup", Models: []string{"shared-model"}, State: true},
	}}
	for _, item := range manager.Sequence {
		item.Load()
	}
	market := &Market{Models: MarketModelList{{Id: "shared-model", Name: "Shared"}}}

	if !market.MigrateChannelModels(manager) {
		t.Fatal("MigrateChannelModels() = false, want true")
	}
	if len(market.Models) != 2 {
		t.Fatalf("len(market.Models) = %d, want 2", len(market.Models))
	}
	for index, channelId := range []int{1, 2} {
		if market.Models[index].ChannelId == nil || *market.Models[index].ChannelId != channelId {
			t.Fatalf("model %d channel = %v, want %d", index, market.Models[index].ChannelId, channelId)
		}
	}
}

func TestMarketModelEnabledWhenAnyChannelEntryIsEnabled(t *testing.T) {
	disabled := false
	enabled := true
	market := &Market{Models: MarketModelList{
		{Id: "shared-model", ChannelId: intPointer(1), Enabled: &disabled},
		{Id: "shared-model", ChannelId: intPointer(2), Enabled: &enabled},
	}}
	if !market.IsModelEnabled("shared-model") {
		t.Fatal("IsModelEnabled() = false, want true when one channel entry is enabled")
	}
}

func TestMarketMigrationDoesNotDuplicateExistingChannelEntry(t *testing.T) {
	manager := &channel.Manager{Sequence: channel.Sequence{
		&channel.Channel{Id: 1, Name: "primary", Models: []string{"shared-model"}, State: true},
		&channel.Channel{Id: 2, Name: "backup", Models: []string{"shared-model"}, State: true},
	}}
	for _, item := range manager.Sequence {
		item.Load()
	}
	market := &Market{Models: MarketModelList{
		{Id: "shared-model", Name: "Primary", ChannelId: intPointer(1)},
		{Id: "shared-model", Name: "Legacy"},
	}}

	if !market.MigrateChannelModels(manager) {
		t.Fatal("MigrateChannelModels() = false, want true")
	}
	if len(market.Models) != 2 {
		t.Fatalf("len(market.Models) = %d, want one entry per channel", len(market.Models))
	}
}

func TestMarketMigrationCreatesEntriesForNewChannelModels(t *testing.T) {
	manager := &channel.Manager{Sequence: channel.Sequence{
		&channel.Channel{Id: 7, Name: "upstream", Models: []string{"new-model"}, State: true},
	}}
	manager.Sequence[0].Load()
	market := &Market{Models: MarketModelList{}}

	if !market.MigrateChannelModels(manager) {
		t.Fatal("MigrateChannelModels() = false, want true")
	}
	if len(market.Models) != 1 {
		t.Fatalf("len(market.Models) = %d, want 1", len(market.Models))
	}
	model := market.Models[0]
	if model.Id != "new-model" || model.Name != "new-model" {
		t.Fatalf("model = %+v, want generated metadata for new-model", model)
	}
	if model.ChannelId == nil || *model.ChannelId != 7 || model.ChannelName != "upstream" {
		t.Fatalf("channel metadata = %+v, want channel 7/upstream", model)
	}
}

func TestActiveChannelModelsHidesDisabledChannelEntries(t *testing.T) {
	disabled := false
	manager := &channel.Manager{Sequence: channel.Sequence{
		&channel.Channel{Id: 1, Name: "active", Models: []string{"shared-model"}, State: true},
		&channel.Channel{Id: 2, Name: "disabled", Models: []string{"hidden-model", "shared-model"}, State: false},
	}}
	manager.Load()
	market := MarketModelList{
		{Id: "shared-model", ChannelId: intPointer(1)},
		{Id: "shared-model", ChannelId: intPointer(2)},
		{Id: "hidden-model", ChannelId: intPointer(2)},
		{Id: "shared-model", Enabled: &disabled},
	}

	models := market.ActiveChannelModels(manager, false)
	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want active channel entry and active legacy entry", len(models))
	}
	for _, model := range models {
		if model.ChannelId != nil && *model.ChannelId == 2 {
			t.Fatalf("disabled channel model leaked into result: %+v", model)
		}
	}
	if enabled := market.ActiveChannelModels(manager, true); len(enabled) != 1 {
		t.Fatalf("len(enabled models) = %d, want 1", len(enabled))
	}
}

func TestMarketEnabledCheckIgnoresDisabledChannels(t *testing.T) {
	disabled := false
	enabled := true
	manager := &channel.Manager{Sequence: channel.Sequence{
		&channel.Channel{Id: 1, Models: []string{"shared-model"}, State: true},
		&channel.Channel{Id: 2, Models: []string{"shared-model"}, State: false},
	}}
	manager.Load()
	market := &Market{Models: MarketModelList{
		{Id: "shared-model", ChannelId: intPointer(1), Enabled: &disabled},
		{Id: "shared-model", ChannelId: intPointer(2), Enabled: &enabled},
	}}

	if market.IsModelEnabledForActiveChannels("shared-model", manager) {
		t.Fatal("disabled channel entry must not enable the active channel model")
	}
}

func TestMergeActiveModelsPreservesDisabledChannelConfiguration(t *testing.T) {
	manager := &channel.Manager{Sequence: channel.Sequence{
		&channel.Channel{Id: 1, Models: []string{"active-model"}, State: true},
		&channel.Channel{Id: 2, Models: []string{"hidden-model"}, State: false},
	}}
	manager.Load()
	market := &Market{Models: MarketModelList{
		{Id: "active-model", Name: "Old active", ChannelId: intPointer(1)},
		{Id: "hidden-model", Name: "Hidden", ChannelId: intPointer(2)},
	}}

	merged := market.MergeActiveModels(MarketModelList{
		{Id: "active-model", Name: "Updated active", ChannelId: intPointer(1)},
	}, manager)
	if len(merged) != 2 {
		t.Fatalf("len(merged) = %d, want 2", len(merged))
	}
	if merged[0].Name != "Updated active" {
		t.Fatalf("active entry was not replaced: %+v", merged[0])
	}
	if merged[1].Id != "hidden-model" || merged[1].ChannelId == nil || *merged[1].ChannelId != 2 {
		t.Fatalf("disabled channel entry was not preserved: %+v", merged[1])
	}
}

func intPointer(value int) *int {
	return &value
}
