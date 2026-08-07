package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// wechatSession 微信 code2session 结果
type wechatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// wechatCode2Session 用 wx.login 的 code 换取 openid + session_key。
// 返回错误时 err 包含微信错误码；session_key 为空属正常（不参与用户态签名时无需）。
func wechatCode2Session(appID, secret, code string) (wechatSession, error) {
	var ws wechatSession
	if appID == "" || secret == "" {
		return ws, fmt.Errorf("微信小程序未配置 WX_APPID/WX_SECRET")
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, secret, code)
	resp, err := http.Get(url)
	if err != nil {
		return ws, fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &ws); err != nil {
		return ws, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if ws.ErrCode != 0 {
		return ws, fmt.Errorf("微信错误 %d: %s", ws.ErrCode, ws.ErrMsg)
	}
	if ws.OpenID == "" {
		return ws, fmt.Errorf("openid 为空")
	}
	return ws, nil
}

// firstOpenID 兼容历史脏数据：openid 可能为逗号分隔多值（账号合并遗留），取第一个
func firstOpenID(openid string) string {
	for i := 0; i < len(openid); i++ {
		if openid[i] == ',' {
			return openid[:i]
		}
	}
	return openid
}
