package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"math"
	"reflect"
	"server-purchase/models"
	_ "server-purchase/utils"
	"strconv"
	"time"
)

type CatController struct {
	CommonController
}


type get_data struct {
	*models.SpPurchaseCategory
	Children []interface{} `json:"children"`
}


//func (c*CatController)  Put(){
//	var post_data models.SpPurchaseCategory
//	var map_data map[string]interface{}
//	o := orm.NewOrm()
//	//获取表单数据
//	err := json.Unmarshal((c.Ctx.Input.RequestBody),&post_data)
//	err = json.Unmarshal((c.Ctx.Input.RequestBody),&map_data)
//	post_data.UpdateTime = int(time.Now().Unix())
//	if err != nil{
//		logs.Error(err)
//		c.ResponseData(400,"格式有误","")
//		return
//	}
//	//校验表单数据
//	v := validation.Validation{}
//	b,err :=v.Valid(post_data)
//	if err != nil{
//		logs.Error(err)
//		c.ResponseData(400,"格式有误","")
//		return
//	}
//	//输出错误信息
//	msg := []string{}
//	if !b {
//		st := reflect.TypeOf(post_data)
//		for _,e :=range v.Errors{
//			filed,_ := st.FieldByName(e.Field)
//			description := filed.Tag.Get("description")
//			logs.Error(description,":",e.Message)
//			msg=append(msg, description + ":" + e.Message)
//		}
//		c.ResponseData(400,msg,"")
//		return
//	}
//	categoryname := post_data.CategoryName
//	//判断分类名是否存在
//	if categoryname != "" {
//		err = o.QueryTable("SpPurchaseCategory").Filter("category_name",categoryname).Exclude("category_id",post_data.Id).Filter("is_deleted",0).One(&post_data)
//		if err == nil{
//			logs.Error(err)
//			c.ResponseData(400,"该分类名已存在","")
//			return
//		}
//	}
//	param_list := make([]string,0)
//	for i := range map_data{
//		param_list=append(param_list,i)
//	}
//	_,err = o.Update(&post_data,param_list...)
//	if err != nil {
//		logs.Error(err)
//		c.ResponseData(400,"创建失败","")
//		return
//	}
//	c.ResponseData(200,"ok","")
//	return
//
//}

//func (c*CatController) Get() {
//	var a_l []*models.SpPurchaseCategory
//	o := orm.NewOrm()
//	//num,err:=o.QueryTable("SpPurchaseCategory").All(&a_l)
//	num ,err:=o.QueryTable("SpPurchaseCategory").Filter("level",1).Filter("is_deleted",0).All(&a_l)
//	fmt.Println(err)
//	fmt.Println(num)
//	for _,i := range a_l{
//		fmt.Println(i)
//		_,err :=o.LoadRelated(i,"Ch")
//		if err != nil{
//			logs.Error(err)
//			c.ResponseData(400,"查询失败","")
//		}
//		for _,j := range i.Ch{
//			_,err := o.LoadRelated(j,"Ch")
//			if err != nil{
//				logs.Error(err)
//				c.ResponseData(400,"查询失败","")
//			}
//		}
//	}
//	c.Data["json"] = a_l
//	c.ServeJSON()
//}


//func (c*CatController)  Get(){
//	var l []orm.Params
//	data_list :=make([]interface{},0)
//	o := orm.NewOrm()
//	_,err:=o.QueryTable("sp_purchase_category").Filter("is_deleted",0).Values(&l)
//	if err != nil{
//		logs.Error(err)
//		c.ResponseData(400,"获取分类失败","")
//		return
//	}
//	for _,i := range  l{
//		if i["Parent"] == int64(0){
//			l_1 := make([]interface{},0)
//			for _,j :=range l{
//				if i["Id"] == j["Parent"]{
//					l_1=append(l_1,j)
//					l_2 := make([]interface{},0)
//					for _,k := range l{
//						if j["Id"] == k["Parent"]{
//							l_2=append(l_2,k)
//						}
//					j["Ch"] = l_2
//					}
//				}
//			i["Ch"] = l_1
//			}
//			data_list=append(data_list,i)
//		}
//	}
//	c.Data["json"] = data_list
//	c.ServeJSON()
//}

