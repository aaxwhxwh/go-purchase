package models

// 个人中心采购单列表
type UserPurchaseList struct {
	SpSellerQuotedBuyer
	TransCommon
	AttrName []string
}
type UserPurchaseListResp struct {
	Include          []int      `json:"Include"`
	MobileNumber	 string		   `json:"MobileNumber"`
	AddressInfo      map[string]interface{}      `json:"AddressInfo"`

	PurchaseSn       string    `orm:"column(purchase_sn);size(128)" description:"采购编号"`
	City             string    `orm:"column(city);size(255);null" description:"采购城市"`
	GoodsName        string    `orm:"column(goods_name);size(255);null" description:"采购商品"`
	QuotedEndTime    int64     `orm:"column(quoted_end_time);null" valid:"Required" description:"报价截止"`
	DeliveryEndTime  int64     `orm:"column(delivery_end_time);null" valid:"Required" description:"交货日期"`
	Price            float64   `orm:"column(price);null;digits(10);decimals(3)" description:"报价"`
	AddressId        int       `orm:"column(address_id);null" valid:"Required" description:"地址id"`
	PurchaseType     int      `orm:"column(purchase_type)" description:"采购单类型 1 单商品  2 多商品"`
	Id               int       `orm:"column(quoted_id);auto"`
	AuditStatus      int      `orm:"column(audit_status);null" description:"审核状态0平台审核中，审核通过，审核不通过"`
	Status           int      `orm:"column(status);null" description:"报价状态0报价中，1已成交，2已暂停，3已取消"`
	SuccessSellerId  int       `orm:"column(success_seller_id);null" description:"成交的商家报价id"`
	BuyerId          int       `orm:"column(buyer_id)" valid:"Required" description:"买家id"`
	QuoteCount       int       `orm:"column(quote_count)" description:"报价数量统计"`
	CreateTime       int64     `orm:"column(create_time);null" description:"采购单创建时间"`
	UpdateTime       int64     `orm:"column(update_time);null" description:"采购单更新时间"`
	AuditTime        int64     `orm:"column(audit_time);null" description:"采购单审核时间"`
	AttrName []string
	AuditMessage     string
	OrderSn 		 string
	PurchaseStatus   int
	PurchaseListDesc string
	Invoice			 int
	BackstageEdit int
	BackstageEditTime int64
}

// 采购单列表返回
type QuoteRespList struct {
	Id      int    `json:"Id"`
	QuoteCommon
}

// 获取采购单编辑详情
type EditDetailResp struct {
	Template TmpPurchaseCategory `json:"Template"`
	Purchase QuoteCommon         `json:"Purchase"`
	CatList  []string
	AttrName []string
	History   map[string]interface{}
	User  interface{}
}

type CompanyPrintEditDetailResp struct {
	QuoteCommon
	History   map[string]interface{}
	User  interface{}
}

type OperationListResp struct {
	Id               int
	PurchaseSn string
	BuyerId int
	PurchaseType     int
	GoodsQty         int
	QuoteCount       int
	Status           int
	OrderSn          string
	QuotedEndTimes    string
	DeliveryEndTimes  string
	CreateTimes       string
	FixtureDates      string
	CatName			  string
	User             interface{}
}

type PurchaseListNumResp struct {
	Invalid			int64
	Available		int64
	Traded			int64
}

// @Title 查询采购单详情
type QuoteDetail struct {
	QuoteCommon
	PurchaseOrder
	Status           int
	QuoteCount       int
	MobileNumber     string
	AttrName []string
	BuyerId int
	CatList	[]string
	User  interface{}

	QuotedEndTimes    string
	DeliveryEndTimes  string
	CreateTimes       string
	FixtureDates      string
	PurchaseStatus    int
}

type ExchangeListResp struct {
	Id               int
	QuoteId				int
	PurchaseSn string
	PurchaseType     int
	GoodsQty         int
	QuotedEndTimes    string
	DeliveryEndTimes  string
	ExchangeTimes     string
	CatName			  string
	OperatorId        int
	User             interface{}
	OriginalUser	 interface{}
}

type ExchangeDetailResp struct {
	QuoteCommon
	QuoteCount       int
	MobileNumber     string
	AttrName []string
	BuyerId int
	CatList	[]string
	User  interface{}

	QuotedEndTimes    string
	DeliveryEndTimes  string
	CreateTimes       string
}