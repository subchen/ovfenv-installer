package gls

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

var (
	goroutinePrefix = []byte("goroutine ")

	bufPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 64)
			return &buf
		},
	}
)

// GoroutineID returns a goroutine ID
// Reference: http://blog.sgmansfield.com/2015/12/goroutine-ids/
func GoroutineID() uint64 {
	bp := bufPool.Get().(*[]byte)
	defer bufPool.Put(bp)
	b := *bp
	b = b[:runtime.Stack(b, false)]
	// Parse the 4707 out of "goroutine 4707 ["
	b = bytes.TrimPrefix(b, goroutinePrefix)
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		panic(fmt.Sprintf("No space found in %q", b))
	}
	b = b[:i]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse goroutine ID out of %q: %v", b, err))
	}
	return n
}
