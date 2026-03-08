package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// GlobalSetting stores key-value pairs for global configurations (e.g., DNS, Ports)
type GlobalSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string // JSON string
}

// User represents an administrator
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex" json:"username"`
	Password string `json:"-"`                        // Bcrypt hash, never send to client
	Token    string `gorm:"uniqueIndex" json:"token"` // Subscription token
}

// Node represents a proxy node (SS, Vmess, Trojan, etc.)
type Node struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex"`
	Type        string // ss, vmess, hysteria2, etc.
	Server      string
	Port        int
	UUID        string // VMess, VLESS, Hysteria2 (password sometimes stored here or in Password)
	Password    string // SS, Trojan, Hysteria2
	Cipher      string // SS, VMess
	UDP         bool
	TLS         bool
	SkipCert    bool
	Network     string // ws, grpc, etc.
	Path        string // WebSocket path
	Host        string // HTTP Host / SNI
	ALPN        string // h3
	ExtraConfig string // JSON for other fields like headers, flow, etc.
	Source      string `gorm:"default:manual"` // Source: manual (手动创建) or subscription (订阅同步)
	SourceID    uint   // ID of the subscription source if from subscription
	SourceName  string // Name of the subscription source for display
}

// Rule represents a routing rule
type Rule struct {
	gorm.Model
	Type       string // DOMAIN-SUFFIX, IP-CIDR, etc.
	Payload    string // google.com, US
	Target     string // ProxyGroupName or specifics (Legacy)
	TargetID   uint   `gorm:"default:0"` // ID of the target Node or Group
	TargetType string // "node", "group", "builtin"
	Priority   int    `gorm:"default:0"` // Lower number = higher priority
	NoResolve  bool
	Tag        string // Tag for categorization: shopping, carrier, video, game, etc.
	Remark     string // Remark for display only, not used in config generation
}

// ProxyGroupDB represents a user-defined proxy group (optional if we want dynamic groups)
// For simplicity, we might store groups in GlobalSetting or a separate table
type ProxyGroupModel struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex" json:"name"`
	Type     string `json:"type"`     // select, url-test
	ProxyIDs string `json:"proxyIDs"` // JSON array of node IDs: [1, 2, 3]
	Use      string `json:"use"`      // JSON array of provider names/other groups
	URL      string `json:"url"`
	Interval int    `json:"interval"`
	Source   string `gorm:"default:user" json:"source"` // Source: user (手动创建) or subscription (订阅同步)
}

// MarshalJSON implements custom JSON marshaling to capitalize first letter
func (p ProxyGroupModel) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        uint      `json:"ID"`
		CreatedAt time.Time `json:"CreatedAt"`
		UpdatedAt time.Time `json:"UpdatedAt"`
		Name      string    `json:"Name"`
		Type      string    `json:"Type"`
		ProxyIDs  string    `json:"ProxyIDs"`
		Use       string    `json:"Use"`
		URL       string    `json:"URL"`
		Interval  int       `json:"Interval"`
		Source    string    `json:"Source"`
	}{
		ID:        p.ID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Name:      p.Name,
		Type:      p.Type,
		ProxyIDs:  p.ProxyIDs,
		Use:       p.Use,
		URL:       p.URL,
		Interval:  p.Interval,
		Source:    p.Source,
	})
}

// ProxyNode represents a node reference in a proxy group
type ProxyNode struct {
	ID   uint   `json:"id"`
	Type string `json:"type"` // "node" or "group"
	Name string `json:"name"` // populated by API for display
}

// SubscriptionLog represents a subscription access log
type SubscriptionLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	UserID    uint           `gorm:"index" json:"userId"`                     // User who subscribed
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"` // Association to User
	Token     string         `gorm:"index" json:"token"`                      // Token used (without exposing full token)
	IP        string         `gorm:"index" json:"ip"`                         // Client IP
	UserAgent string         `json:"userAgent"`                               // Client user agent
	Success   bool           `json:"success"`                                 // Whether subscription was successful
	Error     string         `json:"error"`                                   // Error message if failed
}

// TableName specifies the table name for SubscriptionLog
func (SubscriptionLog) TableName() string {
	return "subscription_logs"
}

// SubscriptionSource represents a third-party subscription source
type SubscriptionSource struct {
	gorm.Model
	Name           string     `gorm:"uniqueIndex" json:"name"`          // 订阅源名称
	URL            string     `json:"url"`                              // 订阅链接
	Enabled        bool       `gorm:"default:true" json:"enabled"`      // 是否启用
	UpdateInterval int        `gorm:"default:24" json:"updateInterval"` // 更新间隔(小时)
	LastSync       *time.Time `json:"lastSync"`                         // 最后同步时间
	NodeTag        string     `json:"nodeTag"`                          // 导入节点的标签
	SyncMode       string     `gorm:"default:append" json:"syncMode"`   // 同步模式: append, replace, smart
	Error          string     `json:"error"`                            // 最后一次错误信息
}

// MarshalJSON implements custom JSON marshaling to capitalize first letter
func (s SubscriptionSource) MarshalJSON() ([]byte, error) {
	type Alias SubscriptionSource
	return json.Marshal(struct {
		ID             uint       `json:"ID"`
		CreatedAt      time.Time  `json:"CreatedAt"`
		UpdatedAt      time.Time  `json:"UpdatedAt"`
		Name           string     `json:"Name"`
		URL            string     `json:"URL"`
		Enabled        bool       `json:"Enabled"`
		UpdateInterval int        `json:"UpdateInterval"`
		LastSync       *time.Time `json:"LastSync"`
		NodeTag        string     `json:"NodeTag"`
		SyncMode       string     `json:"SyncMode"`
		Error          string     `json:"Error"`
	}{
		ID:             s.ID,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		Name:           s.Name,
		URL:            s.URL,
		Enabled:        s.Enabled,
		UpdateInterval: s.UpdateInterval,
		LastSync:       s.LastSync,
		NodeTag:        s.NodeTag,
		SyncMode:       s.SyncMode,
		Error:          s.Error,
	})
}

// TableName specifies the table name for SubscriptionSource
func (SubscriptionSource) TableName() string {
	return "subscription_sources"
}
