package investing

import (
	"math/rand"
	"time"
)

const charsLength = 63

var (
	chars = []byte("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_")
	r     = rand.New(rand.NewSource(time.Now().UnixNano()))

	//ips = make([]int, 256)
)

func randomString(l int) string {
	ret := make([]byte, l)
	for i := 0; i < l; i++ {
		ret[i] = chars[r.Intn(charsLength)]
	}
	return string(ret)
}
