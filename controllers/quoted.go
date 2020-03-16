// @Time    : 2019-09-02 16:25
// @Author  : Frank
// @Email   : frank@163.com
// @File    : quoted.go
// @Software: GoLand
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"math/rand"
	"server-purchase/models"
	"server-purchase/utils"
	"server-purchase/utils/constant"
	"strconv"
	"strings"
	"time"
)

type Quoted struct {
	CommonController
}

// @Title get purchase info for edit and republish
func (q *Quoted) Get() {
	var quoteId, userId int
	var respData models.EditDetailResp
	var o = orm.NewOrm()
	_ = q.Ctx.Input.Bind(&quoteId, "quote_id")
	_ = q.Ctx.Input.Bind(&userId, "user_id")

	// 查询采购单详情
	var quote models.SpSellerQuotedBuyer
	if err := quote.GetQuoteByIdAndBuyerId(quoteId, userId); err != nil {
		logs.Error("Get quote info failed! Reason: ", err)
		q.ResponseData(400, "查询数据失败", nil)
		return
	}
	if constant.OutUse == quote.Status {
		q.ResponseData(400, "采购单不存在，请确认后重试", nil)
		return
	}

	// 列表处理
	resp1 := &quote.QuoteResp
	resp2 := &respData.Purchase
	resp2.QuoteResp = *resp1

	resp2.TransWithGet(*resp1)
	respData.Purchase.TransCommon.MobileNumber = resp1.MobileNumber

	// 获取用户信息
	var userChan = make(chan map[int]interface{}, 1)
	go utils.GetUserInfoAsyc([]int{quote.BuyerId}, userChan)
	//userInfoList := utils.GetUserInfo([]int{quote.BuyerId})
	//respData.User = userInfoList[quote.BuyerId]

	// TODO get backstage edit history
	var editHistory models.SpPurchaseEditLog
	if err := o.QueryTable(&models.SpPurchaseEditLog{}).Filter("QuoteId", quoteId).Filter("IsDeleted", constant.ZERO).One(&editHistory); err == nil {
		respData.History = utils.StringToMap(editHistory.LogDetail)
	}

	// 多商品采购时，无分类及属性信息，直接返回
	var companyPrintEditResp models.CompanyPrintEditDetailResp
	companyPrintEditResp.QuoteCommon = respData.Purchase
	companyPrintEditResp.History = respData.History
	companyPrintEditResp.User = respData.User
	if constant.Batch == quote.PurchaseType || constant.CompanyPrint == quote.PurchaseType {
		q.ResponseData(200, "ok", companyPrintEditResp)
		return
	}

	// 判断分类
	catInfo, err := models.GetTmpCat(quote.TmpCatId.CatId)
	if err != nil {
		logs.Error("Get quote info failed! Reason: ", err)
		q.ResponseData(400, "查询模板数据失败", nil)
		return
	}
	// if the cat is others, then return the response
	if catInfo.SpCatId == constant.OtherCat {
		var tmpData = new(models.TmpPurchaseAttr)
		catInfo.Attr = []*models.TmpPurchaseAttr{tmpData}
		respData.Template = *catInfo
		respData.CatList = []string{catInfo.CatName}
		q.ResponseData(200, "ok", respData)
		return
	}

	d, err, _ := models.GetOneCat(catInfo.SpCatId)
	if err != nil {
		logs.Error("Get quote info failed! Reason: ", err)
		q.ResponseData(400, "查询分类信息失败", nil)
		return
	}
	var resp = make([]string, 0)
	models.GetCatName(d, &resp)
	respData.CatList = resp

	// 获取属性详情
	attrValue, err := models.GetAttrValueName(quote.NormsRequirement)
	if err != nil {
		logs.Error("Get attribute value failed! Reason: ", err.Error())
		q.ResponseData(400, "查询采购商品属性失败", nil)
		return
	}
	respData.AttrName = attrValue

	//查询分类详情
	cat := models.TmpPurchaseCategory{CatId: quote.TmpCatId.CatId}
	_ = o.Read(&cat)
	respData.Template = cat
	num, err := o.LoadRelated(&cat, "Attr")
	logs.Info(num)

	respData.Template.Attr = cat.Attr
	for i, attr := range respData.Template.Attr {
		respData.Template.Attr[i].AttrValues = attr.AttrValues
		attr := models.TmpPurchaseAttr{Id: attr.Id}
		_ = o.Read(&attr)
		num, err = o.LoadRelated(&attr, "AttrValues")
		// 删除非关联自定义属性值
		for j := 0; j < len(attr.AttrValues); j++ {
			if attr.AttrValues[j].Customize == 1 && quote.PurchaseSn != attr.AttrValues[j].PurchaseSn {
				attr.AttrValues = append(attr.AttrValues[:j], attr.AttrValues[j+1:]...)
				j--
			}
		}

		respData.Template.Attr[i].AttrValues = attr.AttrValues
	}
	userInfoList := <-userChan
	respData.User = userInfoList[quote.BuyerId]
	q.ResponseData(200, "ok", respData)
}

