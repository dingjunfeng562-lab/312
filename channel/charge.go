package channel

import (
	"chat/globals"
	"chat/utils"
	"fmt"

	"github.com/spf13/viper"
)

func NewChargeManager() *ChargeManager {
	var seq ChargeSequence
	if err := viper.UnmarshalKey("charge", &seq); err != nil {
		panic(err)
	}

	m := &ChargeManager{
		Sequence:         seq,
		Models:           map[string]*Charge{},
		NonBillingModels: []string{},
	}
	m.Load()
	if m.MigrateChannelPrices(ConduitInstance) {
		if err := m.SaveConfig(); err != nil {
			globals.Warn(fmt.Sprintf("[charge] failed to save channel price migration: %s", err.Error()))
		}
	}

	return m
}

func chargeKey(model string, channelId *int) string {
	if channelId == nil {
		return model
	}
	return fmt.Sprintf("%d:%s", *channelId, model)
}

func sameChannel(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

// MigrateChannelPrices expands legacy model-only prices into one rule per
// channel/model pair. The legacy rule is retained for callers that do not yet
// know the selected channel and as a fallback for newly added channels.
func (m *ChargeManager) MigrateChannelPrices(manager *Manager) bool {
	if manager == nil {
		return false
	}

	changed := false
	legacy := append(ChargeSequence(nil), m.Sequence...)
	for _, rule := range legacy {
		if rule == nil || rule.ChannelId != nil {
			continue
		}
		for _, model := range rule.Models {
			for _, ch := range manager.GetSequence() {
				if ch == nil || !ch.IsHit(model) || m.GetRuleByModelAndChannel(model, ch.GetId()) != nil {
					continue
				}
				instance := rule.New(model)
				instance.ChannelId = utils.ToPtr(ch.GetId())
				m.AddRawRule(instance)
				changed = true
			}
		}
	}
	if changed {
		m.Load()
	}
	return changed
}

func (m *ChargeManager) Load() {
	seq := make(ChargeSequence, 0)
	for _, charge := range m.Sequence {
		if charge == nil {
			continue
		}
		if charge.Id == -1 {
			charge.Id = m.GetMaxId() + 1
		}
		seq = append(seq, charge)
	}
	m.Sequence = seq

	// init support models
	m.Models = map[string]*Charge{}
	for _, charge := range m.Sequence {
		for _, model := range charge.Models {
			key := chargeKey(model, charge.ChannelId)
			if _, ok := m.Models[key]; !ok {
				m.Models[key] = charge
			}
		}
	}

	m.NonBillingModels = []string{}
	for _, charge := range m.Sequence {
		if !charge.IsBilling() {
			for _, model := range charge.Models {
				m.NonBillingModels = append(m.NonBillingModels, model)
			}
		}
	}
}

func (m *ChargeManager) GetModels() map[string]*Charge {
	return m.Models
}

func (m *ChargeManager) GetNonBillingModels() []string {
	return m.NonBillingModels
}

func (m *ChargeManager) IsBilling(model string) bool {
	return !utils.Contains(model, m.NonBillingModels)
}

func (m *ChargeManager) GetCharge(model string) *Charge {
	return m.GetChargeByChannel(model, nil)
}

func (m *ChargeManager) GetChargeByChannel(model string, channelId *int) *Charge {
	if channelId != nil {
		if charge, ok := m.Models[chargeKey(model, channelId)]; ok {
			return charge
		}
	}
	if charge, ok := m.Models[chargeKey(model, nil)]; ok {
		return charge
	}
	for _, charge := range m.Sequence {
		if charge.ChannelId == nil && charge.Contains(model) {
			return charge
		}
	}
	if channelId == nil {
		var fallback *Charge
		for _, charge := range m.Sequence {
			if charge.ChannelId == nil || !charge.Contains(model) {
				continue
			}
			if fallback == nil || charge.GetLimit() > fallback.GetLimit() {
				fallback = charge
			}
		}
		if fallback != nil {
			return fallback
		}
	}
	return &Charge{
		Type:      globals.NonBilling,
		Anonymous: false,
		Unset:     true,
	}
}

func (m *ChargeManager) SaveConfig() error {
	return utils.SaveConfig("charge", m.Sequence)
}

func (m *ChargeManager) GetMaxId() int {
	max := 0
	for _, charge := range m.Sequence {
		if charge.Id > max {
			max = charge.Id
		}
	}
	return max
}

func (m *ChargeManager) AddRawRule(charge *Charge) {
	charge.Id = m.GetMaxId() + 1
	m.Sequence = append(m.Sequence, charge)
}

func (m *ChargeManager) AddRule(charge Charge) error {
	m.AddRawRule(&charge)
	m.Load()
	return m.SaveConfig()
}

func (m *ChargeManager) UpdateRawRule(charge *Charge) {
	for _, item := range m.Sequence {
		if item.Id == charge.Id {
			*item = *charge
			break
		}
	}
}

func (m *ChargeManager) UpdateRule(charge Charge) error {
	m.UpdateRawRule(&charge)
	m.Load()
	return m.SaveConfig()
}

func (m *ChargeManager) SetRawRule(charge *Charge) {
	if charge.Id == -1 {
		m.AddRawRule(charge)
	} else {
		m.UpdateRawRule(charge)
	}
}

func (m *ChargeManager) SetRule(charge Charge) error {
	m.SetRawRule(&charge)
	m.Load()
	return m.SaveConfig()
}

func (m *ChargeManager) DeleteRawRule(id int) {
	for i, item := range m.Sequence {
		if item.Id == id {
			m.Sequence = append(m.Sequence[:i], m.Sequence[i+1:]...)
			break
		}
	}
}

func (m *ChargeManager) DeleteRule(id int) error {
	m.DeleteRawRule(id)
	m.Load()
	return m.SaveConfig()
}

func (m *ChargeManager) SyncRules(charge ChargeSequence, overwrite bool) error {
	for _, item := range charge {
		m.SyncRule(item, overwrite)
	}

	m.Load()
	return m.SaveConfig()
}

func (m *ChargeManager) SyncRule(charge *Charge, overwrite bool) {
	if overwrite {
		m.SyncRuleWithOverwrite(charge)
	} else {
		m.SyncRuleWithoutOverwrite(charge)
	}
}

func (m *ChargeManager) SyncRuleWithOverwrite(charge *Charge) {
	if len(charge.Models) == 0 {
		return
	}

	for _, model := range charge.GetModels() {
		if raw := m.GetRuleByModelAndOptionalChannel(model, charge.ChannelId); raw != nil {
			if len(raw.Models) == 1 {
				// rule is already exist and only contains this model, just delete it

				m.DeleteRawRule(raw.Id)
			} else {
				// rule is already exist and contains other models, delete this model from it and add a new rule
				// delete model from raw rule
				raw.Models = utils.Filter(raw.Models, func(m string) bool {
					return m != model
				})
				m.UpdateRawRule(raw)
			}
		}
	}

	instance := charge.New("")
	instance.Models = charge.Models
	m.AddRawRule(instance)
}

func (m *ChargeManager) SyncRuleWithoutOverwrite(charge *Charge) {
	models := utils.Filter(charge.GetModels(), func(model string) bool {
		return !m.ContainsChannel(model, charge.ChannelId)
	})

	if len(models) > 0 {
		charge.Models = models
		m.AddRawRule(charge)
	}
}

func (m *ChargeManager) ListRules() ChargeSequence {
	return m.Sequence
}

// ListActiveRules returns display/API rules without mutating the persisted
// configuration. Rules belonging to disabled channels are kept in config so
// they become available again when the channel is re-enabled.
func (m *ChargeManager) ListActiveRules(manager *Manager) ChargeSequence {
	if manager == nil {
		return ChargeSequence{}
	}

	rules := make(ChargeSequence, 0, len(m.Sequence))
	for _, rule := range m.Sequence {
		if rule == nil {
			continue
		}

		var models []string
		if rule.ChannelId != nil {
			instance := manager.GetSequence().GetChannelById(*rule.ChannelId)
			if instance == nil || !instance.GetState() {
				continue
			}
			models = utils.Filter(rule.Models, func(model string) bool {
				return model == "*" || instance.IsHit(model)
			})
		} else {
			models = utils.Filter(rule.Models, func(model string) bool {
				return model == "*" && len(manager.GetModels()) > 0 || manager.HasChannel(model)
			})
		}

		if len(models) == 0 {
			continue
		}
		copy := *rule
		copy.Models = append([]string(nil), models...)
		rules = append(rules, &copy)
	}
	return rules
}

func (m *ChargeManager) Contains(model string) bool {
	return m.ContainsChannel(model, nil)
}

func (m *ChargeManager) ContainsChannel(model string, channelId *int) bool {
	for _, item := range m.Sequence {
		if sameChannel(item.ChannelId, channelId) && item.Contains(model) {
			return true
		}
	}
	return false
}

func (m *ChargeManager) GetRule(id int) *Charge {
	for _, item := range m.Sequence {
		if item.Id == id {
			return item
		}
	}
	return nil
}

func (m *ChargeManager) GetRuleByModel(model string) *Charge {
	return m.GetRuleByModelAndOptionalChannel(model, nil)
}

func (m *ChargeManager) GetRuleByModelAndChannel(model string, channelId int) *Charge {
	return m.GetRuleByModelAndOptionalChannel(model, utils.ToPtr(channelId))
}

func (m *ChargeManager) GetRuleByModelAndOptionalChannel(model string, channelId *int) *Charge {
	for _, item := range m.Sequence {
		if sameChannel(item.ChannelId, channelId) && item.Contains(model) {
			return item
		}
	}
	return nil
}

func (c *Charge) IsUnsetType() bool {
	return c.Unset
}

func (c *Charge) GetType() string {
	if c.Type == "" {
		return globals.NonBilling
	}
	return c.Type
}

func (c *Charge) GetModels() []string {
	return c.Models
}

func (c *Charge) GetInput() float32 {
	if c.Input <= 0 {
		return 0
	}
	return c.Input
}

func (c *Charge) GetOutput() float32 {
	if c.Output <= 0 {
		return 0
	}
	return c.Output
}

func (c *Charge) SupportAnonymous() bool {
	return c.Anonymous
}

func (c *Charge) IsBilling() bool {
	return c.GetType() != globals.NonBilling
}

func (c *Charge) IsBillingType(t string) bool {
	return c.GetType() == t
}

func (c *Charge) GetLimit() float32 {
	switch c.GetType() {
	case globals.NonBilling:
		return 0
	case globals.TimesBilling:
		return c.GetOutput()
	case globals.TokenBilling:
		// 1k input tokens + 1k output tokens
		return c.GetInput() + c.GetOutput()
	default:
		return 0
	}
}

func (c *Charge) Contains(model string) bool {
	return utils.Contains("*", c.Models) || utils.Contains(model, c.Models)
}

func (c *Charge) New(model string) *Charge {
	return &Charge{
		ChannelId: c.ChannelId,
		Type:      c.Type,
		Models:    []string{model},
		Input:     c.Input,
		Output:    c.Output,
		Anonymous: c.Anonymous,
	}
}
