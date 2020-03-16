// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"server-purchase/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/api/server_purchase",
		beego.NSRouter("/purchase", &controllers.Quoted{}),
		beego.NSRouter("/purchase/republish", &controllers.Quoted{}, "Post:RePublish"),
		beego.NSRouter("/purchase/detail", &controllers.Quoted{}, "Get:GetPurchaseDetail"),
		beego.NSRouter("/purchase/user/list", &controllers.Quoted{}, "Get:GetUserPurchaseList"),
		beego.NSRouter("/purchase/user/list/count", &controllers.Quoted{}, "Get:GetUserPurchaseListNum"),
		beego.NSRouter("/purchase/operation/list", &controllers.Quoted{}, "Get:OperationList"),
		beego.NSRouter("/purchase/orderInfo", &controllers.Quoted{}, "Put:PurchaseOrder"),
		beego.NSRouter("/purchase/on_audit/count", &controllers.Quoted{}, "Get:OnAuditListNum"),
		beego.NSRouter("/purchase/upload", &controllers.Quoted{}, "Put:UploadFileFromOrder"),
		beego.NSRouter("/purchase/change_owner", &controllers.Quoted{}, "Put:ChangePurchaseOwner"),
		beego.NSRouter("/purchase/exchange_list", &controllers.Quoted{}, "Get:ChangeList"),

		beego.NSRouter("/purchase/admin/edit", &controllers.Quoted{}, "Put:BackstageEdit"),
		beego.NSRouter("/purchase/admin/submit", &controllers.Quoted{}, "Put:BackstageSubmit"),

		beego.NSRouter("/purchase/status/operation", &controllers.Quoted{}, "Put:ChangeQuoteOrderStatus"),
		beego.NSRouter("/purchase/auditStatus/operation", &controllers.Quoted{}, "Put:PurchaseAudit"),
		beego.NSRouter("/cat",&controllers.CatController{},),
		beego.NSRouter("/cat/check",&controllers.CatController{}, "Get:CheckTempUpdate"),
		beego.NSRouter("/user_cat",&controllers.CatController{},"Get:UserGet"),
		beego.NSRouter("/get_attr_cat",&controllers.CatController{},"Get:GetAttrCat"),
		beego.NSRouter("/get_attr_name_cat",&controllers.CatController{},"Get:GetAttrNameCat"),


		beego.NSRouter("/cat_one",&controllers.CatController{},"Get:GetCh"),

		beego.NSRouter("/operate_quoted", &controllers.OperateQuoted{}),
		beego.NSRouter("/goods_attr", &controllers.GoodsAttr{}),
		beego.NSRouter("/template", &controllers.GoodsAttr{}, "Get:CatAttrTmp"),
		beego.NSRouter("/shop_purchase", &controllers.Shop_controllers{}),
		beego.NSRouter("/shop_purchase_content", &controllers.Shop_controllers{},"Get:GetContent"),
		beego.NSRouter("/history_purchase_list", &controllers.Shop_controllers{},"Get:HistoryList"),
		beego.NSRouter("/admin_purchase_content", &controllers.AdminController{}),
		beego.NSRouter("/admin_purchase_list", &controllers.AdminController{},"Get:GetList"),


	)
	beego.AddNamespace(ns)
}
