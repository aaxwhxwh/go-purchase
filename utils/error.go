package utils

const (
	RECODE_OK         = "0"
	RECODE_DBERR      = "401"
	RECODE_NODATA     = "402"
	RECODE_DATAEXIST  = "403"
	RECODE_DATAERR    = "404"
	RECODE_SESSIONERR = "411"
	RECODE_LOGINERR   = "412"
	RECODE_PARAMERR   = "413"
	RECODE_USERERR    = "414"
	RECODE_ROLEERR    = "415"
	RECODE_PWDERR     = "416"
	RECODE_SMSERR     = "417"
	RECODE_REQERR     = "421"
	RECODE_IPERR      = "422"
	RECODE_THIRDERR   = "431"
	RECODE_IOERR      = "432"
	RECODE_SERVERERR  = "450"
	RECODE_UNKNOWERR  = "451"
)

var recodeText = map[string]string{
	RECODE_OK:         "成功",
	RECODE_DBERR:      "数据库查询错误",
	RECODE_NODATA:     "无数据",
	RECODE_DATAEXIST:  "数据已存在",
	RECODE_DATAERR:    "数据错误",
	RECODE_SESSIONERR: "用户未登录",
	RECODE_LOGINERR:   "用户登录失败",
	RECODE_PARAMERR:   "参数错误",
	RECODE_USERERR:    "用户不存在或未激活",
	RECODE_ROLEERR:    "用户身份错误",
	RECODE_PWDERR:     "密码错误",
	RECODE_REQERR:     "非法请求或请求次数受限",
	RECODE_IPERR:      "IP受限",
	RECODE_THIRDERR:   "第三方系统错误",
	RECODE_IOERR:      "文件读写错误",
	RECODE_SERVERERR:  "内部错误",
	RECODE_UNKNOWERR:  "未知错误",
	RECODE_SMSERR:     "短信失败",
}

func RecodeText(code string) string {
	str, ok := recodeText[code]
	if ok {
		return str
	}
	return recodeText[RECODE_UNKNOWERR]
}
