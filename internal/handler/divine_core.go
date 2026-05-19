package handler

import (
	"encoding/json"
	"net/http"

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
	remaining,err:=st.GetRemainingQuota(userID)
	if err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"查询配额失败"});return nil}
	if remaining<=0{writeJSON(w,http.StatusPaymentRequired,map[string]interface{}{"error":"次数不足","remaining_quota":0});return nil}
	var req divineReq
	if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{writeJSON(w,http.StatusBadRequest,map[string]string{"error":"请求格式错误"});return nil}
	if req.Question==""{writeJSON(w,http.StatusBadRequest,map[string]string{"error":"请输入问题"});return nil}
	if err:=st.ConsumeQuota(userID);err!=nil{writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"扣减配额失败"});return nil}
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
