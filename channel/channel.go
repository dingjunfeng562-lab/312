package channel

import (
	"chat/globals"
	"chat/utils"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var defaultMaxRetries = 1
var defaultReplacer = []string{
	"openai_api", "anthropic_api",
	"api2d", "closeai_api",
	"one_api", "new_api", "shell_api",
}

func (c *Channel) GetId() int {
	return c.Id
}

func (c *Channel) GetName() string {
	return c.Name
}

func (c *Channel) GetType() string {
	return c.Type
}

func (c *Channel) GetPriority() int {
	return c.Priority
}

func (c *Channel) GetWeight() int {
	if c.Weight <= 0 {
		return 1
	}
	return c.Weight
}

func (c *Channel) GetModels() []string {
	return c.Models
}

func (c *Channel) GetRetry() int {
	if c.Retry <= 0 {
		return defaultMaxRetries
	}
	return c.Retry
}

func (c *Channel) GetSecret() string {
	return c.Secret
}

func (c *Channel) GetCurrentSecret() *string {
	return c.CurrentSecret
}

// GetRandomSecret returns a random secret from the secret list
func (c *Channel) GetRandomSecret() string {
	arr := strings.Split(c.GetSecret(), "\n")
	if len(arr) == 0 {
		return ""
	}

	idx := utils.Intn(len(arr))
	secret := arr[idx]

	c.CurrentSecret = &secret
	return secret
}

func (c *Channel) GetCurrentSecretValue() string {
	if c.CurrentSecret == nil {
		return ""
	}

	return *c.CurrentSecret
}

func (c *Channel) SplitRandomSecret(num int) []string {
	secret := c.GetRandomSecret()
	arr := strings.Split(secret, "|")
	if len(arr) == num {
		return arr
	} else if len(arr) > num {
		return arr[:num]
	}

	for i := len(arr); i < num; i++ {
		arr = append(arr, "")
	}

	return arr
}

func (c *Channel) GetEndpoint() string {
	return c.Endpoint
}

func (c *Channel) GetDomain() string {
	if instance, err := url.Parse(c.GetEndpoint()); err == nil {
		return instance.Host
	}

	return c.GetEndpoint()
}

func (c *Channel) GetMapper() string {
	return c.Mapper
}

func (c *Channel) GetAutoModelsEndpoint() string {
	endpoint := strings.TrimRight(c.GetEndpoint(), "/")
	if strings.HasSuffix(endpoint, "/v1") {
		return endpoint + "/models"
	}
	return endpoint + "/v1/models"
}

func (c *Channel) GetAutoPriceEndpoint() string {
	endpoint := strings.TrimRight(c.GetEndpoint(), "/")
	if strings.HasSuffix(endpoint, "/v1") {
		return endpoint + "/charge"
	}
	return endpoint + "/v1/charge"
}

func (c *Channel) GetNewAPIPriceEndpoint() string {
	return strings.TrimSuffix(strings.TrimRight(c.GetEndpoint(), "/"), "/v1") + "/api/pricing"
}

func (c *Channel) GetNewAPIStatusEndpoint() string {
	return strings.TrimSuffix(strings.TrimRight(c.GetEndpoint(), "/"), "/v1") + "/api/status"
}

func (c *Channel) GetAgnesSubscriptionEndpoint() string {
	return strings.TrimSuffix(strings.TrimRight(c.GetEndpoint(), "/"), "/v1") + "/v1/dashboard/billing/subscription"
}

func (c *Channel) isAgnesSubscriptionChannel() bool {
	domain := strings.ToLower(strings.TrimSpace(c.GetDomain()))
	return domain == "apihub.agnes-ai.com" || strings.HasSuffix(domain, ".apihub.agnes-ai.com")
}

func (c *Channel) GetSkillAPIBaseEndpoint() string {
	endpoint := strings.TrimRight(c.GetEndpoint(), "/")
	endpoint = strings.TrimSuffix(endpoint, "/api/v1")
	endpoint = strings.TrimSuffix(endpoint, "/v1")
	endpoint = strings.TrimSuffix(endpoint, "/api")
	return endpoint
}

func (c *Channel) GetSkillAPIModelsEndpoint() string {
	return c.GetSkillAPIBaseEndpoint() + "/v1/skills/models"
}

func (c *Channel) GetSkillAPIPriceEndpoint(model string) string {
	return c.GetSkillAPIModelsEndpoint() + "/" + url.PathEscape(model) + "/pricing?status=active"
}

func (c *Channel) getSkillAPIModelsEndpoints() []string {
	base := c.GetSkillAPIBaseEndpoint()
	return []string{
		base + "/v1/skills/models",
		base + "/api/v1/skills/models",
	}
}

func (c *Channel) getSkillAPIPriceEndpoints(model string) []string {
	escapedModel := url.PathEscape(model)
	endpoints := make([]string, 0, 2)
	for _, endpoint := range c.getSkillAPIModelsEndpoints() {
		endpoints = append(endpoints, endpoint+"/"+escapedModel+"/pricing?status=active")
	}
	return endpoints
}

func (c *Channel) getUpstreamHeaders() map[string]string {
	secret := c.GetRandomSecret()
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", secret),
		"x-api-key":     secret,
	}
}

