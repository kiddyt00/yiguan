package main

import (
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Qianwen struct {
		APIKey   string `yaml:"api_key"`
		Model    string `yaml:"model"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"qianwen"`
}

func loadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	return &cfg
}

func main() {
	cfg := loadConfig("config.yaml")
	if key := os.Getenv("DASHSCOPE_API_KEY"); key != "" {
		cfg.Qianwen.APIKey = key
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("易观 API OK"))
	})

	log.Printf("☯ 易观服务启动 http://localhost:%s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}
