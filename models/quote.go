package models

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"math/rand"
	"reflect"
	"server-purchase/utils"
	"server-purchase/utils/constant"
	"strconv"
	"sync"
	"time"
)

type SpSellerQuotedBuyer struct {
	QuoteResp
	Id          int                  `orm:"column(quoted_id);auto"`
	TmpCatId    *TmpPurchaseCategory `orm:"column(tmp_cat_id);null;rel(fk)" json:"-" description:"tmp表分类模板id"`
	AuditStatus int                  `orm:"column(audit_status);null" description:"审核状态0平台审核中，审核通过，审核不通过"`
	Status      int                  `orm:"column(status);null" description:"报价状态0报价中，1已成交，2已暂停，3已取消"`
	BuyerId     int                  `orm:"column(buyer_id)" valid:"Required" description:"买家id"`
	QuoteCount  int                  `orm:"column(quote_count)" description:"报价数量统计"`
	CreateAndUpdateTime
	AuditTime         int    `orm:"column(audit_time);null" description:"采购单审核时间"`
	VirtualOrderUser  string `orm:"column(virtual_order_user)" description:"虚拟订单添加人id"`
	VirtualShop       string `orm:"column(virtual_shop)" description:"虚拟店铺名"`
	BackstageEdit     int    `orm:"column(backstage_edit)" description:"代编辑状态"`
	BackstageEditTime int64  `orm:"column(backstage_edit_time)" description:"代编辑时间"`
	QuoteStatus       int    `orm:"column(quote_status);null" description:"采购单状态 0 待审核，1 审核未通过，2 报价中，3 已暂停，4 已取消，5 已成交"`
	PurchaseOrder
	SpSellerQuotedSet []*SpSellerQuoted `json:"-" orm:"reverse(many)"`
}

type CreateAndUpdateTime struct {
	CreateTime int64 `orm:"column(create_time);null" description:"采购单创建时间"`
	UpdateTime int64 `orm:"column(update_time);null" description:"采购单更新时间"`
}

func (t *SpSellerQuotedBuyer) TableName() string {
	return "sp_seller_quoted_buyer"
}

type SpExchangePurchaseLog struct {
	QuoteResp
	Id          int                  `orm:"column(log_id);auto"`
	TmpCatId    *TmpPurchaseCategory `orm:"column(tmp_cat_id);null;rel(fk)" json:"-" description:"tmp表分类模板id"`
	BuyerId     int                  `orm:"column(buyer_id)" valid:"Required" description:"原采购单买家id"`
	QuoteId     int                  `orm:"column(quote_id)" valid:"Required" description:"采购单id"`
	OperatorId     int               `orm:"column(operator_id)" valid:"Required" description:"转单操作人id"`
	CreateAndUpdateTime
	ExchangeTime int64				 `orm:"column(exchange_time);null" description:"转单时间"`
	OriginalUserId int				 `orm:"column(original_buyer_id)" valid:"Required" description:"接受转单买家id"`
}


func (t *SpExchangePurchaseLog) TableName() string {
	return "sp_exchange_purchase_log"
}

type QuoteResp struct {
	PurchaseSn       string  `orm:"column(purchase_sn);size(128)" description:"采购编号"`
	City             string  `orm:"column(city);size(255);null" description:"采购城市"`
	GoodsName        string  `orm:"column(goods_name);size(255);null" description:"采购商品"`
	GoodsQty         int     `orm:"column(goods_qty)" valid:"Required" description:"采购总数量"`
	QuotedEndTime    int64   `orm:"column(quoted_end_time);null" valid:"Required" description:"报价截止"`
	DeliveryEndTime  int64   `orm:"column(delivery_end_time);null" valid:"Required" description:"交货日期"`
	DetailsDesc      string  `orm:"column(details_desc);size(255);null" description:"详情描述"`
	Code             string  `orm:"column(code);size(255);null" description:"物料编码"`
	MaterielName     string  `orm:"column(materiel_name);size(255);null" description:"物料名称"`
	NormsRequirement string  `orm:"column(norms_requirement);size(255);null" description:"规格要求(产品类别)"`
	GoodsPicture     string  `orm:"column(goods_picture)" description:"商品图片"`
	Documents        string  `orm:"column(documents)" valid:"Required" description:"印刷文件"`
	SpecialNote      string  `orm:"column(special_note);size(255);null" description:"特别说明"`
	Price            float64 `orm:"column(price);null;digits(10);decimals(3)" description:"报价"`
	Remark           string  `orm:"column(remark);size(512);null" description:"备注"`
	MobileNumber     string  `orm:"column(mobile_number);size(20);null" valid:"Required" json:"-" description:"手机号"`
	Include          string  `orm:"column(include);size(128);null" description:"报价包含0含税，1含运费，2含打样"`
	AddressId        int     `orm:"column(address_id);null" valid:"Required" description:"地址id"`
	PurchaseList     string  `orm:"column(purchase_list);size(128)" description:"采购清单"`
	PurchaseType     int     `orm:"column(purchase_type)" description:"采购单类型 1 单商品  2 多商品 3 企业购"`
	AuditMessage     string  `orm:"column(audit_message);null" description:"审核未通过原因"`
	Unit             string  `orm:"column(unit);null" description:"商品单位"`
	AddressInfo      string  `orm:"column(address_info);null" description:"地址信息"`
	Invoice          int     `orm:"column(invoice)" description:"是否需要发票0不需要，1普通，2增值税"`
	QQ               string  `orm:"column(qq)" description:"用户qq"`
	PurchaseListDesc string  `orm:"column(purchase_list_desc)" description:"企业印采购清单描述"`
	BackstageEdit
}

type BackstageEdit struct {
	ExpressDesc         string `orm:"column(express_desc)" description:"运费说明"`
	PackingRequirements string `orm:"column(packing_requirements)" description:"包装要求"`
	OtherSupports       string `orm:"column(other_supports)" description:"其它服务"`
}