func (c *Channel) isHaoduomiProvider() bool {
	endpoint := strings.ToLower(c.GetEndpoint())
	return strings.Contains(endpoint, "lk888.ai") || strings.Contains(endpoint, "haoduomi.ai")
}

func (c *Channel) FetchModels() ([]string, error) {
	models := make([]string, 0)
	headers := c.getUpstreamHeaders()
	var lastErr error

	// Standard OpenAI-compatible model discovery remains the primary source.
	if data, err := utils.Get(c.GetAutoModelsEndpoint(), headers, c.GetProxy()); err == nil {
		form := utils.MapToStruct[struct {
			Data []interface{} `json:"data"`
		}](data)
		if form != nil {
			for _, raw := range form.Data {
				var model string
				switch item := raw.(type) {
				case string:
					model = item
				case map[string]interface{}:
					model = utils.ToString(item["id"])
				}
				model = strings.TrimSpace(model)
				if model != "" && !utils.Contains(model, models) {
					models = append(models, model)
				}
			}
		}
	} else {
		lastErr = err
	}

	// Channel providers implementing the Skill API expose media/video models
	// separately, so merge that list instead of losing those models.
	var skillEndpoints []string
	if c.isHaoduomiProvider() {
		skillEndpoints = c.getSkillAPIModelsEndpoints()
	}
	for _, endpoint := range skillEndpoints {
		data, err := utils.Get(endpoint, headers, c.GetProxy())
		if err != nil {
			lastErr = err
			continue
		}
		form := utils.MapToStruct[skillAPIModelsResponse](data)
		if form == nil {
			continue
		}
		for _, item := range form.Models {
			model := strings.TrimSpace(item.Name)
			if model != "" && !utils.Contains(model, models) {
				models = append(models, model)
			}
		}
		break
	}

	if len(models) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errors.New("upstream model list is empty")
	}
	return models, nil
}

func (c *Channel) FetchCharge() (ChargeSequence, error) {
	if len(c.GetModels()) == 0 {
		return nil, errors.New("channel has no selected models")
	}

	headers := c.getUpstreamHeaders()
	data, chargeErr := utils.Get(c.GetAutoPriceEndpoint(), headers, c.GetProxy())
	if chargeErr == nil {
		if charge, err := c.parseChargeSequence(data); err == nil {
			if filtered, filterErr := c.filterChargeSequence(charge); filterErr == nil {
				return c.bindChargeChannel(filtered), nil
			} else {
				chargeErr = filterErr
			}
		} else {
			chargeErr = err
		}
	}

	data, pricingErr := utils.Get(c.GetNewAPIPriceEndpoint(), headers, c.GetProxy())
	if pricingErr == nil {
		if charge, err := c.parseNewAPIPrice(data, headers); err == nil {
			if filtered, filterErr := c.filterChargeSequence(charge); filterErr == nil {
				return c.bindChargeChannel(filtered), nil
			} else {
				pricingErr = filterErr
			}
		} else {
			pricingErr = err
		}
	}

	agnesCharge, agnesErr := c.fetchAgnesSubscriptionCharge(headers)
	if agnesErr == nil {
		return c.bindChargeChannel(agnesCharge), nil
	}

	charge, skillErr := c.fetchSkillAPICharge(headers)
	if skillErr == nil {
		filtered, err := c.filterChargeSequence(charge)
		if err != nil {
			return nil, err
		}
		return c.bindChargeChannel(filtered), nil
	}
	return nil, fmt.Errorf(
		"cannot fetch upstream price list: charge endpoint: %v; New API pricing endpoint: %v; Skill API pricing endpoint: %v; Agnes subscription endpoint: %v",
		chargeErr, pricingErr, skillErr, agnesErr,
	)
}

