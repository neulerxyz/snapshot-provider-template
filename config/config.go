// config/config.go

package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Execution (Geth) config
	GethRPCURL       string `yaml:"geth_rpc_url"`
	GethClientName   string `yaml:"geth_client_name"`
	GethServiceName  string `yaml:"geth_service_name"`
	GethDataDir      string `yaml:"geth_data_dir"`
	GethSnapshotType string `yaml:"geth_snapshot_type"`
	// Consensus (Cosmos) config
	CosmosRPCURL       string `yaml:"cosmos_rpc_url"`
	CosmosClientName   string `yaml:"cosmos_client_name"`
	CosmosServiceName  string `yaml:"cosmos_service_name"`
	CosmosDataDir      string `yaml:"cosmos_data_dir"`
	CosmosSnapshotType string `yaml:"cosmos_snapshot_type"`
	// Snapshot config
	SnapshotDir      string        `yaml:"snapshot_dir"`
	SnapshotInterval time.Duration `yaml:"snapshot_interval_hours"`
	LogFile          string        `yaml:"log_file"`
	ServerPort       int           `yaml:"server_port"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		// Execution (Geth) config
		GethRPCURL:       viper.GetString("geth_rpc_url"),
		GethClientName:   viper.GetString("geth_client_name"),
		GethServiceName:  viper.GetString("geth_service_name"),
		GethDataDir:      viper.GetString("geth_data_dir"),
		GethSnapshotType: viper.GetString("geth_snapshot_type"),
		// Consensus (Cosmos) config
		CosmosRPCURL:       viper.GetString("cosmos_rpc_url"),
		CosmosClientName:   viper.GetString("cosmos_client_name"),
		CosmosServiceName:  viper.GetString("cosmos_service_name"),
		CosmosDataDir:      viper.GetString("cosmos_data_dir"),
		CosmosSnapshotType: viper.GetString("cosmos_snapshot_type"),
		// Snapshot config
		SnapshotDir:      viper.GetString("snapshot_dir"),
		SnapshotInterval: viper.GetDuration("snapshot_interval_hours") * time.Hour,
		LogFile:          viper.GetString("log_file"),
		ServerPort:       viper.GetInt("server_port"),
	}

	return config, nil
}
