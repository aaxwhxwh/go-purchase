// @Time    : 2019-09-02 16:51
// @Author  : Frank
// @Email   : frank@163.com
// @File    : log.go
// @Software: GoLand
package utils

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// beego 日志配置结构体
type LoggerConfig struct {
	FileName            string `json:"file_name"`
	Level               int    `json:"level"`    // 日志保存的时候的级别，默认是 Trace 级别
	Maxlines            int    `json:"maxlines"` // 每个文件保存的最大行数，默认值 1000000
	Maxsize             int    `json:"maxsize"`  // 每个文件保存的最大尺寸，默认值是 1 << 28, //256 MB
	Daily               bool   `json:"daily"`    // 是否按照每天 logrotate，默认是 true
	Maxdays             int    `json:"maxdays"`  // 文件最多保存多少天，默认保存 7 天
	Rotate              bool   `json:"rotate"`   // 是否开启 logrotate，默认是 true
	Perm                string `json:"perm"`     // 日志文件权限
	RotatePerm          string `json:"rotateperm"`
	EnableFuncCallDepth bool   `json:"-"` // 输出文件名和行号
	LogFuncCallDepth    int    `json:"-"` // 函数调用层级
	Separate            string `json:"separate"`
}

func init() {
	var logCfg = LoggerConfig{
		FileName:            beego.AppConfig.String("logfile"),
		Daily:               true,
		Level:               7,
		EnableFuncCallDepth: true,
		LogFuncCallDepth:    3,
		RotatePerm:          "777",
		Perm:                "777",
		Separate:            beego.AppConfig.String("separate"),
	}
	setting := fmt.Sprintf(`{"filename": "%s", "separate":%s}`, logCfg.FileName, logCfg.Separate)
	logs.SetLogger(logs.AdapterMultiFile, setting)
	logs.EnableFuncCallDepth(logCfg.EnableFuncCallDepth)
	logs.SetLogFuncCallDepth(logCfg.LogFuncCallDepth)
	logs.Info("program start...")
}
