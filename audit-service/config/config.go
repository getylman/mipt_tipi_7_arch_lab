package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
)

type Config struct {
    ServerPort int    `json:"server_port"`
    DBHost     string `json:"db_host"`
    DBPort     int    `json:"db_port"`
    DBUser     string `json:"db_user"`
    DBPassword string `json:"db_password"`
    DBName     string `json:"db_name"`
    LogLevel   string `json:"log_level"`
    AppVersion string `json:"app_version"`
}

func Load() (*Config, error) {
    port, _ := strconv.Atoi(getEnv("APP_PORT", "8080"))
    dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
    
    cfg := &Config{
        ServerPort: port,
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     dbPort,
        DBUser:     getEnv("DB_USER", "audit_user"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "audit_db"),
        LogLevel:   strings.ToUpper(getEnv("LOG_LEVEL", "INFO")),
        AppVersion: getEnv("APP_VERSION", "1.0.0"),
    }
    
    if cfg.DBPassword == "" {
        return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
    }
    
    return cfg, nil
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}