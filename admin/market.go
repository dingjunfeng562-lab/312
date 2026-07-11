package admin

import (
	"chat/channel"
	"chat/globals"
	"chat/utils"
	"fmt"

	"github.com/spf13/viper"
)

type ModelTag []string
type MarketModel struct {
	Id            string   `json:"id" mapstructure:"id" required:"true"`
	ChannelId     *int     `json:"channel_id,omitempty" mapstructure:"channel_id"`     // 渠道ID，nil表示所有渠道
	ChannelName   string   `json:"channel_name,omitempty" mapstructure:"channel_name"` // 渠道名称（显示用）
	Name          string   `json:"name" mapstructure:"name" required:"true"`
	Description   string   `json:"description" mapstructure:"description"`
	Enabled       *bool    `json:"enabled,omitempty" mapstructure:"enabled"`
	Default       bool     `json:"default" mapstructure:"default"`
	HighContext   bool     `json:"high_context" mapstructure:"highcontext"`
	ResponseSpeed string   `json:"response_speed" mapstructure:"responsespeed"`
	ModelType     string   `json:"model_type" mapstructure:"modeltype"`
	Avatar        string   `json:"avatar" mapstructure:"avatar"`
	Tag           ModelTag `json:"tag" mapstructure:"tag"`
	Channels      []string `json:"channels,omitempty" mapstructure:"-"` // 兼容字段：管理后台显示支持的渠道列表
}
type MarketModelList []MarketModel

func (m MarketModel) IsEnabled() bool {
	return m.Enabled == nil || *m.Enabled
}

func (m MarketModelList) EnabledModels() MarketModelList {
	models := MarketModelList{}
	for _, model := range m {
		if model.IsEnabled() {
			models = append(models, model)
		}
	}
	return models
}

// ActiveChannelModels filters presentation/API data by channel state while
// leaving the stored market configuration untouched.
func (m MarketModelList) ActiveChannelModels(manager *channel.Manager, enabledOnly bool) MarketModelList {
	models := MarketModelList{}
	if manager == nil {
		return models
	}
	for _, model := range m {
		if enabledOnly && !model.IsEnabled() {
			continue
		}
		if model.ChannelId == nil {
			if !manager.HasChannel(model.Id) {
				continue
			}
		} else {
			instance := manager.GetSequence().GetChannelById(*model.ChannelId)
			if instance == nil || !instance.GetState() || !instance.IsHit(model.Id) {
				continue
			}
		}
		models = append(models, model)
	}
	return models
}

func marketModelKey(model MarketModel) string {
	if model.ChannelId == nil {
		return "legacy:" + model.Id
	}
	return fmt.Sprintf("%d:%s", *model.ChannelId, model.Id)
}

func isActiveChannelModel(model MarketModel, manager *channel.Manager) bool {
	if manager == nil {
		return false
	}
	if model.ChannelId == nil {
		return manager.HasChannel(model.Id)
	}
	instance := manager.GetSequence().GetChannelById(*model.ChannelId)
	return instance != nil && instance.GetState() && instance.IsHit(model.Id)
}

type Market struct {
	Models MarketModelList `json:"models" mapstructure:"models"`
}

func NewMarket() *Market {
	var models MarketModelList
	if err := viper.UnmarshalKey("market", &models); err != nil {
		globals.Warn(fmt.Sprintf("[market] read config error: %s, use default config", err.Error()))
		models = MarketModelList{}
	}

	market := &Market{
		Models: models,
	}
	if market.MigrateChannelModels(channel.ConduitInstance) {
		if err := market.SaveConfig(); err != nil {
			globals.Warn(fmt.Sprintf("[market] failed to save channel model migration: %s", err.Error()))
		}
	}
	market.LoadModelTypes()
	return market
}

func (m *Market) GetModels() MarketModelList {
	return m.Models.EnabledModels()
}

func (m *Market) GetAllModels() MarketModelList {
	return m.Models
}

func (m *Market) GetModel(id string) *MarketModel {
	var fallback *MarketModel
	for i := range m.Models {
		if m.Models[i].Id == id {
			if m.Models[i].ChannelId == nil {
				return &m.Models[i]
			}
			if fallback == nil {
				fallback = &m.Models[i]
			}
		}
	}
	return fallback
}

