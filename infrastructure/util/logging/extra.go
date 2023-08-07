package logging

import (
	"bytes"
	"context"
	"encoding/base64"
	"sort"
	"strconv"
	"strings"
)

const XRequestID = "X-Request-Id"

type __ctx__ struct {
	m map[string]string
	context.Context
}

func (c *__ctx__) Value(key any) any {
	ks, _ := key.(string)
	if v, ok := c.m[ks]; ok {
		return v
	}
	return c.Context.Value(key)
}

func (l *Logger) GetContext() context.Context {
	return &__ctx__{
		map[string]string{
			XRequestID: l.RequestID,
			"v1":       l.V1,
			"v2":       l.V2,
			"v3":       l.V3,
		},
		context.Background(),
	}
}

func NewContext(reqid, v1, v2, v3 string) context.Context {
	return &__ctx__{
		map[string]string{
			XRequestID: reqid,
			"v1":       v1,
			"v2":       v2,
			"v3":       v3,
		},
		context.Background(),
	}
}

func FromContext(ctx context.Context) *Logger {
	l := &Logger{}
	l.RequestID, _ = ctx.Value(XRequestID).(string)
	l.V1, _ = ctx.Value("v1").(string)
	l.V2, _ = ctx.Value("v2").(string)
	l.V3, _ = ctx.Value("v3").(string)
	return l
}

// SpreadMaps 将url.Values或http.Header值的数组展开为字符串
func SpreadMaps(m map[string][]string) map[string]any {
	res := make(map[string]any, len(m))
	for k, v := range m {
		if len(v) == 1 {
			res[k] = v[0]
		} else {
			res[k] = v
		}
	}
	return res
}

// Compress 超过2048字节返回截断中间的内容
func Compress(b []byte) string {
	if l := len(b); l > 2048 {
		buf := bytes.NewBuffer(nil)
		buf.Grow(2048)
		buf.Write(b[:1000])
		buf.WriteString("***省略{")
		buf.WriteString(strconv.Itoa(l - 2000))
		buf.WriteString("}字符***")
		buf.Write(b[l-1000:])
		return buf.String()
	}
	return string(b)
}

type field struct{ left, right int }

func matchReplace(src []byte) []byte {
	match := make(map[field]struct{})
	for _, reg := range secret {
		allIndex := reg.FindAllIndex(src, -1)
		count := len(allIndex)
		if count == 0 {
			continue
		}
		for _, i := range allIndex {
			part := string(src[i[0]:i[1]])
			arr := sep.Split(part, -1)
			txtLen := len(arr[3])
			if txtLen == 0 {
				continue
			}
			l := strings.LastIndex(part, arr[3]) + i[0]
			r := l + txtLen
			match[field{left: l, right: r}] = struct{}{}
		}
	}
	count := len(match)
	if count == 0 {
		return src
	}
	items := make([]field, 0, count)
	for k := range match {
		items = append(items, k)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].left < items[j].left
	})
	buf := bytes.NewBuffer(nil)
	pre := 0
	for _, f := range items {
		buf.Write(src[pre:f.left])
		buf.WriteString(ecbEncrypt(src[f.left:f.right]))
		pre = f.right
	}
	buf.Write(src[pre:])
	return buf.Bytes()
}

func ecbEncrypt(src []byte) string {
	bs := block.BlockSize()
	length := len(src)
	pad := bs - length%bs
	padText := bytes.Repeat([]byte{byte(pad)}, pad)
	plain := make([]byte, length, length+pad)
	copy(plain, src) //复制一份防止当src仍有cap的时候append修改其后的内容
	plain = append(plain, padText...)
	length = len(plain)
	dst := make([]byte, length)
	for i := 0; i < length; i += bs {
		block.Encrypt(dst[i:i+bs], plain[i:i+bs])
	}
	return base64.StdEncoding.EncodeToString(dst)
}