// @Title create purchase
func (q *Quoted) Post() {
	var ReqBody models.QuoteReq

	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &ReqBody)
	o := orm.NewOrm()
	now := time.Now()

	// compare the category info between request info and backstage info
	if ReqBody.Purchase.PurchaseType == constant.Single && ReqBody.Template.SpCatId != constant.OtherCat {
		if len(ReqBody.Purchase.TransCommon.NormsRequirement) == 0 {
			q.ResponseData(400, "未选择分类属性", nil)
			return
		}
		var Attrs []models.TmpPurchaseAttr
		cat := GetCatAttrValue(ReqBody.Template.SpCatId)
		d, _ := json.Marshal(cat)
		_ = json.Unmarshal(d, &Attrs)
		tmpAttrs := ReqBody.Template.Attr
		if result := CompareAttrAndAttrs(Attrs, tmpAttrs); !result {
			q.ResponseData(413, "分类信息已更新，请重新获取编辑采购单", nil)
			return
		}
		logs.Info(cat)
	}

	// check request data
	var userId = ReqBody.BuyerId
	// 数据提取，组合
	tmpCatId := ReqBody.Template.TmpCatId
	c := new(models.TmpPurchaseCategory)
	result, err := o.QueryTable(&models.TmpPurchaseCategory{}).Filter("CatId", tmpCatId).All(c)
	if err != nil {
		q.ResponseData(400, "数据查询失败", nil)
		return
	}

	var purchase = new(models.SpSellerQuotedBuyer)
	ReqBody.Purchase.QuoteResp.BackstageEdit = *new(models.BackstageEdit)
	quoteResp := ReqBody.Purchase.QuoteResp
	quoteResp1 := ReqBody.Purchase
	// 公共参数及多商品采购参数校验
	if err = ReqBody.PurchasePreSave(now.Unix()); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}

	//处理切片
	quoteResp.TransWithSave(quoteResp1.TransCommon)
	quoteResp.MobileNumber = quoteResp1.TransCommon.MobileNumber

	purchase.BuyerId = userId

	rand.Seed(now.Unix())
	quoteResp.PurchaseSn = fmt.Sprintf("QD%d%02d%02d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), rand.Intn(99))

	purchase.QuoteResp = quoteResp
	var purchaseName = "求购"
	if quoteResp.PurchaseType == constant.Single {
		purchaseName += ReqBody.Template.CatName
	} else {
		purchaseName += "多种商品共"
	}
	unit := quoteResp.Unit
	if unit == "" {
		unit = "件"
	}
	purchaseName += strconv.Itoa(quoteResp.GoodsQty) + unit
	purchase.GoodsName = purchaseName

	// 创建时间及更新时间
	purchase.CreateTime = now.Unix()
	purchase.UpdateTime = now.Unix()

	// save batch purchase or companyPrint purchase
	if ReqBody.Purchase.PurchaseType == constant.Batch || ReqBody.Purchase.PurchaseType == constant.CompanyPrint {
		// TODO auto audit when the purchase type is company purchase
		//if ReqBody.Purchase.PurchaseType == constant.CompanyPrint {
		//	purchase.AuditStatus = constant.APPROVED
		//}
		if _, err = o.Insert(purchase); err != nil {
			logs.Error("Purchase order create failed! reason: ", err.Error())
			q.ResponseData(400, "采购单创建失败", nil)
			return
		}
		q.ResponseData(200, "采购发布成功",
			map[string]interface{}{
				"PurchaseSn": purchase.PurchaseSn,
				"PurchaseId": purchase.Id})
		return
	}

	// save others type purchase
	if ReqBody.Template.SpCatId == constant.OtherCat {
		var cat = new(models.TmpPurchaseCategory)
		if err = o.QueryTable(&models.TmpPurchaseCategory{}).Filter("SpCatId", constant.OtherCat).One(cat); err == nil {
			purchase.TmpCatId = cat
			_, err = o.Insert(purchase)
			if err != nil {
				logs.Error("save purchase info failed! reason: ", err.Error())
				q.ResponseData(400, "数据存储失败", nil)
				return
			}
			q.ResponseData(200, "采购发布成功",
				map[string]interface{}{
					"PurchaseSn": purchase.PurchaseSn,
					"PurchaseId": purchase.Id})
			return
		}
	}
	// save
	if result != 0 {
		// TODO save customize attribute value
		var AttrValues = make([]models.TmpPurchaseAttrValue, 0)
		attrs := ReqBody.Template.Attr
		for _, attr := range attrs {
			if attr.IsLimitUser == 1 {
				// TODO get attribute object
				var CustomizeAttr models.TmpPurchaseAttr
				if err = o.QueryTable(models.TmpPurchaseAttr{}).Filter("AttrName", attr.AttrName).Filter("Cat", ReqBody.Template.TmpCatId).One(&CustomizeAttr); err != nil {
					logs.Error("select tmp customize attr failed! reason: ", err.Error())
					q.ResponseData(400, "自定义属性异常，请联系管理员确认模板", nil)
					return
				}
				for _, attrValue := range attr.AttrValues {
					if attrValue.Customize == 1 {
						attrValue.Id = constant.ZERO
						attrValue.Cat = c
						attrValue.Attr = &CustomizeAttr
						attrValue.PurchaseSn = purchase.PurchaseSn
						AttrValues = append(AttrValues, *attrValue)
					}
				}
			}
		}

		// save
		if err := o.Begin(); err != nil {
			logs.Error("start transaction failed!")
			q.ResponseData(400, "采购单发布失败，请重试", nil)
			return
		}
		if len(AttrValues) > constant.ZERO {
			if _, errs := o.InsertMulti(len(AttrValues), &AttrValues); errs != nil {
				logs.Error("Save Customize AttrValue failed! Reason: ", errs.Error())
				_ = o.Rollback()
				q.ResponseData(400, "采购单发布失败，请重试", nil)
				return
			}
		}

		var list []orm.ParamsList
		// 查询所有tmp属性值表，获取选中id
		qs := o.QueryTable(models.TmpPurchaseAttrValue{})
		cond := orm.NewCondition()
		// select * from tmp_purchase_attr_value where (value_id in () and value_id > 0 and cat_id=?) or (value_id < 0 and value_id in () and order_sn=?)
		or := cond.AndCond(cond.And("ValueId__in", quoteResp1.TransCommon.NormsRequirement).And(
			"Cat", tmpCatId).And("ValueId__gt", constant.ZERO)).OrCond(cond.And(
			"PurchaseSn", purchase.PurchaseSn).And("ValueId__in", quoteResp1.TransCommon.NormsRequirement))
		r, err := qs.SetCond(or).ValuesList(&list, "Id", "ValueId")
		logs.Info(r, err)

		// 按原有顺序保存属性值id
		var li = make([]int64, 0)
		for _, valueId := range quoteResp1.TransCommon.NormsRequirement {
			// jump to the next cycle when unnecessary attribute without choose a value
			if constant.ZERO == valueId {
				li = append(li, int64(constant.ZERO))
				continue
			}
			for _, result := range list {
				if valueId == int(result[1].(int64)) {
					li = append(li, result[0].(int64))
				}
			}
		}
		normal, err := json.Marshal(li)
		purchase.QuoteResp.NormsRequirement = string(normal)
		purchase.TmpCatId = c

		// 存储采购单信息
		if _, err = o.Insert(purchase); err != nil {
			logs.Error("Save Purchase Info Failed! PurchaseSn: ", purchase.PurchaseSn, ", Reason: ", err.Error())
			_ = o.Rollback()
			q.ResponseData(400, "采购单发布失败，请重试", nil)
			return
		}
		if err = o.Commit(); err != nil {
			logs.Error("commit transaction failed! reason: ", err.Error())
			q.ResponseData(400, "采购单发布失败，请重试", nil)
			return
		}
	} else {
		var attrList = make([]*models.TmpPurchaseAttr, constant.ZERO)
		var valueList = make([]*models.TmpPurchaseAttrValue, constant.ZERO)
		var cat = new(models.TmpPurchaseCategory)
		cat = &ReqBody.Template.TmpPurchaseCategory
		purchase.TmpCatId = cat
		attrs := ReqBody.Template.Attr
		logs.Info(attrs)
		for _, attr := range attrs {
			purchaseAttr := attr
			purchaseAttr.Cat = cat
			attrList = append(attrList, purchaseAttr)
			for _, value := range attr.AttrValues {
				if constant.ZERO == int(attr.IsLimitUser) && constant.ZERO >= value.ValueId {
					continue
				}
				a := value
				a.Attr = purchaseAttr
				a.Cat = cat
				valueList = append(valueList, a)
			}
		}
		logs.Info(valueList)
		_ = o.Begin()
		catId, err := o.Insert(cat)
		if err != nil {
			_ = o.Rollback()
			logs.Error("save category info failed! reason: ", err.Error())
			q.ResponseData(400, "分类信息异常，请重试", nil)
			return
		}
		logs.Info("保存分类成功，分类id：", catId)
		// 修改sp分类关联tmp分类
		spCat := models.SpPurchaseCategory{Id: int(ReqBody.Template.SpCatId)}
		err = o.Read(&spCat)
		if err == nil {
			spCat.PurchaseCatId = int8(catId)
			if _, err := o.Update(&spCat, "PurchaseCatId"); err != nil {
				_ = o.Rollback()
				logs.Error("save category info failed! reason: ", err.Error())
				q.ResponseData(400, "分类信息异常，请重试", nil)
				return
			}
		}

		qs := o.QueryTable(&models.TmpPurchaseAttr{})
		i, _ := qs.PrepareInsert()
		for _, attribute := range attrList {
			attrId, _ := i.Insert(attribute)
			logs.Info("保存属性成功，Id: ", attrId)
		}
		if err = i.Close(); err != nil {
			_ = o.Rollback()
			logs.Error("Save TmpAttr failed! Reason: ", err.Error())
			q.ResponseData(400, "数据存储失败", nil)
			return
		}

		qs = o.QueryTable(&models.TmpPurchaseAttrValue{})
		i, _ = qs.PrepareInsert()
		for _, value := range valueList {
			// TODO 自动以属性值添加关联
			if value.Customize == 1 {
				value.PurchaseSn = purchase.PurchaseSn
			}
			valueId, _ := i.Insert(value)
			logs.Info("保存属性值成功，Id: ", valueId)
		}
		if err = i.Close(); err != nil {
			_ = o.Rollback()
			logs.Error("Save TmpAttrValue failed! Reason: ", err.Error())
			q.ResponseData(400, "数据存储失败", nil)
			return
		}

		var newValueList = make([]int, 0)
		for _, i := range ReqBody.Purchase.TransCommon.NormsRequirement {
			if constant.ZERO == i {
				newValueList = append(newValueList, constant.ZERO)
				continue
			}
			for _, j := range valueList {
				if j.ValueId == i {
					newValueList = append(newValueList, j.Id)
				}
			}
		}

		normal, _ := json.Marshal(newValueList)
		purchase.NormsRequirement = string(normal)
		_, err = o.Insert(purchase)
		if err != nil {
			logs.Error("数据存储失败, ", err.Error())
			_ = o.Rollback()
			q.ResponseData(400, "数据存储失败", nil)
			return
		}
		if err := o.Commit(); err != nil {
			logs.Error("commit transaction failed!")
			q.ResponseData(400, "采购单发布失败，请重试", nil)
			return
		}
	}
	q.ResponseData(200, "采购发布成功",
		map[string]interface{}{
			"PurchaseSn": purchase.PurchaseSn,
			"PurchaseId": purchase.Id})
}

