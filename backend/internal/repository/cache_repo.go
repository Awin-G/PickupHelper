package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ActiveCode represents an unexpired, unused SMS verification code in Redis.
type ActiveCode struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	ExpireIn int64  `json:"expire_in"` // remaining TTL in seconds
}

// SMSCodeCache abstracts the per-phone SMS verification code store backed
// by Redis, plus the per-phone / per-IP rate limit counters. Keys:
//   - sms:code:{phone}        → 6-digit code, TTL 300s
//   - sms:rate:phone:{phone}  → request counter, TTL 60s (max 1)
//   - sms:rate:ip:{ip}        → request counter, TTL 600s (max 10)
//
// The rate-limit methods use INCR + conditional EXPIRE so the very first
// request in a window seeds the counter and sets the TTL; subsequent
// requests within the window just increment.
type SMSCodeCache interface {
	SetCode(ctx context.Context, phone, code string, ttl time.Duration) error
	GetCode(ctx context.Context, phone string) (string, error)
	DelCode(ctx context.Context, phone string) error
	// CheckAndIncrPhoneRate increments the per-phone rate counter and
	// returns the new value. Caller treats count > 1 as "60s already sent".
	CheckAndIncrPhoneRate(ctx context.Context, phone string) (int, error)
	// CheckAndIncrIPRate increments the per-IP rate counter and returns
	// the new value. Caller treats count > 10 as "10min IP over limit".
	CheckAndIncrIPRate(ctx context.Context, ip string) (int, error)
	// ListCodes returns all unexpired SMS codes stored in Redis.
	// Expired keys (TTL <= 0) are filtered out. Returns an empty slice
	// (not nil) when no codes exist.
	ListCodes(ctx context.Context) ([]ActiveCode, error)
}

// redisSMSCodeCache implements SMSCodeCache against *redis.Client.
type redisSMSCodeCache struct {
	rdb *redis.Client
}

// NewSMSCodeCache returns a Redis-backed SMSCodeCache.
func NewSMSCodeCache(rdb *redis.Client) SMSCodeCache {
	return &redisSMSCodeCache{rdb: rdb}
}

const (
	smsCodeKeyPrefix     = "sms:code:"
	smsPhoneRateKeyPrefix = "sms:rate:phone:"
	smsIPRateKeyPrefix   = "sms:rate:ip:"

	smsPhoneRateTTL = 60 * time.Second
	smsIPRateTTL    = 600 * time.Second
)

func (c *redisSMSCodeCache) SetCode(ctx context.Context, phone, code string, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, smsCodeKeyPrefix+phone, code, ttl).Err(); err != nil {
		return fmt.Errorf("sms_cache.SetCode: %w", err)
	}
	return nil
}

func (c *redisSMSCodeCache) GetCode(ctx context.Context, phone string) (string, error) {
	val, err := c.rdb.Get(ctx, smsCodeKeyPrefix+phone).Result()
	if err == redis.Nil {
		return "", redis.Nil
	}
	if err != nil {
		return "", fmt.Errorf("sms_cache.GetCode: %w", err)
	}
	return val, nil
}

func (c *redisSMSCodeCache) DelCode(ctx context.Context, phone string) error {
	if err := c.rdb.Del(ctx, smsCodeKeyPrefix+phone).Err(); err != nil {
		return fmt.Errorf("sms_cache.DelCode: %w", err)
	}
	return nil
}

// incrRate performs INCR; on the first increment (count==1) it also sets
// the TTL so the counter expires at the end of the window. Returns the
// post-increment count.
func (c *redisSMSCodeCache) incrRate(ctx context.Context, key string, ttl time.Duration) (int, error) {
	cnt, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("sms_cache incr %s: %w", key, err)
	}
	if cnt == 1 {
		if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, fmt.Errorf("sms_cache expire %s: %w", key, err)
		}
	}
	return int(cnt), nil
}

func (c *redisSMSCodeCache) CheckAndIncrPhoneRate(ctx context.Context, phone string) (int, error) {
	return c.incrRate(ctx, smsPhoneRateKeyPrefix+phone, smsPhoneRateTTL)
}

func (c *redisSMSCodeCache) CheckAndIncrIPRate(ctx context.Context, ip string) (int, error) {
	return c.incrRate(ctx, smsIPRateKeyPrefix+ip, smsIPRateTTL)
}

// ListCodes scans all sms:code:* keys and returns every active (unexpired)
// verification code along with its remaining TTL.
func (c *redisSMSCodeCache) ListCodes(ctx context.Context) ([]ActiveCode, error) {
	keys, err := c.rdb.Keys(ctx, smsCodeKeyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("sms_cache.ListCodes keys: %w", err)
	}
	if len(keys) == 0 {
		return []ActiveCode{}, nil
	}

	out := make([]ActiveCode, 0, len(keys))
	for _, key := range keys {
		phone := strings.TrimPrefix(key, smsCodeKeyPrefix)
		val, err := c.rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			continue // expired between KEYS and GET
		}
		if err != nil {
			return nil, fmt.Errorf("sms_cache.ListCodes get %s: %w", key, err)
		}
		ttl, err := c.rdb.TTL(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("sms_cache.ListCodes ttl %s: %w", key, err)
		}
		if ttl <= 0 {
			continue // expired, skip
		}
		out = append(out, ActiveCode{
			Phone:    phone,
			Code:     val,
			ExpireIn: int64(ttl.Seconds()),
		})
	}
	if out == nil {
		out = []ActiveCode{}
	}
	return out, nil
}
