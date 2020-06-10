package tencentcloud

const (
	API_GATEWAY_KEY_ENABLED  = "on"
	API_GATEWAY_KEY_DISABLED = "off"
)

var API_GATEWAY_KEYS = []string{
	API_GATEWAY_KEY_ENABLED,
	API_GATEWAY_KEY_DISABLED,
}
var API_GATEWAY_KEY_STR2INTS = map[string]int64{
	API_GATEWAY_KEY_ENABLED:  1,
	API_GATEWAY_KEY_DISABLED: 0,
}
var API_GATEWAY_KEY_INT2STRS = map[int64]string{
	1: API_GATEWAY_KEY_ENABLED,
	0: API_GATEWAY_KEY_DISABLED,
}

const (
	API_GATEWAY_TYPE_SERVICE = "SERVICE"
	API_GATEWAY_TYPE_API     = "API"
)

var API_GATEWAY_TYPES = []string{
	API_GATEWAY_TYPE_SERVICE,
	API_GATEWAY_TYPE_API,
}
