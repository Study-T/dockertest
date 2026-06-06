package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	DataSource string `json:",optional"`
	RedisConf  struct {
		Host string
		Pass string `json:",optional"`
		Db   int    `json:",default=0"`
	}

	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
	}

	Webhook struct {
		EncryptKey     string `json:",optional"`
		ReplayWindow   int64  `json:",default=300000"`
		MaxPayloadSize int64  `json:",default=1048576"`
	}

	Sync struct {
		IntervalMinutes int `json:",default=60"`
		BatchSize       int `json:",default=100"`
	}

	Retry struct {
		IntervalMinutes int `json:",default=5"`
		MaxRetries      int `json:",default=5"`
		BatchSize       int `json:",default=50"`
	}

	Grayscale struct {
		Enabled    bool   `json:",default=false"`
		Mode       string `json:",default=whitelist"`
		Whitelist  string `json:",optional"`
		Percentage int    `json:",default=10"`
	}

	Queue struct {
		Key     string `json:",default=queue:yun_express_webhook_track"`
		Timeout int64  `json:",default=0"`
	}
}
