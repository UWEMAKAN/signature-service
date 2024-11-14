package utils

import (
	"strings"
	"math/rand"
)

func RandomString(n int) string {
	var sb strings.Builder
	k := len(Alphabets)

	for i := 0; i < n; i++ {
		c := Alphabets[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}