type PurchaseOrder struct {
	SuccessSellerId int    `orm:"column(success_seller_id);null" description:"成交的商家报价id"`
	FixtureDate     int64  `orm:"column(fixture_date)" description:"成交日期"`
	OrderSn         string `orm:"column(order_sn);null" description:"订单号"`
}

// 出入参数据结构转换struct
type TransCommon struct {
	Include          []int                  `json:"Include"`
	GoodsPicture     []map[string]string    `json:"GoodsPicture"`
	Documents        []map[string]string    `json:"Documents"`
	PurchaseList     []map[string]string    `json:"PurchaseList"`
	NormsRequirement []int                  `json:"NormsRequirement"`
	MobileNumber     string                 `json:"MobileNumber" valid:"Required"`
	AddressInfo      map[string]interface{} `json:"AddressInfo"`
}

type TmpPurchaseAttr struct {
	Id            int                     `orm:"column(attr_id);auto"`
	AttrName      string                  `orm:"column(attr_name);size(255)" description:"属性名称"`
	AttrAliasName string                  `orm:"column(attr_alias_name);size(255)" description:"属性别名"`
	IsLimitUser   int8                    `orm:"column(is_limit_user)" description:"是否允许用户自定义0不允许1允许"`
	IsMust        int                     `orm:"column(is_must)" description:"是否必选字段 0 否 1 是"`
	Cat           *TmpPurchaseCategory    `orm:"rel(fk)" json:"-"`
	AttrValues    []*TmpPurchaseAttrValue `orm:"reverse(many)" json:"AttrValues"`
}

func (t *TmpPurchaseAttr) TableName() string {
	return "tmp_purchase_attr"
}

type TmpPurchaseAttrValue struct {
	Id         int                  `orm:"column(attr_value_id);auto"`
	AttrValue  string               `orm:"column(attr_value);size(255)" description:"属性值"`
	UnitTypeId int                  `orm:"column(unit_type_id)" description:"单位:0mm,1米,2吋"`
	Sort       int                  `orm:"column(sort)" description:"排序"`
	Customize  int                  `orm:"column(customize)" description:"是否自定义属性值"`
	PurchaseSn string               `orm:"column(purchase_sn)" description:"关联采购单"`
	ValueId    int                  `orm:"column(sp_attr_value_id)" json:"ValueId" description:"运营后台属性id"`
	Attr       *TmpPurchaseAttr     `orm:"rel(fk)" json:"omitempty"`
	Cat        *TmpPurchaseCategory `orm:"rel(fk)" json:"omitempty"`
}

func (t *TmpPurchaseAttrValue) TableName() string {
	return "tmp_purchase_attr_value"
}

type TmpPurchaseCategory struct {
	CatId   int                `orm:"column(cat_id);auto"`
	CatName string             `orm:"column(cat_name);size(255)" valid:"Required" description:"分类名称"`
	SpCatId int                `orm:"column(sp_cat_id)" valid:"Required" description:"运营后台分类id"`
	Attr    []*TmpPurchaseAttr `orm:"reverse(many)" json:"Attr"`
}

func (t *TmpPurchaseCategory) TableName() string {
	return "tmp_purchase_category"
}

type SpPurchaseEditLog struct {
	Id         int    `orm:"column(id);auto"`
	QuoteId    int    `orm:"column(quote_id)" valid:"Required" description:"采购单id"`
	LogDetail  string `orm:"column(log_detail)" valid:"Required" description:"代编辑记录"`
	CreateTime int64  `orm:"column(create_time);null" description:"创建时间"`
	UpdateTime int64  `orm:"column(update_time);null" description:"更新时间"`
	IsDeleted  int    `orm:"column(is_deleted)" description:"是否删除 0 未删除, 1 已删除"`
}

func (t *SpPurchaseEditLog) TableName() string {
	return "sp_purchase_edit_log"
}

func GetTmpCat(id int) (t *TmpPurchaseCategory, err error) {
	var o = orm.NewOrm()
	t = &TmpPurchaseCategory{CatId: id}
	err = o.Read(t)
	return
}

func (q *SpSellerQuotedBuyer) GetQuoteById(id int) (err error) {
	var o = orm.NewOrm()
	q.Id = id
	if err = o.Read(q); err != nil {
		return
	}
	if constant.OutUse == q.Status {
		err = errors.New("采购单不存在，请确认后重试")
		return
	}
	return
}

func (s *SpSellerQuotedBuyer) GetQuoteByIdAndBuyerId(id, buyerId int) (err error) {
	var o = orm.NewOrm()
	err = o.QueryTable(&SpSellerQuotedBuyer{}).Filter("BuyerId", buyerId).Filter("Id", id).One(s)
	return
}

