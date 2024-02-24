package cmd

import "os"

type Config struct {
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = &Config{}
	}
	return config
}

func (this *Config) GetBackupDir() string {
	backupDir := os.Getenv("BACKUP_DIR")
	if len(backupDir) == 0 {
		return "./docker-tars"
	}
	return backupDir
}