type agnesSubscriptionResponse struct {
	Object             string `json:"object"`
	HasPaymentMethod   bool   `json:"has_payment_method"`
	HardLimitUSD       int64  `json:"hard_limit_usd"`
	SystemHardLimitUSD int64  `json:"system_hard_limit_usd"`
}

func (c *Channel) fetchAgnesSubscriptionCharge(headers map[string]string) (ChargeSequence, error) {
	if !c.isAgnesSubscriptionChannel() {
		return nil, errors.New("channel is not an Agnes subscription API endpoint")
	}
	data, err := utils.Get(c.GetAgnesSubscriptionEndpoint(), headers, c.GetProxy())
	if err != nil {
		return nil, fmt.Errorf("cannot fetch Agnes subscription: %w", err)
	}
	return c.parseAgnesSubscriptionCharge(data)
}

func (c *Channel) parseAgnesSubscriptionCharge(data interface{}) (ChargeSequence, error) {
	subscription := utils.MapToStruct[agnesSubscriptionResponse](data)
	if subscription == nil || subscription.Object != "billing_subscription" {
		return nil, errors.New("invalid Agnes subscription response")
	}

	models := append([]string(nil), c.GetModels()...)
	if len(models) == 0 {
		return nil, errors.New("Agnes subscription channel has no models")
	}
	return ChargeSequence{&Charge{
		Type:   globals.NonBilling,
		Models: models,
	}}, nil
}

func (c *Channel) bindChargeChannel(charge ChargeSequence) ChargeSequence {
	channelId := c.GetId()
	for _, item := range charge {
		if item != nil {
			item.ChannelId = utils.ToPtr(channelId)
		}
	}
	return charge
}

func (c *Channel) filterChargeSequence(charge ChargeSequence) (ChargeSequence, error) {
	selected := c.GetModels()
	result := make(ChargeSequence, 0, len(charge))
	for _, item := range charge {
		if item == nil {
			continue
		}

		models := utils.Filter(item.GetModels(), func(model string) bool {
			return utils.Contains(model, selected)
		})
		if len(models) == 0 {
			continue
		}

		instance := item.New("")
		instance.Models = models
		result = append(result, instance)
	}
	if len(result) == 0 {
		return nil, errors.New("upstream price list has no prices for the channel's selected models")
	}
	return result, nil
}

func (c *Channel) parseChargeSequence(data interface{}) (ChargeSequence, error) {
	charge := utils.MapToStruct[ChargeSequence](data)
	if charge == nil {
		return nil, errors.New("cannot parse upstream price list")
	}

	result := make(ChargeSequence, 0, len(*charge))
	for _, item := range *charge {
		if item == nil || len(item.Models) == 0 {
			continue
		}
		switch item.GetType() {
		case globals.NonBilling, globals.TimesBilling, globals.TokenBilling:
		default:
			continue
		}

		instance := item.New("")
		instance.Models = append([]string(nil), item.Models...)
		switch instance.GetType() {
		case globals.TokenBilling:
			instance.Input = adjustedPrice(instance.Input, c.PriceAdjust)
			instance.Output = adjustedPrice(instance.Output, c.PriceAdjust)
		case globals.TimesBilling:
			instance.Output = adjustedPrice(instance.Output, c.PriceAdjust)
		}
		result = append(result, instance)
	}

	if len(result) == 0 {
		return nil, errors.New("upstream price list is empty")
	}
	return result, nil
}

