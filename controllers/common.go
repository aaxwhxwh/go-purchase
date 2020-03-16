// @Time    : 2019-09-02 16:26
// @Author  : Frank
// @Email   : frank@163.com
// @File    : common.go
// @Software: GoLand
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"math"
)

// 定义通用结构体
type CommonController struct {
	beego.Controller
	Pagination
}

type Pagination struct {
	count int64
	pageCount float64
	pageIndex int
	pageSize  int
	start int
}

func (c *CommonController) ResponseData(code int, msg interface{}, data interface{}) {
	if c.count == 0 {
		RespData := map[string]interface{}{
			"code": code,
			"msg":  msg,
			"data": data,
		}
		if data == nil {
			delete(RespData, "data")
		}
		c.Data["json"] = RespData
		c.ServeJSON()
	} else {
		RespData := map[string]interface{}{
			"code": code,
			"msg":  msg,
			"total": c.count,
			"pages": c.pageCount,
			"index": c.pageIndex,
			"data": data,
		}
		if c.count == 0 {
			delete(RespData, "data")
			delete(RespData, "total")
			delete(RespData, "index")
			delete(RespData, "pages")
		}
		//if c.pageCount == 1 {
		//	delete(RespData, "total")
		//	delete(RespData, "index")
		//	delete(RespData, "pages")
		//}
		c.Data["json"] = RespData
		c.ServeJSON()
	}

}

func (c *CommonController) Prepare() {
	var param string
	var post_params interface{}
	//获取当前url
	path := c.Ctx.Request.URL.Path
	method := c.Ctx.Request.Method
	logs.Info(method + ":" + path)
	switch method {
	case "POST":
		//获取url传参
		err :=c.Ctx.Input.Context.Request.ParseForm()
		if err != nil{
			logs.Error(err)
		}
		get_params := c.Ctx.Input.Context.Request.Form
		err =json.Unmarshal(c.Ctx.Input.RequestBody,&post_params)
		if err != nil{
			logs.Error(err)
		}
		g,_ := json.Marshal(get_params)
		p,_ := json.Marshal(post_params)
		str_g := string(g)
		str_p := string(p)
		param ="url参数：" + str_g +" body参数:"+str_p
	case "PUT":
		//获取url传参
		err :=c.Ctx.Input.Context.Request.ParseForm()
		if err != nil{
			logs.Error(err)
		}
		get_params := c.Ctx.Input.Context.Request.Form
		err =json.Unmarshal(c.Ctx.Input.RequestBody,&post_params)
		if err != nil{
			logs.Error(err)
		}
		g,_ := json.Marshal(get_params)
		p,_ := json.Marshal(post_params)
		str_g := string(g)
		str_p := string(p)
		param ="url参数：" + str_g +" body参数:"+str_p

	default:
		err :=c.Ctx.Input.Context.Request.ParseForm()
		if err != nil{
			logs.Error(err)
		}
		get_params := c.Ctx.Input.Context.Request.Form
		err =c.Ctx.Input.Context.Request.ParseForm()
		if err != nil{
			logs.Error(err)
		}
		g,_ := json.Marshal(get_params)
		str_g := string(g)
		param ="url参数：" + str_g
	}
	logs.Info(method + ":" + path + param)
}

func (c *CommonController) Finish() {
		respons:=c.Data["json"]
		respone_json,err := json.Marshal(respons)
		if err != nil{
			logs.Error(err)
		}
		respons_str := string(respone_json)
		path := c.Ctx.Request.URL.Path
		method := c.Ctx.Request.Method
		logs.Info("url:"+method+":"+path+" " + "响应:" + respons_str)
}

func (c *CommonController) GetPagination() {
	var err error
	c.pageSize, err = c.GetInt("pageSize")
	if err != nil{
		c.pageSize = 5
	}
	//获取总页数
	c.pageCount = math.Ceil(float64(c.count) / float64(c.pageSize))
	//获取页码
	c.pageIndex, err = c.GetInt("pageIndex")
	if err != nil{
		c.pageIndex = 1
	}
	//确定数据的起始位置
	c.start = (c.pageIndex - 1) * c.pageSize
	return
}

// 参数校验，入参必须为结构体或结构体指针
func (c *CommonController) Validate(obj interface{}) (b bool, msg []string) {
	valid := validation.Validation{}
	var err error
	if b, err = valid.Valid(obj); err != nil {
		b = false
		msg = []string{"参数校验失败"}
		logs.Error("参数校验失败", err.Error())
		return
	}

	if !b {
		for _, err := range valid.Errors {
			s := fmt.Sprintf("%s: %s", err.Field, err.Message)
			msg = append(msg, s)
			logs.Info(err.Key, err.Message)
		}
	}
	return
}