func (quote *SpSellerQuotedBuyer) UpdateQuote(reqBody UpdateQuoteReq, now int64, backstage int) (purchaseId int, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			logs.Error("update purchase failed! reason: ", rec)
			err = errors.New("采购单编辑失败，请重试")
		}
	}()
	o := orm.NewOrm()
	var oldQuote = new(SpSellerQuotedBuyer)
	if errs := deepCopy(oldQuote, quote); errs != nil {
		logs.Error("copy purchase info failed!", errs.Error())
		err = errors.New("采购单编辑失败，请重试")
	}

	quoteName := quote.GoodsName
	if reqBody.Purchase.GoodsQty != quote.GoodsQty {
		var purchaseName = "求购"
		if reqBody.Purchase.PurchaseType == constant.Single {
			purchaseName += reqBody.Template.CatName
		} else {
			purchaseName += "多种商品共"
		}
		unit := reqBody.Purchase.Unit
		if unit == "" {
			unit = "件"
		}
		purchaseName += strconv.Itoa(reqBody.Purchase.GoodsQty) + unit
		quoteName = purchaseName
	}
	quoteSn := quote.PurchaseSn
	quote.QuoteResp = reqBody.Purchase.QuoteResp
	quote.GoodsName = quoteName
	quote.PurchaseSn = quoteSn
	quote.MobileNumber = reqBody.Purchase.TransCommon.MobileNumber
	quoteResp := &quote.QuoteResp
	quoteResp1 := reqBody.Purchase.TransCommon

	//alter slice to string for saving
	quoteResp.TransWithSave(quoteResp1)

	// initial status and quote count info
	quote.UpdateTime = now
	quote.AuditMessage = constant.NULL
	quote.QuoteCount = constant.ZERO

	// TODO update individual attribute value
	if errs := o.Begin(); errs != nil {
		logs.Error("start transaction failed! reason: ", errs.Error())
		err = errors.New("采购单信息更新失败")
		return
	}
	attrs := reqBody.Template.Attr
	for _, attr := range attrs {
		customValueNum := constant.ZERO
		if int(attr.IsLimitUser) == constant.ONE {
			for _, attrValue := range attr.AttrValues {
				if attrValue.Customize == constant.ONE {
					logs.Info("attr, ", attr.Id, "customize attribute value: ", attrValue.AttrValue)
					// judge customize attr value if is exists, if not update value, otherwise save it
					if attrValue.Id > constant.ZERO {
						customValue := TmpPurchaseAttrValue{Id: attrValue.Id}
						_ = o.Read(&customValue)
						if customValue.Customize != attrValue.Customize {
							logs.Error("customize attr value did not matched! PurchaseSn: ", quote.PurchaseSn, "AttrValueId: ", customValue.Id)
							continue
						}
						if customValue.AttrValue == attrValue.AttrValue {
							customValueNum += constant.ONE
							continue
						}
						customValue.AttrValue = attrValue.AttrValue
						if _, errs := o.Update(&customValue); errs != nil {
							logs.Error("Update customize attr value failed! QuotedId: ", quote.Id, "AttrValueId: ", attrValue.AttrValue, ", Reason: ", errs.Error())
							_ = o.Rollback()
							err = errors.New("采购单信息更新失败")
							return
						}
					} else {
						attrValue.Id = constant.ZERO
						attrValue.PurchaseSn = quote.PurchaseSn
						attrValue.Attr = attr
						cat := TmpPurchaseCategory{CatId: reqBody.Template.CatId}
						if errs := o.Read(&cat); errs != nil {
							logs.Error("Get Cat Info Failed! QuotedId: ", quote.Id, "CatId: ", reqBody.Template.CatId, ", Reason: ", errs.Error())
							_ = o.Rollback()
							err = errors.New("采购单信息更新失败")
							return
						}
						attrValue.Cat = &cat
						if _, errs := o.Insert(attrValue); errs != nil {
							logs.Error("Update customize attr value failed! QuotedId: ", quote.Id, "AttrValueId: ", attrValue.AttrValue, ", Reason: ", errs.Error())
							_ = o.Rollback()
							err = errors.New("采购单信息更新失败")
							return
						}
						normalRequirement := reqBody.Purchase.TransCommon.NormsRequirement
						for i, v := range normalRequirement {
							if v < constant.ZERO && attrValue.ValueId == v {
								normalRequirement[i] = attrValue.Id
							}
						}
						quote.NormsRequirement = utils.ToString(normalRequirement)
					}
					customValueNum += constant.ONE
				}
				if customValueNum > constant.ONE {
					_ = o.Rollback()
					logs.Error("exist over one customer attribute value, ", attrValue.AttrValue)
					err = errors.New("采购单信息更新失败")
					return
				}
			}
		}
	}
	logs.Info("Save Customize Attribute Value Succeed! QuotedId: ", quote.Id)

	// TODO if the purchase type is companyPrint, set the audit status as APPROVED
	//if quote.PurchaseType == constant.CompanyPrint {
	//	quote.AuditStatus = constant.APPROVED
	//}
	if _, errs := o.Update(quote); errs != nil {
		logs.Error("Update quote info failed! QuotedId: ", quote.Id, ", Reason: ", errs.Error())
		_ = o.Rollback()
		err = errors.New("采购单信息更新失败")
		return
	}
	// TODO if the request is from backstage, then compare the difference between new and older purchase detail
	//if constant.Yes == backstage {
	//	var attrList orm.ParamsList
	//	var valueList = make([]string, 0)
	//	if _, errs := o.QueryTable(&TmpPurchaseAttrValue{}).Filter(
	//		"Id__in", utils.StringToSliceInt(oldQuote.NormsRequirement)).ValuesFlat(
	//			&attrList, "AttrValue"); errs != nil {
	//			logs.Error("get old purchase attribute value failed! reason: ", errs.Error())
	//		_ = o.Rollback()
	//		err = errors.New("采购单信息更新失败")
	//		return
	//	}
	//	var i = 0
	//	for _, value := range utils.StringToSliceInt(oldQuote.NormsRequirement) {
	//		if constant.ZERO == value {
	//			valueList = append(valueList, constant.NULL)
	//		} else {
	//			valueList = append(valueList, attrList[i].(string))
	//			i += constant.ONE
	//		}
	//	}
	//	if err = oldQuote.BackstageEditDetail(*quote, o, valueList); err != nil {
	//		if errs := o.Rollback(); errs != nil {
	//			logs.Error("save edit history failed! reason: ", errs.Error())
	//			err = errors.New("数据操作异常，请稍候重试")
	//			return
	//		}
	//		return
	//	}
	//}

	// TODO if the request is from user, delete edit history
	//if constant.No == backstage {
	//	var editHistory SpPurchaseEditLog
	//	if errs := o.QueryTable(&SpPurchaseEditLog{}).Filter("QuoteId", quote.Id).Filter("IsDeleted", constant.ZERO).One(&editHistory); errs == nil {
	//		editHistory.IsDeleted = constant.ONE
	//		editHistory.UpdateTime = now
	//		if _, errs = o.Update(&editHistory); errs != nil {
	//			logs.Error("delete edit history failed! reason:", errs.Error())
	//			_ = o.Rollback()
	//			err = errors.New("采购单信息更新失败")
	//			return
	//		}
	//	}
	//}

	_ = o.Commit()
	err = nil
	purchaseId = quote.Id
	return
}