func (c*CatController)  Get(){

	type_data := c.GetString("type","1")
	code,msg,data,err := Getcat(type_data)
	if err != nil{
		logs.Error(err)
	}
	c.ResponseData(code,msg,data)
	//var get_data_list []*models.SpPurchaseCategory
	//var get_list []*get_data
	//var data []interface{}
	//o := orm.NewOrm()
	//type_data := c.GetString("type","1")
	//if type_data == "1" {
	//	num,err:=o.QueryTable("SpPurchaseCategory").Filter("is_deleted",0).OrderBy("-sort").All(&get_data_list)
	//	if err != nil {
	//		c.ResponseData(400,"获取错误","")
	//	}
	//	if num == 0{
	//		c.ResponseData(200,"ok",make([]interface{},0))
	//	}
	//}else {
	//	num,err:=o.QueryTable("SpPurchaseCategory").Filter("is_deleted",0).Filter("category_status",0).OrderBy("-sort").All(&get_data_list)
	//	if err != nil {
	//		c.ResponseData(400,"获取错误","")
	//	}
	//	if num == 0{
	//		c.ResponseData(200,"ok",make([]interface{},0))
	//	}
	//}
	//fmt.Println(get_data_list)
	//for _ , i := range get_data_list{
	//	j := get_data{i,make([]interface{},0)}
	//	get_list = append(get_list,&j)
	//}
	//for _,i := range get_list{
	//	if i.Parent.Id ==0 {
	//		data = append(data, i)
	//		for _,j :=range get_list{
	//			if j.Parent.Id == i.Id{
	//				i.Children=append(i.Children,j)
	//				for _,k := range get_list{
	//					if k.Parent.Id == j.Id{
	//						j.Children=append(j.Children,k)
	//					}
	//				}
	//			}
	//		}
	//
	//	}
	//}
	//c.ResponseData(200,"ok",data)
}


func (c*CatController)  UserGet() {

	type_data := c.GetString("type", "1")
	code, msg, data, err := UserGetcat(type_data)
	if err != nil {
		logs.Error(err)
	}
	c.ResponseData(code,msg,data)
}
func (c*CatController)  Post(){
	var post_data models.SpPurchaseCategory
	var cat_parent models.SpPurchaseCategory
	var buyer models.SpCatLinkAttr
	var level int
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"服务暂时无法使用，请稍后再试","")
		return
	}
	//获取表单数据
	err = json.Unmarshal((c.Ctx.Input.RequestBody),&post_data)
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"格式有误","")
		return
	}
	post_data.IsEnd = 1
	post_data.CreateTime = int(time.Now().Unix())
	post_data.UpdateTime = int(time.Now().Unix())
	post_data.Sort = 0
	var cat models.SpPurchaseCategory
	err = o.QueryTable("SpPurchaseCategory").OrderBy("-sort").One(&cat)
	if err == nil{
		post_data.Sort = cat.Sort + 1
	}
	//校验表单数据
	v := validation.Validation{}
	b,err :=v.Valid(post_data)
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"格式有误","")
		o.Rollback()
		return
	}
	//输出错误信息
	msg := []string{}
	if !b {
		st := reflect.TypeOf(post_data)
		for _,e :=range v.Errors{
			filed,_ := st.FieldByName(e.Field)
			description := filed.Tag.Get("description")
			logs.Error(description,":",e.Message)
			msg=append(msg, description + ":" + e.Message)
		}
		c.ResponseData(400,msg,"")
		o.Rollback()
		return
	}
	categoryname := post_data.CategoryName
	fmt.Println(categoryname)
	//判断分类名是否存在
	if categoryname != "" {
		err = o.QueryTable("SpPurchaseCategory").Filter("category_name",categoryname).Filter("is_deleted",0).One(&post_data)
		if err == nil{
			logs.Error(err)
			c.ResponseData(400,"该分类名已存在","")
			o.Rollback()
			return
		}
	}
	//判断父类是否存在
	z := post_data.Parent
	fmt.Println("=============",z)

	parent_id := post_data.Parent.Id
	if parent_id > 0 {
		err = o.QueryTable("SpPurchaseCategory").Filter("category_id",parent_id).Filter("is_deleted",0).One(&cat_parent)
		if err != nil{
			logs.Error(err)
			c.ResponseData(400,"父级分类不存在","")
			o.Rollback()
			return
		}
		tmp_cat_id := cat_parent.Id
		err = o.QueryTable("sp_cat_link_attr").Filter("cat_id",tmp_cat_id).One(&buyer)
		if err == nil{
			c.ResponseData(400,"父级分类上已绑定模板，无法添加子级","")
			o.Rollback()
			return
		}
		//如果父类存在，判断是否为3级分类
		level = cat_parent.Level
		if level >= 3 {
			c.ResponseData(400,"分类只能存三级","")
			o.Rollback()
			return
		}
		//如果父类存在，is_end 改为 0
		data := orm.Params{}
		data["is_end"] = 0
		num,err:=o.QueryTable("SpPurchaseCategory").Filter("category_id",parent_id).Filter("is_deleted",0).Update(data)
		if err != nil && num == 0{
			o.Rollback()
			logs.Error(err)
			c.ResponseData(400,"创建失败","")
			return
		}
	}
	post_data.Level = level + 1
	_,err = o.Insert(&post_data)
	if err != nil {
		o.Rollback()
		logs.Error(err)
		c.ResponseData(400,"创建失败","")
		return
	}
	o.Commit()
	c.ResponseData(200,"ok","")
	return
}

