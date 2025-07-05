package main

import (
	"sync"
	"time"
)

type KeyVal struct {
	val string
	px  int       // expiry in milliseconds
	t   time.Time // time when key was set
}

type DB struct {
	mp sync.Map
}

func (db *DB) set(key string, val string, px int) {
	kv := &KeyVal{
		val: val,
		px:  px,
		t:   time.Now(),
	}
	db.mp.Store(key, kv)
}

func (db *DB) get(key string) (string, bool) {
	v, ok := db.mp.Load(key)
	if !ok {
		return "", false
	}

	kv := v.(*KeyVal)

	// Check for expiry
	if kv.px > 0 {
		expireAt := kv.t.Add(time.Duration(kv.px) * time.Millisecond)
		if time.Now().After(expireAt) {
			db.mp.Delete(key)
			return "", false
		}
	}

	return kv.val, true
}