func (reqBody *UpdateQuoteReq) UpdateQuote() (purchaseId int, purchaseSn string, err error) {
	o := orm.NewOrm()
	now := time.Now().Unix()
	var quote SpSellerQuotedBuyer
	var newQuote SpSellerQuotedBuyer
	if err = quote.GetQuoteByIdAndBuyerId(reqBody.Id, reqBody.BuyerId); err != nil {
		logs.Error("Get quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: ", err.Error())
		err = errors.New("获取采购单信息失败")
		return
	}

	if constant.OnPurchase == quote.Status && constant.InReview == quote.AuditStatus && quote.QuotedEndTime >= now {
		logs.Error("Update quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: UnAudit Purchase Order!")
		err = errors.New("暂未通过审核，不允许重新发布")
		return
	}

	if constant.CompanyPrint == quote.PurchaseType || quote.IsValid(now, quote.Status) || constant.Traded == quote.Status ||
		(now < quote.QuotedEndTime && constant.UnApprove == quote.AuditStatus) {
		logs.Error("Update quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: the purchase is active!")
		err = errors.New("该采购单不允许重新发布")
		return
	}


	var purchaseName = "求购"
	if reqBody.Purchase.PurchaseType == constant.Single {
		purchaseName += reqBody.Template.CatName
	} else {
		purchaseName += "多种商品共"
	}
	unit := reqBody.Purchase.Unit
	if unit == "" {
		unit = "件"
	}
	purchaseName += strconv.Itoa(reqBody.Purchase.GoodsQty) + unit
	quote.QuoteResp = reqBody.Purchase.QuoteResp
	quote.GoodsName = purchaseName
	rand.Seed(now)
	Now := time.Now()
	quote.PurchaseSn = fmt.Sprintf("QD%d%02d%02d%02d%02d%02d%02d", Now.Year(), Now.Month(), Now.Day(), Now.Hour(), Now.Minute(), Now.Second(), rand.Intn(99))
	quote.MobileNumber = reqBody.Purchase.TransCommon.MobileNumber
	quoteResp := &quote.QuoteResp
	quoteResp1 := reqBody.Purchase.TransCommon
	quote.Id = newQuote.Id

	// transform slice to string for saving
	quoteResp.TransWithSave(quoteResp1)

	// initial purchase status detail
	quote.UpdateTime = now
	quote.CreateTime = now
	quote.AuditStatus = constant.InReview
	quote.Status = constant.OnPurchase
	quote.AuditMessage = constant.NULL
	quote.QuoteCount = constant.ZERO
	quote.BackstageEdit = constant.No
	quote.BackstageEditTime = int64(constant.ZERO)
	// TODO new status fields
	quote.QuoteStatus = constant.AuditingStatus
	//quote.QuoteResp.BackstageEdit = *new(BackstageEdit)

	// save all customize attribute value
	_ = o.Begin()
	attrs := reqBody.Template.Attr
	for _, attr := range attrs {
		if attr.IsLimitUser == 1 {
			for _, attrValue := range attr.AttrValues {
				if attrValue.Customize == 1 {
					attrValue.ValueId = attrValue.Id
					attrValue.Id = constant.ZERO
					attrValue.PurchaseSn = quote.PurchaseSn
					attrValue.Attr = attr
					cat := TmpPurchaseCategory{CatId: reqBody.Template.CatId}
					if errs := o.Read(&cat); errs != nil {
						logs.Error("Get Cat Info Failed! QuotedId: ", quote.Id, "CatId: ", reqBody.Template.CatId, ", Reason: ", errs.Error())
						_ = o.Rollback()
						err = errors.New("采购单信息更新失败")
						return
					}
					attrValue.Cat = &cat
					if _, errs := o.Insert(attrValue); errs != nil {
						logs.Error("Update customize attr value failed! QuotedId: ", quote.Id, "AttrValueId: ", attrValue.AttrValue, ", Reason: ", errs.Error())
						_ = o.Rollback()
						err = errors.New("采购单信息更新失败")
						return
					}
					normalRequirement := reqBody.Purchase.TransCommon.NormsRequirement
					for i, v := range normalRequirement {
						if attrValue.ValueId == v {
							normalRequirement[i] = attrValue.Id
						}
					}
					quote.NormsRequirement = utils.ToString(normalRequirement)
				}
			}
		}
	}
	logs.Info("Save Customize Attribute Value Succeed! QuotedId: ", quote.Id)

	// determine the purchase type if is CompanyPrint, if is true, set the audit status as APPROVED
	//if quote.PurchaseType == constant.CompanyPrint {
	//	quote.AuditStatus = constant.APPROVED
	//}
	if _, errs := o.Insert(&quote); errs != nil {
		logs.Error("Update quote info failed! QuotedId: ", quote.Id, ", Reason: ", errs.Error())
		_ = o.Rollback()
		err = errors.New("采购单信息更新失败")
		return
	}
	_ = o.Commit()
	err = nil
	purchaseId = quote.Id
	purchaseSn = quote.PurchaseSn
	return
}

