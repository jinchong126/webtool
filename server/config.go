package server

type Http struct {
	Server []Server `json:"server"`
}

type Server struct {
	Addr string `json:"addr"` // localhost:8080
	Type string `json:"Type"` // echo: 回应; waf: 配套pandafence测试请求和响应.
}

type ResponseCodes struct {
	RequestCode  int // 请求的响应，默认200
	ResponseCode int // 响应的响应，默认200
}
