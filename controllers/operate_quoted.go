// @Time    : 2019-09-05 12:41
// @Author  : Frank
// @Email   : frank@163.com
// @File    : operate_quoted.go
// @Software: GoLand
package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"server-purchase/models"
	"server-purchase/utils"
)

type OperateQuoted struct {
	CommonController
}

// 模板查询接口
func (this *OperateQuoted) Get() {
	/*
		--by frank
		:params cat_id 分类id int
		:return map[string]interface
	*/
	cat_id, _ := this.GetInt("category_id", 0)
	logs.Info(cat_id)
	o := orm.NewOrm()
	var maps []orm.Params
	_, err := o.QueryTable("SpCatLinkAttr").Filter("cat_id", cat_id).OrderBy("sort").Values(&maps, "cat__category_id", "attr__attr_name", "attr__AttrAliasName", "attr_id", "sort", "cat__unit", "cat__remark","is_must")
	if err != nil {
		logs.Error(err)
		return
	}
	for k, v := range maps {
		maps[k]["attr_alias_name"] = v["Attr__AttrAliasName"]
		maps[k]["attr_name"] = v["Attr__AttrName"]
		maps[k]["cat_id"] = v["Cat__Id"]
		maps[k]["attr_id"] = v["Attr__Attr"]
		maps[k]["sort"] = v["Sort"]
		maps[k]["unit"] = v["Cat__Unit"]
		maps[k]["remark"] = v["Cat__Remark"]
		maps[k]["is_must"] = v["IsMust"]
		delete(v, "Cat__Id")
		delete(v, "Attr__AttrAliasName")
		delete(v, "Attr__AttrName")
		delete(v, "Attr__Attr")
		delete(v, "Sort")
		delete(v, "Cat__Unit")
		delete(v, "Cat__Remark")
		delete(v, "IsMust")
	}
	this.ResponseData(200, "ok", maps)
}

type TempParam struct {
	AttrID       []int  `json:"attr_id"`
	CatID        int    `json:"cat_id"`
	CateGoryName string `json:"category_name"`
	Remark       string `json:"remark"`
	Unit         string `json:"unit"`
}



type AttrValueParam struct {
	AttrValue []struct {
		AttrID int `json:"attr_id"`
		IsMust int `json:"is_must"`
		} `json:"attr_value"`
	CategoryName string `json:"category_name"`
	Remark       string `json:"remark"`
	Unit         string `json:"unit"`
	CategoryID   int    `json:"category_id"`
}

func (this *OperateQuoted) Post()  {
	var reqBody AttrValueParam
	_ = json.Unmarshal(this.Ctx.Input.RequestBody, &reqBody)
	o:=orm.NewOrm()
	var category models.SpPurchaseCategory
	var attr models.SpPurchaseAttr
	category.Id = reqBody.CategoryID
	_ = o.Read(&category)
	_ = o.Begin()
	category.Unit = reqBody.Unit
	category.CategoryName = reqBody.CategoryName
	category.Remark = reqBody.Remark
	category.PurchaseCatId = 0
	if _, err := o.Update(&category, "Unit", "CategoryName","Remark","PurchaseCatId"); err != nil{
		_ = o.Rollback()
		logs.Error("update purchase category failed! reason:", err.Error())
		this.ResponseData(400, "分类更新失败", nil)
		return
	}
	for _,v :=range reqBody.AttrValue{
		var linkSort models.SpCatLinkAttr
		_ = o.QueryTable("SpCatLinkAttr").OrderBy("-id").One(&linkSort)
		var link models.SpCatLinkAttr
		attr.Id = v.AttrID
		_ = o.Read(&attr)
		link.Cat = &category
		link.Attr = &attr
		link.IsMust = v.IsMust
		link.Sort = linkSort.Sort + 1
		_, err := o.Insert(&link)
		if err != nil {
			_ = o.Rollback()
			logs.Error(err)
			this.ResponseData(400, "模板插入失败", nil)
			return
		}
	}
	_ = o.Commit()
	this.ResponseData(200, "ok", "")
}