// @Title edit purchase
func (q *Quoted) Put() {
	var reqBody models.UpdateQuoteReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	now := time.Now().Unix()

	// 入参校验
	if reqBody.Id <= 0 || reqBody.BuyerId <= 0 {
		q.ResponseData(400, "用户或采购单信息不全", nil)
		return
	}
	if err := reqBody.PurchasePreSave(now); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}

	// 获取采购单详情
	var quote models.SpSellerQuotedBuyer
	if err := quote.GetQuoteByIdAndBuyerId(reqBody.Id, reqBody.BuyerId); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	// only on audit or on purchase without quote if purchase type is company print, or backstage edit purchase can be edit
	if !(constant.OnPurchase == quote.Status && quote.QuotedEndTime >= now && ((((constant.InReview == quote.AuditStatus) ||
		(constant.APPROVED == quote.AuditStatus && quote.QuoteCount == constant.ZERO)) && constant.CompanyPrint == quote.PurchaseType) ||
		(constant.UnApprove == quote.AuditStatus))) {
		logs.Error("Update quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: purchase status is not allowed edit!")
		q.ResponseData(400, "该采购单不允许编辑，请确认状态", nil)
		return
	}

	// 更新数据
	//if quote.BackstageEdit == constant.No {
	//	reqBody.Purchase.BackstageEdit = *new(models.BackstageEdit)
	//}
	quote.AuditStatus = constant.InReview
	quote.BackstageEdit = constant.No
	quote.BackstageEditTime = int64(constant.ZERO)
	quote.Status = constant.OnPurchase
	// TODO new status field
	quote.QuoteStatus = constant.AuditingStatus
	quoteId, err := quote.UpdateQuote(reqBody, now, constant.No)
	if err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	q.ResponseData(200, "ok", map[string]interface{}{
		"PurchaseId": quoteId, "PurchaseSn": quote.PurchaseSn})
}

// @Title republish purchase
func (q *Quoted) RePublish() {
	var reqBody models.UpdateQuoteReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	now := time.Now().Unix()

	// 入参校验
	if reqBody.Id <= 0 || reqBody.BuyerId <= 0 {
		q.ResponseData(400, "用户或采购单信息不全", nil)
		return
	}
	if err := reqBody.PurchasePreSave(now); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}

	// check whether the template is changed
	var spCat models.SpPurchaseCategory
	o := orm.NewOrm()
	if err := o.QueryTable(&models.SpPurchaseCategory{}).Filter("Id", reqBody.Template.SpCatId).Filter("PurchaseCatId", reqBody.Template.CatId).One(&spCat); err != nil {
		q.ResponseData(413, "分类信息已修改，请重新获取", nil)
		return
	}
	// republic purchase
	var quoteId int
	var purchaseSn string
	var err error
	quoteId, purchaseSn, err = reqBody.UpdateQuote()
	if err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	q.ResponseData(200, "ok", map[string]interface{}{
		"PurchaseId": quoteId, "PurchaseSn": purchaseSn})
}

// @Title get purchase detail
func (q *Quoted) GetPurchaseDetail() {
	var quoteId int
	_ = q.Ctx.Input.Bind(&quoteId, "quote_id")
	var data models.QuoteDetail
	data, err := models.GetPurchaseDetail(quoteId, constant.OPERATION)
	if err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}

	if constant.OutUse == data.Status {
		q.ResponseData(400, "采购单不存在，请确认后重试", nil)
		return
	}
	q.ResponseData(200, "ok", data)
}

// @Title purchase list in user center
func (q *Quoted) GetUserPurchaseList() {
	var buyerId int
	var PurchaseType int
	if err := q.Ctx.Input.Bind(&buyerId, "BuyerId"); err != nil {
		q.ResponseData(400, "用户信息异常", nil)
		return
	}
	if err := q.Ctx.Input.Bind(&PurchaseType, "PurchaseType"); err != nil {
		PurchaseType = constant.ZERO
	}
	if !utils.ContainsInt([]interface{}{constant.ZERO, constant.Single, constant.Batch, constant.CompanyPrint}, PurchaseType) {
		q.ResponseData(400, "请输入正确的采购单类型", nil)
		return
	}
	if constant.ZERO >= buyerId {
		q.ResponseData(400, "用户信息异常", nil)
		return
	}
	status := q.Ctx.Input.Query("status")
	var data []models.SpSellerQuotedBuyer
	o := orm.NewOrm()
	qs := o.QueryTable(&models.SpSellerQuotedBuyer{}).OrderBy("-CreateTime")
	now := time.Now().Unix()

	if status == "traded" {
		qs = qs.Filter("Status", constant.Traded)
	} else if status == "invalid" {
		//select * from sp_seller_quoted_buyer where buyer_id = 623 and ((((`quoted_end_time` > 1568531832 and quoted_end_time < 1568618232 and order_sn <= 0) or quoted_end_time < 1568531832) and status in (0, 2)) or status=3);
		cond := orm.NewCondition()
		invalid := cond.AndCond(cond.AndCond(cond.AndCond(cond.AndCond(cond.And("QuotedEndTime__lt", now-constant.PurchaseOverTime).
			And("SuccessSellerId__gt", constant.ZERO)).OrCond(cond.And("QuotedEndTime__lt", now-constant.PurchaseFreezeTime).And("QuoteCount__gt", 0).And("SuccessSellerId__lte", constant.ZERO)).
			OrCond(cond.And("QuotedEndTime__lt", now).And("QuoteCount", 0))).
			AndCond(cond.And("Status__in", constant.OnPurchase, constant.Paused)).OrCond(cond.And("Status", constant.Canceled))))
		qs = qs.SetCond(invalid)
	} else if status == "available" {
		//select * from sp_seller_quoted_buyer where buyer_id = 623 and ((quoted_end_time > 1568618232) or (quoted_end_time > 1568531832 and quoted_end_time < 1568618232 and order_sn > 0)) and status in (0, 2);
		cond := orm.NewCondition()
		available := cond.AndCond(cond.And("Status__in", constant.OnPurchase, constant.Paused)).
			AndCond(cond.AndCond(cond.And("QuotedEndTime__gt", now-constant.PurchaseFreezeTime).And("QuoteCount__gt", 0)).
				OrCond(cond.And("QuotedEndTime__gt", now-constant.PurchaseOverTime).
					And("QuotedEndTime__lt", now-constant.PurchaseFreezeTime).
					And("SuccessSellerId__gt", constant.ZERO)).OrCond(cond.And("QuotedEndTIme__gt", now)))
		qs = qs.SetCond(available)
	} else if status == "all" {
		qs = qs
	} else {
		q.ResponseData(400, "无此状态", nil)
		return
	}
	var allStatus = []int{constant.OnPurchase, constant.Traded, constant.Paused, constant.Canceled}
	qs = qs.Filter("BuyerId", buyerId).Filter("Status__in", allStatus)

	if utils.ContainsInt([]interface{}{constant.Single, constant.Batch, constant.CompanyPrint}, PurchaseType) {
		qs = qs.Filter("PurchaseType", PurchaseType)
	}
	q.count, _ = qs.Count()
	//确定每页显示数
	//获取页码
	q.GetPagination()
	//查询数据库部分数据
	_, _ = qs.Limit(q.pageSize, q.start).All(&data)

	var d []models.UserPurchaseList
	for _, v := range data {
		var tmpResp models.UserPurchaseList
		tmpResp.SpSellerQuotedBuyer = v
		tmpResp.TransCommon.TransWithGet(v.QuoteResp)
		tmpResp.MobileNumber = v.MobileNumber

		// 获取属性详情
		if v.PurchaseType == constant.Single && len(tmpResp.NormsRequirement) > constant.ZERO {
			attrValue, err := models.GetAttrValueName(tmpResp.NormsRequirement)
			if err != nil {
				q.ResponseData(400, "查询采购商品属性失败", nil)
				return
			}
			tmpResp.AttrName = attrValue
		}
		d = append(d, tmpResp)
	}
	if status == "available" {
		for i, _ := range d {
			s := &d[i]
			if s.QuotedEndTime > (now-constant.PurchaseOverTime) && s.QuotedEndTime < (now-constant.PurchaseFreezeTime) {
				d[i].Status = constant.Freeze
			}
		}
	}
	var res []models.UserPurchaseListResp
	if err := utils.TransferToStruct(d, &res); err != nil {
		logs.Error("Transfer to UserPurchaseListResp failed! reason:", err.Error())
		q.ResponseData(400, "查询数据转换失败", nil)
		return
	}

	// H5全部采购列表状态
	for i := 0; i < len(res); i++ {
		v := &res[i]
		v.PurchaseStatus = v.Status
		result, _ := utils.Contain(v.Status, []int{constant.Traded, constant.Paused, constant.Canceled})
		// 报价中状态
		if !result && constant.APPROVED == v.AuditStatus &&
			(now < v.QuotedEndTime || (now-constant.PurchaseFreezeTime < v.QuotedEndTime && v.QuotedEndTime < now && constant.ZERO < v.QuoteCount) ||
				(now-constant.PurchaseOverTime < v.QuotedEndTime && v.QuotedEndTime < now-constant.PurchaseFreezeTime && constant.ZERO < v.SuccessSellerId)) {
			v.PurchaseStatus = constant.OnPurchase
		}
		// 审核中状态
		if constant.InReview == v.AuditStatus && constant.OnPurchase == v.Status {
			v.PurchaseStatus = constant.Auditing
		}
		// 已失效状态
		if v.Status != constant.Canceled && (((constant.ZERO >= v.QuoteCount && now > v.QuotedEndTime) ||
			(now-constant.PurchaseFreezeTime >= v.QuotedEndTime && constant.ZERO >= v.SuccessSellerId) ||
			(now-constant.PurchaseOverTime >= v.QuotedEndTime)) && constant.Traded != v.Status) {
			v.PurchaseStatus = constant.Expired
		}
		// 已取消状态
		if constant.Canceled == v.Status {
			v.PurchaseStatus = constant.Canceled
		}
		// 审核未通过状态
		if constant.UnApprove == v.AuditStatus {
			v.PurchaseStatus = constant.AuditFailed
		}
		// 冻结状态
		if v.QuotedEndTime > (now-constant.PurchaseOverTime) && v.QuotedEndTime < (now-constant.PurchaseFreezeTime) && v.SuccessSellerId > constant.ZERO {
			v.PurchaseStatus = constant.Freeze
		}
		// 代编辑状态
		if constant.Yes == v.BackstageEdit && constant.UnApprove == v.AuditStatus && v.QuotedEndTime >= now {
			v.PurchaseStatus = constant.BackstageEdit
			v.Status = constant.BackstageEdit
		}
	}

	q.ResponseData(200, "ok", res)
}