// GetModelByChannel 根据模型ID和渠道ID获取模型（新增）
func (m *Market) GetModelByChannel(id string, channelId *int) *MarketModel {
	for i := range m.Models {
		if m.Models[i].Id == id {
			// 如果指定了渠道ID，需要精确匹配
			if channelId != nil {
				if m.Models[i].ChannelId != nil && *m.Models[i].ChannelId == *channelId {
					return &m.Models[i]
				}
			} else {
				// 如果没有指定渠道，返回第一个匹配的（或者ChannelId为nil的通用配置）
				if m.Models[i].ChannelId == nil {
					return &m.Models[i]
				}
			}
		}
	}
	return nil
}

func (m *Market) MigrateChannelModels(manager *channel.Manager) bool {
	if manager == nil {
		return false
	}

	migrated := make(MarketModelList, 0, len(m.Models))
	existing := map[string]bool{}
	for _, model := range m.Models {
		if model.ChannelId != nil {
			existing[fmt.Sprintf("%d:%s", *model.ChannelId, model.Id)] = true
		}
	}
	changed := false
	for _, model := range m.Models {
		if model.ChannelId != nil {
			migrated = append(migrated, model)
			continue
		}

		matched := false
		for _, ch := range manager.GetSequence() {
			if ch == nil || !ch.IsHit(model.Id) {
				continue
			}
			key := fmt.Sprintf("%d:%s", ch.GetId(), model.Id)
			if existing[key] {
				matched = true
				continue
			}
			instance := model
			instance.ChannelId = utils.ToPtr(ch.GetId())
			instance.ChannelName = ch.GetName()
			instance.Channels = nil
			migrated = append(migrated, instance)
			existing[key] = true
			matched = true
		}
		if matched {
			changed = true
		} else {
			migrated = append(migrated, model)
		}
	}

	for _, ch := range manager.GetSequence() {
		if ch == nil {
			continue
		}
		for _, modelId := range ch.GetModels() {
			key := fmt.Sprintf("%d:%s", ch.GetId(), modelId)
			if existing[key] {
				continue
			}
			var template *MarketModel
			for index := range migrated {
				if migrated[index].Id == modelId {
					template = &migrated[index]
					break
				}
			}
			instance := MarketModel{Id: modelId, Name: modelId}
			if template != nil {
				instance = *template
			}
			instance.ChannelId = utils.ToPtr(ch.GetId())
			instance.ChannelName = ch.GetName()
			instance.Channels = nil
			migrated = append(migrated, instance)
			existing[key] = true
			changed = true
		}
	}

	if changed {
		m.Models = migrated
	}
	return changed
}

func (m *Market) IsModelEnabled(id string) bool {
	found := false
	for _, model := range m.Models {
		if model.Id != id {
			continue
		}
		found = true
		if model.IsEnabled() {
			return true
		}
	}
	return !found
}

func (m *Market) IsModelEnabledForActiveChannels(id string, manager *channel.Manager) bool {
	found := false
	for _, model := range m.Models.ActiveChannelModels(manager, false) {
		if model.Id != id {
			continue
		}
		found = true
		if model.IsEnabled() {
			return true
		}
	}
	return !found
}

func (m *Market) SaveConfig() error {
	return utils.SaveConfig("market", m.Models)
}

func (m *Market) LoadModelTypes() {
	types := map[string]string{}
	for _, model := range m.Models {
		if _, exists := types[model.Id]; !exists || model.ChannelId == nil {
			types[model.Id] = model.ModelType
		}
	}
	globals.SetModelTypes(types)
}

func (m *Market) SetModels(models MarketModelList) error {
	for index := range models {
		// 清空 Channels 字段（不保存到配置）
		models[index].Channels = nil
		if models[index].ChannelId != nil {
			if instance := channel.ConduitInstance.GetSequence().GetChannelById(*models[index].ChannelId); instance != nil {
				models[index].ChannelName = instance.GetName()
			}
		}
	}
	m.Models = models
	if err := m.SaveConfig(); err != nil {
		return err
	}
	m.LoadModelTypes()
	return nil
}

// SetActiveModels replaces the entries currently exposed to the admin while
// retaining configuration owned by disabled channels. This prevents saving an
// edited visible list from permanently deleting temporarily hidden entries.
func (m *Market) MergeActiveModels(models MarketModelList, manager *channel.Manager) MarketModelList {
	keys := make(map[string]bool, len(models))
	for _, model := range models {
		keys[marketModelKey(model)] = true
	}

	merged := append(MarketModelList(nil), models...)
	for _, model := range m.Models {
		if isActiveChannelModel(model, manager) || keys[marketModelKey(model)] {
			continue
		}
		merged = append(merged, model)
	}
	return merged
}

func (m *Market) SetActiveModels(models MarketModelList, manager *channel.Manager) error {
	return m.SetModels(m.MergeActiveModels(models, manager))
}
