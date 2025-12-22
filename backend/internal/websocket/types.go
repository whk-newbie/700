package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

// ClientType 客户端类型
type ClientType string

const (
	ClientTypeWindows  ClientType = "windows"  // Windows客户端
	ClientTypeDashboard ClientType = "dashboard" // 前端看板
)

// Client WebSocket客户端
type Client struct {
	ID             string          // 客户端唯一ID
	Type           ClientType      // 客户端类型
	ActivationCode string          // 激活码（Windows客户端使用）
	GroupID        uint            // 分组ID
	UserID         uint            // 用户ID（前端看板使用）
	Conn           *websocket.Conn // WebSocket连接
	Send           chan []byte      // 发送消息通道
	LastHeartbeat  time.Time       // 最后心跳时间
	RegisteredAt   time.Time       // 注册时间
}

// Message WebSocket消息结构
type Message struct {
	Type          string      `json:"type"`           // 消息类型
	ActivationCode string      `json:"activation_code,omitempty"` // 激活码（客户端发送时包含）
	Data          interface{} `json:"data,omitempty"` // 消息数据
	Timestamp     int64       `json:"timestamp,omitempty"` // 时间戳
	Error         string      `json:"error,omitempty"` // 错误信息
}

// HeartbeatMessage 心跳消息
type HeartbeatMessage struct {
	Type          string `json:"type"`
	ActivationCode string `json:"activation_code,omitempty"`
	Timestamp     int64  `json:"timestamp"`
}

// SyncLineAccountsMessage 同步Line账号消息
type SyncLineAccountsMessage struct {
	Type          string                   `json:"type"`
	ActivationCode string                   `json:"activation_code"`
	Data          []LineAccountSyncData    `json:"data"`
}

// LineAccountSyncData Line账号同步数据
type LineAccountSyncData struct {
	LineID        string `json:"line_id"`
	DisplayName   string `json:"display_name,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	PlatformType  string `json:"platform_type"`
	ProfileURL    string `json:"profile_url,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	Bio           string `json:"bio,omitempty"`
	StatusMessage string `json:"status_message,omitempty"`
	OnlineStatus  string `json:"online_status,omitempty"`
}

// IncomingMessage 进线消息
type IncomingMessage struct {
	Type          string            `json:"type"`
	ActivationCode string            `json:"activation_code"`
	Data          IncomingData     `json:"data"`
}

// IncomingData 进线数据
type IncomingData struct {
	LineAccountID  string `json:"line_account_id"`  // Line账号的line_id
	IncomingLineID string `json:"incoming_line_id"` // 进线客户的Line User ID
	Timestamp      string `json:"timestamp"`
	DisplayName    string `json:"display_name,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
}

// CustomerSyncMessage 客户同步消息
type CustomerSyncMessage struct {
	Type          string         `json:"type"`
	ActivationCode string         `json:"activation_code"`
	Data          CustomerData  `json:"data"`
}

// CustomerData 客户数据
type CustomerData struct {
	LineAccountID string `json:"line_account_id"`
	CustomerID    string `json:"customer_id"`
	DisplayName   string `json:"display_name,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Country       string `json:"country,omitempty"`
	Birthday      string `json:"birthday,omitempty"`
	Address       string `json:"address,omitempty"`
	Remark        string `json:"remark,omitempty"`
}

// FollowUpSyncMessage 跟进记录同步消息
type FollowUpSyncMessage struct {
	Type          string        `json:"type"`
	ActivationCode string        `json:"activation_code"`
	Data          FollowUpData `json:"data"`
}

// FollowUpData 跟进记录数据
type FollowUpData struct {
	LineAccountID string `json:"line_account_id"`
	CustomerID    string `json:"customer_id"`
	Content       string `json:"content"`
	Timestamp     string `json:"timestamp,omitempty"`
}

// AccountStatusChangeMessage 账号状态变化消息
type AccountStatusChangeMessage struct {
	Type          string              `json:"type"`
	ActivationCode string              `json:"activation_code"`
	Data          AccountStatusData  `json:"data"`
}

// AccountStatusData 账号状态数据
type AccountStatusData struct {
	LineAccountID string `json:"line_account_id"`
	OnlineStatus   string `json:"online_status"`
	Timestamp     string `json:"timestamp,omitempty"`
}