// @Title purchase count in user center
func (q *Quoted) GetUserPurchaseListNum() {
	var buyerId, purchaseType int
	var purchaseTypeList = make([]int, 0)
	if err := q.Ctx.Input.Bind(&buyerId, "BuyerId"); err != nil {
		q.ResponseData(400, "用户信息异常", nil)
		return
	}
	if err := q.Ctx.Input.Bind(&purchaseType, "PurchaseType"); err != nil {
		q.ResponseData(400, "请输入采购单类型", nil)
		return
	}
	if !utils.ContainsInt([]interface{}{constant.ZERO, constant.Single, constant.Batch, constant.CompanyPrint}, purchaseType) {
		q.ResponseData(400, "请输入正确的采购单类型", nil)
		return
	}
	if utils.ContainsInt([]interface{}{constant.Single, constant.Batch, constant.CompanyPrint}, purchaseType) {
		purchaseTypeList = append(purchaseTypeList, purchaseType)
	}
	if constant.ZERO == purchaseType {
		purchaseTypeList = append(purchaseTypeList, []int{constant.Single, constant.Batch, constant.CompanyPrint}...)
	}
	// 按条件查询列表数量
	o := orm.NewOrm()
	qs := o.QueryTable(&models.SpSellerQuotedBuyer{})
	now := time.Now().Unix()
	var allStatus = []int{constant.OnPurchase, constant.Traded, constant.Paused, constant.Canceled}

	traded, err := qs.Filter("Status", constant.Traded).Filter("BuyerId", buyerId).Filter("PurchaseType__in", purchaseTypeList).Filter("Status__in", allStatus).Count()
	if err != nil {
		logs.Error("select traded purchase list failed! buyer_id: ", buyerId, ", reason: ", err.Error())
		q.ResponseData(400, "查询采购单列表失败", nil)
		return
	}

	cond := orm.NewCondition()
	invalid := cond.AndCond(cond.AndCond(cond.AndCond(cond.AndCond(cond.And("QuotedEndTime__lt", now-constant.PurchaseOverTime).
		And("SuccessSellerId__gt", constant.ZERO)).OrCond(cond.And("QuotedEndTime__lt", now-constant.PurchaseFreezeTime).And("QuoteCount__gt", 0).And("SuccessSellerId__lte", constant.ZERO)).
		OrCond(cond.And("QuotedEndTime__lt", now).And("QuoteCount", constant.ZERO))).
		AndCond(cond.And("Status__in", constant.OnPurchase, constant.Paused)).OrCond(cond.And("Status", constant.Canceled))).AndCond(cond.And("PurchaseType__in", purchaseTypeList)).AndCond(cond.And("BuyerId", buyerId).And("Status__in", allStatus)))
	invalidCount, err := qs.SetCond(invalid).Count()
	if err != nil {
		logs.Error("select traded purchase list failed! buyer_id: ", buyerId, ", reason: ", err.Error())
		q.ResponseData(400, "查询采购单列表失败", nil)
		return
	}

	available := cond.AndCond(cond.And("Status__in", constant.OnPurchase, constant.Paused)).
		AndCond(cond.AndCond(cond.And("QuotedEndTime__gt", now-constant.PurchaseFreezeTime).And("QuoteCount__gt", constant.ZERO)).
			OrCond(cond.And("QuotedEndTime__gt", now-constant.PurchaseOverTime).
				And("QuotedEndTime__lt", now-constant.PurchaseFreezeTime).
				And("SuccessSellerId__gt", constant.ZERO)).OrCond(cond.And("QuotedEndTIme__gt", now))).AndCond(cond.And("PurchaseType__in", purchaseTypeList)).AndCond(cond.And("BuyerId", buyerId).And("Status__in", allStatus))
	availableCount, err := qs.SetCond(available).Count()
	if err != nil {
		logs.Error("select traded purchase list failed! buyer_id: ", buyerId, ", reason: ", err.Error())
		q.ResponseData(400, "查询采购单列表失败", nil)
		return
	}

	ListNumResp := models.PurchaseListNumResp{
		Traded:    traded,
		Invalid:   invalidCount,
		Available: availableCount,
	}

	q.ResponseData(200, "ok", ListNumResp)

}

