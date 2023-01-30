package app

import "webtool/server"

type Config struct {
	Logger Logger      `json:"logger"`
	Http   server.Http `json:"http"`
}

type Logger struct {
	Level  string `json:"level"` // info, warn, error, debug
	Enable bool   `json:"enable"`
}
