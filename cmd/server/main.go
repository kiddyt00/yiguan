package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kiddyt00/yiguan/internal/handler"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/qianwen"
	"github.com/kiddyt00/yiguan/internal/store/sqlite"
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
	if key := os.Getenv("DASHSCOPE_API_KEY"); key != "" {
		cfg.Qianwen.APIKey = key
	}
	if s := os.Getenv("JWT_SECRET"); s != "" {
		cfg.JWTSecret = s
	}

	// 数据库
	st, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer st.Close()
	log.Printf("数据库已连接: %s", cfg.DBPath)

	// 千问客户端
	qw := qianwen.NewClient(cfg.Qianwen.APIKey, cfg.Qianwen.Model, cfg.Qianwen.Endpoint)

	// 中间件
	authMW := middleware.AuthRequired(cfg.JWTSecret)

	// 路由
	mux := http.NewServeMux()

	// CORS 预检
	mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(204)
	})

	// 认证路由（无需登录）
	authHandler := handler.NewAuthHandler(st, cfg.JWTSecret)
	mux.Handle("/api/auth/", corsMiddleware(authHandler.ServeMux()))

	// 需要登录的路由
	uh := handler.NewUserHandler(st)
	hh := handler.NewHistoryHandler(st)
	dh := handler.NewDivineHandler(st, qw)

	mux.Handle("GET /api/user", authMW(corsHandler(http.HandlerFunc(uh.GetUser))))
	mux.Handle("PUT /api/user", authMW(corsHandler(http.HandlerFunc(uh.UpdateUser))))
	mux.Handle("GET /api/history", authMW(corsHandler(http.HandlerFunc(hh.GetHistory))))
	mux.Handle("POST /api/divine", authMW(corsHandler(dh)))

	log.Printf("☯ 易观 v2.0 API 启动 http://localhost:%s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(w, r)
	})
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
