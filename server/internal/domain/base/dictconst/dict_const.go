package dictconst

// 字典编码常量
const (
	DictUserType   = "user_type"
	DictGender     = "gender"
	DictNoticeType = "notice_type"
	DictTaskType   = "task_type"
	DictConfigType = "config_type"
)

// 用户类型（user_type）字典项 Key 常量
const (
	UserTypeAdmin  = "admin"
	UserTypeNormal = "normal"
	UserTypeWechat = "wechat"
	UserTypeAlipay = "alipay"
)

// 性别（gender）字典项 Key 常量
const (
	GenderUnknown = "unknown"
	GenderMale    = "male"
	GenderFemale  = "female"
)

// 公告类型（notice_type）字典项 Key 常量
const (
	NoticeTypeSystem  = "system"
	NoticeTypeActivity = "activity"
	NoticeTypeNotice   = "notice"
)

// 任务类型（task_type）字典项 Key 常量
const (
	TaskTypeHTTP             = "http"
	TaskTypeDatabase         = "database"
	TaskTypeCache            = "cache"
	TaskTypeScript           = "script"
	TaskTypeStreamMaintenance = "stream_maintenance"
)

// 配置类型（config_type）字典项 Key 常量
const (
	ConfigTypeSystem   = "system"
	ConfigTypeWechat   = "wechat"
	ConfigTypeAlipay   = "alipay"
	ConfigTypeBusiness = "business"
)
