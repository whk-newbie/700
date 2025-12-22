package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"line-management/pkg/logger"
	redisPkg "line-management/pkg/redis"

	"github.com/go-redis/redis/v8"
)

// SessionInfo 会话信息
type SessionInfo struct {
	UserID         uint      `json:"user_id"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	GroupID        uint      `json:"group_id,omitempty"`
	ActivationCode string    `json:"activation_code,omitempty"`
	LoginTime      time.Time `json:"login_time"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
}

// SessionService Session管理服务
type SessionService struct {
	rdb *redis.Client
}

// NewSessionService 创建Session服务实例
func NewSessionService() *SessionService {
	return &SessionService{
		rdb: redisPkg.GetClient(),
	}
}

// generateSessionKey 生成Session Key
func (s *SessionService) generateSessionKey(userID uint, token string) string {
	// 使用Token的哈希值作为Session标识
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:16]) // 取前16字节
	return fmt.Sprintf("session:%d:%s", userID, tokenHash)
}

// generateUserSessionsKey 生成用户所有Session的Key
func (s *SessionService) generateUserSessionsKey(userID uint) string {
	return fmt.Sprintf("user_sessions:%d", userID)
}

// CreateSession 创建Session
func (s *SessionService) CreateSession(userID uint, token string, info *SessionInfo, expireTime time.Duration) error {
	ctx := context.Background()
	
	// 生成Session Key
	sessionKey := s.generateSessionKey(userID, token)
	userSessionsKey := s.generateUserSessionsKey(userID)

	// 序列化Session信息
	sessionData, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化Session信息失败: %w", err)
	}

	// 存储Session信息
	if err := s.rdb.Set(ctx, sessionKey, sessionData, expireTime).Err(); err != nil {
		return fmt.Errorf("存储Session失败: %w", err)
	}

	// 将Session Key添加到用户的Session集合中（用于管理用户的所有Session）
	if err := s.rdb.SAdd(ctx, userSessionsKey, sessionKey).Err(); err != nil {
		logger.Warnf("添加Session到用户集合失败: %v", err)
	}
	
	// 设置用户Session集合的过期时间（比Session稍长一些）
	s.rdb.Expire(ctx, userSessionsKey, expireTime+time.Hour)

	logger.Infof("创建Session成功: user_id=%d, session_key=%s", userID, sessionKey)
	return nil
}

// GetSession 获取Session信息
func (s *SessionService) GetSession(userID uint, token string) (*SessionInfo, error) {
	ctx := context.Background()
	sessionKey := s.generateSessionKey(userID, token)

	// 从Redis获取Session数据
	data, err := s.rdb.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("Session不存在或已过期")
		}
		return nil, fmt.Errorf("获取Session失败: %w", err)
	}

	// 反序列化Session信息
	var info SessionInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, fmt.Errorf("反序列化Session信息失败: %w", err)
	}

	return &info, nil
}

// DeleteSession 删除Session
func (s *SessionService) DeleteSession(userID uint, token string) error {
	ctx := context.Background()
	sessionKey := s.generateSessionKey(userID, token)
	userSessionsKey := s.generateUserSessionsKey(userID)

	// 删除Session
	if err := s.rdb.Del(ctx, sessionKey).Err(); err != nil {
		return fmt.Errorf("删除Session失败: %w", err)
	}

	// 从用户Session集合中移除
	s.rdb.SRem(ctx, userSessionsKey, sessionKey)

	logger.Infof("删除Session成功: user_id=%d, session_key=%s", userID, sessionKey)
	return nil
}

// DeleteAllUserSessions 删除用户的所有Session（用于强制下线）
func (s *SessionService) DeleteAllUserSessions(userID uint) error {
	ctx := context.Background()
	userSessionsKey := s.generateUserSessionsKey(userID)

	// 获取用户的所有Session Key
	sessionKeys, err := s.rdb.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("获取用户Session列表失败: %w", err)
	}

	// 删除所有Session
	if len(sessionKeys) > 0 {
		if err := s.rdb.Del(ctx, sessionKeys...).Err(); err != nil {
			return fmt.Errorf("删除Session失败: %w", err)
		}
	}

	// 删除用户Session集合
	s.rdb.Del(ctx, userSessionsKey)

	logger.Infof("删除用户所有Session成功: user_id=%d, count=%d", userID, len(sessionKeys))
	return nil
}

// GetUserSessions 获取用户的所有活跃Session
func (s *SessionService) GetUserSessions(userID uint) ([]*SessionInfo, error) {
	ctx := context.Background()
	userSessionsKey := s.generateUserSessionsKey(userID)

	// 获取用户的所有Session Key
	sessionKeys, err := s.rdb.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取用户Session列表失败: %w", err)
	}

	var sessions []*SessionInfo
	for _, sessionKey := range sessionKeys {
		// 检查Session是否存在
		data, err := s.rdb.Get(ctx, sessionKey).Result()
		if err != nil {
			if err == redis.Nil {
				// Session已过期，从集合中移除
				s.rdb.SRem(ctx, userSessionsKey, sessionKey)
				continue
			}
			continue
		}

		// 反序列化Session信息
		var info SessionInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			continue
		}

		sessions = append(sessions, &info)
	}

	return sessions, nil
}

// CheckSession 检查Session是否存在
func (s *SessionService) CheckSession(userID uint, token string) bool {
	ctx := context.Background()
	sessionKey := s.generateSessionKey(userID, token)
	
	exists, err := s.rdb.Exists(ctx, sessionKey).Result()
	if err != nil {
		logger.Warnf("检查Session失败: %v", err)
		return false
	}
	
	return exists > 0
}

