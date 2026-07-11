package channel

import (
	"chat/globals"
)

type Channel struct {
	Id             int                 `json:"id" mapstructure:"id"`
	Name           string              `json:"name" mapstructure:"name"`
	Type           string              `json:"type" mapstructure:"type"`
	Priority       int                 `json:"priority" mapstructure:"priority"`
	Weight         int                 `json:"weight" mapstructure:"weight"`
	Models         []string            `json:"models" mapstructure:"models"`
	Retry          int                 `json:"retry" mapstructure:"retry"`
	Secret         string              `json:"secret" mapstructure:"secret"`
	Endpoint       string              `json:"endpoint" mapstructure:"endpoint"`
	Mapper         string              `json:"mapper" mapstructure:"mapper"`
	AutoModels     bool                `json:"auto_models" mapstructure:"automodels"`
	AutoPrice      bool                `json:"auto_price" mapstructure:"autoprice"`
	PriceAdjust    float32             `json:"price_adjust" mapstructure:"priceadjust"`
	PriceOverwrite bool                `json:"price_overwrite" mapstructure:"priceoverwrite"`
	State          bool                `json:"state" mapstructure:"state"`
	Group          []string            `json:"group" mapstructure:"group"`
	Proxy          globals.ProxyConfig `json:"proxy" mapstructure:"proxy"`
	Reflect        *map[string]string  `json:"-"`
	HitModels      *[]string           `json:"-"`
	ExcludeModels  *[]string           `json:"-"`
	CurrentSecret  *string             `json:"-"`
}

type Sequence []*Channel

type Manager struct {
	Sequence          Sequence            `json:"sequence"`
	PreflightSequence map[string]Sequence `json:"preflight_sequence"`
	Models            []string            `json:"models"`
}

type Ticker struct {
	Sequence Sequence `json:"sequence"`
	Cursor   int      `json:"cursor"`
}

type Charge struct {
	Id        int      `json:"id" mapstructure:"id"`
	ChannelId *int     `json:"channel_id,omitempty" mapstructure:"channelid"`
	Type      string   `json:"type" mapstructure:"type"`
	Models    []string `json:"models" mapstructure:"models"`
	Input     float32  `json:"input" mapstructure:"input"`
	Output    float32  `json:"output" mapstructure:"output"`
	Anonymous bool     `json:"anonymous" mapstructure:"anonymous"`
	Unset     bool     `json:"-" mapstructure:"-"`
}

type ChargeSequence []*Charge

type ChargeManager struct {
	Sequence         ChargeSequence     `json:"sequence"`
	Models           map[string]*Charge `json:"models"`
	NonBillingModels []string           `json:"non_billing_models"`
}