// 模板修改接口
func (this *OperateQuoted) Put() {
	categoryId, _ := this.GetInt("category_id", 0)
	var reqBody AttrValueParam
	// 查询是否存在
	o := orm.NewOrm()
	var cat models.SpPurchaseCategory
	cat.Id = categoryId
	count, _ := o.QueryTable("SpPurchaseCategory").Filter("category_id", categoryId).Count()
	if count < 0 {
		this.ResponseData(400, "分类参数错误", "")
		return
	}
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &reqBody); err != nil{
		logs.Error(err)
		this.ResponseData(400, "参数错误", "")
		return
	}

	var category models.SpPurchaseCategory
	var attr models.SpPurchaseAttr
	category.Id = categoryId
	_ = o.Read(&category)
	_ = o.Begin()
	// TODO 判断是否需要修改分类名称
	if reqBody.CategoryName != "" {
		category.Remark = reqBody.Remark
		category.PurchaseCatId = 0
		category.CategoryName = reqBody.CategoryName
		category.Unit = reqBody.Unit
		if _, err := o.Update(&category, "Remark", "CategoryName", "PurchaseCatId","Unit"); err!=nil{
			logs.Error(err)
			this.ResponseData(400, "修改分类失败", "")
			return
		}
	}
	// TODO:先删除再添加
	if _, err:= o.QueryTable("SpCatLinkAttr").Filter("cat_id", categoryId).Delete();err!=nil{
		logs.Error(err)
		this.ResponseData(400, "操作失败", "")
		return
	}
	for _, v := range reqBody.AttrValue {
		var linkSort models.SpCatLinkAttr
		_ = o.QueryTable("SpCatLinkAttr").OrderBy("-id").One(&linkSort)
		var link models.SpCatLinkAttr
		attr.Id = v.AttrID
		_ = o.Read(&attr)
		link.Cat = &category
		link.Attr = &attr
		link.IsMust = v.IsMust
		link.Sort = linkSort.Sort + 1
		if _, err:= o.Insert(&link);err!=nil{
			_ = o.Rollback()
			logs.Error(err)
			this.ResponseData(400, "修改模板失败", "")
		}
	}
	_ = o.Commit()
	this.ResponseData(200, "ok", "")
}

// 模板属性删除接口
func (this *OperateQuoted) Delete() {
	category_id, _ := this.GetInt("category_id")
	attr_id, _ := this.GetInt("attr_id")
	if category_id == 0 || attr_id == 0 {
		this.ResponseData(400, "请求参数不能为空", nil)
	}
	o := orm.NewOrm()
	var link models.SpCatLinkAttr
	var category models.SpPurchaseCategory
	var attr models.SpPurchaseAttr
	category.Id = category_id
	attr.Id = attr_id
	link.Cat = &category
	link.Attr = &attr
	_, err := o.Delete(&link)
	if err != nil {
		this.ResponseData(400, "删除失败", nil)
	}
	this.ResponseData(200, "ok", nil)
}

// 属性结构体
type GoodsAttr struct {
	CommonController
}

// 属性名联想
func (this *GoodsAttr) Get() {
	keyword := this.GetString("keyword")
	o := orm.NewOrm()
	var attr []models.SpPurchaseAttr
	o.QueryTable("SpPurchaseAttr").Filter("attr_alias_name__contains", keyword).Filter("is_deleted", 0).Filter("shop_id",0).All(&attr)
	this.ResponseData(200, "ok", attr)
}

