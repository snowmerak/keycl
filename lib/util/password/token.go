package password

import (
	"encoding/base64"
	"hash"
	"time"
)

func HashToken(hasher hash.Hash, email string, password string) string {
	buf := [8]byte{}
	now := time.Now().UnixNano()
	buf[0] = byte(now)
	buf[1] = byte(now >> 8)
	buf[2] = byte(now >> 16)
	buf[3] = byte(now >> 24)
	buf[4] = byte(now >> 32)
	buf[5] = byte(now >> 40)
	buf[6] = byte(now >> 48)
	buf[7] = byte(now >> 56)
	hasher.Reset()
	data := []byte(email + password)
	count := 0
	for count < 8 {
		hasher.Write(data)
		hasher.Write(buf[:])
		data = hasher.Sum(nil)
		count++
	}
	return base64.URLEncoding.EncodeToString(data)
}
