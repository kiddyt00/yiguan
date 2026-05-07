package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kiddyt00/yiguan/internal/handler"
	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store/sqlite"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	LLM struct {
		Default   string `yaml:"default"`
		Providers map[string]struct {
			APIKey   string `yaml:"api_key"`
			Endpoint string `yaml:"endpoint"`
			Model    string `yaml:"model"`
		} `yaml:"providers"`
	} `yaml:"llm"`
	JWTSecret string `yaml:"jwt_secret"`
	DBPath    string `yaml:"db_path"`
}

func loadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}
	cfg := &Config{
		JWTSecret: "yiguan-dev-secret",
		DBPath:    "yiguan.db",
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	return cfg
}

func main() {
	cfg := loadConfig("config.yaml")

	// 环境变量覆盖
	if s := os.Getenv("JWT_SECRET"); s != "" {
		cfg.JWTSecret = s
	}
	if key := os.Getenv("LLM_API_KEY"); key != "" {
		for k := range cfg.LLM.Providers {
			p := cfg.LLM.Providers[k]
			p.APIKey = key
			cfg.LLM.Providers[k] = p
		}
	}
	if db := os.Getenv("DB_PATH"); db != "" {
		cfg.DBPath = db
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}

	// 数据库
	st, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer st.Close()

	// LLM 客户端
	defaultProvider := cfg.LLM.Providers[cfg.LLM.Default]
	llmClient := llm.New(llm.Config{
		APIKey:   defaultProvider.APIKey,
		Endpoint: defaultProvider.Endpoint,
		Model:    defaultProvider.Model,
	})
	log.Printf("LLM: %s (%s)", cfg.LLM.Default, defaultProvider.Model)

	// 中间件
	authMW := middleware.AuthRequired(cfg.JWTSecret)

	// 路由
	mux := http.NewServeMux()
	// OPTIONS preflight 已统一由 corsWrap 中间件处理（见文件底部）

	authHandler := handler.NewAuthHandler(st, cfg.JWTSecret)
	mux.Handle("/api/auth/", corsWrap(authHandler.ServeMux()))

	uh := handler.NewUserHandler(st)
	hh := handler.NewHistoryHandler(st)
	dh := handler.NewDivineHandler(st, llmClient)
	ah := handler.NewAdminHandler(st)

	mux.Handle("GET /api/user", authMW(corsWrap(http.HandlerFunc(uh.GetUser))))
	mux.Handle("PUT /api/user", authMW(corsWrap(http.HandlerFunc(uh.UpdateUser))))
	mux.Handle("GET /api/history", authMW(corsWrap(http.HandlerFunc(hh.GetHistory))))
	mux.Handle("POST /api/divine", authMW(corsWrap(dh)))
	mux.Handle("GET /api/admin/dashboard", authMW(corsWrap(http.HandlerFunc(ah.Dashboard))))
	mux.Handle("GET /api/admin/users", authMW(corsWrap(http.HandlerFunc(ah.ListUsers))))

	log.Printf("☯ 易观 v2.0 http://localhost:%s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}

func corsWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
