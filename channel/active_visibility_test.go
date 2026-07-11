package channel

import "testing"

func TestReloadActiveModelsDropsDisabledChannelModels(t *testing.T) {
	active := &Channel{Id: 1, Models: []string{"shared-model"}, State: true}
	disabled := &Channel{Id: 2, Models: []string{"hidden-model", "shared-model"}, State: false}
	manager := &Manager{Sequence: Sequence{active, disabled}}
	manager.Load()

	if !manager.HasChannel("shared-model") {
		t.Fatal("active shared model is missing")
	}
	if manager.HasChannel("hidden-model") {
		t.Fatal("disabled channel model leaked into active model index")
	}

	active.State = false
	manager.reloadActiveModels()
	if manager.HasChannel("shared-model") {
		t.Fatal("model remained in active index after its last channel was disabled")
	}
}

func TestListActiveRulesFiltersDisabledChannelsAndModels(t *testing.T) {
	active := &Channel{Id: 1, Models: []string{"active-model"}, State: true}
	disabled := &Channel{Id: 2, Models: []string{"hidden-model"}, State: false}
	manager := &Manager{Sequence: Sequence{active, disabled}}
	manager.Load()
	one, two := 1, 2
	charges := &ChargeManager{Sequence: ChargeSequence{
		&Charge{Id: 1, ChannelId: &one, Models: []string{"active-model"}},
		&Charge{Id: 2, ChannelId: &two, Models: []string{"hidden-model"}},
		&Charge{Id: 3, Models: []string{"active-model", "hidden-model"}},
	}}

	rules := charges.ListActiveRules(manager)
	if len(rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(rules))
	}
	for _, rule := range rules {
		for _, model := range rule.Models {
			if model == "hidden-model" {
				t.Fatalf("disabled model leaked into price rules: %+v", rule)
			}
		}
	}
	if len(charges.Sequence[2].Models) != 2 {
		t.Fatal("filtering active rules mutated persisted price configuration")
	}
}