func (c*CatController)  Delete(){
	var queryset models.SpPurchaseCategory
	var tm models.SpCatLinkAttr
	var parent_queryset models.SpPurchaseCategory
	o := orm.NewOrm()
	err := o.Begin()
	id_str:=c.GetString("category_id")
	id, err := strconv.Atoi(id_str)
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"传入类型有误","")
		return
	}
	//判断是否传id
	if id == 0{
		c.ResponseData(400,"缺少参数","")
		return
	}
	//判断分类是否存在
	err =o.QueryTable("sp_purchase_category").Filter("category_id",id).Filter("is_deleted",0).One(&queryset)
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"无此分类","")
		return
	}
	//判断是否有子集分类
	if queryset.IsEnd == 0 {
		logs.Error(err)
		c.ResponseData(400,"不是最后一级分类，无法删除","")
		return
	}
	//判断是否存在模板
	err = o.QueryTable("sp_cat_link_attr").Filter("cat_id",id).Filter("is_deleted",0).One(&tm)
	if err == nil{
		logs.Error(err)
		c.ResponseData(400,"分类下存在模板无法删除","")
		return
	}
	//删除分类,判断有无父级
	if queryset.Parent.Id > 0{
		//获取父类
		err = o.QueryTable("sp_purchase_category").Filter("category_id",queryset.Parent.Id).Filter("is_deleted",0).One(&parent_queryset)
		if err != nil{
			logs.Error(err)
			c.ResponseData(400,"删除失败，获取父级失败","")
			return
		}
		//判断该父类下，有无其他分类
		var queryset_a models.SpPurchaseCategory
		err = o.QueryTable("sp_purchase_category").Filter("parent_id",queryset.Parent.Id).Filter("is_deleted",0).Exclude("category_id",queryset.Id).One(&queryset_a)
		if err != nil{
			//如果没有，父类is_end 改为1
			parent_queryset.IsEnd = 1
			_,err = o.Update(&parent_queryset,"IsEnd")
			if err != nil{
				logs.Error(err)
				o.Rollback()
				c.ResponseData(400,"删除失败","")
				return
			}
		}
	}
	//2删除该分类
	queryset.IsDeleted = 1
	_,err=o.Update(&queryset,"IsDeleted")
	if err != nil{
		logs.Error(err)
		o.Rollback()
		c.ResponseData(400,"删除失败","")
		return
	}
	o.Commit()
	c.ResponseData(200,"删除成功","")


}

