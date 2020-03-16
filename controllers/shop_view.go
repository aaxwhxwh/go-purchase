package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"reflect"
	"server-purchase/models"
	"server-purchase/utils/constant"
	"strconv"
	"time"

)

type Shop_controllers struct {
	CommonController
}


func (s*Shop_controllers)  Post(){
	//获取商家端报价
	type body struct {
		QuotedId   int `json:"quoted_id" valid:"Required;" description:"采购id"`
		ShopId     int `json:"shop_id" valid:"Required;" description:"商家id"`
		TotalPrice string  `json:"total_price" valid:"Required;" description:"总价"`
		GoodsQty   int `json:"goods_qty"  description:"采购总数量"`
		Remark     string `json:"remark" description:"备注"`
		File     string `json:"file" description:"上传文件"`
		FileName     string  `json:"file_name"  description:"文件名称"`
	}
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"服务展示无法使用，请稍后重试","")
		return
	}
	var post_body body
	err = json.Unmarshal(s.Ctx.Input.RequestBody,&post_body)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"格式有误","")
		return
	}
	//表单校验
	v := validation.Validation{}
	b,err :=v.Valid(post_body)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"格式有误","")
		return
	}
	//输出错误信息
	msg := []string{}
	if !b {
		st := reflect.TypeOf(post_body)
		for _,e :=range v.Errors{
			filed,_ := st.FieldByName(e.Field)
			description := filed.Tag.Get("description")
			logs.Error(description,":",e.Message)
			msg=append(msg, description + ":" + e.Message)
		}
		s.ResponseData(400,msg,"")
		return
	}

	var queryset models.SpSellerQuoted
	queryset.ShopId = post_body.ShopId
	queryset.File = post_body.File
	queryset.FileName = post_body.FileName
	quoted_id := models.SpSellerQuotedBuyer{Id:post_body.QuotedId}
	queryset.Quoted = &quoted_id
	//判断是否对该采购单报过价
	var valid_queryset models.SpSellerQuoted
	err = o.QueryTable("sp_seller_quoted").Filter("quoted__quoted_id",post_body.QuotedId).Filter("shop_id",post_body.ShopId).One(&valid_queryset)
	fmt.Println(valid_queryset)
	if err == nil{
		s.ResponseData(400,"报价已存在","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	//判断采购单是否存在
	var buyer models.SpSellerQuotedBuyer
	err = o.QueryTable("sp_seller_quoted_buyer").Filter("quoted_id",post_body.QuotedId).Filter("audit_status",1).Filter("status",0).Filter("quoted_end_time__gt",int(time.Now().Unix())).One(&buyer)
	if err != nil {
		s.ResponseData(400,"采购单不存在","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	//如果存在计数 +1
	buyer.QuoteCount = buyer.QuoteCount + 1
	_,err = o.Update(&buyer,"QuoteCount")
	if err != nil{
		s.ResponseData(400,"创建失败","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	queryset.Remark = post_body.Remark
	totalprice ,err := strconv.ParseFloat( post_body.TotalPrice,64)
	queryset.TotalPrice = totalprice
	queryset.QuotedTime = int(time.Now().Unix())
	//当GoodsQty大于0的时候 需计算单价
	if post_body.GoodsQty > 0{
		qty,err := strconv.ParseFloat(strconv.Itoa(post_body.GoodsQty),64)
		if err != nil{
			logs.Error(err)
			s.ResponseData(400,"格式有误","")
			err = o.Rollback()
			return
		}
		queryset.UnitPrice=totalprice / qty
	}
	//插入数据
	_,err=o.Insert(&queryset)
	if err != nil {
		logs.Error(err)
		s.ResponseData(400,"创建失败","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	o.Commit()
	//给中台记录日志使用
	log_data := map[string]interface{}{"quoted_id": post_body.QuotedId,"buyer_id":buyer.BuyerId,"purchase_sn":buyer.PurchaseSn}
	s.ResponseData(200,"ok",log_data)
}

func (s*Shop_controllers) Get(){
	//获取商家端采购单列表
	o := orm.NewOrm()
	shop_id ,err:= strconv.Atoi(s.GetString("shop_id"))
	if err != nil{
		s.ResponseData(400,"参数有误","")
		return
	}
	type_str := s.GetString("type","0")
	var seller_id_list orm.ParamsList
	//查出商家一对那些采购单做出报价
	_, err =o.QueryTable("sp_seller_quoted").Filter("shop_id",shop_id).ValuesFlat(&seller_id_list,"quoted_id")
	if len(seller_id_list) == 0{
		seller_id_list = append(seller_id_list,0)
	}
	var buyer_queryset []orm.Params
	switch type_str{
		case "0":
			//type 0 表示待报价
			queryset_0 := o.QueryTable("sp_seller_quoted_buyer").Filter("audit_status",1).Filter("status",0).Filter("quoted_end_time__gt",int(time.Now().Unix())).Exclude("quoted_id__in",seller_id_list).OrderBy("-update_time")
			//分页
			s.count,err = queryset_0.Count()
			if err != nil{
				logs.Error(err)
				s.ResponseData(400,"查询失败","")
			}
			s.GetPagination()
			_,err = queryset_0.Limit(s.pageSize,s.start).Values(&buyer_queryset,"quoted_id","purchase_sn","city","goods_name","goods_qty","quoted_end_time","delivery_end_time","details_desc","create_time")
			if err != nil{
				logs.Error(err)
				s.ResponseData(400,"查询失败","")
			}
		case "1":
			//type 1 表示已报价
			queryset_1 := o.QueryTable("sp_seller_quoted_buyer").Filter("audit_status",1).Filter("SpSellerQuotedSet__shop_id",shop_id).OrderBy("-update_time")
			//分页
			s.count,err = queryset_1.Count()
			if err != nil{
				logs.Error(err)
				s.ResponseData(400,"查询失败","")
			}
			s.GetPagination()
			_,err:=queryset_1.Limit(s.pageSize,s.start).Values(&buyer_queryset,"quoted_id","purchase_sn","city","goods_name","goods_qty","quoted_end_time","delivery_end_time","details_desc","status","success_seller_id","quoted_end_time","SpSellerQuotedSet__TotalPrice","SpSellerQuotedSet__Id","create_time")
			if err != nil{
				logs.Error(err)
				s.ResponseData(400,"查询失败","")
				}

			// 报价状态
			for _,i := range buyer_queryset{
				//价格保留两位小数
				i["SpSellerQuotedSet__TotalPrice"] = fmt.Sprintf("%.2f", i["SpSellerQuotedSet__TotalPrice"])
				endtime,ok := i["QuotedEndTime"].(int64)
				History_Count,err :=o.QueryTable("history_seller_quoted").Filter("seller_quoted_id",i["SpSellerQuotedSet__Id"]).Filter("shop_id",shop_id).Count()
				if err != nil{
					s.ResponseData(400,"查询有误","")
					return
				}
				i["History_Count"] = History_Count+1
				if !ok{
					s.ResponseData(400,"类型有误","")
					return
				}
				sellerid,ok := i["SuccessSellerId"].(int64)
				if !ok{
					s.ResponseData(400,"类型有误","")
					return
				}
				switch i["Status"] {
					//采购单状态0 ，如果报价超时  抢单失败 如果未超时 抢单中
					case int64(0):
						if endtime + int64(4*24*60*60)  < time.Now().Unix(){
							i["seller_status"] = 2
						}else {
							i["seller_status"] = 0
						}
					//采购单状态1（已成交），如果选中本商家，抢单成功  否则抢单失败
					case int64(1):
						if sellerid == int64(shop_id){
							i["seller_status"] = 1
						}else {
							i["seller_status"] = 2
						}
					//采购单状态2（已暂停），如果超时了三天 ，抢单失败，否则报价中
					case int64(2):
						if endtime + int64(4*24*60*60) < time.Now().Unix(){
							i["seller_status"] = 2
						}else {
							i["seller_status"] = 0
						}
					//采购单状态3（失效），抢单失败
					case int64(3):
						i["seller_status"] = 2
					//采购单状态7（失效），抢单失败
					case int64(7):
						i["seller_status"] = 2
				}
			}
	}
	s.ResponseData(200,"ok",buyer_queryset)
}

func (s*Shop_controllers) Put(){
	type body struct {
		SellerQuotedId         int     `json:"seller_quoted_id" valid:"Required;" description:"商家报价id"`
		QuotedId   int `json:"quoted_id" valid:"Required;" description:"采购id"`
		ShopId     int `json:"shop_id" valid:"Required;" description:"商家id"`
		TotalPrice string  `json:"total_price" valid:"Required;" description:"总价"`
		GoodsQty   int `json:"goods_qty"  description:"采购总数量"`
		Remark     string `json:"remark" description:"备注"`
		File     string `json:"file" description:"上传文件"`
		FileName     string  `json:"file_name" description:"文件名称"`
	}
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"服务展示无法使用，请稍后重试","")
		return
	}
	var post_body body
	err = json.Unmarshal(s.Ctx.Input.RequestBody,&post_body)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"格式有误","")
		return
	}
	//表单校验
	v := validation.Validation{}
	b,err :=v.Valid(post_body)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"格式有误","")
		return
	}
	//输出错误信息
	msg := []string{}
	if !b {
		st := reflect.TypeOf(post_body)
		for _,e :=range v.Errors{
			filed,_ := st.FieldByName(e.Field)
			description := filed.Tag.Get("description")
			logs.Error(description,":",e.Message)
			msg=append(msg, description + ":" + e.Message)
		}
		s.ResponseData(400,msg,"")
		return
	}
	var queryset models.SpSellerQuoted
	//判断是否存在商家报价单
	err = o.QueryTable("sp_seller_quoted").Filter("seller_quoted_id",post_body.SellerQuotedId).Filter("shop_id",post_body.ShopId).One(&queryset)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"商家报价单不存在","")
		err = o.Rollback()
		return
	}
	//判断采购单是否存在
	var buyer models.SpSellerQuotedBuyer
	err = o.QueryTable("sp_seller_quoted_buyer").Filter("quoted_id",post_body.QuotedId).Filter("audit_status",1).Filter("status",0).Filter("quoted_end_time__gt",int(time.Now().Unix())).One(&buyer)
	if err != nil {
		s.ResponseData(400,"采购单不存在","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	//保存历史报价
	var history_seller_quoted models.HistorySellerQuoted
	history_seller_quoted.SellerQuoted = &queryset
	history_seller_quoted.Quoted = queryset.Quoted
	history_seller_quoted.ShopId = queryset.ShopId
	history_seller_quoted.UnitPrice = queryset.UnitPrice
	history_seller_quoted.TotalPrice = queryset.TotalPrice
	history_seller_quoted.Remark = queryset.Remark
	history_seller_quoted.QuotedTime = queryset.QuotedTime
	history_seller_quoted.File = queryset.File
	history_seller_quoted.FileName = queryset.FileName
	_,err = o.Insert(&history_seller_quoted)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"修改失败","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	//修改报价
	queryset.Remark = post_body.Remark
	totalprice ,err := strconv.ParseFloat( post_body.TotalPrice,64)
	queryset.TotalPrice = totalprice
	queryset.QuotedTime = int(time.Now().Unix())
	queryset.File = post_body.File
	queryset.FileName = post_body.FileName
	if post_body.GoodsQty > 0{
		qty,err := strconv.ParseFloat(strconv.Itoa(post_body.GoodsQty),64)
		if err != nil{
			logs.Error(err)
			s.ResponseData(400,"格式有误","")
			err = o.Rollback()
			logs.Error(err)
			return
		}
		queryset.UnitPrice=totalprice / qty
	}
	_,err = o.Update(&queryset)
	if err != nil{
		s.ResponseData(400,"修改失败","")
		err = o.Rollback()
		logs.Error(err)
		return
	}
	o.Commit()
	log_data := map[string]interface{}{"quoted_id": post_body.QuotedId,"buyer_id":buyer.BuyerId,"purchase_sn":buyer.PurchaseSn}
	s.ResponseData(200,"ok",log_data)

}

func (s*Shop_controllers) GetContent(){
	//获取商家端采购单详情
	o := orm.NewOrm()
	shop_id ,err:= strconv.Atoi(s.GetString("shop_id"))
	if err != nil{
		s.ResponseData(400,"参数有误","")
		return
	}
	quoted_id ,err:= strconv.Atoi(s.GetString("quoted_id"))
	if err != nil{
		s.ResponseData(400,"参数有误","")
		return
	}
	
	//获取采购单详情
	data,err:= models.GetPurchaseDetail(quoted_id, constant.SHOP)
	if err != nil{
		s.ResponseData(400,"查询错误","")
		return
	}
	var data_dic map[string]interface{}
	var queryset models.SpSellerQuoted
	//var queryset []orm.Params
	data_dic = make(map[string]interface{})
	fmt.Println(shop_id)
	data_dic["buyer"] = data
	//获取报价单
	err=o.QueryTable("sp_seller_quoted").Filter("shop_id",shop_id).Filter("quoted_id",quoted_id).One(&queryset)
	if err != nil{
		data_dic["seller"] = ""
		s.ResponseData(200,"ok",data_dic)
	}
	data_dic["seller"] = queryset
	s.ResponseData(200,"ok",data_dic)
}

func (s*Shop_controllers) HistoryList(){
	shop_id,err := s.GetInt("shop_id")
	if err != nil {
		s.ResponseData(400,"传参有误","")
		return
	}
	quoted_id,err := s.GetInt("quoted_id")
	if err != nil {
		s.ResponseData(400,"传参有误","")
		return
	}
	o := orm.NewOrm()
	var querset_list []models.HistorySellerQuoted
	_,err = o.QueryTable("history_seller_quoted").Filter("shop_id",shop_id).Filter("quoted_id",quoted_id).OrderBy("-quoted_time").All(&querset_list)
	if err != nil{
		logs.Error(err)
		s.ResponseData(400,"查询失败","")
		return
	}
	s.ResponseData(200,"ok",querset_list)
}






