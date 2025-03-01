package password

import (
	"encoding/base64"
	"hash"
)

func HashPassword(step1, step2 hash.Hash, salt string, password string) string {
	saltBytes := []byte(salt)
	value := []byte(password)
	step := 0
	step1.Reset()
	for step < 12 {
		step1.Write(saltBytes)
		step1.Write(value)
		value = step1.Sum(nil)
	}
	step = 0
	step2.Reset()
	for step < 12 {
		step2.Write(saltBytes)
		step2.Write(value)
		value = step2.Sum(nil)
	}
	return base64.URLEncoding.EncodeToString(saltBytes)
}