func (reqBody *UpdateQuoteReq) UpdatePurchase() (purchaseId int, err error) {
	o := orm.NewOrm()
	var quote SpSellerQuotedBuyer
	now := time.Now().Unix()
	if err = o.QueryTable(&SpSellerQuotedBuyer{}).Filter("BuyerId", reqBody.BuyerId).Filter("Id", reqBody.Id).One(&quote); err != nil {
		logs.Error("Get quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: ", err.Error())
		err = errors.New("获取采购单信息失败")
		return
	}

	if constant.Canceled != quote.Status && constant.InReview == quote.AuditStatus && quote.QuotedEndTime >= (now-constant.PurchaseFreezeTime) {
		logs.Error("Update quote info failed! quoteId:", quote.Id, ", User: ", quote.BuyerId, ",Reason: UnAudit Purchase Order!")
		err = errors.New("暂未通过审核，不允许编辑")
		return
	}

	if (constant.Paused == quote.Status || ((constant.Single == quote.PurchaseType || constant.Batch == quote.PurchaseType) ||
		(constant.CompanyPrint == quote.PurchaseType && constant.ZERO < quote.QuoteCount))  && constant.OnPurchase == quote.Status) && quote.AuditStatus == constant.APPROVED && (
		(quote.QuotedEndTime > now || (quote.QuotedEndTime > now - constant.PurchaseFreezeTime && quote.QuoteCount > constant.ZERO)) ||
			(now - constant.PurchaseOverTime < quote.QuotedEndTime && now - constant.PurchaseFreezeTime > quote.QuotedEndTime && constant.NULL != quote.OrderSn)) {
		logs.Error("Update quote info failed! quoteId:",quote.Id,", User: ",quote.BuyerId,",Reason: UnAudit Purchase Order!")
		err = errors.New("该采购单不允许编辑")
		return
	}

	quoteName := quote.GoodsName
	quoteSn := quote.PurchaseSn
	quote.QuoteResp = reqBody.Purchase.QuoteResp
	quote.GoodsName = quoteName
	quote.PurchaseSn = quoteSn
	quote.MobileNumber = reqBody.Purchase.TransCommon.MobileNumber
	quoteResp := &quote.QuoteResp
	quoteResp1 := reqBody.Purchase.TransCommon

	//处理切片
	quoteResp.TransWithSave(quoteResp1)

	quote.UpdateTime = now
	//quote.CreateTime = now
	quote.AuditStatus = constant.InReview
	quote.Status = constant.OnPurchase
	quote.AuditMessage = constant.NULL
	quote.QuoteCount = constant.ZERO

	// TODO 更新自定义属性值
	if errs := o.Begin(); errs != nil {
		logs.Error("start transaction failed! reason: ", errs.Error())
		err = errors.New("采购单信息更新失败")
		return
	}
	attrs := reqBody.Template.Attr
	for _, attr := range attrs {
		customValueNum := constant.ZERO
		if int(attr.IsLimitUser) == constant.ONE {
			for _, attrValue := range attr.AttrValues {
				if attrValue.Customize == constant.ONE {
					logs.Info("attr, ", attr.Id, "customize attribute value: ", attrValue.AttrValue)
					// 判断是否为新增自定义属性值，新增则保存，非新增则更新
					if attrValue.Id > constant.ZERO {
						customValue := TmpPurchaseAttrValue{Id: attrValue.Id}
						_ = o.Read(&customValue)
						if customValue.Customize != attrValue.Customize {
							logs.Error("customize attr value did not matched! PurchaseSn: ", quote.PurchaseSn, "AttrValueId: ", customValue.Id)
							continue
						}
						if customValue.AttrValue == attrValue.AttrValue {
							customValueNum += constant.ONE
							continue
						}
						customValue.AttrValue = attrValue.AttrValue
						if _, errs := o.Update(&customValue); errs != nil {
							logs.Error("Update customize attr value failed! QuotedId: ", quote.Id, "AttrValueId: ", attrValue.AttrValue, ", Reason: ", errs.Error())
							_ = o.Rollback()
							err = errors.New("采购单信息更新失败")
							return
						}
					} else {
						attrValue.Id = constant.ZERO
						attrValue.PurchaseSn = quote.PurchaseSn
						attrValue.Attr = attr
						cat := TmpPurchaseCategory{CatId: reqBody.Template.CatId}
						if errs := o.Read(&cat); errs != nil {
							logs.Error("Get Cat Info Failed! QuotedId: ", quote.Id, "CatId: ", reqBody.Template.CatId, ", Reason: ", errs.Error())
							_ = o.Rollback()
							err = errors.New("采购单信息更新失败")
							return
						}
						attrValue.Cat = &cat
						if _, errs := o.Insert(attrValue); errs != nil {
							logs.Error("Update customize attr value failed! QuotedId: ", quote.Id, "AttrValueId: ", attrValue.AttrValue, ", Reason: ", errs.Error())
							_ = o.Rollback()
							err = errors.New("采购单信息更新失败")
							return
						}
						normalRequirement := reqBody.Purchase.TransCommon.NormsRequirement
						for i, v := range normalRequirement {
							if v < constant.ZERO && attrValue.ValueId == v {
								normalRequirement[i] = attrValue.Id
							}
						}
						quote.NormsRequirement = utils.ToString(normalRequirement)
					}
					customValueNum += constant.ONE
				}
				if customValueNum > constant.ONE {
					_ = o.Rollback()
					logs.Error("exist over one customer attribute value, ", attrValue.AttrValue)
					err = errors.New("采购单信息更新失败")
					return
				}
			}
		}
	}
	logs.Info("Save Customize Attribute Value Succeed! QuotedId: ", quote.Id)

	// 企业印自动审核
	//if quote.PurchaseType == constant.CompanyPrint {
	//	quote.AuditStatus = constant.APPROVED
	//}
	if _, errs := o.Update(&quote); errs != nil {
		logs.Error("Update quote info failed! QuotedId: ", quote.Id, ", Reason: ", errs.Error())
		_ = o.Rollback()
		err = errors.New("采购单信息更新失败")
		return
	}
	_ = o.Commit()
	err = nil
	purchaseId = quote.Id
	return
}