type newAPIPriceItem struct {
	ModelName       string  `json:"model_name"`
	QuotaType       int     `json:"quota_type"`
	ModelRatio      float64 `json:"model_ratio"`
	ModelPrice      float64 `json:"model_price"`
	CompletionRatio float64 `json:"completion_ratio"`
}

type newAPIPricingResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    []newAPIPriceItem `json:"data"`
}

type newAPIStatusResponse struct {
	Success bool `json:"success"`
	Data    struct {
		QuotaPerUnit    float64 `json:"quota_per_unit"`
		USDExchangeRate float64 `json:"usd_exchange_rate"`
	} `json:"data"`
}

type skillAPIModel struct {
	Name string `json:"name"`
}

type skillAPIModelsResponse struct {
	Models []skillAPIModel `json:"models"`
}

type skillAPIChannelGroup struct {
	IsActive         bool    `json:"is_active"`
	InKeyWhitelist   bool    `json:"in_key_whitelist"`
	BillingMethod    string  `json:"billing_method"`
	BasePrice        float64 `json:"base_price"`
	InputTokenPrice  float64 `json:"input_token_price"`
	OutputTokenPrice float64 `json:"output_token_price"`
}

type skillAPIPriceResponse struct {
	AvailableForThisKey bool                   `json:"available_for_this_key"`
	Name                string                 `json:"name"`
	ChannelGroups       []skillAPIChannelGroup `json:"channel_groups"`
}

func (c *Channel) fetchSkillAPICharge(headers map[string]string) (ChargeSequence, error) {
	models := append([]string(nil), c.GetModels()...)
	if len(models) == 0 {
		var data interface{}
		var lastErr error
		for _, endpoint := range c.getSkillAPIModelsEndpoints() {
			data, lastErr = utils.Get(endpoint, headers, c.GetProxy())
			if lastErr == nil {
				break
			}
		}
		if lastErr != nil {
			return nil, fmt.Errorf("cannot fetch Skill API model list: %w", lastErr)
		}
		form := utils.MapToStruct[skillAPIModelsResponse](data)
		if form == nil {
			return nil, errors.New("cannot parse Skill API model list")
		}
		for _, item := range form.Models {
			model := strings.TrimSpace(item.Name)
			if model != "" && !utils.Contains(model, models) {
				models = append(models, model)
			}
		}
	}
	if len(models) == 0 {
		return nil, errors.New("Skill API model list is empty")
	}

	type skillAPIPriceJob struct {
		index int
		model string
	}
	type skillAPIPriceResult struct {
		index  int
		charge *Charge
		err    error
	}

	workerCount := len(models)
	if workerCount > 8 {
		workerCount = 8
	}
	jobs := make(chan skillAPIPriceJob)
	prices := make(chan skillAPIPriceResult, len(models))
	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				var lastErr error
				for _, endpoint := range c.getSkillAPIPriceEndpoints(job.model) {
					data, err := utils.Get(endpoint, headers, c.GetProxy())
					if err != nil {
						lastErr = err
						continue
					}
					charge, err := c.parseSkillAPIPrice(data, job.model)
					if err != nil {
						lastErr = err
						continue
					}
					prices <- skillAPIPriceResult{index: job.index, charge: charge}
					lastErr = nil
					break
				}
				if lastErr != nil {
					prices <- skillAPIPriceResult{index: job.index, err: lastErr}
				}
			}
		}()
	}
	go func() {
		for index, model := range models {
			jobs <- skillAPIPriceJob{index: index, model: model}
		}
		close(jobs)
		workers.Wait()
		close(prices)
	}()

	ordered := make(ChargeSequence, len(models))
	var lastErr error
	for price := range prices {
		if price.err != nil {
			lastErr = price.err
			continue
		}
		ordered[price.index] = price.charge
	}
	result := make(ChargeSequence, 0, len(models))
	for _, price := range ordered {
		if price != nil {
			result = append(result, price)
		}
	}

	if len(result) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("Skill API pricing list has no supported model prices: %w", lastErr)
		}
		return nil, errors.New("Skill API pricing list has no supported model prices")
	}
	return result, nil
}

