// @Time    : 2019-09-02 15:25
// @Author  : Frank
// @Email   : frank@163.com
// @File    : models.go
// @Software: GoLand
package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type SpCatLinkAttr struct {
	Id        int  `orm:"column(id);auto"`
	Cat     *SpPurchaseCategory `valid:"Required" orm:"rel(fk);column(cat_id)" description:"分类id"`
	Attr    *SpPurchaseAttr `valid:"Required" orm:"rel(fk);column(attr_id)" description:"属性id"`
	Sort      int   `orm:"column(sort)"`
	IsMust    int  `orm:"column(is_must)" description:"是否必填 0 不是 1是"`
	IsDeleted int8 `orm:"column(is_deleted)" description:"是否删除 0 不是 1是"`
}

func (t *SpCatLinkAttr) TableName() string {
	return "sp_cat_link_attr"
}

type SpPurchaseAttr struct {
	Id            int    `json:"id" orm:"column(attr_id);auto"`
	AttrName      string `json:"attr_name" orm:"column(attr_name);size(255)" description:"属性名称"`
	AttrAliasName string `json:"attr_alias_name" orm:"column(attr_alias_name);size(255)" description:"属性别名"`
	AttrType      int8   `json:"attr_type" orm:"column(attr_type)" description:"属性类型 0.名片,1.装订,2.海报"`
	AttrStatus    int8   `json:"attr_status" orm:"column(attr_status)" description:"属性状态 0.启用,1.停用"`
	AddUserId     int    `json:"add_user_id" orm:"column(add_user_id)" description:"添加人id"`
	Remark        string `json:"remark" orm:"column(remark);size(255);null" description:"备注"`
	UpdateTime    int    `json:"update_time" orm:"column(update_time)" description:"修改时间"`
	CreateTime    int    `json:"create_time" orm:"column(create_time)" description:"新建时间"`
	Desc          int    `json:"desc" orm:"column(desc)" description:"排序"`
	IsDeleted     int8   `json:"is_deleted" orm:"column(is_deleted)" description:"0:未删除      1:已删除"`
	IsLimitUser   int8   `json:"is_limit_user" orm:"column(is_limit_user)" description:"是否允许用户自定义0不允许1允许"`
	ShopId        int    `json:"shop_id" orm:"column(shop_id)" description:"店铺id"`
	Category      []*SpPurchaseCategory `orm:"reverse(many)"`
	AttrValue      []*SpPurchaseAttrValue `orm:"reverse(many)"`
}

func (t *SpPurchaseAttr) TableName() string {
	return "sp_purchase_attr"
}

type SpPurchaseAttrValue struct {
	Id         int    `orm:"column(attr_value_id);auto"`
	AttrValue  string `orm:"column(attr_value);size(255)" description:"属性值"`
	UnitTypeId int8   `orm:"column(unit_type_id)" description:"单位:0mm,1米,2吋"`
	Sort       int32  `orm:"column(sort)" description:"排序"`
	Customize  int  `orm:"column(customize)" description:"是否自定义属性值"`
	Remark     string `orm:"column(remark);size(128)" description:"备注"`
	CreateTime int    `orm:"column(create_time)" description:"创建时间"`
	UpdateTime int    `orm:"column(update_time)" description:"更新时间"`
	IsDeleted  int    `orm:"column(is_deleted)" description:"0:未删除 1:已删除"`
	Attr    *SpPurchaseAttr `valid:"Required" orm:"rel(fk);column(attr_id)" description:"属性id"`
}

func (t *SpPurchaseAttrValue) TableName() string {
	return "sp_purchase_attr_value"
}

type SpPurchaseCategory struct {
	Id             int    `json:"category_id" orm:"column(category_id);pk;auto"`
	CategoryName   string `json:"category_name" valid:"Required;MaxSize(128);",orm:"column(category_name);size(128)" description:"分类名称"`
	//ParentId       int    `orm:"column(parent_id);" description:"父级id"`
	Parent         *SpPurchaseCategory    `json:"parent_id" valid:"Required" orm:"rel(fk)" description:"父级id"`
	CategoryStatus int8   `json:"category_status" orm:"column(category_status)" description:"状态 0.在用,1.停用"`
	Sort           int32  `json:"sort" orm:"column(sort);" description:"排序字段"`
	PurchaseCatId  int8   `json:"purchase_cat_id" orm:"column(purchase_cat_id)" description:"关联采购使用分类表id"`
	Remark         string `json:"remark" valid:"MaxSize(255)",orm:"column(remark);size(512);null" description:"备注"`
	UpdateTime     int    `json:"update_time" orm:"column(update_time)" description:"更新时间"`
	CreateTime     int    `json:"create_time" orm:"column(create_time)" description:"创建时间"`
	IsEnd          int8   `json:"is_end" orm:"column(is_end);default(1)" description:"是否为最后一级 0不是1是"`
	IsDeleted      int8   `json:"is_deleted" orm:"column(is_deleted)" description:"是否删除 0 不是 1是"`
	Level          int    `json:"level" orm:"column(level);" description:"父级id"`
	Unit		   string  `json:"unit" orm:"column(unit);" description:"单位"`
	Ch            []*SpPurchaseCategory `json:"children" orm:"reverse(many)" useDefaultRelsDepth:"3"`
	SpPurchaseAttr  []*SpPurchaseAttr `orm:"rel(m2m)"`
}

