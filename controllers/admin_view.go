package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"reflect"
	"server-purchase/models"
	"time"
)

type AdminController struct {
	CommonController
}


func (a*AdminController)  Get(){
	//获取采购单对应报价单列表
	quoted_id,err:=a.GetInt("quoted_id")
	if err != nil {
		logs.Error(err)
		a.ResponseData(400,"获取quoted_id有误","")
	}
	o := orm.NewOrm()
	var data []orm.Params
	queryset :=o.QueryTable("sp_seller_quoted").Filter("quoted_id",quoted_id).OrderBy("-quoted_time")
	a.count,err = queryset.Count()
	if err != nil {
		logs.Error(err)
		a.ResponseData(400,"服务有误","")
		return
	}
	pageSize:=a.GetString("pageSize")
	pageIndex := a.GetString("pageIndex")
	if pageSize != "" && pageIndex != ""{
		a.GetPagination()
	}
	_,err = queryset.Limit(a.pageSize,a.start).Values(&data)
	if err != nil {
		logs.Error(err)
	}
	if data == nil{
		a.ResponseData(200,"ok",make([]interface{},0))
		return
	}
	for _, i := range data{
		tm := time.Unix(i["QuotedTime"].(int64),0)
		i["QuotedTime"] = tm.Format("2006-01-02 15:04:05")
		var history []orm.Params
		_,err := o.QueryTable("history_seller_quoted").Filter("quoted_id",i["Quoted"]).Filter("ShopId",i["ShopId"]).OrderBy("-quoted_time").Values(&history)
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"服务有误","")
			return
		}
		if history != nil{
			for _,h := range history{
				tm := time.Unix(h["QuotedTime"].(int64), 0)
				h["QuotedTime"] = tm.Format("2006-01-02 15:04:05")
			}
		}
		i["history"] = history
	}
	a.ResponseData(200,"ok",data)
}

func (a*AdminController)  Post(){
	//创建虚拟采购单
	type post_data struct {
		CreateTime int64 `json:"create_time" valid:"Required;" description:"求购时间"`
		GoodsName string `json:"goods_name" valid:"Required;" description:"求购内容"`
		SuccessSellerId string `json:"success_seller_id" valid:"Required;" description:"虚拟成交店铺"`
		FixtureDate int64 `json:"fixture_date" valid:"Required;" description:"成交时间"`
		VirtualOrderUser string `json:"virtual_order_user" valid:"Required;" description:"虚拟订单添加人"`

	}
	var body post_data
	o := orm.NewOrm()
	//获取body
	err := json.Unmarshal(a.Ctx.Input.RequestBody,&body)
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"参数格式有误","")
		return
	}
	//校验表单数据
	v := validation.Validation{}
	b,err :=v.Valid(body)
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"格式有误","")
		return
	}
	//输出错误信息
	msg := []string{}
	if !b {
		st := reflect.TypeOf(body)
		for _,e :=range v.Errors{
			filed,_ := st.FieldByName(e.Field)
			description := filed.Tag.Get("description")
			logs.Error(description,":",e.Message)
			msg=append(msg, description + ":" + e.Message)
		}
		a.ResponseData(400,msg,"")
		return
	}
	//插入数据
	var querset models.SpSellerQuotedBuyer
	querset.CreateTime = body.CreateTime
	querset.UpdateTime = time.Now().Unix()
	querset.GoodsName = body.GoodsName
	querset.FixtureDate = body.FixtureDate
	querset.VirtualOrderUser = body.VirtualOrderUser
	querset.VirtualShop = body.SuccessSellerId
	querset.Status = 1
	_,err = o.Insert(&querset)
	if err != nil{
		a.ResponseData(400,"创建失败","")
		return
	}
	a.ResponseData(200,"ok","")
}

