// Copyright 2018 The Janus Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"
	"testing"
)

func TestMultiplier4xBool(t *testing.T) {
	circuit := Multiplier4()
	device := circuit.NewDeviceBool()
	for y := uint64(0); y < 16; y++ {
		for x := uint64(0); x < 16; x++ {
			device.SetUint64("Y", y)
			device.SetUint64("X", x)
			device.Execute(false)
			r := device.Uint64("P")
			if r != x*y {
				t.Fatalf("%d * %d != %d (%d)", x, y, r, x*y)
			}
			device.Execute(true)
			for i := 0; i < 16; i++ {
				name := fmt.Sprintf("A%d", i)
				if device.Get(name) {
					t.Fatal("should be zero")
				}
			}
			for i := 0; i < 12; i++ {
				name := fmt.Sprintf("Z%d", i)
				if device.Get(name) {
					t.Fatal("should be zero")
				}
			}
			device.Reset()
		}
	}
}

func TestMultiplier4xFloat32(t *testing.T) {
	circuit := Multiplier4()
	device := circuit.NewDeviceFloat32()
	for y := uint64(0); y < 16; y++ {
		for x := uint64(0); x < 16; x++ {
			device.SetUint64("Y", y)
			device.SetUint64("X", x)
			device.Execute(false)
			r := device.Uint64("P")
			if r != x*y {
				t.Fatalf("%d * %d != %d (%d)", x, y, r, x*y)
			}
			device.Execute(true)
			for i := 0; i < 16; i++ {
				name := fmt.Sprintf("A%d", i)
				if signal := device.Get(name); signal > 0.5 {
					t.Fatal("should be zero", name, signal)
				}
			}
			for i := 0; i < 12; i++ {
				name := fmt.Sprintf("Z%d", i)
				if signal := device.Get(name); signal > 0.5 {
					t.Fatal("should be zero", name, signal)
				}
			}
			device.Reset()
		}
	}
}

func TestMultiplier4xDual(t *testing.T) {
	circuit := Multiplier4()
	device := circuit.NewDeviceDual()
	for y := uint64(0); y < 16; y++ {
		for x := uint64(0); x < 16; x++ {
			device.SetUint64("Y", y)
			device.SetUint64("X", x)
			device.Execute(false)
			r := device.Uint64("P")
			if r != x*y {
				t.Fatalf("%d * %d != %d (%d)", x, y, r, x*y)
			}
			device.Execute(true)
			for i := 0; i < 16; i++ {
				name := fmt.Sprintf("A%d", i)
				if signal := device.Get(name); signal.Val > 0.5 {
					t.Fatal("should be zero", name, signal)
				}
			}
			for i := 0; i < 12; i++ {
				name := fmt.Sprintf("Z%d", i)
				if signal := device.Get(name); signal.Val > 0.5 {
					t.Fatal("should be zero", name, signal)
				}
			}
			device.Reset()
		}
	}
}

func TestDual(t *testing.T) {
	x, y := Dual{Val: 5, Der: 1}, Dual{Val: 6}
	f := Mul(Pow(x, 2), y)
	if math.Round(float64(f.Der)) != 60.0 {
		t.Fatal("derivative should be 60")
	}
}

func TestMultiplier(t *testing.T) {
	test := func(size int, full FullAdder, half HalfAdder) {
		circuit := Multiplier(size, full, half)
		max := uint64(1)
		for i := 0; i < size; i++ {
			max *= 2
		}
		device := circuit.NewDeviceBool()
		for y := uint64(0); y < max; y++ {
			for x := uint64(0); x < max; x++ {
				device.SetUint64("Y", y)
				device.SetUint64("X", x)
				device.Execute(false)
				r := device.Uint64("P")
				if r != x*y {
					t.Fatalf("%d * %d != %d (%d)", x, y, r, x*y)
				}
				device.Execute(true)
				a := int(circuit.Buses["A"])
				for i := 0; i < a; i++ {
					name := fmt.Sprintf("A%d", i)
					if device.Get(name) {
						t.Fatal("should be zero")
					}
				}
				z := int(circuit.Buses["Z"])
				for i := 0; i < z; i++ {
					name := fmt.Sprintf("Z%d", i)
					if device.Get(name) {
						t.Fatal("should be zero")
					}
				}
				device.Reset()
			}
		}
	}
	test(4, FullAdderA1, HalfAdderA1)
	test(8, FullAdderA1, HalfAdderA1)
}