func (c*CatController) Put(){
	body := make(map[string]int)
	o := orm.NewOrm()
	err:=o.Begin()
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"服务暂时无法使用","")
		return
	}
	request_body := c.Ctx.Input.RequestBody
	err = json.Unmarshal(request_body,&body)
	if err != nil {
		logs.Error(err)
		c.ResponseData(400,"格式有误","")
		return
	}
	//校验入参
	category_status,ok_c := body["category_status"]
	move,ok_m := body["move"]
	fmt.Println(move)
	category_id,ok_id := body["category_id"]
	if !ok_id{
		c.ResponseData(400,"参数不齐","")
		return
	}
	if !(ok_c || ok_m) {
		c.ResponseData(400,"参数不齐","")
		return
	}
	//判断分类是否存在
	var queryset models.SpPurchaseCategory
	err = o.QueryTable("sp_purchase_category").Filter("category_id",category_id).Filter("is_deleted",0).One(&queryset)
	if err != nil{
		c.ResponseData(400,"分类不存在","")
		return
	}
	//分类 启用 停用
	if ok_c{
		switch category_status {
			//0,启用分类
			case 0:
				//如有父类，需把父类开启
				err=open_cat(queryset,o)
				if err != nil {
					logs.Error(err)
					c.ResponseData(400, "失败", "")
					}
			case 1:
				//如有子集，需关闭
				err=close_cat(queryset,o)
				if err != nil {
					logs.Error(err)
					c.ResponseData(400, "失败", "")
				}
			default: c.ResponseData(400,"参数格式有误","")
		}
	}
	if ok_m {
		//move  1.上移,0.下移
		switch move {
			case 0:
				 //获取同级比本身小的最大分类
				 var max_queryset models.SpPurchaseCategory
				 if queryset.Parent.Id == 0{
					 err = o.QueryTable("sp_purchase_category").Filter("sort__lt",queryset.Sort).Filter("is_deleted",0).Filter("level",1).OrderBy("-sort").One(&max_queryset)
				 }else {
					 err = o.QueryTable("sp_purchase_category").Filter("parent_id",queryset.Parent.Id).Filter("sort__lt",queryset.Sort).Filter("is_deleted",0).OrderBy("-sort").One(&max_queryset)
				 }
				 //err = o.QueryTable("sp_purchase_category").Filter("parent_id",queryset.Parent.Id).Filter("sort__lt",queryset.Sort).Filter("is_deleted",0).OrderBy("-sort").One(&max_queryset)
				 if err != nil{
				 	logs.Error(err)
				 	c.ResponseData(400,"无法移动","")
				 	return
				 }
				 max_queryset.Sort,queryset.Sort = queryset.Sort , max_queryset.Sort
				 _,err = o.Update(&max_queryset,"Sort")
				if err != nil{
					logs.Error(err)
					o.Rollback()
					c.ResponseData(400,"查询错误","")
					return
				}
				 _,err = o.Update(&queryset,"Sort")
				if err != nil{
					logs.Error(err)
					o.Rollback()
					c.ResponseData(400,"查询错误","")
					return
				}
			case 1:
				//获取同级比本身大的最小分类
				var min_queryset models.SpPurchaseCategory
				if queryset.Parent.Id == 0 {
					err = o.QueryTable("sp_purchase_category").Filter("sort__gt",queryset.Sort).Filter("is_deleted",0).Filter("level",1).OrderBy("sort").One(&min_queryset)
				}else {
					err = o.QueryTable("sp_purchase_category").Filter("parent_id",queryset.Parent.Id).Filter("sort__gt",queryset.Sort).Filter("is_deleted",0).OrderBy("sort").One(&min_queryset)
				}
				//err = o.QueryTable("sp_purchase_category").Filter("parent_id",queryset.Parent.Id).Filter("sort__gt",queryset.Sort).Filter("is_deleted",0).OrderBy("sort").One(&min_queryset)
				if err != nil{
					logs.Error(err)
					o.Rollback()
					c.ResponseData(400,"无法移动","")
					return
				}
				//交换sort字段
				min_queryset.Sort,queryset.Sort = queryset.Sort,min_queryset.Sort
				_,err = o.Update(&min_queryset,"Sort")
				if err != nil{
					logs.Error(err)
					o.Rollback()
					c.ResponseData(400,"查询错误","")
					return
				}
				_,err = o.Update(&queryset,"Sort")
				if err != nil{
					logs.Error(err)
					o.Rollback()
					c.ResponseData(400,"查询错误","")
					return
				}
			default: c.ResponseData(400,"参数格式有误","")
		}
	}
	o.Commit()
	c.ResponseData(200,"ok","")
}

