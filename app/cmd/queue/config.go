package main

type Config struct {
	Name       string `json:",default=queue-worker"`
	DataSource string `json:",optional"`
	RedisConf  struct {
		Host string
		Pass string `json:",optional"`
		Db   int    `json:",default=0"`
	}
	Queue struct {
		Key     string `json:",default=queue:yun_express_webhook_track"`
		Timeout int64  `json:",default=0"`
	}
}
