package memory

import (
	"github.com/lucperkins/strato/internal/oops"
	"time"

	"github.com/lucperkins/strato/internal/services/cache"
	"github.com/lucperkins/strato/internal/services/counter"
	"github.com/lucperkins/strato/internal/services/kv"
	"github.com/lucperkins/strato/internal/services/set"

	"github.com/lucperkins/strato"
)

type Memory struct {
	cache    map[string]*cache.Item
	counters map[string]int64
	kv       map[string]*kv.Value
	sets     map[string][]string
}

var (
	_ cache.Cache     = (*Memory)(nil)
	_ counter.Counter = (*Memory)(nil)
	_ kv.KV           = (*Memory)(nil)
	_ set.Set         = (*Memory)(nil)
)

func NewMemoryBackend() *Memory {
	cacheMem := make(map[string]*cache.Item)

	counterMem := make(map[string]int64)

	setMem := make(map[string][]string)

	kvMem := make(map[string]*kv.Value)

	return &Memory{
		cache:    cacheMem,
		counters: counterMem,
		kv:       kvMem,
		sets:     setMem,
	}
}

// Interface methods
func (m *Memory) Close() error {
	return nil
}

func (m *Memory) Flush() error {
	return nil
}

// Cache
func (m *Memory) CacheGet(key string) (string, error) {
	val, ok := m.cache[key]

	if !ok {
		return "", oops.NotFound(key)
	}

	now := time.Now().Unix()

	expired := (now - val.Timestamp) > int64(val.TTLSeconds)

	if expired {
		delete(m.cache, key)

		return "", oops.NotFound(key)
	}

	return val.Value, nil
}

func (m *Memory) CacheSet(key, value string, ttl int32) error {
	if key == "" {
		return oops.ErrNoKey
	}

	if value == "" {
		return oops.ErrNoValue
	}

	item := &cache.Item{
		Value:      value,
		Timestamp:  time.Now().Unix(),
		TTLSeconds: parseTtl(ttl),
	}

	m.cache[key] = item

	return nil
}

func parseTtl(ttl int32) int32 {
	if ttl == 0 {
		return cache.DefaultTtl
	} else {
		return ttl
	}
}

// Counter
func (m *Memory) CounterIncrement(key string, increment int64) error {
	count, ok := m.counters[key]
	if !ok {
		m.counters[key] = increment
	} else {
		m.counters[key] = count + increment
	}

	return nil
}

func (m *Memory) CounterGet(key string) (int64, error) {
	return m.counters[key], nil
}

func (m *Memory) KVGet(key string) (*kv.Value, error) {
	val, ok := m.kv[key]
	if !ok {
		return nil, strato.NotFound(key)
	}

	return val, nil
}

func (m *Memory) KVPut(key string, value *kv.Value) error {
	m.kv[key] = value
	return nil
}

func (m *Memory) KVDelete(key string) error {
	delete(m.kv, key)
	return nil
}

func (m *Memory) GetSet(set string) ([]string, error) {
	s, ok := m.sets[set]

	if !ok {
		return []string{}, nil
	}

	return s, nil
}

func (m *Memory) AddToSet(set, item string) ([]string, error) {
	if _, ok := m.sets[set]; ok {
		m.sets[set] = append(m.sets[set], item)
	} else {
		m.sets[set] = []string{item}
	}

	return m.sets[set], nil
}

func (m *Memory) RemoveFromSet(set, item string) ([]string, error) {
	_, ok := m.sets[set]
	if ok {
		for idx, it := range m.sets[set] {
			if it == item {
				m.sets[set] = append(m.sets[set][:idx], m.sets[set][idx+1:]...)
			}
		}

		return m.sets[set], nil
	} else {
		return []string{}, nil
	}
}
