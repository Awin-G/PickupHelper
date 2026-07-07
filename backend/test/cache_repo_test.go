//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	"pickup-helper/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSMSCache(env *TestEnv) repository.SMSCodeCache {
	return repository.NewSMSCodeCache(env.Rdb)
}

func TestSMSCodeCache_SetGetDel(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()
	cache := newSMSCache(env)

	t.Cleanup(func() { env.Rdb.Del(ctx, "sms:code:13800001111") })

	require.NoError(t, cache.SetCode(ctx, "13800001111", "123456", 60*time.Second))

	got, err := cache.GetCode(ctx, "13800001111")
	require.NoError(t, err)
	assert.Equal(t, "123456", got)

	require.NoError(t, cache.DelCode(ctx, "13800001111"))

	_, err = cache.GetCode(ctx, "13800001111")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestSMSCodeCache_GetCode_NotFound(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()
	cache := newSMSCache(env)

	_, err := cache.GetCode(ctx, "13900000000")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestSMSCodeCache_TTLExpires(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()
	cache := newSMSCache(env)

	t.Cleanup(func() { env.Rdb.Del(ctx, "sms:code:13800002222") })

	require.NoError(t, cache.SetCode(ctx, "13800002222", "999999", 1*time.Second))
	time.Sleep(1500 * time.Millisecond)

	_, err := cache.GetCode(ctx, "13800002222")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestSMSCodeCache_PhoneRate_FirstCallSeeds(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()
	cache := newSMSCache(env)
	key := "13800003333"

	t.Cleanup(func() { env.Rdb.Del(ctx, "sms:rate:phone:"+key) })

	cnt, err := cache.CheckAndIncrPhoneRate(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt, "first call seeds the counter")

	// Verify TTL was set on the key.
	ttl, err := env.Rdb.TTL(ctx, "sms:rate:phone:"+key).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0), "TTL should be set on first call")
	assert.LessOrEqual(t, ttl, 60*time.Second)

	cnt2, err := cache.CheckAndIncrPhoneRate(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, 2, cnt2, "second call increments without resetting TTL")
}

func TestSMSCodeCache_IPRate_FirstCallSeeds(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()
	cache := newSMSCache(env)
	ip := "10.0.0.1"

	t.Cleanup(func() { env.Rdb.Del(ctx, "sms:rate:ip:"+ip) })

	for i := 1; i <= 10; i++ {
		cnt, err := cache.CheckAndIncrIPRate(ctx, ip)
		require.NoError(t, err)
		assert.Equal(t, i, cnt)
	}
	cnt, err := cache.CheckAndIncrIPRate(ctx, ip)
	require.NoError(t, err)
	assert.Equal(t, 11, cnt, "11th call returns 11 — caller enforces max")

	// Verify TTL was set.
	ttl, err := env.Rdb.TTL(ctx, "sms:rate:ip:"+ip).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 600*time.Second)
}

func TestSMSCodeCache_PhoneRate_IsolatedPerPhone(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()
	cache := newSMSCache(env)
	t.Cleanup(func() {
		env.Rdb.Del(ctx, "sms:rate:phone:13800004444", "sms:rate:phone:13800005555")
	})

	c1, err := cache.CheckAndIncrPhoneRate(ctx, "13800004444")
	require.NoError(t, err)
	assert.Equal(t, 1, c1)

	c2, err := cache.CheckAndIncrPhoneRate(ctx, "13800005555")
	require.NoError(t, err)
	assert.Equal(t, 1, c2, "different phones have independent counters")
}