func open_cat(queryset models.SpPurchaseCategory,o orm.Ormer) (error){
	//开启分类
	//开启该分类
	queryset.CategoryStatus =0
	_,err :=o.Update(&queryset,"CategoryStatus")
	if err != nil{
		return err
	}
	if queryset.Parent.Id == 0{
		//当没有父类id递归结束
		return nil
	}
	//获取父类
	var aa models.SpPurchaseCategory
	err = o.QueryTable("sp_purchase_category").Filter("category_id",queryset.Parent.Id).One(&aa)
	if err != nil{
		return err
	}
	 return open_cat(aa,o)
}

func close_cat(queryset models.SpPurchaseCategory,o orm.Ormer)(error){
	//关闭分类
	//反向关联，获取子集对象
	_,err :=o.LoadRelated(&queryset,"Ch")
	if err != nil{
		return err
	}
	//修改改分类状态，CategoryStatus1为关闭
	queryset.CategoryStatus = 1
	_,err =o.Update(&queryset,"CategoryStatus")
	if err != nil{
		return err
	}
	//如果子集为空，return
	if queryset.Ch == nil{
		return nil
	}
	//递归所有子集
	for _, i :=range queryset.Ch{
		err = close_cat(*i,o)
		if err!= nil{
			return err
		}
	}
	return nil
}


func Getcat(type_data string) (code int ,msg interface{},response_data interface{},err interface{}){
	var get_data_list []*models.SpPurchaseCategory
	var get_list []*get_data
	var data []interface{}
	o := orm.NewOrm()
	if type_data == "1" {
		num,err:=o.QueryTable("SpPurchaseCategory").Filter("is_deleted",0).OrderBy("-sort").All(&get_data_list)
		if err != nil {
			return 400,"获取错误","",err
		}
		if num == 0{
			return 200,"ok",make([]interface{},0),nil
		}
	}else {
		num,err:=o.QueryTable("SpPurchaseCategory").Filter("is_deleted",0).Filter("category_status",0).OrderBy("-sort").All(&get_data_list)
		if err != nil {
			return 400,"获取错误","",err
		}
		if num == 0{
			return 200,"ok",make([]interface{},0),nil
		}
	}
	for _ , i := range get_data_list{
		j := get_data{i,make([]interface{},0)}
		get_list = append(get_list,&j)
	}
	for _,i := range get_list{
		if i.Parent.Id ==0 {
			data = append(data, i)
			for _,j :=range get_list{
				if j.Parent.Id == i.Id{
					i.Children=append(i.Children,j)
					for _,k := range get_list{
						if k.Parent.Id == j.Id{
							j.Children=append(j.Children,k)
						}
					}
				}
			}
		}
	}
	return 200,"ok",data ,nil
}


func UserGetcat(type_data string) (code int ,msg interface{},response_data interface{},err interface{}){
	var get_data_list []*models.SpPurchaseCategory
	var get_list []*get_data
	var data []interface{}
	o := orm.NewOrm()
	if type_data == "1" {
		num,err:=o.QueryTable("SpPurchaseCategory").Filter("is_deleted",0).OrderBy("-sort").All(&get_data_list)
		if err != nil {
			return 400,"获取错误","",err
		}
		if num == 0{
			return 200,"ok",make([]interface{},0),nil
		}
	}else {
		num,err:=o.QueryTable("SpPurchaseCategory").Filter("is_deleted",0).Filter("category_status",0).OrderBy("-sort").All(&get_data_list)
		if err != nil {
			return 400,"获取错误","",err
		}
		if num == 0{
			return 200,"ok",make([]interface{},0),nil
		}
	}
	for _ , i := range get_data_list{
		j := get_data{i,make([]interface{},0)}
		get_list = append(get_list,&j)
	}
	for _,i := range get_list{
		if i.Parent.Id ==0 {
			for _,j :=range get_list{
				if j.Parent.Id == i.Id{

					for _,k := range get_list{
						if k.Parent.Id == j.Id{
							//j.Children=append(j.Children,k)
							var catlinkattr models.SpCatLinkAttr
							err = o.QueryTable("sp_cat_link_attr").Filter("cat_id",k.Id).Filter("is_deleted",0).One(&catlinkattr)
							if err != nil {
								logs.Error(err)
							}else {
								j.Children=append(j.Children,k)
							}
						}
					}
					if j.IsEnd == 0{
						if len(j.Children) != 0{
							i.Children = append(i.Children, j)
						}
					}else {
						var catlinkattr models.SpCatLinkAttr
						err = o.QueryTable("sp_cat_link_attr").Filter("cat_id",j.Id).Filter("is_deleted",0).One(&catlinkattr)
						if err != nil {
							logs.Error(err)
						}else {
							i.Children = append(i.Children, j)
						}
					}
				}
			}
			if i.IsEnd == 0{
				if len(i.Children) != 0{
					data = append(data, i)
				}
			}else {
				var catlinkattr models.SpCatLinkAttr
				err = o.QueryTable("sp_cat_link_attr").Filter("cat_id",i.Id).Filter("is_deleted",0).One(&catlinkattr)
				if err != nil {
					logs.Error(err)
				}else {
					data = append(data, i)
				}
			}

		}
	}
	return 200,"ok",data ,nil
}

