package constant


// 采购单冻结时间
const PurchaseFreezeTime int64 = 60 * 60 * 24 * 3

// 采购单失效时间
const PurchaseOverTime int64 = 60 * 60 * 24 * 4

// 采购单类型
const Single int = 1			// 单商品采购
const Batch int = 2				// 多商品采购
const CompanyPrint int = 3		// 企业印

// 采购单状态
const (
	OnPurchase int = 0		// 报价中
	Traded int = 1			// 已成交
	Paused int = 2			// 已暂停
	Canceled int = 3		// 采购单已取消
	Freeze int = 4			// 冻结状态
	Auditing int = 5		// 审核中
	AuditFailed int = 6		// 审核未通过
	OutUse int = 7			// 重新编辑后失效
	Expired int = 8			// 过期失效
	BackstageEdit int = 9   // 已代编辑
)



// 获取采购单详情平台
const SHOP int = 2				// 商家端
const OPERATION int = 1			// 运营平台及用户端

// 数字常量
const (
	ZERO int = 0
	ONE int = 1
)

// 空字符串
const NULL string = ""


// 采购单所有审核状态数组
const (
	UnApprove int = 2			//审核未通过
	APPROVED int = 1			//审核已通过
	InReview int = 0			//审核中
)

// 上传文件最大限制
const MaxDocuments int = 3		// 最大设计文件数
const MaxPictures int = 9		// 最大商品图片数量
const MinFileNum int = 1

// 采购单订单状态
const CreateOrder int = 1
const CancelOrder int = 2

// 手机号正则
const MobileReg string = "^[1](([3][0-9])|([4][5-9])|([5][0-3,5-9])|([6][5,6])|([7][0-8])|([8][0-9])|([9][1,8,9]))[0-9]{8}$"

// 代编辑状态
const (
	No int = 0
	Yes int = 1
)

// category
const (
	OtherCat int = -1
)

// new status 采购单状态 0 待审核，1 审核未通过，2 报价中，3 已暂停，4 已取消，5 已成交
const (
	AuditingStatus int = 0
	OnPurchaseStatus int = 1
	AuditFailedStatus int = 2
	CanceledStatus int = 3
	PausedStatus int = 4
	TradedStatus int = 5
	FreezeStatus int = 6
	OutUseStatus int = 7
	ExpiredStatus int = 8
	BackstageEditStatus int = 9

)