// @Title change purchase status
func (q *Quoted) ChangeQuoteOrderStatus() {
	var reqBody models.ChangeQuoteOrderStatusReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	if result, _ := utils.Contain(reqBody.Status, []int{constant.OnPurchase, constant.Traded, constant.Paused, constant.Canceled}); !result {
		logs.Error("入参状态异常，Status:", reqBody.Status)
		q.ResponseData(400, "非法状态", nil)
		return
	}
	// 获取采购单详情
	var purchase models.SpSellerQuotedBuyer
	o := orm.NewOrm()
	now := time.Now().Unix()
	if err := purchase.GetQuoteByIdAndBuyerId(reqBody.Id, reqBody.BuyerId); err != nil {
		logs.Error("获取采购单信息失败，", err.Error())
		q.ResponseData(400, "获取采购单信息失败", nil)
		return
	}
	// TODO new status field
	//if constant.OutUseStatus == purchase.QuoteStatus {
	//	q.ResponseData(400, "采购单不存在，请确认后重试", nil)
	//	return
	//}
	if constant.OutUse == purchase.Status {
		q.ResponseData(400, "采购单不存在，请确认后重试", nil)
		return
	}
	// TODO new status field
	//if constant.AuditingStatus == purchase.QuoteStatus {
	//	q.ResponseData(400, "该采购单暂未通过审核", nil)
	//	return
	//}
	if constant.APPROVED != purchase.AuditStatus {
		q.ResponseData(400, "该采购单暂未通过审核", nil)
		return
	}

	if reqBody.Status == purchase.Status {
		q.ResponseData(400, "重复操作", nil)
		return
	}
	// TODO new status field
	//oldStatus := purchase.QuoteStatus
	// 修改采购单状态
	// 报价中
	if reqBody.Status == constant.OnPurchase {
		if constant.Paused != purchase.Status || (purchase.QuotedEndTime < now && 0 >= purchase.QuoteCount) || (now-constant.PurchaseFreezeTime > purchase.DeliveryEndTime) {
			q.ResponseData(400, "非法操作", nil)
			return
		}
		purchase.QuoteStatus = constant.OnPurchaseStatus
	}
	// 已成交
	if reqBody.Status == constant.Traded {
		if constant.NULL == purchase.OrderSn || constant.ZERO > purchase.SuccessSellerId {
			q.ResponseData(400, "非法操作", nil)
			return
		}
		purchase.QuoteStatus = constant.TradedStatus
	}
	// 暂停
	if reqBody.Status == constant.Paused {
		if constant.OnPurchase != purchase.Status {
			q.ResponseData(400, "非法操作", nil)
			return
		}
		purchase.QuoteStatus = constant.PausedStatus
	}
	// 已取消
	if reqBody.Status == constant.Canceled {
		result, _ := utils.Contain(purchase.Status, []int{constant.OnPurchase, constant.Paused})
		if !result || constant.NULL != purchase.OrderSn {
			q.ResponseData(400, "非法操作", nil)
			return
		}
		purchase.QuoteStatus = constant.PausedStatus
	}
	// TODO new status field
	//if  oldStatus == purchase.QuoteStatus {
	//	q.ResponseData(400, "重复操作", nil)
	//	return
	//}
	// 更新状态
	purchase.Status = reqBody.Status
	if constant.Traded == reqBody.Status {
		purchase.FixtureDate = now
	}
	if _, err := o.Update(&purchase, "Status", "FixtureDate"); err != nil {
		logs.Error("update purchase status failed! reason:", err.Error())
		q.ResponseData(400, "更新采购单状态失败", nil)
		return
	}
	q.ResponseData(200, "Operation Success!", nil)
}

// @Title audit purchase
func (q *Quoted) PurchaseAudit() {
	var reqBody models.PurchaseAuditReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)

	if result, err := q.Validate(&reqBody); !result {
		q.ResponseData(400, err, nil)
		return
	}

	if result, _ := utils.Contain(reqBody.AuditStatus, []int{constant.APPROVED, constant.UnApprove}); !result {
		logs.Error("入参状态异常，Status:", reqBody.AuditStatus)
		q.ResponseData(400, "非法状态", nil)
		return
	}

	purchase := models.SpSellerQuotedBuyer{Id: reqBody.PurchaseId}
	o := orm.NewOrm()
	if err := o.Read(&purchase); err != nil {
		logs.Error("获取采购单信息失败，", err.Error())
		q.ResponseData(400, "获取采购单信息失败", nil)
		return
	}
	if constant.OutUse == purchase.Status {
		q.ResponseData(400, "采购单不存在，请确认后重试", nil)
		return
	}
	if !(constant.OnPurchase == purchase.Status && constant.InReview == purchase.AuditStatus && time.Now().Unix() <= purchase.QuotedEndTime) {
		q.ResponseData(400, "非待审核采购单，请确认", nil)
		return
	}
	// TODO 设计缺陷，后续产品提需求后放开即可
	//if constant.APPROVED == reqBody.AuditStatus && constant.Yes == purchase.BackstageEdit {
	//	q.ResponseData(400, "请先提交代编辑内容", nil)
	//	return
	//}

	// TODO new status field
	purchase.QuoteStatus = reqBody.AuditStatus

	purchase.AuditStatus = reqBody.AuditStatus
	purchase.AuditMessage = constant.NULL
	if reqBody.AuditStatus == constant.UnApprove {
		reqBody.AuditMessage = strings.TrimSpace(reqBody.AuditMessage)
		if reqBody.AuditMessage == constant.NULL {
			q.ResponseData(400, "请填写未通过审核原因", nil)
			return
		}
		purchase.AuditMessage = reqBody.AuditMessage
	}
	if _, err := o.Update(&purchase, "AuditStatus", "AuditMessage"); err != nil {
		logs.Error("更新采购单审核状态失败, message:", err.Error())
		q.ResponseData(400, "更新采购单审核状态失败", nil)
		return
	}
	q.ResponseData(200, "Operation Success!", nil)
}