func GetPurchaseDetail(quoteId, quoteType int) (data QuoteDetail, err error) {
	// get purchase detail
	var quote = new(SpSellerQuotedBuyer)
	var userChan = make(chan map[int]interface{}, 1)
	now := time.Now().Unix()
	if errs := quote.GetQuoteById(quoteId); errs != nil {
		logs.Error("Get quote info failed! Reason: ", err)
		err = errors.New("查询数据失败")
		return
	}

	go utils.GetUserInfoAsyc([]int{quote.BuyerId}, userChan)
	status := quote.Status
	if quote.BuyerId == constant.ZERO || (quote.AuditStatus != constant.APPROVED && quoteType == constant.SHOP) {
		logs.Error("PurchaseId:", quote.Id, "has not pass the audit!")
		err = errors.New("该商品暂未通过审核")
		return
	}

	// get category info
	if quote.PurchaseType == constant.Single {
		var resp = make([]string, 0)
		resp, err = GetCatInfo(quote)
		if err != nil {
			return
		}
		data.CatList = resp
	}

	data.QuoteResp = quote.QuoteResp
	data.QuoteCount = quote.QuoteCount
	data.PurchaseOrder = quote.PurchaseOrder
	if constant.InReview == quote.AuditStatus {
		quote.Status = constant.Auditing
		data.PurchaseStatus = constant.Auditing
	}
	if constant.UnApprove == quote.AuditStatus {
		quote.Status = constant.AuditFailed
		data.PurchaseStatus = constant.AuditFailed
	}

	data.PurchaseStatus = quote.Status
	result, _ := utils.Contain(status, []int{constant.Traded, constant.Paused, constant.Canceled})
	// on purchase status
	if !result && constant.APPROVED == quote.AuditStatus &&
		(now < quote.QuotedEndTime || (now-constant.PurchaseFreezeTime < quote.QuotedEndTime && quote.QuotedEndTime < now && constant.ZERO < quote.QuoteCount) ||
			(now-constant.PurchaseOverTime < quote.QuotedEndTime && quote.QuotedEndTime < now-constant.PurchaseFreezeTime && constant.ZERO < quote.SuccessSellerId)) {
		quote.Status = constant.OnPurchase
		data.PurchaseStatus = constant.OnPurchase
	}
	// on auditing status
	if constant.InReview == quote.AuditStatus && constant.OnPurchase == quote.Status {
		quote.Status = constant.Auditing
		data.PurchaseStatus = constant.Auditing
	}
	// 已失效状态
	if quote.Status == constant.Canceled || (((constant.ZERO >= quote.QuoteCount && now > quote.QuotedEndTime) ||
		(now-constant.PurchaseFreezeTime >= quote.QuotedEndTime && constant.ZERO >= quote.SuccessSellerId) ||
		(now-constant.PurchaseOverTime >= quote.QuotedEndTime)) && constant.Traded != quote.Status) {
		quote.Status = constant.Canceled
		data.PurchaseStatus = constant.Canceled
	}
	// 审核未通过状态
	if constant.UnApprove == quote.AuditStatus {
		quote.Status = constant.AuditFailed
		data.PurchaseStatus = constant.AuditFailed
	}
	if constant.OutUse == status {
		quote.Status = constant.OutUse
		data.PurchaseStatus = constant.OutUse
	}

	data.Status = quote.Status
	// 兼容H5状态 PurchaseStatus
	if constant.Canceled != status && (((constant.ZERO >= quote.QuoteCount && now > quote.QuotedEndTime) ||
		(now-constant.PurchaseFreezeTime >= quote.QuotedEndTime && constant.ZERO >= quote.SuccessSellerId) ||
		(now-constant.PurchaseOverTime >= quote.QuotedEndTime)) && constant.Traded != quote.Status) {
		data.PurchaseStatus = constant.Expired
	}
	// freeze status
	if quote.QuotedEndTime > (now-constant.PurchaseOverTime) && quote.QuotedEndTime < (now-constant.PurchaseFreezeTime) && quote.SuccessSellerId > constant.ZERO {
		data.PurchaseStatus = constant.Freeze
	}
	// Backstage edit status
	if constant.UnApprove == quote.AuditStatus && constant.Yes == quote.BackstageEdit && now <= quote.QuotedEndTime {
		data.Status = constant.BackstageEdit
		data.PurchaseStatus = constant.BackstageEdit
	}
	data.MobileNumber = quote.MobileNumber
	data.BuyerId = quote.BuyerId
	data.QuoteCommon.TransWithGet(quote.QuoteResp)
	data.QuotedEndTimes = TimeStampToLocalString(quote.QuotedEndTime)
	data.DeliveryEndTimes = TimeStampToLocalString(quote.DeliveryEndTime)
	data.CreateTimes = TimeStampToLocalString(quote.CreateTime)
	if quote.FixtureDate > 0 {
		data.FixtureDates = TimeStampToLocalString(quote.FixtureDate)
	}

	// get user info
	//userInfoList := utils.GetUserInfo([]int{quote.BuyerId})
	userInfoList := <-userChan
	data.User = userInfoList[quote.BuyerId]

	// get attribute and attribute value name choice list
	if quote.PurchaseType == constant.Single && data.TransCommon.NormsRequirement != nil {
		attrValue, errs := GetAttrValueName(quote.NormsRequirement)
		if errs != nil {
			err = errors.New("查询采购商品属性失败")
			return
		}
		data.AttrName = attrValue
	}
	return
}

// 根据属性id，获取属性值展示
func GetAttrValueName(normal interface{}) (attrValue []string, err error) {
	var attrName []orm.Params
	var ls = new([]int)
	switch normal.(type) {
	case string:
		_ = json.Unmarshal([]byte(normal.(string)), &ls)
	case []int:
		*ls = normal.([]int)
	}
	if ls != nil && len(*ls) > 0 {
		o := orm.NewOrm()
		qs := o.QueryTable(&TmpPurchaseAttrValue{})
		_, err = qs.Filter("Id__in", ls).Values(&attrName, "Attr__AttrName", "AttrValue")
		if len(attrName) < 1 {
			return
		}
		for _, attrList := range attrName {
			attrValue = append(attrValue, fmt.Sprintf("%s: %s", attrList["Attr__AttrName"], attrList["AttrValue"]))
		}
	}
	return
}

func GetCatInfo(quote *SpSellerQuotedBuyer) ([]string, error) {
	var resp = make([]string, 0)
	if quote.TmpCatId == nil {
		logs.Error("Get quote tmpCat error! ", quote)
		err := errors.New("采购单数据异常")
		return resp, err
	}
	cat, err := GetTmpCat(quote.TmpCatId.CatId)
	if err != nil {
		logs.Error("Get quote info failed! Reason: ", err)
		err = errors.New("查询数据失败")
		return resp, err
	}
	if cat.SpCatId != constant.OtherCat {
		d, errs, _ := GetOneCat(cat.SpCatId)
		if errs != nil {
			err = errors.New("获取模板分类错误")
			return resp, err
		}
		GetCatName(d, &resp)
	} else {
		resp = append(resp, cat.CatName)
	}
	return resp, err
}

