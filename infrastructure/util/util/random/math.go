package random

import (
	"math/rand"
	"time"
)

// 使用私有的随机生成器
var irand = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower  = "abcdefghijklmnopqrstuvwxyz"
	number = "0123456789"
	char62 = upper + lower + number
)

func Uppers(n int) string {
	return Chars(upper, n)
}

func Lowers(n int) string {
	return Chars(lower, n)
}

func Numbers(n int) string {
	return Chars(number, n)
}

func Strings(n int) string {
	return Chars(char62, n)
}

func Chars(chars string, n int) string {
	length := len(chars)
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = chars[irand.Intn(length)]
	}
	return string(buf)
}

// 生成[min,max]的随机数，可包含负数
func RangeNum(min, max int) int {
	if min > max {
		min, max = max, min
	}
	num := irand.Intn(max - min + 1)
	return num + min
}

// 打乱一个数组
func Shuffle[T any](sli []T) {
	irand.Shuffle(len(sli), func(i, j int) {
		sli[i], sli[j] = sli[j], sli[i]
	})
}
