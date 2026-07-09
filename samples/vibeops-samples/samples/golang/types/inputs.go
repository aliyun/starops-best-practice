package types

type UserInputParams struct {
	// 基础输入
	RegionID    string `json:"region,omitempty"`
	Workspace   string `json:"workspace,omitempty"`
	Project     string `json:"project,omitempty"`
	LogStore    string `json:"logstore,omitempty"`
	MetricStore string `json:"metricstore,omitempty"`

	// 用户语言、时区、时间戳
	Language  string `json:"language,omitempty"`
	TimeZone  string `json:"timeZone,omitempty"`  // 标准的TimeZone格式
	TimeStamp string `json:"timeStamp,omitempty"` // Unix时间戳，单位秒

	// 值为 []UserContext 的 JSON 字符串
	UserContext string `json:"userContext,omitempty"`

	// 值为 map[string]any 的 JSON 字符串
	Config string `json:"config,omitempty"`

	// 用于 Agent 恢复运行时，读取用户的决策
	UserInteractiveResp map[string]any `json:"userInteractive,omitempty"`
}

// UserContextType 定义上下文类型
type UserContextType string

const (
	UserContextTypeMetadata     UserContextType = "metadata"
	UserContextTypeEntity       UserContextType = "entity"
	UserContextTypeSQLGenerated UserContextType = "sql_generation"
	UserContextTypeSPLGenerated UserContextType = "spl_generation"
)

// MetadataUserData 元数据类型的数据
type MetadataUserData struct {
	FromTime int64 `json:"fromTime,omitempty" xml:"fromTime,omitempty"`
	ToTime   int64 `json:"toTime,omitempty" xml:"toTime,omitempty"`
}

type SQLGenerationUserData struct {
	RawUserQuery string `json:"rawUserQuery,omitempty"`
}

type SPLGenerationUserData struct {
	RawUserQuery string `json:"rawUserQuery,omitempty"`
	Data         string `json:"data,omitempty"`
}

// EntityUserData 实体类型的数据
type EntityUserData struct {
	EntityId     string `json:"entityId,omitempty"`
	EntityDomain string `json:"entityDomain,omitempty"`
	EntityType   string `json:"entityType,omitempty"`
	FromTime     int64  `json:"fromTime,omitempty"`
	ToTime       int64  `json:"toTime,omitempty"`
	Title        string `json:"title,omitempty"`
}

// ContextData 上下文数据接口，所有 UserContext 的 Data 类型都应实现此接口
type ContextData any

// UserContext 用户上下文，根据Type决定Data的类型
type UserContext struct {
	Type UserContextType `json:"type" xml:"type"`
	Data ContextData     `json:"data" xml:"data"`
}

// CommonUserData map 类型的通用数据
type CommonUserData map[string]any