// 列表字段处理
func (resp2 *TransCommon) TransWithGet(resp1 QuoteResp) {
	resp2.NormsRequirement = utils.StringToSliceInt(resp1.NormsRequirement)
	resp2.Include = utils.StringToSliceInt(resp1.Include)
	resp2.GoodsPicture = utils.StringToSliceMap(resp1.GoodsPicture)
	resp2.PurchaseList = utils.StringToSliceMap(resp1.PurchaseList)
	resp2.Documents = utils.StringToSliceMap(resp1.Documents)
	resp2.AddressInfo = utils.StringToMap(resp1.AddressInfo)
}

func (quoteResp *QuoteResp) TransWithSave(quoteResp1 TransCommon) {
	//处理切片
	quoteResp.GoodsPicture = utils.ToString(quoteResp1.GoodsPicture)
	quoteResp.Documents = utils.ToString(quoteResp1.Documents)
	quoteResp.PurchaseList = utils.ToString(quoteResp1.PurchaseList)
	quoteResp.NormsRequirement = utils.ToString(quoteResp1.NormsRequirement)
	quoteResp.Include = utils.ToString(quoteResp1.Include)
	quoteResp.AddressInfo = utils.ToString(quoteResp1.AddressInfo)
}

func TimeStampToLocalString(t int64) string {
	timeLayout := "2006-01-02 15:04:05"
	dataTimeStr := time.Unix(t, 0).Format(timeLayout)
	return dataTimeStr
}

func GetCatName(a interface{}, data *[]string) {
	if a.(SpPurchaseCategory).Level < 1 {
		return
	}
	GetCatName(*a.(SpPurchaseCategory).Parent, data)
	*data = append(*data, a.(SpPurchaseCategory).CategoryName)
	return
}

// check where has changed when backstage edit happened
func (s *SpSellerQuotedBuyer) BackstageEditDetail(newQuote SpSellerQuotedBuyer, o orm.Ormer, oldValueList []string) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			logs.Error("update purchase failed! reason: ", rec)
			err = errors.New("采购单代编辑失败，请重试")
		}
	}()
	var editHistory SpPurchaseEditLog
	var map1 = make(map[string]interface{})
	var history = M{Map:map1}
	now := time.Now().Unix()
	// compare backstage edit detail
	key := reflect.TypeOf(newQuote.QuoteResp)
	newValue := reflect.ValueOf(newQuote.QuoteResp)
	oldValue := reflect.ValueOf(s.QuoteResp)
	keyNum := key.NumField()
	var endChan = make(chan struct{}, 1)
	logs.Info("keyNum: ", keyNum)
	for i := 0; i < keyNum; i++ {
		// compare values between new and old
		name := key.Field(i).Name
		newV1 := newValue.Field(i).Interface()
		oldV1 := oldValue.Field(i).Interface()
		go func(newV, oldV interface{}, history M) {

			if newV != oldV {
				// record difference from different key
				switch name {
				case "NormsRequirement":
					oldNorm := utils.StringToSliceInt(oldV.(string))
					newNorm := utils.StringToSliceInt(newV.(string))
					if len(oldNorm) != len(newNorm) {
						err = errors.New("属性异常，请确认")
						return
					}
					change := false
					var changeList = make([]string, 0)
					for i := 0; i < len(oldNorm); i++ {
						if oldNorm[i] != newNorm[i] {
							changeList = append(changeList, oldValueList[i])
							change = true
						} else {
							changeList = append(changeList, constant.NULL)
						}
					}
					if change {
						history.Set(name, changeList)
					}
				case "AddressInfo":
					oldInfo := utils.StringToMap(oldV.(string))
					newInfo := utils.StringToMap(newV.(string))
					for k, v := range oldInfo {
						if v != newInfo[k] {
							history.Set(name, oldInfo)
							break
						}
					}
				case "PurchaseList":
					if result := CompareFileList(newV, oldV); result != nil {
						history.Set(name, result)
					}
				case "GoodsPicture":
					if result := CompareFileList(newV, oldV); result != nil {
						history.Set(name, result)
					}
				case "Documents":
					if result := CompareFileList(newV, oldV); result != nil {
						history.Set(name, result)
					}
				case "Include":
					if result := CompareIntList(oldV.(string), newV.(string)); result != nil {
						history.Set(name, result)
					}
				case "BackstageEdit":
					//endChan <- struct{}{}
					//return
				default:
					history.Set(name, oldV)
				}
			}
			endChan <- struct{}{}
		}(newV1, oldV1, history)
	}
	for i := 0; i < keyNum; i++ {
		<- endChan
		if i == keyNum - 1 {
			close(endChan)
		}
	}
	// check if exists a change record with this purchase
	if errs := o.QueryTable(&SpPurchaseEditLog{}).Filter("QuoteId", s.Id).Filter("IsDeleted", constant.ZERO).One(&editHistory); errs != nil {
		logs.Info("there did not exist edit history! start to insert ...")
		editHistory.QuoteId = s.Id
		editHistory.CreateTime = now
		editHistory.UpdateTime = now
		editHistory.LogDetail = utils.ToString(history.Map)
		if _, errs := o.Insert(&editHistory); errs != nil {
			logs.Error("save edit history failed! reason: ", errs.Error())
			if errs := o.Rollback(); errs != nil {
				logs.Error("save edit history failed! reason: ", errs.Error())
				err = errors.New("数据操作异常，请稍候重试")
				return
			}
			err = errors.New("代编辑信息存储失败，请重试")
			return
		}
		return
	}
	oldHistory := utils.StringToMap(editHistory.LogDetail)
	change := false
	for k, v := range history.Map {
		if _, ok := oldHistory[k]; !ok {
			oldHistory[k] = v
			change = true
		}
	}
	if change {
		editHistory.LogDetail = utils.ToString(oldHistory)
		editHistory.UpdateTime = now
		if _, errs := o.Update(&editHistory); errs != nil {
			logs.Error("save edit history failed! reason: ", errs.Error())
			if errs := o.Rollback(); errs != nil {
				logs.Error("save edit history failed! reason: ", errs.Error())
				err = errors.New("数据操作异常，请稍候重试")
				return
			}
			err = errors.New("代编辑信息存储失败，请重试")
			return
		}
	}
	return
}