func (c *Channel) parseSkillAPIPrice(data interface{}, fallbackModel string) (*Charge, error) {
	pricing := utils.MapToStruct[skillAPIPriceResponse](data)
	if pricing == nil || !pricing.AvailableForThisKey {
		return nil, errors.New("Skill API model is not available for this key")
	}

	model := strings.TrimSpace(pricing.Name)
	if model == "" {
		model = strings.TrimSpace(fallbackModel)
	}
	if model == "" {
		return nil, errors.New("Skill API pricing response has no model name")
	}

	var best *Charge
	var bestCost float32
	for _, group := range pricing.ChannelGroups {
		if !group.IsActive || !group.InKeyWhitelist {
			continue
		}

		var charge *Charge
		switch group.BillingMethod {
		case "按token":
			if group.InputTokenPrice < 0 || group.OutputTokenPrice < 0 {
				continue
			}
			charge = &Charge{
				Type:   globals.TokenBilling,
				Models: []string{model},
				Input:  adjustedPrice(float32(group.InputTokenPrice/1000), c.PriceAdjust),
				Output: adjustedPrice(float32(group.OutputTokenPrice/1000), c.PriceAdjust),
			}
		case "按次":
			if group.BasePrice < 0 {
				continue
			}
			charge = &Charge{
				Type:   globals.TimesBilling,
				Models: []string{model},
				Output: adjustedPrice(float32(group.BasePrice), c.PriceAdjust),
			}
		default:
			// CoAI currently has no per-second billing type. Skipping is safer
			// than storing a per-second price as a per-request price.
			continue
		}

		cost := charge.GetLimit()
		if best == nil || cost < bestCost {
			best = charge
			bestCost = cost
		}
	}

	if best == nil {
		return nil, errors.New("Skill API pricing response has no supported active channel group")
	}
	return best, nil
}

func (c *Channel) parseNewAPIPrice(data interface{}, headers map[string]string) (ChargeSequence, error) {
	pricing := utils.MapToStruct[newAPIPricingResponse](data)
	if pricing == nil || len(pricing.Data) == 0 {
		return nil, errors.New("New API pricing list is empty")
	}
	if !pricing.Success && pricing.Message != "" {
		return nil, errors.New(pricing.Message)
	}

	statusData, err := utils.Get(c.GetNewAPIStatusEndpoint(), headers, c.GetProxy())
	if err != nil {
		return nil, fmt.Errorf("cannot fetch New API status: %w", err)
	}
	status := utils.MapToStruct[newAPIStatusResponse](statusData)
	if status == nil || status.Data.QuotaPerUnit <= 0 {
		return nil, errors.New("New API status does not contain a valid quota_per_unit")
	}

	usdExchangeRate := status.Data.USDExchangeRate
	if usdExchangeRate <= 0 {
		usdExchangeRate = 1
	}
	// CoAI stores 10 quota for each CNY. New API token ratios are based on
	// (1,000,000 / quota_per_unit) USD per million tokens.
	quotaPerUSD := 10 * usdExchangeRate
	baseQuotaPer1KTokens := (1000 / status.Data.QuotaPerUnit) * quotaPerUSD

	result := make(ChargeSequence, 0, len(pricing.Data))
	for _, item := range pricing.Data {
		model := strings.TrimSpace(item.ModelName)
		if model == "" {
			continue
		}

		var charge *Charge
		switch item.QuotaType {
		case 0:
			if item.ModelRatio < 0 || item.CompletionRatio < 0 {
				continue
			}
			input := float32(item.ModelRatio * baseQuotaPer1KTokens)
			output := float32(item.ModelRatio * item.CompletionRatio * baseQuotaPer1KTokens)
			charge = &Charge{
				Type:   globals.TokenBilling,
				Models: []string{model},
				Input:  adjustedPrice(input, c.PriceAdjust),
				Output: adjustedPrice(output, c.PriceAdjust),
			}
		case 1:
			if item.ModelPrice < 0 {
				continue
			}
			price := float32(item.ModelPrice * quotaPerUSD)
			charge = &Charge{
				Type:   globals.TimesBilling,
				Models: []string{model},
				Output: adjustedPrice(price, c.PriceAdjust),
			}
		default:
			continue
		}
		result = append(result, charge)
	}

	if len(result) == 0 {
		return nil, errors.New("New API pricing list has no supported model prices")
	}
	return result, nil
}

