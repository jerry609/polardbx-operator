package config

import (
	"log"
	"sync"
)

// AppConfig is the unified application configuration
// It provides a single entry point to all configuration
type AppConfig struct {
	Server     *ServerConfig
	Kubernetes *KubeConfig
	Image      *ImageRegistryConfig
	AutoFix    *AutoFixOverlayConfig
}

var (
	appConfig     *AppConfig
	appConfigOnce sync.Once
)

// GetAppConfig returns the unified application configuration (singleton)
func GetAppConfig() *AppConfig {
	appConfigOnce.Do(func() {
		appConfig = &AppConfig{
			Server:     GetServerConfig(),
			Kubernetes: GetKubeConfig(),
			Image:      GetGlobalConfig(),
			AutoFix:    GetAutoFixOverlayConfig(),
		}
		log.Println("AppConfig initialized: All configuration modules loaded")
	})
	return appConfig
}

// PrintConfig prints the current configuration (for debugging)
func (c *AppConfig) PrintConfig() {
	log.Println("=== Application Configuration ===")
	log.Printf("Server: Port=%d, Mode=%s, LogLevel=%s",
		c.Server.Port, c.Server.Mode, c.Server.LogLevel)

	configSource := "in-cluster"
	if !c.Kubernetes.IsInCluster() {
		configSource = c.Kubernetes.ConfigPath
	}
	log.Printf("Kubernetes: Source=%s, DefaultNamespace=%s",
		configSource, c.Kubernetes.DefaultNamespace)

	log.Printf("Image: DefaultRegistry=%s, Mirrors=%v",
		c.Image.GetDefaultRegistry(), c.Image.GetMirrors())

	log.Printf("AutoFix: Enabled=%t, Namespace=%s",
		c.AutoFix.Enabled, c.AutoFix.Namespace)

	log.Println("=================================")
}

// Validate validates the configuration and returns any errors
func (c *AppConfig) Validate() []string {
	var errors []string

	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errors = append(errors, "Invalid server port")
	}

	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		errors = append(errors, "Invalid GIN_MODE, must be 'debug', 'release', or 'test'")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Server.LogLevel] {
		errors = append(errors, "Invalid LOG_LEVEL, must be 'debug', 'info', 'warn', or 'error'")
	}

	return errors
}
