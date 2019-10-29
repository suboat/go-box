package box

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"
)

const (
	randCharset    = "abcdefghijklmnopqrstuvwxyz0123456789" // "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randCharsetNum = "0123456789"                           //
)

//
func PubJsonMust(v_ interface{}) (r string) {
	r = "{}"
	if v_ == nil {
		return
	}
	switch v := v_.(type) {
	case string:
		if len(v) > 2 && v[0] == '{' && v[len(v)-1] == '}' {
			r = v
		}
		break
	case *string:
		_v := *v
		if len(_v) > 2 && _v[0] == '{' && _v[len(_v)-1] == '}' {
			r = _v
		}
		break
	default:
		// null: v_不为nil但其实是空
		if r1, _err := json.Marshal(v_); _err == nil && string(r1) != "null" {
			r = string(r1)
		}
		break
	}
	return
}

//
func PubGetBoolPoint(v bool) *bool {
	return &v
}

//
func PubGetStringPoint(v string) *string {
	return &v
}

//
func PubGetFloatPoint(v float64) *float64 {
	return &v
}

//
func PubGetIntPoint(v int) *int {
	return &v
}

//
func PubGetInt32Point(v int32) *int32 {
	return &v
}

// 取环境变量或默认值
func PubGetEnv(key_, default_ string) (ret string) {
	ret = os.Getenv(key_)
	if len(ret) == 0 {
		ret = default_
	}
	return
}

//
func PubGetEnvBool(key_ string, default_ bool) (ret bool) {
	val := strings.ToLower(os.Getenv(key_))
	if val == "true" {
		ret = true
	} else if val == "false" {
		ret = false
	} else {
		ret = default_
	}
	return
}

// 将Uid缩减为32位
func PubUidShort32(org string) (s string) {
	s = strings.Replace(org, "-", "", -1)
	return
}

// 将uid由32位变回36位
func PubUidBack36(s string) (org string) {
	if len(s) == 32 {
		org = fmt.Sprintf("%s-%s-%s-%s-%s", s[0:8], s[8:12], s[12:16], s[16:20], s[20:32])
	} else if len(s) == 16 {
		org = fmt.Sprintf("%s-%s", s[0:4], s[4:16])
	} else if len(s) == 20 {
		org = fmt.Sprintf("%s-%s-%s", s[0:4], s[4:8], s[8:20])
	} else if len(s) == 24 {
		org = fmt.Sprintf("%s-%s-%s-%s", s[0:4], s[4:8], s[8:12], s[12:24])
	} else {
		org = s
	}
	return
}

// 将时间格式转为前端友好的string
func PubTimeToStr(t time.Time) (s string) {
	return t.Local().Format(time.RFC3339Nano) // 使用服务器时区时间
}

// 将时间格式转为前端友好的string
func PubStrToTime(data string) (r time.Time, err error) {
	if r, err = time.Parse(time.RFC3339Nano, data); err != nil {
		return
	}
	r = r.Local() // 使用服务器时区时间
	return
}

// uuid转slug
func PubUuidToSlug(id string) string {
	return base64.RawURLEncoding.EncodeToString(PubUuidToByte(id))
}

// slug转uuid
func PubSlugToUuid(in string) string {
	id32, _ := base64.RawURLEncoding.DecodeString(in)
	s := fmt.Sprintf("%x", id32)
	l := len(s)
	if l < 32 {
		for i := 0; i < 32-l; i++ {
			s = "0" + s
		}
	}
	return PubUidBack36(s)
}

// 十进制转uuid
func PubDecimalToUuid(s string) (ret string) {
	l := len(s)
	if l > 32 {
		s = s[0:32]
	} else if l < 32 {
		for i := 0; i < 32-l; i++ {
			s = "0" + s
		}
	}
	return PubUidBack36(s)
}

// uuid转十进制
func PubUuidToDecimal(s string) (ret string) {
	var (
		idx = -1
	)
	s = PubUidShort32(s)
	for i, v := range s {
		if v > '0' && v <= '9' {
			idx = i
			break
		}
	}
	if idx > -1 {
		ret = s[idx:]
	} else {
		ret = s
	}
	return
}

// uuid转字符串
func PubUuidToStr(id string) (ret string) {
	return string(PubUuidToByte(id))
}

// 字符串转uuid
func PubStrToUuid(s string) (ret string) {
	// 转小写
	s = strings.ToLower(s)
	return PubByteToUuid([]byte(s))
}

// 二进制转uuid
func PubByteToUuid(s []byte) (ret string) {
	strHex := fmt.Sprintf("%x", s)
	if len(strHex) > 32 {
		strHex = strHex[0:32]
	} else {
		//strHex = strHex + strings.Repeat("0", 32-len(strHex))
		strHex = strings.Repeat("0", 32-len(strHex)) + strHex
	}
	ret = PubUidBack36(strHex)
	return
}

// uuid转二进制: 会去除头部的零值
func PubUuidToByte(id string) []byte {
	idx := 0
	id = PubUidShort32(id)
	for i := 0; i < len(id); i++ {
		if id[i] != '0' {
			idx = i
			break
		}
	}
	if idx%2 == 1 {
		// 奇数
		idx--
	}
	b, _ := hex.DecodeString(id[idx:])
	return b
}

// 随机字符串
func PubRandomCode(length int) (ret string) {
	b := make([]byte, length)
	for i := range b {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(randCharset))))
		b[i] = randCharset[r.Int64()]
	}
	return string(b)
}

// 随机数字字符串
func PubRandomCodeNum(length int) (ret string) {
	b := make([]byte, length)
	for i := range b {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(randCharsetNum))))
		b[i] = randCharsetNum[r.Int64()]
	}
	return string(b)
}
