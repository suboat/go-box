package box

import (
	"github.com/shopspring/decimal"
)

var (
	CfgAutoRound       = true // true: 自动约
	CfgRound     int32 = 4    // 默认保留小数位
)

// 浮点数加
func PubFloatAdd(a float64, b float64, vals ...float64) (c float64) {
	n := decimal.NewFromFloat(a)
	n = n.Add(decimal.NewFromFloat(b))
	for _, v := range vals {
		n = n.Add(decimal.NewFromFloat(v))
	}
	c, _ = n.Float64()
	if CfgAutoRound {
		c = PubFloatRoundAuto(c)
	}
	return
}

// 浮点数加,并保留有效位数
func PubFloatAddRound(a float64, b float64, vals ...float64) (c float64) {
	c = PubFloatRoundAuto(PubFloatAdd(a, b, vals...))
	return
}

// 浮点数乘
func PubFloatMul(a float64, b float64) (c float64) {
	n := decimal.NewFromFloat(a)
	n = n.Mul(decimal.NewFromFloat(b))
	c, _ = n.Float64()
	return
}

// 浮点数乘,并保留有效位数
func PubFloatMulRound(a float64, b float64) (c float64) {
	c = PubFloatRoundAuto(PubFloatMul(a, b))
	return
}

// 浮点数除
func PubFloatDiv(a float64, b float64) (c float64) {
	n := decimal.NewFromFloat(a)
	n = n.Div(decimal.NewFromFloat(b))
	c, _ = n.Float64()
	return
}

// 浮点数除,并保留有效位数
func PubFloatDivRound(a float64, b float64) (c float64) {
	c = PubFloatRoundAuto(PubFloatDiv(a, b))
	return
}

// 浮点数保留位数
func PubFloatRound(a float64, r int32) (c float64) {
	// 四舍五入
	//n := decimal.NewFromFloat(a).Round(r)
	//c, _ = n.Float64()
	c = PubIntRoundToFloat(PubFloatRoundToInt(a, r), r)
	return
}

// 浮点数截尾
func PubFloatToInt(a float64, r int32) (c int64) {
	return decimal.New(1, r).Mul(decimal.NewFromFloat(a)).IntPart()
}
func PubIntToFloat(a int64, r int32) (c float64) {
	c, _ = decimal.New(a, 0).Div(decimal.New(1, r)).Float64()
	return
}

// 浮点数进位
func PubFloatRoundToInt(a float64, r int32) (c int64) {
	// 四舍六入五留双
	i0 := decimal.New(1, r).Mul(decimal.NewFromFloat(a)).IntPart()
	i1 := decimal.New(1, r+1).Mul(decimal.NewFromFloat(a)).IntPart() - i0*10
	if i1 == 5 {
		i2 := i0 - decimal.New(1, r-1).Mul(decimal.NewFromFloat(a)).IntPart()*10
		if i2%2 == 0 {
			c = decimal.New(i0+1, 0).IntPart()
		} else {
			c = decimal.New(i0, 0).IntPart()
		}
	} else if i1 > 5 {
		c = decimal.New(i0+1, 0).IntPart()
	} else {
		c = decimal.New(i0, 0).IntPart()
	}
	return
}

func PubIntRoundToFloat(a int64, r int32) (c float64) {
	c, _ = decimal.New(a, 0).Div(decimal.New(1, r)).Float64()
	return
}

// 浮点数保留系统默认位数
func PubFloatRoundAuto(a float64) (c float64) {
	return PubFloatRound(a, CfgRound)
}
