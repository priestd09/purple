package redis

import (
	"github.com/purpledb/purple/internal/data"
	"strconv"
	"time"

	"github.com/purpledb/purple/internal/services/flag"

	"github.com/purpledb/purple/internal/services/cache"
	"github.com/purpledb/purple/internal/services/counter"
	"github.com/purpledb/purple/internal/services/kv"
	"github.com/purpledb/purple/internal/services/set"

	"github.com/go-redis/redis"
	"github.com/purpledb/purple"
)

const defaultUrl = "localhost:6379"

type Redis struct {
	cache, counters, flags, kv, sets *redis.Client
}

func (r *Redis) Name() string {
	return "redis"
}

var (
	_ cache.Cache     = (*Redis)(nil)
	_ counter.Counter = (*Redis)(nil)
	_ flag.Flag       = (*Redis)(nil)
	_ kv.KV           = (*Redis)(nil)
	_ set.Set         = (*Redis)(nil)
)

func NewRedisBackend(addr string) (*Redis, error) {
	if addr == "" {
		addr = defaultUrl
	}

	cacheCl, err := newRedisClient(addr, 0)
	if err != nil {
		return nil, err
	}

	counterCl, err := newRedisClient(addr, 1)
	if err != nil {
		return nil, err
	}

	flagCl, err := newRedisClient(addr, 2)
	if err != nil {
		return nil, err
	}

	kvCl, err := newRedisClient(addr, 3)
	if err != nil {
		return nil, err
	}

	setCl, err := newRedisClient(addr, 4)
	if err != nil {
		return nil, err
	}

	return &Redis{
		cache:    cacheCl,
		counters: counterCl,
		flags:    flagCl,
		kv:       kvCl,
		sets:     setCl,
	}, nil
}

func newRedisClient(addr string, i int) (*redis.Client, error) {
	opts, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}

	opts.DB = i

	cl := redis.NewClient(opts)

	if err := cl.Ping().Err(); err != nil {
		return nil, err
	}

	return cl, nil
}

// Service methods
func (r *Redis) Close() error {
	for _, db := range []*redis.Client{
		r.cache,
	} {
		if err := db.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Redis) Flush() error {
	for _, db := range []*redis.Client{
		r.cache,
	} {
		if err := db.FlushAll().Err(); err != nil {
			return err
		}
	}

	return nil
}

// Cache operations
func (r *Redis) CacheGet(key string) (string, error) {
	s, err := r.cache.Get(key).Result()

	if err != nil {
		if err == redis.Nil {
			return "", purple.NotFound(key)
		} else {
			return "", err
		}
	}

	return s, nil
}

func (r *Redis) CacheSet(key, value string, ttl int32) error {
	t := time.Duration(ttl) * time.Second

	return r.cache.Set(key, value, t).Err()
}

// Counter operations
func (r *Redis) CounterGet(key string) (int64, error) {
	i, err := r.counters.Get(key).Int64()

	if err == redis.Nil {
		return 0, nil
	} else {
		return i, err
	}
}

func (r *Redis) CounterIncrement(key string, increment int64) (int64, error) {
	return r.counters.IncrBy(key, increment).Result()
}

// Flag operations
func (r *Redis) FlagGet(key string) (bool, error) {
	s := r.flags.Get(key).String()

	if s == "" {
		return false, nil
	}

	val, err := strconv.ParseBool(s)
	if err != nil {
		return false, nil
	}

	return val, nil
}

func (r *Redis) FlagSet(key string, value bool) error {
	val := strconv.FormatBool(value)

	return r.flags.Set(key, val, 0).Err()
}

// KV operations
func (r *Redis) KVGet(key string) (*kv.Value, error) {
	s, err := r.kv.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, purple.NotFound(key)
		} else {
			return nil, err
		}
	}

	return &kv.Value{
		Content: []byte(s),
	}, nil
}

func (r *Redis) KVPut(key string, value *kv.Value) error {
	return r.kv.Set(key, value.Content, 0).Err()
}

func (r *Redis) KVDelete(key string) error {
	return r.kv.Del(key).Err()
}

// Set operations
func (r *Redis) SetGet(set string) ([]string, error) {
	s, err := r.sets.SMembers(set).Result()
	if err != nil {
		return nil, err
	}

	return data.NonNilSet(s), nil
}

func (r *Redis) SetAdd(set, item string) ([]string, error) {
	if err := r.sets.SAdd(set, item).Err(); err != nil {
		return nil, err
	}

	return r.sets.SMembers(set).Result()
}

func (r *Redis) SetRemove(set, item string) ([]string, error) {
	if err := r.sets.SRem(set, item).Err(); err != nil {
		return nil, err
	}

	return r.sets.SMembers(set).Result()
}
