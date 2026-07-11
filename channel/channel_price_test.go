package channel

import (
	"chat/globals"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdjustedPrice(t *testing.T) {
	tests := []struct {
		name       string
		price      float32
		adjustment float32
		want       float32
	}{
		{name: "add", price: 1.25, adjustment: 0.5, want: 1.75},
		{name: "subtract", price: 1.25, adjustment: -0.5, want: 0.75},
		{name: "floor at zero", price: 0.25, adjustment: -0.5, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adjustedPrice(tt.price, tt.adjustment); got != tt.want {
				t.Fatalf("adjustedPrice(%v, %v) = %v, want %v", tt.price, tt.adjustment, got, tt.want)
			}
		})
	}
}

func TestChannelAutoPriceEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
		want     string
	}{
		{endpoint: "https://example.com", want: "https://example.com/v1/charge"},
		{endpoint: "https://example.com/", want: "https://example.com/v1/charge"},
		{endpoint: "https://example.com/v1", want: "https://example.com/v1/charge"},
	}

	for _, tt := range tests {
		channel := Channel{Endpoint: tt.endpoint}
		if got := channel.GetAutoPriceEndpoint(); got != tt.want {
			t.Fatalf("GetAutoPriceEndpoint() = %q, want %q", got, tt.want)
		}
	}
}

func TestFilterChargeSequenceUsesSelectedChannelModels(t *testing.T) {
	instance := Channel{Models: []string{"selected-model"}}
	charge, err := instance.filterChargeSequence(ChargeSequence{
		&Charge{
			Type:   globals.TokenBilling,
			Models: []string{"selected-model", "other-model"},
			Input:  1,
			Output: 2,
		},
		&Charge{
			Type:   globals.TimesBilling,
			Models: []string{"not-selected"},
			Output: 3,
		},
	})
	if err != nil {
		t.Fatalf("filterChargeSequence() error = %v", err)
	}
	if len(charge) != 1 {
		t.Fatalf("len(charge) = %d, want 1", len(charge))
	}
	if len(charge[0].Models) != 1 || charge[0].Models[0] != "selected-model" {
		t.Fatalf("models = %v, want [selected-model]", charge[0].Models)
	}
}

func TestChannelNewAPIEndpoints(t *testing.T) {
	tests := []struct {
		endpoint string
		pricing  string
		status   string
	}{
		{
			endpoint: "https://example.com",
			pricing:  "https://example.com/api/pricing",
			status:   "https://example.com/api/status",
		},
		{
			endpoint: "https://example.com/v1",
			pricing:  "https://example.com/api/pricing",
			status:   "https://example.com/api/status",
		},
	}

	for _, tt := range tests {
		channel := Channel{Endpoint: tt.endpoint}
		if got := channel.GetNewAPIPriceEndpoint(); got != tt.pricing {
			t.Fatalf("GetNewAPIPriceEndpoint() = %q, want %q", got, tt.pricing)
		}
		if got := channel.GetNewAPIStatusEndpoint(); got != tt.status {
			t.Fatalf("GetNewAPIStatusEndpoint() = %q, want %q", got, tt.status)
		}
	}
}

func TestChannelSkillAPIEndpoints(t *testing.T) {
	tests := []struct {
		endpoint string
		base     string
	}{
		{endpoint: "https://api.example.com", base: "https://api.example.com"},
		{endpoint: "https://api.example.com/", base: "https://api.example.com"},
		{endpoint: "https://api.example.com/api", base: "https://api.example.com"},
		{endpoint: "https://api.example.com/api/v1", base: "https://api.example.com"},
	}

	for _, tt := range tests {
		instance := Channel{Endpoint: tt.endpoint}
		if got := instance.GetSkillAPIBaseEndpoint(); got != tt.base {
			t.Fatalf("GetSkillAPIBaseEndpoint() = %q, want %q", got, tt.base)
		}
		if got := instance.GetSkillAPIModelsEndpoint(); got != tt.base+"/v1/skills/models" {
			t.Fatalf("GetSkillAPIModelsEndpoint() = %q", got)
		}
		if got := instance.GetSkillAPIPriceEndpoint("kimi/k2 code"); got != tt.base+"/v1/skills/models/kimi%2Fk2%20code/pricing?status=active" {
			t.Fatalf("GetSkillAPIPriceEndpoint() = %q", got)
		}
	}
}

func TestFetchChargeAdjustsUpstreamPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/charge" {
			t.Fatalf("request path = %q, want /v1/charge", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer upstream-secret" {
			t.Fatalf("Authorization = %q, want Bearer upstream-secret", auth)
		}
		if apiKey := r.Header.Get("x-api-key"); apiKey != "upstream-secret" {
			t.Fatalf("x-api-key = %q, want upstream-secret", apiKey)
		}

		_ = json.NewEncoder(w).Encode(ChargeSequence{
			&Charge{Type: globals.TokenBilling, Models: []string{"model-a"}, Input: 1, Output: 2},
			&Charge{Type: globals.TimesBilling, Models: []string{"image-a"}, Input: 9, Output: 0.25},
			&Charge{Type: globals.NonBilling, Models: []string{"free-a"}, Input: 7, Output: 8},
		})
	}))
	defer server.Close()

	instance := Channel{
		Endpoint:    server.URL,
		Secret:      "upstream-secret",
		Models:      []string{"model-a", "image-a", "free-a"},
		PriceAdjust: -0.5,
	}
	charge, err := instance.FetchCharge()
	if err != nil {
		t.Fatalf("FetchCharge() error = %v", err)
	}
	if len(charge) != 3 {
		t.Fatalf("len(charge) = %d, want 3", len(charge))
	}
	if charge[0].Input != 0.5 || charge[0].Output != 1.5 {
		t.Fatalf("token prices = (%v, %v), want (0.5, 1.5)", charge[0].Input, charge[0].Output)
	}
	if charge[1].Input != 9 || charge[1].Output != 0 {
		t.Fatalf("times prices = (%v, %v), want (9, 0)", charge[1].Input, charge[1].Output)
	}
	if charge[2].Input != 7 || charge[2].Output != 8 {
		t.Fatalf("non-billing prices = (%v, %v), want (7, 8)", charge[2].Input, charge[2].Output)
	}
}

func TestFetchChargeFallsBackToNewAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer upstream-secret" {
			t.Fatalf("Authorization = %q, want Bearer upstream-secret", auth)
		}

		switch r.URL.Path {
		case "/v1/charge":
			http.NotFound(w, r)
		case "/api/pricing":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": []map[string]interface{}{
					{
						"model_name":       "token-model",
						"quota_type":       0,
						"model_ratio":      1.5,
						"model_price":      0,
						"completion_ratio": 5,
					},
					{
						"model_name":       "fixed-model",
						"quota_type":       1,
						"model_ratio":      0,
						"model_price":      0.2,
						"completion_ratio": 0,
					},
				},
			})
		case "/api/status":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"quota_per_unit":    500000,
					"usd_exchange_rate": 7.3,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	instance := Channel{
		Endpoint:    server.URL + "/v1",
		Secret:      "upstream-secret",
		Models:      []string{"token-model", "fixed-model"},
		PriceAdjust: 0.1,
	}
	charge, err := instance.FetchCharge()
	if err != nil {
		t.Fatalf("FetchCharge() error = %v", err)
	}
	if len(charge) != 2 {
		t.Fatalf("len(charge) = %d, want 2", len(charge))
	}

	// token input: 1.5 * (1000 / 500000) * (10 * 7.3) = 0.219, plus 0.1
	assertFloat32Close(t, charge[0].Input, 0.319)
	// token output: input base * completion ratio 5 = 1.095, plus 0.1
	assertFloat32Close(t, charge[0].Output, 1.195)
	if charge[0].Type != globals.TokenBilling || charge[0].Models[0] != "token-model" {
		t.Fatalf("unexpected token charge: %+v", charge[0])
	}

	// fixed price: $0.2 * 7.3 CNY/USD * 10 quota/CNY = 14.6, plus 0.1
	assertFloat32Close(t, charge[1].Output, 14.7)
	if charge[1].Type != globals.TimesBilling || charge[1].Models[0] != "fixed-model" {
		t.Fatalf("unexpected fixed charge: %+v", charge[1])
	}
}

func TestFetchChargeFallsBackToSkillAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer upstream-secret" {
			t.Fatalf("Authorization = %q, want Bearer upstream-secret", auth)
		}

		switch r.URL.Path {
		case "/v1/charge", "/api/pricing":
			http.NotFound(w, r)
		case "/v1/skills/models/token-model/pricing":
			if r.URL.Query().Get("status") != "active" {
				t.Fatalf("status query = %q, want active", r.URL.Query().Get("status"))
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"available_for_this_key": true,
				"name":                   "token-model",
				"channel_groups": []map[string]interface{}{
					{
						"is_active":          true,
						"in_key_whitelist":   true,
						"billing_method":     "按token",
						"input_token_price":  2.5,
						"output_token_price": 10.0,
					},
					{
						"is_active":          true,
						"in_key_whitelist":   true,
						"billing_method":     "按token",
						"input_token_price":  1.5,
						"output_token_price": 5.0,
					},
				},
			})
		case "/v1/skills/models/image-model/pricing":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"available_for_this_key": true,
				"name":                   "image-model",
				"channel_groups": []map[string]interface{}{
					{
						"is_active":        true,
						"in_key_whitelist": true,
						"billing_method":   "按次",
						"base_price":       0.25,
					},
				},
			})
		case "/v1/skills/models/video-model/pricing":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"available_for_this_key": true,
				"name":                   "video-model",
				"channel_groups": []map[string]interface{}{
					{
						"is_active":        true,
						"in_key_whitelist": true,
						"billing_method":   "按秒",
						"base_price":       0.1,
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	instance := Channel{
		Endpoint:    server.URL,
		Secret:      "upstream-secret",
		Models:      []string{"token-model", "image-model", "video-model"},
		PriceAdjust: 0.1,
	}
	charge, err := instance.FetchCharge()
	if err != nil {
		t.Fatalf("FetchCharge() error = %v", err)
	}
	if len(charge) != 2 {
		t.Fatalf("len(charge) = %d, want 2", len(charge))
	}

	if charge[0].Type != globals.TokenBilling || charge[0].Models[0] != "token-model" {
		t.Fatalf("unexpected token charge: %+v", charge[0])
	}
	// Skill API token prices are per million tokens; CoAI stores per 1K.
	assertFloat32Close(t, charge[0].Input, 0.1015)
	assertFloat32Close(t, charge[0].Output, 0.105)

	if charge[1].Type != globals.TimesBilling || charge[1].Models[0] != "image-model" {
		t.Fatalf("unexpected image charge: %+v", charge[1])
	}
	assertFloat32Close(t, charge[1].Output, 0.35)
}

func TestFetchChargeFallsBackToLegacySkillAPIPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/charge", "/api/pricing", "/v1/skills/models/legacy-model/pricing":
			http.NotFound(w, r)
		case "/api/v1/skills/models/legacy-model/pricing":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"available_for_this_key": true,
				"name":                   "legacy-model",
				"channel_groups": []map[string]interface{}{
					{
						"is_active":          true,
						"in_key_whitelist":   true,
						"billing_method":     "按token",
						"input_token_price":  2.0,
						"output_token_price": 4.0,
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	instance := Channel{
		Endpoint: server.URL,
		Secret:   "upstream-secret",
		Models:   []string{"legacy-model"},
	}
	charge, err := instance.FetchCharge()
	if err != nil {
		t.Fatalf("FetchCharge() error = %v", err)
	}
	if len(charge) != 1 || charge[0].Models[0] != "legacy-model" {
		t.Fatalf("unexpected charge: %+v", charge)
	}
	assertFloat32Close(t, charge[0].Input, 0.002)
	assertFloat32Close(t, charge[0].Output, 0.004)
}

func TestFetchAgnesSubscriptionChargeCreatesChannelFreeRule(t *testing.T) {
	instance := Channel{
		Id:       7,
		Endpoint: "https://example.com",
		Secret:   "upstream-secret",
		Models:   []string{"agnes-text", "agnes-image"},
	}
	// The fallback is intentionally restricted to the real Agnes API domain.
	if _, err := instance.fetchAgnesSubscriptionCharge(instance.getUpstreamHeaders()); err == nil {
		t.Fatal("fetchAgnesSubscriptionCharge() succeeded for a non-Agnes domain")
	}

	instance.Endpoint = "https://apihub.agnes-ai.com"
	if !instance.isAgnesSubscriptionChannel() {
		t.Fatal("isAgnesSubscriptionChannel() = false for Agnes APIHub")
	}
	if got := instance.GetAgnesSubscriptionEndpoint(); got != "https://apihub.agnes-ai.com/v1/dashboard/billing/subscription" {
		t.Fatalf("GetAgnesSubscriptionEndpoint() = %q", got)
	}
}

func TestParseAgnesSubscriptionCharge(t *testing.T) {
	instance := Channel{Models: []string{"agnes-text", "agnes-image"}}
	charge, err := instance.parseAgnesSubscriptionCharge(map[string]interface{}{
		"object":             "billing_subscription",
		"has_payment_method": true,
		"hard_limit_usd":     100000000,
	})
	if err != nil {
		t.Fatalf("parseAgnesSubscriptionCharge() error = %v", err)
	}
	if len(charge) != 1 || charge[0].Type != globals.NonBilling || len(charge[0].Models) != 2 {
		t.Fatalf("unexpected Agnes charge: %+v", charge)
	}
}

func assertFloat32Close(t *testing.T, got, want float32) {
	t.Helper()
	delta := got - want
	if delta < 0 {
		delta = -delta
	}
	if delta > 0.0001 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestChargeManagerKeepsChannelPricesIndependent(t *testing.T) {
	channelOne := 1
	channelTwo := 2
	manager := &ChargeManager{Sequence: ChargeSequence{
		&Charge{Id: 1, ChannelId: &channelOne, Type: globals.TokenBilling, Models: []string{"shared-model"}, Input: 1, Output: 2},
		&Charge{Id: 2, ChannelId: &channelTwo, Type: globals.TokenBilling, Models: []string{"shared-model"}, Input: 3, Output: 4},
	}}
	manager.Load()

	if got := manager.GetChargeByChannel("shared-model", &channelOne); got.GetInput() != 1 {
		t.Fatalf("channel one input = %v, want 1", got.GetInput())
	}
	if got := manager.GetChargeByChannel("shared-model", &channelTwo); got.GetInput() != 3 {
		t.Fatalf("channel two input = %v, want 3", got.GetInput())
	}
	if got := manager.GetCharge("shared-model"); got.GetInput() != 3 {
		t.Fatalf("generic fallback input = %v, want conservative channel price 3", got.GetInput())
	}
}

func TestChargeManagerMigratesLegacyPricesByChannel(t *testing.T) {
	manager := &Manager{Sequence: Sequence{
		&Channel{Id: 1, Models: []string{"shared-model"}, State: true},
		&Channel{Id: 2, Models: []string{"shared-model"}, State: true},
	}}
	for _, item := range manager.Sequence {
		item.Load()
	}
	prices := &ChargeManager{Sequence: ChargeSequence{
		&Charge{Id: 1, Type: globals.TimesBilling, Models: []string{"shared-model"}, Output: 2},
	}}
	prices.Load()

	if !prices.MigrateChannelPrices(manager) {
		t.Fatal("MigrateChannelPrices() = false, want true")
	}
	if len(prices.Sequence) != 3 {
		t.Fatalf("len(prices.Sequence) = %d, want legacy + 2 channel rules", len(prices.Sequence))
	}
	for _, channelId := range []int{1, 2} {
		if rule := prices.GetRuleByModelAndChannel("shared-model", channelId); rule == nil || rule.GetOutput() != 2 {
			t.Fatalf("channel %d migrated rule = %+v, want output 2", channelId, rule)
		}
	}
}