func (c*CatController)  GetCh(){
	cat_id,err := c.GetInt("cat_id")
	if err != nil {
		c.ResponseData(400,"参数格式有误","")
		return
	}
	data,err,msg := models.GetOneCat(cat_id)
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,msg,"")
		return
	}
	c.ResponseData(200,msg,data)
}


type parent_cat struct {
	models.SpPurchaseCategory
	Children interface{} `json:"children"`
}

func (c*CatController) GetAttrCat(){
	//根据属性查分类接口
	attr_id,err:= c.GetInt("attr_id")
	if err != nil{
		c.ResponseData(400,"参数有误","")
		return
	}
	o := orm.NewOrm()
	//var cat_list []orm.Params
	var cat_list []*models.SpPurchaseCategory
	//var DefaultRelsDepth = 2
	_,err = o.QueryTable("sp_purchase_category").Filter("SpCatLinkAttr__Attr",attr_id).Filter("is_deleted",0).Filter("category_status",0).Distinct().All(&cat_list)
	if err != nil {
		logs.Error(err)
		c.ResponseData(200,"查询失败","")
		return
	}

	//data_list := make([]*models.SpPurchaseCategory,0)
	//for _,i := range cat_list{
	//	var cat models.SpPurchaseCategory
	//	id := i.Id
	//	DefaultRelsDepth := i.Level - 1
	//	err = o.QueryTable("sp_purchase_category").Filter("id",id).RelatedSel(DefaultRelsDepth).One(&cat)
	//	if err != nil {
	//		logs.Error(err)
	//		c.ResponseData(400,"查询失败","")
	//		return
	//	}
	//	data_list = append(data_list, &cat)
	//}
	data_list := make([]parent_cat,0)
	for _,i := range cat_list{
		var cat models.SpPurchaseCategory
		id := i.Id
		err = o.QueryTable("sp_purchase_category").Filter("id",id).One(&cat)
		if err != nil {
			logs.Error(err)
			c.ResponseData(400,"查询失败","")
			return
		}
		cat_data := parent_cat{cat,nil}
		cat_data,err:=attr_id_get_cat(cat_data,cat_data.Level,o)
		if err != nil{
			logs.Error(err)
			c.ResponseData(400,"查询失败","")
			return
		}
		data_list = append(data_list, cat_data)
	}
	//a,err := json.Marshal(data_list)
	//fmt.Println(err)
	//c.Data["json"] = data_list
	//c.ServeJSON()

	c.ResponseData(200,"ok",data_list)
}

func  attr_id_get_cat(cat parent_cat,l int,o orm.Ormer) (parent_cat_data parent_cat,err interface{} ){
	if l <= 1{
		return cat,nil
	}
	parent_id := cat.Parent.Id
	var parent models.SpPurchaseCategory
	err = o.QueryTable("sp_purchase_category").Filter("id",parent_id).One(&parent)

	parent_cat_data = parent_cat{parent,cat}
	l -= 1
	return attr_id_get_cat(parent_cat_data,l,o)
}

