package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证太频繁")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 redis 出了问题
		return err
	}
	switch res {
	case -2:
		return errors.New("验证码存在，但是没有过期时间")
	case -1:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 redis 出了问题
		return false, err
	}
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return true, nil
	}
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

// LocalCodeCache 本地缓存实现，直接放内存
type LocalCodeCache struct {
	data map[string]cacheEntry
	lock sync.RWMutex
}

type cacheEntry struct {
	code     string
	expireAt time.Time
	attempts int
}

const maxAttempts = 3
const codeTTL = 5 * time.Minute

func NewLocalCodeCache() *LocalCodeCache {
	return &LocalCodeCache{
		data: make(map[string]cacheEntry),
	}
}

func (c *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := c.key(biz, phone)
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, exists := c.data[key]
	if exists && time.Now().Before(entry.expireAt) {
		entry.attempts++
		if entry.attempts > maxAttempts {
			return ErrCodeSendTooMany
		}
	} else {
		entry = cacheEntry{
			code:     code,
			expireAt: time.Now().Add(codeTTL),
		}
	}
	c.data[key] = entry
	return nil
}

func (c *LocalCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	key := c.key(biz, phone)
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expireAt) {
		return false, nil
	}

	if entry.code == code {
		delete(c.data, key)
		return true, nil
	}

	entry.attempts++
	if entry.attempts > maxAttempts {
		delete(c.data, key)
		return false, ErrCodeVerifyTooMany
	}

	c.data[key] = entry
	return false, nil
}

func (c *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
