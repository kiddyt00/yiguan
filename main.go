package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/kiddyt00/yiguan/internal/handler"
	"github.com/kiddyt00/yiguan/internal/qianwen"
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

	// 模板函数
	funcMap := template.FuncMap{
		"yaoLabel": handler.YaoLabelFunc(),
	}

	// 解析所有模板
	tmpl := template.Must(
		template.New("").Funcs(funcMap).ParseFiles(
			"templates/layout.html",
			"templates/home.html",
			"templates/result.html",
		),
	)

	// 千问客户端
	qw := qianwen.NewClient(cfg.Qianwen.APIKey, cfg.Qianwen.Model, cfg.Qianwen.Endpoint)

	mux := http.NewServeMux()

	// 静态文件
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 页面路由
	home := &handler.HomeHandler{Tmpl: tmpl}
	divine := &handler.DivineHandler{Tmpl: tmpl, Qianwen: qw}

	mux.Handle("GET /", home)
	mux.Handle("POST /divine", divine)

	log.Printf("☯ 易观服务启动 http://localhost:%s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}
