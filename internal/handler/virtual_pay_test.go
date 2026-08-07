package handler

import "testing"

// 验证值来自官方文档《2.6 参考的 Python 脚本》：
//   post_body  = {"openid": "xxx", "user_ip": "127.0.0.1", "env": 0}
//   appkey     = "12345"
//   uri        = /xpay/query_user_balance
//   session_key = "9hAb/NEYUlkaMBEsmFgzig=="
func TestCalcVirtualPaySig(t *testing.T) {
	postBody := `{"openid": "xxx", "user_ip": "127.0.0.1", "env": 0}`
	got := calcVirtualPaySig("12345", "/xpay/query_user_balance", postBody)
	want := "c37809f27c6d7fd1837ad2500a04512b66b34fd793a39a385fade56dca89a4b5"
	if got != want {
		t.Errorf("calcVirtualPaySig() = %s, want %s", got, want)
	}
}

func TestCalcVirtualUserSignature(t *testing.T) {
	postBody := `{"openid": "xxx", "user_ip": "127.0.0.1", "env": 0}`
	sessionKey := "9hAb/NEYUlkaMBEsmFgzig=="
	got := calcVirtualUserSignature(sessionKey, postBody)
	want := "089d9e8dc5d308977360c4b79ec600a93d736802802a807d634192328032f6c7"
	if got != want {
		t.Errorf("calcVirtualUserSignature() = %s, want %s", got, want)
	}
}
