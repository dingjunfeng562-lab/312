package admin

import (
	"chat/globals"
	"chat/utils"
	"fmt"

	"github.com/spf13/viper"
)

type ModelTag []string
type MarketModel struct {
	Id            string   `json:"id" mapstructure:"id" required:"true"`
	Name          string   `json:"name" mapstructure:"name" required:"true"`
	Description   string   `json:"description" mapstructure:"description"`
	Enabled       *bool    `json:"enabled,omitempty" mapstructure:"enabled"`
	Default       bool     `json:"default" mapstructure:"default"`
	HighContext   bool     `json:"high_context" mapstructure:"highcontext"`
	ResponseSpeed string   `json:"response_speed" mapstructure:"responsespeed"`
	ModelType     string   `json:"model_type" mapstructure:"modeltype"`
	Avatar        string   `json:"avatar" mapstructure:"avatar"`
	Tag           ModelTag `json:"tag" mapstructure:"tag"`
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
	for i := range m.Models {
		if m.Models[i].Id == id {
			return &m.Models[i]
		}
	}
	return nil
}

func (m *Market) IsModelEnabled(id string) bool {
	model := m.GetModel(id)
	return model == nil || model.IsEnabled()
}

func (m *Market) SaveConfig() error {
	return utils.SaveConfig("market", m.Models)
}

func (m *Market) LoadModelTypes() {
	types := map[string]string{}
	for _, model := range m.Models {
		types[model.Id] = model.ModelType
	}
	globals.SetModelTypes(types)
}

func (m *Market) SetModels(models MarketModelList) error {
	m.Models = models
	if err := m.SaveConfig(); err != nil {
		return err
	}
	m.LoadModelTypes()
	return nil
}
