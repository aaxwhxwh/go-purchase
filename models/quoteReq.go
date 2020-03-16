package models

import (
	"errors"
	"regexp"
	"server-purchase/utils/constant"
	"strings"
)

// 采购单编辑入参
//type UpdateQuoteReq struct {
//	Template TmpPurchaseCategory    `json:"Template"`
//	Purchase QuoteCommon
//	BuyerId          int       `json:"BuyerId"`
//	Id               int       `json:"Id"`
//}
type UpdateQuoteReq struct {
	QuoteReq
	Id               int       `json:"Id"`
}


// 创建采购单
type QuoteReq struct {
	Template Template    `json:"Template"`
	Purchase QuoteCommon `json:"Purchase"`
	BuyerId   int         `json:"BuyerId"`
}
type Template struct {
	TmpPurchaseCategory
	TmpCatId int `valid:"Required"`
}

type QuoteCommon struct {
	QuoteResp
	TransCommon
}
func (q *QuoteReq) PurchasePreSave(now int64) (err error) {
	c := q.Purchase
	reg := regexp.MustCompile(constant.MobileReg)
	if constant.NULL == c.TransCommon.MobileNumber || !reg.MatchString(c.TransCommon.MobileNumber) {
		err = errors.New("非法用户手机")
		return
	}
	if constant.Single == c.PurchaseType && len(c.TransCommon.Documents) > constant.MaxDocuments {
		err = errors.New("设计文件数量超出允许范围")
		return
	}
	if constant.Single == c.PurchaseType && len(c.TransCommon.GoodsPicture) > constant.MaxPictures {
		err = errors.New("成品展示文件数量超出允许范围")
		return
	}
	if constant.ZERO >= c.GoodsQty {
		err = errors.New("商品数量最小值需大于1")
	}
	if (c.PurchaseType == constant.Batch || constant.CompanyPrint == c.PurchaseType)&& (nil == c.TransCommon.PurchaseList || len(c.TransCommon.PurchaseList) < constant.MinFileNum && constant.NULL == c.PurchaseListDesc) {
		err = errors.New("请至少填写一项采购需求详情")
		return
	}
	if now >= c.QuotedEndTime || c.QuotedEndTime >= c.DeliveryEndTime {
		err = errors.New("非法送达时间或报价截止时间")
		return
	}
	if 0 == c.AddressId || nil == c.TransCommon.AddressInfo {
		err = errors.New("请选择用户收货信息")
		return
	}
	if constant.NULL == c.City {
		c.City = c.TransCommon.AddressInfo["city_id"].(string)
	}
	if constant.Single == c.PurchaseType && constant.NULL == c.Unit {
		err = errors.New("未选择正确的产品单位")
		return
	}

	if constant.OtherCat == q.Template.SpCatId && constant.NULL == strings.TrimSpace(q.Purchase.PurchaseListDesc) {
		err = errors.New("请填写详细规格要求")
		return
	}
	if constant.OtherCat != q.Template.SpCatId && constant.Single == c.PurchaseType &&
		(c.TransCommon.NormsRequirement == nil || constant.ZERO > len(c.TransCommon.NormsRequirement)) {
		err = errors.New("请选择相应的商品属性")
		return
	}
	return
}


// 修改采购单状态
type ChangeQuoteOrderStatusReq struct {
	Id int `valid:"Required"`
	Status int `valid:"Required"`
	BuyerId int `valid:"Required"`
}

// 修改采购单审核状态
type PurchaseAuditReq struct {
	AuditStatus int `valid:"Required"`
	PurchaseId int `valid:"Required"`
	AuditMessage string
}

// 修改采购单相关订单信息
type UpdateOrderReq struct {
	OrderSn string
	SuccessSellerId int
	PurchaseId int			`valid:"Required"`
	BuyerId int				`valid:"Required"`
	Price float64
	Status int				`valid:"Required"`
}

// 订单上传文件
type UploadFileReq struct {
	QuoteId int							`valid:"Required"`
	Documents map[string]interface{}	`valid:"Required"`
	UserId int							`valid:"Required"`
}

// @Title change purchase owner
type ChangeOwnerReq struct {
	QuoteId int						`valid:"Required"`
	ReceiveUserId  int				`valid:"Required"`
	OperatorId		int				`valid:"Required"`
}