func (a*AdminController)  Put(){
	//虚拟采购单编辑

	type post_data struct {
		Id         int     `json:"seller_quoted_id" valid:"Required;" description:"seller_quoted_id"`
		CreateTime int64 `json:"create_time" valid:"Required;" description:"求购时间"`
		GoodsName string `json:"goods_name" valid:"Required;" description:"求购内容"`
		SuccessSellerId string `json:"success_seller_id" valid:"Required;" description:"虚拟成交店铺"`
		FixtureDate int64 `json:"fixture_date" valid:"Required;" description:"成交时间"`
		VirtualOrderUser string `json:"virtual_order_user" valid:"Required;" description:"虚拟订单添加人"`

	}
	var body post_data
	o := orm.NewOrm()
	//获取，校验body
	err := json.Unmarshal(a.Ctx.Input.RequestBody,&body)
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"参数格式有误","")
		return
	}
	v := validation.Validation{}
	b,err :=v.Valid(body)
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"格式有误","")
		return
	}
	//输出错误信息
	msg := []string{}
	if !b {
		st := reflect.TypeOf(body)
		for _,e :=range v.Errors{
			filed,_ := st.FieldByName(e.Field)
			description := filed.Tag.Get("description")
			logs.Error(description,":",e.Message)
			msg=append(msg, description + ":" + e.Message)
		}
		a.ResponseData(400,msg,"")
		return
	}
	//校验此单是否为虚拟单
	var querset models.SpSellerQuotedBuyer
	err = o.QueryTable("sp_seller_quoted_buyer").Filter("quoted_id",body.Id).Filter("buyer_id",0).One(&querset)
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"此报价不为虚拟报价，不可修改","")
		return
	}
	//修改数据
	querset.CreateTime = body.CreateTime
	querset.UpdateTime =  time.Now().Unix()
	querset.GoodsName = body.GoodsName
	querset.VirtualShop = body.SuccessSellerId
	querset.FixtureDate = body.FixtureDate
	querset.VirtualOrderUser = body.VirtualOrderUser
	querset.Status = 1
	_,err = o.Update(&querset)
	if err != nil{
		a.ResponseData(400,"创建失败","")
		return
	}
	a.ResponseData(200,"ok","")
}

func (a*AdminController) GetList(){
		//成交动态列表
		o := orm.NewOrm()
		queryset := o.QueryTable("sp_seller_quoted_buyer").Filter("status",1).OrderBy("-update_time")
		quoted_id,err := a.GetInt("quoted_id")
		if err != nil {
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		if quoted_id > 0 {
			queryset = queryset.Filter("quoted_id",quoted_id)
		}
		create_time_start,err := a.GetInt("create_time_start")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		create_time_end,err := a.GetInt("create_time_start")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		if create_time_start != 0 && create_time_end != 0 {
			queryset =  queryset.Filter("create_time__gte",create_time_start).Filter("create_time__lte",create_time_end)
		}
		fixture_date_end,err := a.GetInt("fixture_date_end")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		fixture_date_start,err := a.GetInt("fixture_date_start")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		if fixture_date_start != 0 && fixture_date_end != 0 {
			queryset =  queryset.Filter("fixture_date__gte",fixture_date_start).Filter("fixture_date__lte",fixture_date_end)
		}
		success_seller_id,err := a.GetInt("success_seller_id")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		if success_seller_id > 0 {
			queryset =  queryset.Filter("success_seller_id",success_seller_id)
		}
		virtual_shop := a.GetString("virtual_shop")
		if virtual_shop != "" {
			queryset =  queryset.Filter("virtual_shop__contains",virtual_shop)
		}
		status,err := a.GetInt("status")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"格式有误","")
			return
		}
		switch status {
			case 1:queryset = queryset.Filter("buyer_id",0)
			case 2:queryset = queryset.Filter("buyer_id__gt",0)
		}
		var data []orm.Params
		a.count,err = queryset.Count()
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"查询错误","")
			return
		}
		a.GetPagination()
		_ ,err = queryset.Limit(a.pageSize,a.start).Values(&data,"create_time","fixture_date","goods_name","success_seller_id","virtual_order_user","quoted_id","order_sn","virtual_shop")
		if err != nil{
			logs.Error(err)
			a.ResponseData(400,"查询错误","")
			return
		}
		if data == nil{
			data_nil := make([]interface{},0)
			a.ResponseData(200,"ok",data_nil)
			return
		}
		for _,i := range data{
			tm := time.Unix(i["CreateTime"].(int64), 0)
			i["CreateTime"] = tm.Format("2006-01-02 15:04:05")
			tm  = time.Unix(i["FixtureDate"].(int64), 0)
			i["FixtureDate"] = tm.Format("2006-01-02 15:04:05")
		}
		a.ResponseData(200,"ok",data)
}

func (a*AdminController) Delete(){
	//虚拟单删除
	o := orm.NewOrm()
	var   queryset models.SpSellerQuotedBuyer
	id,err := a.GetInt("id")
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"参数格式有误","")
		return
	}
	 //校验是否为虚拟单
	 err =  o.QueryTable("sp_seller_quoted_buyer").Filter("quoted_id",id).Filter("buyer_id",0).One(&queryset)
	 if err != nil{
		 logs.Error(err)
		 a.ResponseData(400,"不为虚拟单，无法删除","")
		 return
	 }
	 _,err = o.Delete(&queryset)
	if err != nil{
		logs.Error(err)
		a.ResponseData(400,"删除失败","")
		return
	}
	 a.ResponseData(200,"ok","")
}