func CompareFileList(newV, oldV interface{}) (fileList []string) {
	newNorm := utils.StringToSliceMap(newV.(string))
	oldNorm := utils.StringToSliceMap(oldV.(string))
	if len(newNorm) != len(oldNorm) {
		for _, i := range oldNorm {
			fileList = append(fileList, i["imgName"])
		}
	} else {
		listCount := constant.ZERO
		for _, i := range oldNorm {
			fileList = append(fileList, i["imgName"])
			for _, j := range newNorm {
				detailCount := constant.ZERO
				for k, v := range i {
					if v == j[k] {
						detailCount += constant.ONE
					}
				}
				if detailCount == len(i) {
					listCount += constant.ONE
				}
			}
		}
		if listCount == len(oldNorm) {
			return nil
		}
	}
	return
}

func CompareIntList(newV, oldV interface{}) (oldInvoice []int) {
	oldInvoice = utils.StringToSliceInt(oldV.(string))
	newInvoice := utils.StringToSliceInt(newV.(string))
	if len(oldInvoice) != len(newInvoice) {
		return
	}
	count := constant.ZERO
	for _, v := range oldInvoice {
		for _, n := range newInvoice {
			if v == n {
				count += constant.ONE
			}
		}
	}
	if count == len(oldInvoice) {
		return nil
	}
	return
}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func (s *SpSellerQuotedBuyer) IsValid(now int64, status int) (result bool) {
	result = false
	if constant.Canceled != status && s.AuditStatus == constant.APPROVED &&
		(s.QuotedEndTime > now || (s.QuotedEndTime > now-constant.PurchaseFreezeTime && s.QuoteCount > constant.ZERO) ||
			(now-constant.PurchaseOverTime < s.QuotedEndTime && now-constant.PurchaseFreezeTime > s.QuotedEndTime && constant.ZERO < s.SuccessSellerId)) {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsOnAudit(now int64, status int) (result bool) {
	result = false
	if constant.OnPurchase == status && constant.InReview == s.AuditStatus && now < s.QuotedEndTime {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsInReview(now int64, status int) (result bool) {
	result = false
	if constant.OnPurchase == status && constant.InReview == s.AuditStatus && now < s.QuotedEndTime {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsUnApprove(now int64, status int) (result bool) {
	result = false
	if constant.OnPurchase == status && constant.UnApprove == s.AuditStatus && now < s.QuotedEndTime {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsOnPurchase(now int64, status int) (result bool) {
	result = false
	if constant.OnPurchase == status && constant.APPROVED == s.AuditStatus && (now < s.QuotedEndTime ||
		(now-constant.PurchaseFreezeTime < s.QuotedEndTime && now >= s.QuotedEndTime && constant.ZERO < s.QuoteCount) ||
		(now-constant.PurchaseOverTime < s.QuotedEndTime && now-constant.PurchaseFreezeTime >= s.QuotedEndTime &&
			constant.ZERO < s.QuoteCount && constant.ZERO < s.SuccessSellerId)) {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsTraded(status int) (result bool) {
	result = false
	if constant.Traded == status {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsPaused(now int64, status int) (result bool) {
	result = false
	if constant.Paused == status && now < s.QuotedEndTime {
		result = true
		return
	}
	return
}
func (s *SpSellerQuotedBuyer) IsBackstageEdit(now int64, status int) (result bool) {
	result = false
	if constant.OnPurchase == status && constant.UnApprove == s.AuditStatus && constant.Yes == s.BackstageEdit && now < s.QuotedEndTime {
		result = true
		return
	}
	return
}

func (s *SpExchangePurchaseLog) GetDetail() (data ExchangeDetailResp, err error) {
	var userChan = make(chan map[int]interface{}, 1)
	go utils.GetUserInfoAsyc([]int{s.BuyerId}, userChan)
	o := orm.NewOrm()
	if errs := o.Read(s); errs != nil {
		logs.Error("get exchange purchase detail failed! reason: ", errs.Error())
		err = errors.New("获取转单详情失败")
		return
	}
	// get category info
	if s.PurchaseType == constant.Single {
		var resp = make([]string, 0)
		var quote SpSellerQuotedBuyer
		quote.TmpCatId = s.TmpCatId
		resp, err = GetCatInfo(&quote)
		if err != nil {
			return
		}
		data.CatList = resp
	}
	data.QuoteResp = s.QuoteResp
	data.MobileNumber = s.MobileNumber
	data.BuyerId = s.BuyerId
	data.QuoteCommon.TransWithGet(s.QuoteResp)
	data.QuotedEndTimes = TimeStampToLocalString(s.QuotedEndTime)
	data.DeliveryEndTimes = TimeStampToLocalString(s.DeliveryEndTime)
	data.CreateTimes = TimeStampToLocalString(s.CreateTime)

	// get user info
	userInfoList := <-userChan
	data.User = userInfoList[s.BuyerId]

	// get attribute and attribute value name choice list
	if s.PurchaseType == constant.Single && data.TransCommon.NormsRequirement != nil {
		attrValue, errs := GetAttrValueName(s.NormsRequirement)
		if errs != nil {
			err = errors.New("查询采购商品属性失败")
			return
		}
		data.AttrName = attrValue
	}
	return
}


// M
type M struct {
	Map     map[string]interface{}
	lock 	sync.RWMutex // 加锁
}

// Set ...
func (m *M) Set(key string, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Map[key] = value
}