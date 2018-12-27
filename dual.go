// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "math"

var (
	Half = Dual{Val: 0.5}
	One  = Dual{Val: 1.0}
)

type Dual struct {
	Val, Der float32
}

func Add(u, v Dual) Dual {
	return Dual{
		Val: u.Val + v.Val,
		Der: u.Der + v.Der,
	}
}

func Sub(u, v Dual) Dual {
	return Dual{
		Val: u.Val - v.Val,
		Der: u.Der - v.Der,
	}
}

func Mul(u, v Dual) Dual {
	return Dual{
		Val: u.Val * v.Val,
		Der: u.Der*v.Val + u.Val*v.Der,
	}
}

func Div(u, v Dual) Dual {
	return Dual{
		Val: u.Val / v.Val,
		Der: (u.Der*v.Val - u.Val*v.Der) / (v.Val * v.Val),
	}
}

func Sin(d Dual) Dual {
	return Dual{
		Val: float32(math.Sin(float64(d.Val))),
		Der: d.Der * float32(math.Cos(float64(d.Val))),
	}
}

func Cos(d Dual) Dual {
	return Dual{
		Val: float32(math.Cos(float64(d.Val))),
		Der: -d.Der * float32(math.Sin(float64(d.Val))),
	}
}

func Exp(d Dual) Dual {
	return Dual{
		Val: float32(math.Exp(float64(d.Val))),
		Der: d.Der * float32(math.Exp(float64(d.Val))),
	}
}

func Sigmoid(d Dual) Dual {
	e := Exp(d)
	return Div(e, Add(e, One))
}

func Log(d Dual) Dual {
	return Dual{
		Val: float32(math.Log(float64(d.Val))),
		Der: d.Der / d.Val,
	}
}

func Abs(d Dual) Dual {
	var sign float32
	val := float32(math.Abs(float64(d.Val)))
	if d.Val != 0.0 {
		sign = d.Val / val
	}
	return Dual{
		Val: val,
		Der: d.Der * sign,
	}
}

func Pow(d Dual, p float32) Dual {
	return Dual{
		Val: float32(math.Pow(float64(d.Val), float64(p))),
		Der: p * d.Der * float32(math.Pow(float64(d.Val), float64(p-1.0))),
	}
}
