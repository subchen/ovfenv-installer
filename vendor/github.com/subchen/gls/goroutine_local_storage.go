// Package gls implements goroutine-local storage.
package gls

import "sync"

type (
	storage map[string]interface{}
)

var (
	// gls is a map of goroutine IDs that stores the key-value pairs
	gls map[uint64]storage
	// mutex protects access to the gls map
	mutex sync.RWMutex
)

func init() {
	gls = make(map[uint64]storage, 64)
}

// Get gets the value by key as it exists for the current goroutine.
func Get(key string) interface{} {
	gid := GoroutineID()
	mutex.RLock()
	if s := gls[gid]; s == nil {
		mutex.RUnlock()
		return nil
	} else {
		value := s[key]
		mutex.RUnlock()
		return value
	}
}

// GetOK gets the value by key as it exists for the current goroutine.
func GetOK(key string) (value interface{}, ok bool) {
	gid := GoroutineID()
	mutex.RLock()
	if s := gls[gid]; s == nil {
		mutex.RUnlock()
		return nil, false
	} else {
		value, ok = s[key]
		mutex.RUnlock()
		return
	}
}

// Set sets the value by key and associates it with the current goroutine.
func Set(key string, value interface{}) {
	gid := GoroutineID()
	mutex.Lock()
	s := gls[gid]
	if s == nil {
		s = empty()
		gls[gid] = s
	}
	s[key] = value
	mutex.Unlock()
}

// Copy sets multiple values and associates it with the current goroutine.
func Copy(src map[string]interface{}) {
	gid := GoroutineID()
	mutex.Lock()
	s := gls[gid]
	if s == nil {
		s = empty()
		gls[gid] = s
	}
	cp(s, storage(src))
	mutex.Unlock()
}

// Remove removes key from the CopyOnWriteMap.
func Remove(key string) {
	gid := GoroutineID()
	mutex.Lock()
	if d := gls[gid]; d != nil {
		delete(d, key)
	}
	mutex.Unlock()
}

// Clear removes all gls associated with this goroutine. If this is not
// called, the gls may persist for the lifetime of your application. This
// must be called from the very first goroutine to invoke Set
func Clear() {
	gid := GoroutineID()
	mutex.Lock()
	delete(gls, gid)
	mutex.Unlock()
}

func Keys() []string {
	gid := GoroutineID()
	keys := make([]string, len(gls[gid]))
	mutex.RLock()
	for k, _ := range gls[gid] {
		keys = append(keys, k)
	}
	mutex.RUnlock()
	return keys
}

func Values() []interface{} {
	gid := GoroutineID()
	values := make([]interface{}, len(gls[gid]))
	mutex.RLock()
	for _, v := range gls[gid] {
		values = append(values, v)
	}
	mutex.RUnlock()
	return values
}

func RawMap() map[string]interface{} {
	gid := GoroutineID()
	mutex.RLock()
	dst := dup(gls[gid])
	mutex.RUnlock()
	return dst
}

func empty() storage {
	return make(storage)
}

func cp(dst storage, src storage) {
	for k, v := range src {
		dst[k] = v
	}
}

func dup(src storage) storage {
	dst := make(storage, len(src))
	cp(dst, src)
	return dst
}
