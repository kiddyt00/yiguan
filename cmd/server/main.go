package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kiddyt00/yiguan/internal/handler"
	"github.com/kiddyt00/yiguan/internal/llm"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
	sqlitestore "github.com/kiddyt00/yiguan/internal/store/sqlite"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
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

	// 数据库
	st, err := sqlitestore.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer st.Close()

	// 从配置文件导入 LLM provider 到数据库（如果数据库为空）
	seedLLMProviders(st, cfg)

	// LLM 管理器（从数据库读取配置）
	llmMgr := llm.NewManager(st)
	llmClient, err := llmMgr.GetDefault()
	if err != nil {
		log.Fatalf("LLM 初始化失败: %v", err)
	}

	// 中间件
	authMW := middleware.AuthRequired(cfg.JWTSecret)

	// 路由
	mux := http.NewServeMux()
	mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(204)
	})

	authHandler := handler.NewAuthHandler(st, cfg.JWTSecret)
	mux.Handle("/api/auth/", corsWrap(authHandler.ServeMux()))

	uh := handler.NewUserHandler(st)
	hh := handler.NewHistoryHandler(st)
	dh := handler.NewDivineHandler(st, llmClient)
	ah := handler.NewAdminHandler(st, llmMgr)

	mux.Handle("GET /api/user", authMW(corsWrap(http.HandlerFunc(uh.GetUser))))
	mux.Handle("PUT /api/user", authMW(corsWrap(http.HandlerFunc(uh.UpdateUser))))
	mux.Handle("GET /api/history", authMW(corsWrap(http.HandlerFunc(hh.GetHistory))))
	mux.Handle("POST /api/divine", authMW(corsWrap(dh)))
	mux.Handle("GET /api/admin/dashboard", authMW(corsWrap(http.HandlerFunc(ah.Dashboard))))
	mux.Handle("GET /api/admin/users", authMW(corsWrap(http.HandlerFunc(ah.ListUsers))))

	// LLM Provider 管理路由
	mux.Handle("GET /api/admin/llm", authMW(corsWrap(http.HandlerFunc(ah.ListLLMProviders))))
	mux.Handle("POST /api/admin/llm", authMW(corsWrap(http.HandlerFunc(ah.CreateLLMProvider))))
	mux.Handle("PUT /api/admin/llm/", authMW(corsWrap(http.HandlerFunc(ah.UpdateLLMProvider))))
	mux.Handle("DELETE /api/admin/llm/", authMW(corsWrap(http.HandlerFunc(ah.DeleteLLMProvider))))

	log.Printf("☯ 易观 v2.0 http://localhost:%s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, mux))
}

// seedLLMProviders 首次启动时从 config.yaml 导入 LLM 配置到数据库
func seedLLMProviders(st store.Store, cfg *Config) {
	providers, err := st.ListLLMProviders()
	if err != nil || len(providers) > 0 {
		return // 已有数据，跳过
	}

	// 导入 config.yaml 中的 providers
	for key, p := range cfg.LLM.Providers {
		provider := &store.LLMProvider{
			Name:      key,
			Provider:  key,
			APIKey:    p.APIKey,
			Endpoint:  p.Endpoint,
			Model:     p.Model,
			IsDefault: key == cfg.LLM.Default,
		}
		if err := st.CreateLLMProvider(provider); err != nil {
			log.Printf("[seed] 导入 LLM provider '%s' 失败: %v", key, err)
		} else {
			log.Printf("[seed] 已导入 LLM provider: %s (%s)", key, p.Model)
		}
	}
}

func corsWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