// 属性上下排序
func (this *GoodsAttr) Put() {
	category_id, _ := this.GetInt("category_id", 0)
	var params map[string][]int
	o := orm.NewOrm()
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &params)
	if err != nil {
		logs.Error(err)
		this.ResponseData(400, "参数错误", "")
		return
	}
	if len(params["attr_id"]) != 2 {
		this.ResponseData(400, "参数错误", "")
		return
	}
	var goods_link []models.SpCatLinkAttr
	o.QueryTable("SpCatLinkAttr").Filter("attr_id__in", params["attr_id"]).Filter("cat_id", category_id).OrderBy("-id").Limit(2).All(&goods_link)
	goods_attr_first := goods_link[0].Sort
	goods_attr_second := goods_link[1].Sort
	goods_link[0].Sort = goods_attr_second
	goods_link[1].Sort = goods_attr_first
	o.Begin()
	_, err = o.Update(&goods_link[0])
	if err != nil {
		o.Rollback()
		this.ResponseData(400, "操作失败", nil)
		return
	}
	_, err = o.Update(&goods_link[1])
	if err != nil {
		o.Rollback()
		this.ResponseData(400, "操作失败", nil)
		return
	}
	o.Commit()
	this.ResponseData(200, "ok", "")
}

//根据分类id返回属性和属性值
func GetCatAttrValue(cat_id int) (data []interface{}){
	/*
	:TODO params cat_id 分类id return map[string]interface{}
	*/
//func (this *GoodsAttr) GetCatAttrValue() {
	//cat_id:=39
	o := orm.NewOrm()
	var maps []orm.Params
	num, _ := o.QueryTable("SpCatLinkAttr").Filter("cat_id", cat_id).Filter("Attr__shop_id",0).Filter("Attr__IsDeleted",0).Filter("Attr__AttrValue__IsDeleted",0).OrderBy("Sort").Values(&maps, "attr_id","is_must", "Attr__AttrAliasName","attr__attr_name", "attr__is_limit_user","attr__AttrValue__attr_value","attr__AttrValue__attr_value_id","attr__AttrValue__unit_type_id","attr__AttrValue__Sort","Attr__AttrValue__Customize")

	logs.Info(num)
	if num > 0{
		var attr_list []interface{}
		for _,v:=range maps{
			returnData := make(map[string]interface{})
			if utils.ContainsInt(attr_list,v["Attr__Attr"])==false {
				attr_list = append(attr_list, v["Attr__Attr"])
				var attr_value_list []interface{}
				var attr_value_name_list []interface{}
				for _, v1 := range maps {
					attr_value_k_v:=make(map[string]interface{})
					if v["Attr__Attr"] == v1["Attr__Attr"] && utils.ContainsInt(attr_value_list, v1["Attr__AttrValue__Id"]) == false {
						attr_value_list = append(attr_value_list, v1["Attr__AttrValue__Id"])
						attr_value_k_v["ValueId"] = v1["Attr__AttrValue__Id"]
						attr_value_k_v["AttrValue"] = v1["Attr__AttrValue__AttrValue"]
						attr_value_k_v["UnitTypeId"] = v1["Attr__AttrValue__UnitTypeId"]
						attr_value_k_v["Sort"] = v1["Attr__AttrValue__Sort"]
						attr_value_k_v["Customize"] = v1["Attr__AttrValue__Customize"]
						attr_value_name_list = append(attr_value_name_list,attr_value_k_v)
					} else {
						continue
					}
				}
				returnData["AttrId"] = v["Attr__Attr"]
				returnData["AttrName"] = v["Attr__AttrName"]
				returnData["AttrAliasName"] = v["Attr__AttrAliasName"]
				returnData["IsLimitUser"] = v["Attr__IsLimitUser"]
				returnData["IsMust"] = v["IsMust"]
				returnData["AttrValues"] = attr_value_name_list
				//data = append(data,returnData)
			}else {
				continue
			}
			data = append(data,returnData)
		}
	}
	return
	//this.ResponseData(200,"ok",returnData)
}

func (this *GoodsAttr) CatAttrTmp() {
	var catId int
	_ = this.Ctx.Input.Bind(&catId, "cat_id")
	data := GetCatAttrValue(catId)
	this.ResponseData(200, "ok", data)
}