func (t *SpPurchaseCategory) TableName() string {
	return "sp_purchase_category"
}


func GetOneCat(cat_id int) (data interface{},err error,msg string){
	o := orm.NewOrm()
	var category SpPurchaseCategory
	queryset := o.QueryTable("sp_purchase_category").Filter("category_id",cat_id).Filter("is_deleted",0)
	err = queryset.One(&category)
	if err != nil {
		return nil,err,"查无此分类"
	}
	if category.IsEnd == int8(1){
		var DefaultRelsDepth = category.Level-1
		err = queryset.RelatedSel(DefaultRelsDepth).One(&category)
		if err != nil{
			return nil , err ,"查询失败"
		}
	}else {
		err = queryset.One(&category)
		if err != nil{
			return nil , err ,"查询失败"
		}
		err = get_cat_ch(&category,o)
		if err != nil{
			return nil , err ,"查询失败"
		}
	}
	return category ,nil ,"ok"
}

func get_cat_ch(category *SpPurchaseCategory,o orm.Ormer)  (err error){
	if category.IsEnd == 1 {
		return nil
	}

	_,err =o.LoadRelated(category,"Ch")
	if err != nil{
		return err
	}
	ch := category.Ch
	for _,i := range ch{
		err = get_cat_ch(i,o)
		if err != nil {
			return err
		}
	}
	return nil
}


type SpSellerQuoted struct {
	Id         int     `orm:"column(seller_quoted_id);auto"`
	Quoted   *SpSellerQuotedBuyer     `orm:"rel(fk);" description:"用户报价id"`
	ShopId     int     `orm:"column(shop_id)" description:"店铺id"`
	UnitPrice  float64 `orm:"column(unit_price);digits(10);decimals(3)" description:"单价"`
	TotalPrice float64 `orm:"column(total_price);digits(10);decimals(2)" description:"总价"`
	Remark     string  `orm:"column(remark);null" description:"备注"`
	File     string  `orm:"column(file);size(128)" description:"上传文件"`
	FileName     string  `orm:"column(file_name);size(128)" description:"文件名称"`
	Status     int8    `orm:"column(status)" description:"状态0已报价，1抢单成功，2抢单失败"`
	QuotedTime     int    `json:"update_time" orm:"column(quoted_time)" description:"报价时间"`
}

func (t *SpSellerQuoted) TableName() string {
	return "sp_seller_quoted"
}



type HistorySellerQuoted struct {
	Id		   int     `orm:"column(history_quoted_id);auto"`
	SellerQuoted         *SpSellerQuoted     `orm:"rel(fk);" description:"商家报价单"`
	Quoted   *SpSellerQuotedBuyer     `orm:"rel(fk);" description:"用户报价id"`
	ShopId     int     `orm:"column(shop_id)" description:"店铺id"`
	UnitPrice  float64 `orm:"column(unit_price);digits(10);decimals(3)" description:"单价"`
	TotalPrice float64 `orm:"column(total_price);digits(10);decimals(2)" description:"总价"`
	Remark     string  `orm:"column(remark);null" description:"备注"`
	File     string  `orm:"column(file);size(128)" description:"上传文件"`
	FileName     string  `orm:"column(file_name);size(128)" description:"文件名"`
	Status     int8    `orm:"column(status)" description:"状态0已报价，1抢单成功，2抢单失败"`
	QuotedTime     int    `json:"quoted_time" orm:"column(quoted_time)" description:"报价时间"`
}

func (t *HistorySellerQuoted) TableName() string {
	return "history_seller_quoted"
}


func init() {
	// 设置默认数据库
	user := beego.AppConfig.String("mysqluser")
	pwd := beego.AppConfig.String("mysqlpass")
	host := beego.AppConfig.String("mysqlurls")
	port := beego.AppConfig.String("mysqlport")
	name := beego.AppConfig.String("mysqldb")
	orm.RegisterDataBase("default", "mysql", user+":"+pwd+"@tcp("+host+":"+port+")/"+name+"?charset=utf8")
	orm.RegisterModel( new(SpCatLinkAttr), new(SpPurchaseAttr), new(SpPurchaseAttrValue),
		new(SpPurchaseCategory), new(SpSellerQuoted), new(SpSellerQuotedBuyer),
		new(TmpPurchaseAttr), new(TmpPurchaseAttrValue), new(TmpPurchaseCategory),new(HistorySellerQuoted),
		new(SpPurchaseEditLog), new(SpExchangePurchaseLog))
}
