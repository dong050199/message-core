package platform

type GatewayValidationRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type PlatformBaseResponse struct {
	StatusCode int                  `json:"status_code"`
	Code       int                  `json:"code"`
	Message    string               `json:"message"`
	TraceID    string               `json:"trace_id"`
	Data       RulesDevicesResponse `json:"data"`
}

type RulesDevicesResponse struct {
	RulesDevices []RulesDevices `json:"rules"`
}

type RulesDevices struct {
	Atribute   string `json:"attribute"`
	Comparison string `json:"comparison"`
	RuleValue  string `json:"rule_value"`
}