// @Title get purchase list in operation backstage
func (q *Quoted) OperationList() {
	// get params
	reqData := q.Ctx.Request.Form
	logs.Info("reqData: ", reqData)
	quoteStart, _ := q.GetInt("quoteStart")
	quoteEnd, _ := q.GetInt("quoteEnd")
	deliveryStart, _ := q.GetInt("deliveryStart")
	deliveryEnd, _ := q.GetInt("deliveryEnd")
	createStart, _ := q.GetInt("createStart")
	createEnd, _ := q.GetInt("createEnd")
	purchaseSn := q.GetString("purchaseSn")
	fixtureStart, _ := q.GetInt("fixtureStart")
	fixtureEnd, _ := q.GetInt("fixtureEnd")
	qtyStart, _ := q.GetInt("qtyStart")
	qtyEnd, _ := q.GetInt("qtyEnd")
	status, _ := q.GetInt("status")
	purchaseType, _ := q.GetInt("purchaseType")
	userInfo := q.GetString("userInfo")

	// 会员手机及邮箱
	userInfoChan := make(chan utils.UserInfoResp, 1)
	if userInfo != constant.NULL {
		go utils.FindUserInfoAsyc(userInfo, userInfoChan)
	}

	o := orm.NewOrm()
	now := time.Now().Unix()
	var purchaseList = make([]models.OperationListResp, constant.ZERO)
	var data []models.SpSellerQuotedBuyer
	var allStatus = []int{constant.OnPurchase, constant.Traded, constant.Paused, constant.Canceled}
	qs := o.QueryTable(&models.SpSellerQuotedBuyer{}).Filter("BuyerId__gt", constant.ZERO).Filter("Status__in", allStatus).OrderBy("-CreateTime")

	// 按发布日期筛选
	if createStart > constant.ZERO {
		if createEnd > constant.ZERO && createEnd <= createStart {
			q.ResponseData(400, "请确认查询起始时间", nil)
			return
		}
		qs = qs.Filter("CreateTime__gte", createStart)
	}
	if createEnd > constant.ZERO {
		qs = qs.Filter("CreateTime__lte", createEnd)
	}

	// 按交货日期
	if deliveryStart > constant.ZERO {
		if deliveryEnd > constant.ZERO && deliveryEnd <= deliveryStart {
			q.ResponseData(400, "请确认查询起始时间", nil)
			return
		}
		qs = qs.Filter("DeliveryEndTime__gte", deliveryStart)
	}
	if deliveryEnd > constant.ZERO {
		qs = qs.Filter("DeliveryEndTime__lte", deliveryEnd)
	}

	// 按报价截止日期
	if quoteStart > constant.ZERO {
		if quoteEnd > constant.ZERO && quoteEnd <= quoteStart {
			q.ResponseData(400, "请确认查询起始时间", nil)
			return
		}
		qs = qs.Filter("QuotedEndTime__gte", quoteStart)
	}
	if quoteEnd > constant.ZERO {
		qs = qs.Filter("QuotedEndTime__lte", quoteEnd)
	}

	// 按成交日期
	if fixtureStart > constant.ZERO {
		if fixtureEnd > constant.ZERO && fixtureEnd <= fixtureStart {
			q.ResponseData(400, "请确认查询起始时间", nil)
			return
		}
		qs = qs.Filter("FixtureDate__gte", createStart)
	}
	if fixtureEnd > constant.ZERO {
		qs = qs.Filter("FixtureDate__lte", createStart)
	}

	// 按采购数量
	if qtyStart > constant.ZERO {
		if qtyEnd > constant.ZERO && qtyEnd <= qtyStart {
			q.ResponseData(400, "请确认查询采购数量区间", nil)
			return
		}
		qs = qs.Filter("GoodsQty__gte", qtyStart)
	}
	if qtyEnd > constant.ZERO {
		qs = qs.Filter("GoodsQty__lte", qtyEnd)
	}

	// 产品类型
	if result, _ := utils.Contain(purchaseType, []int{constant.Single, constant.Batch, constant.CompanyPrint}); result {
		qs = qs.Filter("PurchaseType", purchaseType)
	}

	// 状态
	if result, _ := utils.Contain(status, []int{constant.OnPurchase, constant.Traded, constant.Paused, constant.Canceled}); result {
		if status == constant.OnPurchase || constant.Paused == status {
			cond := orm.NewCondition()
			available := cond.AndCond(cond.And("Status", status)).
				AndCond(cond.AndCond(cond.And("QuotedEndTime__gt", now-constant.PurchaseFreezeTime).And("QuoteCount__gt", 0)).
					OrCond(cond.And("QuotedEndTime__gt", now-constant.PurchaseOverTime).
						And("QuotedEndTime__lt", now-constant.PurchaseFreezeTime).
						And("SuccessSellerId__gt", constant.ZERO)).OrCond(cond.And("QuotedEndTIme__gt", now))).
				AndCond(cond.And("AuditStatus", constant.APPROVED))
			qs = qs.SetCond(available)
			//qs = qs.Filter("Status", status).Filter("QuotedEndTime__gt", now).Filter("AuditStatus", constant.APPROVED)
		} else if status == constant.Canceled {
			cond := orm.NewCondition()
			or := cond.AndCond(cond.AndCond(cond.AndCond(cond.AndCond(cond.And("QuotedEndTime__lt", now-constant.PurchaseOverTime).
				And("SuccessSellerId__gt", constant.ZERO)).OrCond(cond.And("QuotedEndTime__lt", now-constant.PurchaseFreezeTime).And("QuoteCount__gt", 0).And("SuccessSellerId__lte", constant.ZERO)).
				OrCond(cond.And("QuotedEndTime__lt", now).And("QuoteCount", 0))).
				AndCond(cond.And("Status__in", constant.OnPurchase, constant.Paused)).AndNotCond(cond.And("AuditStatus", constant.UnApprove)).OrCond(cond.And("Status", constant.Canceled))))
			qs = qs.SetCond(or)
		} else {
			qs = qs.Filter("Status", status)
		}
	}
	if result, _ := utils.Contain(status, []int{constant.Auditing, constant.AuditFailed}); result {
		if status == constant.Auditing {
			qs = qs.Filter("AuditStatus", constant.InReview).Filter("QuotedEndTime__gt", now)
		}
		if status == constant.AuditFailed {
			qs = qs.Filter("AuditStatus", constant.UnApprove)
		}
	}
	// 采购单号 支持模糊查询
	if purchaseSn != constant.NULL {
		qs = qs.Filter("PurchaseSn__icontains", purchaseSn)
	}

	// 会员手机及邮箱
	if userInfo != constant.NULL {
		user := <-userInfoChan
		userId := user.Data.UserID
		qs = qs.Filter("BuyerId", userId)
	}

	q.count, _ = qs.Count()
	q.GetPagination()

	if count, err := qs.Limit(q.pageSize, q.start).RelatedSel().All(&data); err != nil {
		logs.Error("Query purchase list error, reason:", err.Error())
		q.ResponseData(400, "查询采购单异常", nil)
		return
	} else {
		logs.Info("共查询采购单", q.count, "条", "本次返回", count, "条")
	}
	// 获取用户信息
	var userList = make([]int, constant.ZERO)
	for _, i := range data {
		if result, _ := utils.Contain(i.BuyerId, userList); !result {
			userList = append(userList, i.BuyerId)
		}
	}

	userInfoList := utils.GetUserInfo(userList)
	//userChan := make(chan map[int]interface{}, 1)
	//middleChan := make(chan map[int]interface{}, 1)
	//go utils.GetUserInfoAsyc(userList, userChan)

	for _, i := range data {
		// 修改状态
		if result, _ := utils.Contain(status, []int{constant.Auditing, constant.AuditFailed, constant.OnPurchase, constant.Traded, constant.Paused, constant.Canceled}); result {
			i.Status = status
		}
		result, _ := utils.Contain(i.Status, []int{constant.Traded, constant.Paused, constant.Canceled})
		// 报价中状态
		if !result && constant.APPROVED == i.AuditStatus &&
			(now < i.QuotedEndTime || (now-constant.PurchaseFreezeTime < i.QuotedEndTime && i.QuotedEndTime < now && constant.ZERO < i.QuoteCount) ||
				(now-constant.PurchaseOverTime < i.QuotedEndTime && i.QuotedEndTime < now-constant.PurchaseFreezeTime && constant.NULL != i.OrderSn)) {
			i.Status = constant.OnPurchase
		}
		// 审核中状态
		if constant.InReview == i.AuditStatus && constant.OnPurchase == i.Status && now < i.QuotedEndTime {
			i.Status = constant.Auditing
		}
		// 已失效状态
		if i.Status == constant.Canceled || (((constant.ZERO >= i.QuoteCount && now > i.QuotedEndTime) ||
			(now-constant.PurchaseFreezeTime >= i.QuotedEndTime && constant.ZERO >= i.SuccessSellerId) ||
			(now-constant.PurchaseOverTime >= i.QuotedEndTime)) && constant.Traded != i.Status) {
			i.Status = constant.Canceled
		}
		// 审核未通过状态
		if constant.UnApprove == i.AuditStatus {
			i.Status = constant.AuditFailed
		}

		// 组装返回数据
		var j models.OperationListResp
		if err := utils.TransferToStruct(i, &j); err != nil {
			logs.Error("Transfer response data error, reason:", err.Error())
			q.ResponseData(400, "获取数据异常", nil)
			return
		}
		//go func(data *models.OperationListResp, userChan, middleChan chan map[int]interface{}) {
		//	select {
		//	case user := <- userChan:
		//		middleChan <- user
		//		data.User = "user[data.BuyerId]"
		//		//data.User = user[data.BuyerId]
		//		logs.Info("userChan",user[data.BuyerId])
		//		logs.Info("userChan",data.BuyerId)
		//	case user := <- middleChan:
		//		middleChan <- user
		//		data.User = "user[data.BuyerId]"
		//		//data.User = user[data.BuyerId]
		//		logs.Info("middleChan",user[data.BuyerId])
		//		logs.Info("middleChan",data.BuyerId)
		//	}
		//	data.User = "user[data.BuyerId]"
		//}(&j, userChan, middleChan)
		j.User = userInfoList[j.BuyerId]
		j.QuotedEndTimes = models.TimeStampToLocalString(i.QuotedEndTime)
		j.DeliveryEndTimes = models.TimeStampToLocalString(i.DeliveryEndTime)
		j.CreateTimes = models.TimeStampToLocalString(i.CreateTime)
		if i.FixtureDate > 0 {
			j.FixtureDates = models.TimeStampToLocalString(int64(i.FixtureDate))
		}
		if constant.Single == i.PurchaseType {
			j.CatName = i.TmpCatId.CatName
		}
		purchaseList = append(purchaseList, j)
	}
	q.ResponseData(200, "Operation Success!", purchaseList)
}

