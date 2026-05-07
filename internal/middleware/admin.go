package middleware

import (
	"context"
	"net/http"
)

const RoleKey contextKey = "role"

// AdminOnly 管理员权限中间件
func AdminOnly(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearer(r)
			if tokenStr == "" {
				writeAuthError(w, "需要登录")
				return
			}

			claims, err := parseJWT(tokenStr, secret)
			if err != nil {
				writeAuthError(w, "token无效")
				return
			}

			role, _ := claims["role"].(string)
			if role != "admin" {
				writeAuthError(w, "需要管理员权限")
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				writeAuthError(w, "token无效")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
