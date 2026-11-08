package configs

import (
	"192.168.100.6/coreinno/common.git/configs"
)

type AppConfigs struct {
	ServiceName             string `config:"SERVICE_NAME"`
	DatabaseURI             string `config:"DB_URI"`
	DatabaseMaxConnSize     uint   `config:"DB_MAX_CONN_SIZE"`
	DatabaseMinConnSize     uint   `config:"DB_MIN_CONN_SIZE"`
	DatabaseMaxConnIdleTime string `config:"DB_MAX_CONN_IDLE_TIME"`
	DatabaseMaxConnLifeTime string `config:"DB_MAX_CONN_LIFE_TIME"`
	DatabaseConnectTimeout  string `config:"DB_CONNECT_TIMEOUT"`
	GRPCHost                string `config:"GRPC_HOST"`
	GRPCPort                uint   `config:"GRPC_PORT"`
	HTTPHost                string `config:"HTTP_HOST"`
	HTTPPort                uint   `config:"HTTP_PORT"`
	EventBrokerHost         string `config:"EVENT_BROKER_HOST"`
	EventBrokerPort         uint   `config:"EVENT_BROKER_PORT"`
	EventBrokerTimeout      string `config:"EVENT_BROKER_TIMEOUT"`
}

var ApplicationConfigs AppConfigs

func init() {
	loader := configs.New(configs.NewEnvProvider())
	err := loader.Load(&ApplicationConfigs)
	if err != nil {
		return
	}
}