// @Title update order info to purchase from order
func (q *Quoted) PurchaseOrder() {
	var orderInfos []models.UpdateOrderReq
	var quoted models.SpSellerQuotedBuyer
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &orderInfos)
	if orderInfos == nil || constant.ZERO >= len(orderInfos) {
		q.ResponseData(400, "缺少参数", nil)
		return
	}
	for _, orderInfo := range orderInfos {
		if _, err := q.Validate(orderInfo); err != nil {
			logs.Error("update purchase order failed! reason: ", err)
			q.ResponseData(400, err, nil)
			return
		}
		if !utils.ContainsInt([]interface{}{constant.CancelOrder, constant.CreateOrder}, orderInfo.Status) {
			q.ResponseData(400, "非法操作", nil)
			return
		}
		if constant.CreateOrder == orderInfo.Status && (constant.NULL == orderInfo.OrderSn || orderInfo.SuccessSellerId <= 0) {
			q.ResponseData(400, "订单号或商家id异常", nil)
			return
		}
		o := orm.NewOrm()
		if err := o.QueryTable(&models.SpSellerQuotedBuyer{}).Filter("Id", orderInfo.PurchaseId).Filter("BuyerId", orderInfo.BuyerId).One(&quoted); err != nil {
			logs.Error("query purchase info failed! reason: ", err.Error())
			q.ResponseData(400, "查询采购单失败", nil)
			return
		}
		if constant.OutUse == quoted.Status {
			q.ResponseData(400, "采购单不存在，请确认后重试", nil)
			return
		}

		if constant.CancelOrder == orderInfo.Status {
			orderInfo.SuccessSellerId = 0
			orderInfo.Price = 0
			orderInfo.OrderSn = ""
		}
		quoted.SuccessSellerId = orderInfo.SuccessSellerId
		quoted.Price = orderInfo.Price
		quoted.OrderSn = orderInfo.OrderSn
		if _, err := o.Update(&quoted); err != nil {
			logs.Error("update purchase info failed! reason: ", err.Error())
			q.ResponseData(400, "更新采购单订单信息失败", nil)
			return
		}
	}
	q.ResponseData(200, "Operation Success!", nil)
}

// @Title backstage edit save
func (q *Quoted) BackstageEdit() {
	var reqBody models.UpdateQuoteReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	now := time.Now().Unix()

	// 入参校验
	if err := reqBody.PurchasePreSave(now); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}

	// 获取采购单详情
	var quote models.SpSellerQuotedBuyer
	if err := quote.GetQuoteById(reqBody.Id); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	// company purchase is not allowed backstage edit
	if constant.CompanyPrint == quote.PurchaseType {
		q.ResponseData(400, "企业印采购单不允许代编辑", nil)
		return
	}
	if !(constant.OnPurchase == quote.Status && constant.InReview == quote.AuditStatus && quote.QuotedEndTime >= now) {
		logs.Error("Update quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: Unable Backstage Edit Purchase!")
		q.ResponseData(400, "该采购单不允许代编辑", nil)
		return
	}
	// update purchase info
	quote.BackstageEdit = constant.Yes
	quote.BackstageEditTime = now
	quote.AuditStatus = constant.InReview
	quoteId, err := quote.UpdateQuote(reqBody, now, constant.Yes)
	if err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	q.ResponseData(200, "ok", map[string]int{"PurchaseId": quoteId})
}

// @Title backstage edit submit
func (q *Quoted) BackstageSubmit() {
	var reqBody models.UpdateQuoteReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	now := time.Now().Unix()

	// 入参校验
	if err := reqBody.PurchasePreSave(now); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}

	// 获取采购单详情
	var quote models.SpSellerQuotedBuyer
	if err := quote.GetQuoteById(reqBody.Id); err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	// company purchase is not allowed backstage edit
	if constant.CompanyPrint == quote.PurchaseType {
		q.ResponseData(400, "企业印采购单不允许代编辑", nil)
		return
	}
	if !(constant.OnPurchase == quote.Status && constant.InReview == quote.AuditStatus && quote.QuotedEndTime >= now) {
		logs.Error("Update quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: Unable Backstage Edit Purchase!")
		q.ResponseData(400, "该采购单不允许代编辑", nil)
		return
	}
	// 更新数据
	quote.AuditStatus = constant.UnApprove
	quote.BackstageEdit = constant.Yes
	quote.BackstageEditTime = now
	// TODO new status field
	quote.QuoteStatus = constant.BackstageEditStatus
	quoteId, err := quote.UpdateQuote(reqBody, now, constant.Yes)
	if err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	q.ResponseData(200, "ok", map[string]int{"PurchaseId": quoteId})
}

// @Title get the count of on auditing purchase
func (q *Quoted) OnAuditListNum() {
	defer func() {
		if rec := recover(); rec != nil {
			logs.Error("")
			q.ResponseData(400, "服务异常", nil)
		}
	}()
	// 按条件查询列表数量
	o := orm.NewOrm()
	qs := o.QueryTable(&models.SpSellerQuotedBuyer{})
	now := time.Now().Unix()
	count, err := qs.Filter("AuditStatus", constant.InReview).Filter("Status", constant.OnPurchase).Filter("QuotedEndTime__gt", now).Count()
	if err != nil {
		logs.Error("get inreview purchase failed! reason: ", err.Error())
		q.ResponseData(400, "获取采购单信息失败", nil)
		return
	}
	// 返回数据
	q.ResponseData(200, "ok", map[string]int64{"AuditingCount": count})
}

// @Title change document file from order
func (q *Quoted) UploadFileFromOrder() {
	defer func() {
		if rec := recover(); rec != nil {
			logs.Error("upload file handling error, reason:", rec)
			q.ResponseData(400, "服务异常", nil)
		}
	}()
	// 1 获取入参
	var uploadFile = make([]map[string]interface{}, 0)
	var reqBody models.UploadFileReq
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	uploadFile = append(uploadFile, reqBody.Documents)
	if reqBody.Documents == nil || constant.ONE != len(uploadFile) {
		q.ResponseData(400, "请填写正确的上传文件", nil)
		return
	}

	//表单校验
	if result, err := q.Validate(&reqBody); !result {
		q.ResponseData(400, err, nil)
		return
	}

	// 查询采购单
	var quote models.SpSellerQuotedBuyer
	if err := quote.GetQuoteById(reqBody.QuoteId); err != nil {
		logs.Error("Get quote info failed! Reason: ", err)
		q.ResponseData(400, "获取采购单信息失败", nil)
		return
	}
	// 判断采购单是否已下单
	if constant.ZERO >= quote.SuccessSellerId || quote.BuyerId != reqBody.UserId {
		q.ResponseData(400, "采购单状态异常，请确认", nil)
		return
	}
	quote.Documents = utils.ToString(uploadFile)

	o := orm.NewOrm()
	if _, err := o.Update(&quote, "Documents"); err != nil {
		logs.Error("update upload file failed! reason: ", err.Error())
		q.ResponseData(400, "更新印刷文件失败", nil)
		return
	}
	q.ResponseData(200, "ok", nil)
}

// @Title change purchase owner
func (q *Quoted) ChangePurchaseOwner() {
	defer func() {
		if rec := recover(); rec != nil {
			logs.Error("upload file handling error, reason:", rec)
			q.ResponseData(400, "服务异常", nil)
		}
	}()
	now := time.Now().Unix()
	var reqBody models.ChangeOwnerReq
	var quote models.SpSellerQuotedBuyer
	var exchangePurchase models.SpExchangePurchaseLog
	_ = json.Unmarshal(q.Ctx.Input.RequestBody, &reqBody)
	//表单校验
	if result, err := q.Validate(&reqBody); !result {
		q.ResponseData(400, err, nil)
		return
	}
	if reqBody.QuoteId <= constant.ZERO || reqBody.ReceiveUserId <= constant.ZERO {
		q.ResponseData(400, "请求参数异常，请确认", nil)
		return
	}
	if err := quote.GetQuoteById(reqBody.QuoteId); err != nil {
		logs.Error("get purchase info failed! reason:", err.Error(), ", quote_id: ", reqBody.QuoteId)
		q.ResponseData(400, "获取采购单信息失败，请重试", nil)
		return
	}
	if !((now <= quote.QuotedEndTime || (now-constant.PurchaseFreezeTime <= quote.QuotedEndTime && constant.ZERO < quote.QuoteCount) ||
		(now-constant.PurchaseOverTime <= quote.QuotedEndTime && constant.ZERO < quote.SuccessSellerId)) &&
		constant.OnPurchase == quote.Status && constant.APPROVED == quote.AuditStatus) {
		q.ResponseData(400, "转单失败，请确认采购单状态后重试", nil)
		return
	}
	if constant.ZERO < quote.SuccessSellerId {
		q.ResponseData(400, "该采购单存在有效订单，请联系用户取消该订单", nil)
		return
	}
	if reqBody.ReceiveUserId == quote.BuyerId {
		q.ResponseData(400, "请填写其它账户!", nil)
		return
	}
	o := orm.NewOrm()
	if errs := o.Begin(); errs != nil {
		q.ResponseData(400, "数据操作异常，请重试", nil)
		return
	}
	exchangePurchase.QuoteResp = quote.QuoteResp
	exchangePurchase.OriginalUserId = quote.BuyerId
	exchangePurchase.CreateAndUpdateTime = quote.CreateAndUpdateTime
	exchangePurchase.TmpCatId = quote.TmpCatId
	exchangePurchase.QuoteId = quote.Id
	exchangePurchase.BuyerId = reqBody.ReceiveUserId
	exchangePurchase.OperatorId = reqBody.OperatorId
	exchangePurchase.ExchangeTime = now
	if _, err := o.Insert(&exchangePurchase); err != nil {
		logs.Error("save change purchase log failed! reason: ", err.Error())
		if errs := o.Rollback(); errs != nil {
			logs.Error("save change purchase log rollback failed! reason: ", errs.Error())
		}
		q.ResponseData(400, "采购单转单失败，请重试", nil)
		return
	}

	quote.BuyerId = reqBody.ReceiveUserId
	if _, err := o.Update(&quote, "BuyerId"); err != nil {
		logs.Error("change purchase owner failed! reason: ", err.Error())
		if errs := o.Rollback(); errs != nil {
			logs.Error("change purchase owner rollback failed! reason: ", errs.Error())
		}
		q.ResponseData(400, "采购单转单失败，请重试", nil)
		return
	}
	if errs := o.Commit(); errs != nil {
		logs.Error("change purchase owner commit failed! reason: ", errs.Error())
		q.ResponseData(400, "采购单转单失败，请重试", nil)
		return
	}
	q.ResponseData(200, "转单成功", map[string]interface{}{
		"QuoteId": quote.Id,
		"BuyerId": quote.BuyerId,
	})
}

