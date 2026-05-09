package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	Admin     struct {
		Phone    string `yaml:"phone"`
		Password string `yaml:"password"`
	} `yaml:"admin"`
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
	if phone := os.Getenv("ADMIN_PHONE"); phone != "" {
		cfg.Admin.Phone = phone
	}
	if pwd := os.Getenv("ADMIN_PASSWORD"); pwd != "" {
		cfg.Admin.Password = pwd
	}

	// 数据库
	st, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer st.Close()

	// 管理员初始化
	if phone := cfg.Admin.Phone; phone != "" {
		existing, _ := st.GetUserByPhone(phone)
		if existing != nil {
			if existing.Role != "admin" {
				st.UpdateUserRole(existing.ID, "admin")
				log.Printf("已将用户 %s 升级为管理员", phone)
			}
		} else {
			_, err := st.CreateUser(phone, cfg.Admin.Password, "管理员")
			if err != nil {
				log.Printf("创建管理员失败: %v", err)
			} else {
				st.UpdateUserRole(1, "admin")
				log.Printf("已创建管理员账号: %s", phone)
			}
		}
	}

	// LLM Router（模型热切换，数据库无模型时用 config 兜底）
	defaultProvider := cfg.LLM.Providers[cfg.LLM.Default]
	llmRouter, err := llm.NewRouterWithFallback(st, llm.Config{
		APIKey:   defaultProvider.APIKey,
		Endpoint: defaultProvider.Endpoint,
		Model:    defaultProvider.Model,
	})
	if err != nil {
		log.Fatalf("LLM 路由器初始化失败: %v", err)
	}
	log.Printf("LLM Router: %s (%s)", cfg.LLM.Default, llmRouter.Get().ModelName())

	// 中间件
	authMW := middleware.AuthRequired(cfg.JWTSecret)
	adminMW := middleware.AdminOnly(cfg.JWTSecret)

	// 路由
	mux := http.NewServeMux()

	// 健康检查（无需鉴权）
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	// OPTIONS preflight 已统一由 corsWrap 中间件处理（见文件底部）

	authHandler := handler.NewAuthHandler(st, cfg.JWTSecret)
	if appID := os.Getenv("WX_APPID"); appID != "" {
		authHandler.SetWechatConfig(appID, os.Getenv("WX_SECRET"))
		log.Printf("微信小程序登录已配置: %s", appID)
	}
	if appID := os.Getenv("WX_OPEN_APPID"); appID != "" {
		authHandler.SetWechatOpenConfig(appID, os.Getenv("WX_OPEN_SECRET"))
		log.Printf("微信开放平台扫码登录已配置: %s", appID)
	}
	mux.Handle("/api/auth/", corsWrap(authHandler.ServeMux()))

	uh := handler.NewUserHandler(st)
	hh := handler.NewHistoryHandler(st)
	dh := handler.NewDivineHandler(st, llmRouter)
	ah := handler.NewAdminHandler(st)

	// 用户端
	mux.Handle("GET /api/user", authMW(corsWrap(http.HandlerFunc(uh.GetUser))))
	mux.Handle("PUT /api/user", authMW(corsWrap(http.HandlerFunc(uh.UpdateUser))))
	mux.Handle("GET /api/history", authMW(corsWrap(http.HandlerFunc(hh.GetHistory))))
	mux.Handle("POST /api/divine", authMW(corsWrap(dh)))

	// SSE 流式起卦
	streamHandler := handler.NewDivineStreamHandler(st, llmRouter)
	mux.Handle("POST /api/divine/stream", authMW(corsWrap(streamHandler)))

	// 后台管理
	mux.Handle("GET /api/admin/dashboard", adminMW(corsWrap(http.HandlerFunc(ah.Dashboard))))
	mux.Handle("GET /api/admin/users", adminMW(corsWrap(http.HandlerFunc(ah.ListUsers))))
	mux.Handle("POST /api/admin/users/{id}/toggle", adminMW(corsWrap(http.HandlerFunc(ah.ToggleUser))))
	mux.Handle("POST /api/admin/users/{id}/quota", adminMW(corsWrap(http.HandlerFunc(ah.AdjustUserQuota))))
	mux.Handle("GET /api/admin/users/{id}/history", adminMW(corsWrap(http.HandlerFunc(ah.GetUserHistory))))

	// 卦象记录管理
	hh2 := handler.NewHexagramHandler(st)
	mux.Handle("GET /api/admin/hexagrams", adminMW(corsWrap(http.HandlerFunc(hh2.ListHistory))))
	mux.Handle("GET /api/admin/hexagrams/{id}", adminMW(corsWrap(http.HandlerFunc(hh2.GetHistoryDetail))))
	mux.Handle("DELETE /api/admin/hexagrams/{id}", adminMW(corsWrap(http.HandlerFunc(hh2.DeleteHistory))))

	// 模型管理
	mh := handler.NewModelHandler(st, func() { _ = llmRouter.Reload() })
	mux.Handle("GET /api/admin/models", adminMW(corsWrap(http.HandlerFunc(mh.ListModels))))
	mux.Handle("POST /api/admin/models", adminMW(corsWrap(http.HandlerFunc(mh.CreateModel))))
	mux.Handle("PUT /api/admin/models/{id}", adminMW(corsWrap(http.HandlerFunc(mh.UpdateModel))))
	mux.Handle("DELETE /api/admin/models/{id}", adminMW(corsWrap(http.HandlerFunc(mh.DeleteModel))))
	mux.Handle("POST /api/admin/models/{id}/set-default", adminMW(corsWrap(http.HandlerFunc(mh.SetDefaultModel))))
	mux.Handle("POST /api/admin/models/{id}/toggle", adminMW(corsWrap(http.HandlerFunc(mh.ToggleModel))))
	mux.Handle("POST /api/admin/models/fetch", adminMW(corsWrap(http.HandlerFunc(mh.FetchModels))))
	mux.Handle("POST /api/admin/models/test", adminMW(corsWrap(http.HandlerFunc(mh.TestConnection))))

	// 广告管理
	adH := handler.NewAdHandler(st)
	mux.Handle("GET /api/admin/ads", adminMW(corsWrap(http.HandlerFunc(adH.ListAds))))
	mux.Handle("POST /api/admin/ads", adminMW(corsWrap(http.HandlerFunc(adH.CreateAd))))
	mux.Handle("PUT /api/admin/ads/{id}", adminMW(corsWrap(http.HandlerFunc(adH.UpdateAd))))
	mux.Handle("DELETE /api/admin/ads/{id}", adminMW(corsWrap(http.HandlerFunc(adH.DeleteAd))))
	mux.Handle("POST /api/admin/ads/{id}/toggle", adminMW(corsWrap(http.HandlerFunc(adH.ToggleAd))))
	mux.Handle("GET /api/admin/ads/stats", adminMW(corsWrap(http.HandlerFunc(adH.GetAdStats))))
	mux.Handle("GET /api/ads/active", corsWrap(http.HandlerFunc(adH.ListActiveAds)))
	mux.Handle("POST /api/ads/{id}/watch", authMW(corsWrap(http.HandlerFunc(adH.StartWatch))))
	mux.Handle("POST /api/ads/{id}/complete", authMW(corsWrap(http.HandlerFunc(adH.CompleteWatch))))

	// 日志中间件
	logMux := loggingMiddleware(mux)

	// 带超时配置的 HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      logMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second, // SSE 流式响应需要较长写超时
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Printf("收到退出信号，正在关闭服务...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("关闭服务出错: %v", err)
		}
	}()

	log.Printf("☯ 易观 v2.1 http://localhost:%s", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
	log.Println("服务已关闭")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
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
