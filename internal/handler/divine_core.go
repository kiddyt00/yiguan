package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/kiddyt00/yiguan/internal/engine"
	"github.com/kiddyt00/yiguan/internal/middleware"
	"github.com/kiddyt00/yiguan/internal/store"
)

type divineCoreResult struct {
	UserID       int64
	Question     string
	Lines        []int
	Primary      *engine.GuaInfo
	Changing     *engine.GuaInfo
	YaoPositions []yaoPos
	YaoDesc      string
	MasterYao    int
	TossData     string
}

type tossRecord struct {
	Throw      int    `json:"throw"`
	Label      string `json:"label"`
	Result     string `json:"result"`
	Sum        int    `json:"sum"`
	CoinValues []int  `json:"coin_values"`
	Yang       bool   `json:"yang"`
}

func coinsFromLine(v int) []int {
	switch v {
	case 6: return []int{2,2,2}
	case 7: return []int{2,2,3}
	case 8: return []int{2,3,3}
	case 9: return []int{3,3,3}
	default: return []int{0,0,0}
	}
}

func lineType(v int) string {
	switch v {
	case 6: return "老阴"
	case 7: return "少阳"
	case 8: return "少阴"
	case 9: return "老阳"
	default: return ""
	}
}

func formatTossData(lines []int) string {
	if len(lines)==0{return""}
	names:=[]string{"初爻","二爻","三爻","四爻","五爻","上爻"}
	tosses:=make([]tossRecord,0,len(lines))
	for i,v:=range lines{
		tosses=append(tosses,tossRecord{Throw:i+1,Label:names[i],Result:lineType(v),Sum:v,CoinValues:coinsFromLine(v),Yang:v%2!=0})
	}
	b,_:=json.Marshal(tosses)
	return string(b)
}

func divineCore(w http.ResponseWriter, r *http.Request, st store.Store) *divineCoreResult {
	userID:=r.Context().Value(middleware.UserIDKey).(int64)

	// 先解析并校验请求参数（避免无效请求浪费配额）
	var req divineReq
	if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":"请求格式错误"});return nil}
	req.Question=strings.TrimSpace(req.Question)
	if req.Question==""{writeJSON(w,http.StatusBadRequest,map[string]string{"error":"请输入问题"});return nil}
	if utf8.RuneCountInString(req.Question)>500{writeJSON(w,http.StatusBadRequest,map[string]string{"error":"问题过长，请控制在500字以内"});return nil}
	if strings.ContainsAny(req.Question,"<>"){writeJSON(w,http.StatusBadRequest,map[string]string{"error":"问题包含不支持的字符"});return nil}

	// 会员优先判定：有有效会员则直接放行
	hasMember,memberErr:=st.HasActiveMembership(userID)
	if memberErr!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"查询会员状态失败"});return nil}

	if !hasMember{
		// 非会员：按免费 quota 次数判定
		remaining,err:=st.GetRemainingQuota(userID)
		if err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"查询配额失败"});return nil}
		if remaining<=0{writeJSON(w,http.StatusPaymentRequired,map[string]interface{}{"error":"次数不足","remaining_quota":0});return nil}
		// 扣减一次 quota（原子操作，防并发双花）
		if err:=st.ConsumeQuota(userID);err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"扣减配额失败"});return nil}
	}

	linesArr:=engine.CastSixLines()
	lines:=linesArr[:]
	primary,changing,positions,master:=engine.BuildHexagrams(linesArr)
	yaoPositions:=buildYaoPositions(positions,master)
	yaoDesc:=engine.FormatYaoPositions(positions,master)
	return &divineCoreResult{
		UserID:userID,Question:req.Question,Lines:lines,
		Primary:primary,Changing:changing,
		YaoPositions:yaoPositions,YaoDesc:yaoDesc,
		MasterYao:master,TossData:formatTossData(lines),
	}
}