// @Title change owner log
func (q *Quoted) ChangeList() {
	// get params
	reqData := q.Ctx.Request.Form
	logs.Info("reqData: ", reqData)
	changeStart, _ := q.GetInt("changeStart")
	changeEnd, _ := q.GetInt("changeEnd")
	purchaseSn := q.GetString("purchaseSn")
	deliveryStart, _ := q.GetInt("deliveryStart")
	deliveryEnd, _ := q.GetInt("deliveryEnd")
	userInfo := q.GetString("userInfo")
	originalUserInfo := q.GetString("originalUserInfo")

	// 会员手机及邮箱
	var userInfoChan chan utils.UserInfoResp
	if userInfo != constant.NULL {
		go utils.FindUserInfoAsyc(userInfo, userInfoChan)
	}

	o := orm.NewOrm()
	var exchangeList = make([]models.ExchangeListResp, constant.ZERO)
	var data []models.SpExchangePurchaseLog
	qs := o.QueryTable(&models.SpExchangePurchaseLog{}).OrderBy("-ExchangeTime")

	// 按交货日期
	if deliveryStart > constant.ZERO {
		if deliveryEnd > constant.ZERO && deliveryEnd <= deliveryStart {
			q.ResponseData(400, "请确认查询起始时间", nil)
			return
		}
		qs = qs.Filter("DeliveryEndTime__gte", deliveryStart)
	}
	if deliveryEnd > constant.ZERO {
		qs = qs.Filter("DeliveryEndTime__lte", deliveryEnd)
	}

	// 按报价截止日期
	if changeStart > constant.ZERO {
		if changeEnd > constant.ZERO && changeEnd <= changeStart {
			q.ResponseData(400, "请确认查询起始时间", nil)
			return
		}
		qs = qs.Filter("ExchangeTime__gte", changeStart)
	}
	if changeEnd > constant.ZERO {
		qs = qs.Filter("ExchangeTime__lte", changeEnd)
	}

	// 采购单号 支持模糊查询
	if purchaseSn != constant.NULL {
		qs = qs.Filter("PurchaseSn__icontains", purchaseSn)
	}

	// 会员手机及邮箱
	if userInfo != constant.NULL {
		//user := utils.FindUserInfo(userInfo)
		user := <-userInfoChan
		userId := user.Data.UserID
		qs = qs.Filter("BuyerId", userId)
	}

	// 会员手机及邮箱
	if originalUserInfo != constant.NULL {
		user := utils.FindUserInfo(originalUserInfo)
		userId := user.Data.UserID
		qs = qs.Filter("OriginalUserId", userId)
	}
	q.count, _ = qs.Count()
	q.GetPagination()

	if count, err := qs.Limit(q.pageSize, q.start).RelatedSel().All(&data); err != nil {
		logs.Error("Query purchase list error, reason:", err.Error())
		q.ResponseData(400, "查询采购单异常", nil)
		return
	} else {
		logs.Info("共查询采购单", q.count, "条", "本次返回", count, "条")
	}

	// 获取用户信息
	var userList = make([]int, constant.ZERO)
	for _, i := range data {
		if result, _ := utils.Contain(i.BuyerId, userList); !result {
			userList = append(userList, i.BuyerId)
			userList = append(userList, i.OriginalUserId)
		}
	}

	//userInfoList := utils.GetUserInfo(userList)
	var userChan = make(chan map[int]interface{}, 1)
	var userInfoList = make(map[int]interface{})
	go utils.GetUserInfoAsyc(userList, userChan)

	for _, i := range data {
		var j models.ExchangeListResp
		if err := utils.TransferToStruct(i, &j); err != nil {
			logs.Error("Transfer exchange list response data error, reason:", err.Error())
			q.ResponseData(400, "获取数据异常", nil)
			return
		}
		j.QuoteId = i.QuoteId
		//j.User = userInfoList[i.BuyerId]
		//j.OriginalUser = userInfoList[i.OriginalUserId]
		j.DeliveryEndTimes = models.TimeStampToLocalString(i.DeliveryEndTime)
		j.ExchangeTimes = models.TimeStampToLocalString(i.ExchangeTime)
		j.QuotedEndTimes = models.TimeStampToLocalString(i.QuotedEndTime)
		if constant.Single == i.PurchaseType {
			j.CatName = i.TmpCatId.CatName
		}
		exchangeList = append(exchangeList, j)
	}

	userInfoList = <- userChan
	for index, i := range data {
		exchangeList[index].User = userInfoList[i.BuyerId]
		exchangeList[index].OriginalUser = userInfoList[i.OriginalUserId]
	}

	q.ResponseData(200, "Operation Success!", exchangeList)
}

// @Title change purchase detail
func (q *Quoted) ChangePurchaseDetail() {
	var exchangeLogId int
	var exchange models.SpExchangePurchaseLog
	_ = q.Ctx.Input.Bind(&exchangeLogId, "log_id")
	exchange.Id = exchangeLogId

	data, err := exchange.GetDetail()
	if err != nil {
		q.ResponseData(400, err.Error(), nil)
		return
	}
	q.ResponseData(200, "Operation Sucess!", data)
}

// @Title compare the category info between request info and backstage info
func CompareAttrAndAttrs(Attrs []models.TmpPurchaseAttr, tmpAttrs []*models.TmpPurchaseAttr) (result bool) {
	logs.Info(Attrs)
	logs.Info(tmpAttrs)
	result = false
	if len(Attrs) != len(tmpAttrs) {
		return
	}
	count := 0
	for _, attr := range Attrs {
		res := false
		for _, tmpAttr := range tmpAttrs {
			if res {
				continue
			}
			if attr.Id == tmpAttr.Id {

				if (len(attr.AttrValues) != len(tmpAttr.AttrValues) && len(attr.AttrValues)+1 != len(tmpAttr.AttrValues)) ||
					attr.AttrName != tmpAttr.AttrName || attr.AttrAliasName != tmpAttr.AttrAliasName {
					continue
				}
				count += 1
				res = true
				count1 := 0
				for _, value := range attr.AttrValues {
					res1 := false
					for _, tmpValue := range tmpAttr.AttrValues {
						if res1 {
							continue
						}
						if value.ValueId == tmpValue.ValueId {
							if value.AttrValue != tmpValue.AttrValue || value.Sort != tmpValue.Sort || value.UnitTypeId != tmpValue.UnitTypeId {
								continue
							}
							count1 += 1
							res1 = true
						}
					}
				}
				if count1 != len(attr.AttrValues) {
					return
				}
			}
		}
	}
	if count != len(Attrs) {
		return
	}
	result = true
	return
}
