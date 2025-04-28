package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

type Config struct {
	data map[string]string
}

func InitConfig() *Config {
	rootDir := getRootDir()
	err := godotenv.Load(filepath.Join(rootDir, ".env"))
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err.Error())
	}

	cfg := new(Config)
	cfg.data = make(map[string]string)
	cfg.data["ROOT_DIR"] = rootDir
	for _, item := range os.Environ() {
		splits := strings.SplitN(item, "=", 2)
		cfg.data[splits[0]] = splits[1]
	}

	return cfg
}

func (cfg *Config) GetKey(key string) string {
	value, ok := cfg.data[key]

	if !ok {
		log.Fatalf("Undefined config key: %s", key)
	}

	return value
}

func getRootDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goEnvPath := filepath.Join(currentDir, ".env")
		if _, err := os.Stat(goEnvPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return currentDir
}
