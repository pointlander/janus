package main

import (
	"fmt"
	"testing"
)

func TestMultiplier4xBool(t *testing.T) {
	circuit := Multiplier4()
	device := circuit.NewDeviceBool()
	for y := uint64(0); y < 16; y++ {
		for x := uint64(0); x < 16; x++ {
			device.SetUint64("Y", 4, y)
			device.SetUint64("X", 4, x)
			device.Execute(false)
			r := device.Uint64("P", 8)
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
			device.SetUint64("Y", 4, y)
			device.SetUint64("X", 4, x)
			device.Execute(false)
			r := device.Uint64("P", 8)
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