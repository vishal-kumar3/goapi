package main

import "os"

type EnvConfig struct {
	POSTGRESUSER     string
	POSTGRESPASSWORD string
	POSTGRESDB       string
}

func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func LoadEnvConfig() EnvConfig {
	return EnvConfig{
		POSTGRESUSER:     GetEnv("POSTGRESUSER", "godb"),
		POSTGRESPASSWORD: GetEnv("POSTGRESPASSWORD", "godb"),
		POSTGRESDB:       GetEnv("POSTGRESDB", "godb"),
	}
}