func adjustedPrice(price, adjustment float32) float32 {
	price += adjustment
	if price < 0 {
		return 0
	}
	return price
}

func (c *Channel) Load() {
	reflect := make(map[string]string)
	exclude := make([]string, 0)
	if c.AutoModels {
		if models, err := c.FetchModels(); err == nil {
			c.Models = models
			globals.Info(fmt.Sprintf("[channel] loaded %d upstream model(s) for channel #%d", len(models), c.GetId()))
		} else {
			globals.Warn(fmt.Sprintf("[channel] failed to load upstream models for channel #%d: %s", c.GetId(), err.Error()))
		}
	}
	models := c.GetModels()

	arr := strings.Split(c.GetMapper(), "\n")
	for _, item := range arr {
		pair := strings.Split(item, ">")
		if len(pair) != 2 {
			continue
		}

		from, to := pair[0], pair[1]
		if strings.HasPrefix(from, "!") {
			from = strings.TrimPrefix(from, "!")
			exclude = append(exclude, to)
		}

		reflect[from] = to
	}

	c.Reflect = &reflect
	c.ExcludeModels = &exclude

	var hits []string

	for _, model := range models {
		if !utils.Contains(model, hits) && !utils.Contains(model, exclude) {
			hits = append(hits, model)
		}
	}

	for model := range reflect {
		if !utils.Contains(model, hits) && !utils.Contains(model, exclude) {
			hits = append(hits, model)
		}
	}

	c.HitModels = &hits
}

func (c *Channel) GetReflect() map[string]string {
	return *c.Reflect
}

func (c *Channel) GetExcludeModels() []string {
	return *c.ExcludeModels
}

// GetModelReflect returns the reflection model name if it exists, otherwise returns the original model name
func (c *Channel) GetModelReflect(model string) string {
	ref := c.GetReflect()
	if reflect, ok := ref[model]; ok && len(reflect) > 0 {
		return reflect
	}

	return model
}

func (c *Channel) GetHitModels() []string {
	return *c.HitModels
}

func (c *Channel) GetState() bool {
	return c.State
}

func (c *Channel) GetGroup() []string {
	return c.Group
}

func (c *Channel) GetProxy() globals.ProxyConfig {
	return c.Proxy
}

func (c *Channel) IsHitGroup(group string) bool {
	if len(c.GetGroup()) == 0 {
		return true
	}

	return utils.Contains(group, c.GetGroup())
}

func (c *Channel) IsHit(model string) bool {
	return utils.Contains(model, c.GetHitModels())
}

func (c *Channel) ProcessError(err error) error {
	if err == nil {
		return nil
	}
	content := err.Error()

	if strings.Contains(content, c.GetEndpoint()) {
		// hide the endpoint
		replacer := fmt.Sprintf("channel://%d", c.GetId())
		content = strings.Replace(content, c.GetEndpoint(), replacer, -1)
	}

	if domain := c.GetDomain(); len(strings.TrimSpace(domain)) > 0 && strings.Contains(content, domain) {
		content = strings.Replace(content, domain, "channel", -1)
	}

	for _, item := range defaultReplacer {
		content = strings.Replace(content, item, "chatnio_upstream", -1)
	}

	secret := c.GetCurrentSecret()
	if secret != nil {
		content = strings.Replace(content, *secret, utils.HideSecret(*secret), -1)
	}

	return errors.New(content)
}
