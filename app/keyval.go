package main

import (
	"encoding/binary"
	"path/filepath"
	"sync"
	"time"
)

type KeyVal struct {
	val string
	px  int       // expiry in milliseconds
	t   time.Time // time when key was set
}

type MP struct {
	mp sync.Map
}

func (MP *MP) set(key string, val string, px int) {
	kv := &KeyVal{
		val: val,
		px:  px,
		t:   time.Now(),
	}
	MP.mp.Store(key, kv)
}

func (MP *MP) get(key string) (string, bool) {
	v, ok := MP.mp.Load(key)
	if !ok {
		return "", false
	}
	kv := v.(*KeyVal)
	// Check for expiry
	if kv.px > 0 {
		expireAt := kv.t.Add(time.Duration(kv.px) * time.Millisecond)
		if time.Now().After(expireAt) {
			MP.mp.Delete(key)
			return "", false
		}
	}
	return kv.val, true
}

func (rdb *RDB) LoadMap(mp *MP) {
	now := time.Now()

	for _, db := range rdb.DBs {
		// Load all entries from this database
		for _, entry := range db.Entries {
			key := string(entry.Key)
			value := string(entry.Value)

			// Calculate expiry time in milliseconds from now
			var expiryMs int = 0
			var setTime time.Time = now

			if entry.Expire != nil {
				var expireTimestamp int64

				if entry.Expire.Type == 0xFC {
					// Milliseconds timestamp (8 bytes, little-endian)
					expireTimestamp = int64(binary.LittleEndian.Uint64(entry.Expire.Time))
				} else if entry.Expire.Type == 0xFD {
					// Seconds timestamp (4 bytes, little-endian)
					expireTimestamp = int64(binary.LittleEndian.Uint32(entry.Expire.Time)) * 1000
				}

				// Convert Unix timestamp to time.Time
				expireTime := time.Unix(expireTimestamp/1000, (expireTimestamp%1000)*1000000)

				if expireTime.Before(now) {
					// Skip expired keys
					continue
				}

				// Calculate milliseconds until expiry from now
				expiryMs = int(expireTime.Sub(now).Milliseconds())

				setTime = now
			}

			// Create KeyVal and store in map
			kv := &KeyVal{
				val: value,
				px:  expiryMs,
				t:   setTime,
			}
			mp.mp.Store(key, kv)
		}
	}
}

func LoadRDBToMP(rdbData []byte) (*MP, error) {
	rdb, err := ParseRDB(rdbData)
	if err != nil {
		return nil, err
	}

	mp := &MP{}

	rdb.LoadMap(mp)

	return mp, nil
}

func (MP *MP) exists(key string) bool {
	_, ok := MP.get(key)
	return ok
}

func (MP *MP) delete(key string) bool {
	_, ok := MP.mp.LoadAndDelete(key)
	return ok
}

func (MP *MP) keys(pattern string) []string {
	var keys []string
	MP.mp.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		match, _ := filepath.Match(pattern, keyStr)
		if match {
			if MP.exists(keyStr) {
				keys = append(keys, keyStr)
			}
		}
		return true
	})
	return keys
}

func (MP *MP) count() int {
	count := 0
	MP.mp.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if MP.exists(keyStr) {
			count++
		}
		return true
	})
	return count
}