//通过属性名字搜索分类
func (c*CatController)  GetAttrNameCat(){
	type data struct {
		models.SpPurchaseAttr
		Id            int    `json:"attr_id" `
		State            int    `json:"attribute_state" `
		CreateTime string  `json:"create_time"`
		UpdateTime string  `json:"update_time"`
	}
	o := orm.NewOrm()
	data_dict := make(map[string]interface{})
	cat_name :=c.GetString("cat_name")
	page,err := c.GetInt("page",1)
	if err != nil{
		c.ResponseData(400,"传参有误","")
		return
	}
	page_size,err := c.GetInt("page_size",5)
	if err != nil{
		c.ResponseData(400,"传参有误","")
		return
	}
	queryset := o.QueryTable("sp_purchase_attr").Filter("shop_id",0).Filter("is_deleted",0).OrderBy("-create_time")

	attr_alias_name := c.GetString("attr_alias_name")
	if attr_alias_name != ""{
		queryset = queryset.Filter("attr_alias_name__contains",attr_alias_name)
	}
	attr_name := c.GetString("attr_name")
	if attr_name != ""{
		queryset = queryset.Filter("attr_name__contains",attr_name)
	}
	attr_status,err := c.GetInt("attr_status")
	if err == nil{
		queryset = queryset.Filter("attr_status",attr_status)
	}
	attr_type,err := c.GetInt("attr_type")
	if err == nil{
		queryset = queryset.Filter("attr_type",attr_type)
	}
	//if cat_name == ""{
	//	c.ResponseData(400,"缺少参数","")
	//	return
	//}

	var list orm.ParamsList
	_,err = o.QueryTable("sp_cat_link_attr").Filter("Cat__CategoryName__contains",cat_name).Filter("is_deleted",0).Distinct().ValuesFlat(&list,"attr_id")
	if err != nil || len(list)==0{
		logs.Error(err)
		data_dict["code"] = 200
		data_dict["msg"] = "ok"
		data_dict["data"] = make([]int,0)
		data_dict["page_total"] = 1
		data_dict["total"] = 0
		data_dict["page_index"] = 1
		c.Data["json"] = data_dict
		c.Data["json"] = data_dict
		c.ServeJSON()
		return
	}

	var cat_list []*models.SpPurchaseAttr
	_,err = queryset.Filter("attr_id__in",list).Limit(page_size,(page-1)*page_size).All(&cat_list)
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"查询失败","")
		return
	}
	total,err := queryset.Filter("attr_id__in",list).Count()
	if err != nil{
		logs.Error(err)
		c.ResponseData(400,"查询失败","")
		return
	}
	page_total := math.Ceil(float64(total) / float64(page_size))
	var data_list []*data
	for _,i := range cat_list{
		var data_one data
		data_one.Id = i.Id
		data_one.AttrStatus = i.AttrStatus
		data_one.AttrName = i.AttrName
		data_one.AttrAliasName = i.AttrAliasName
		data_one.AttrType = i.AttrType
		data_one.AddUserId = i.AddUserId
		data_one.Remark = i.Remark
		data_one.Desc = i.Desc
		data_one.ShopId = i.ShopId
		data_one.State = 1
		data_one.CreateTime=time.Unix(int64(i.CreateTime), 0).Format("2006-01-02 15:04:05")
		data_one.UpdateTime=time.Unix(int64(i.UpdateTime), 0).Format("2006-01-02 15:04:05")
		data_list = append(data_list, &data_one)
	}
	data_dict["code"] = 200
	data_dict["msg"] = "ok"
	data_dict["data"] = data_list
	data_dict["page_total"] = page_total
	data_dict["total"] = total
	data_dict["page_index"] = page
	c.Data["json"] = data_dict
	c.ServeJSON()
}

// @Title 查询分类模板是否已修改
func (q *CatController) CheckTempUpdate() {
	var spCatId, tmpCatId int
	o := orm.NewOrm()
	_ = q.Ctx.Input.Bind(&spCatId, "SpCatId")
	_ = q.Ctx.Input.Bind(&tmpCatId, "CatId")
	var spCat models.SpPurchaseCategory
	if err := o.QueryTable(&models.SpPurchaseCategory{}).Filter("Id", spCatId).One(&spCat); err != nil {
		logs.Error("get category failed! reason: ", err.Error())
		q.ResponseData(400, "获取分类信息失败，请重试", nil)
		return
	}
	if int(spCat.PurchaseCatId) != tmpCatId {
		q.ResponseData(413, "分类信息已更新", spCat.PurchaseCatId)
	}
	q.ResponseData(200, "ok", nil)
}








