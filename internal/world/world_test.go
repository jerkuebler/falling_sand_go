package World

import (
	"testing"

	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

func NewTestWorld(width, height int) *World {
	area := make([]Material.Grain, width*height)
	next := make([]Material.Grain, width*height)
	zero := make([]Material.Grain, width*height)
	heldGrain := Material.Sand // Default grain type
	w := &World{
		area:      area,
		next:      next,
		zero:      zero,
		width:     width,
		height:    height,
		heldGrain: heldGrain,
		paused:    false,
	}
	w.testInit()
	return w
}

func randMaterial() Material.Grain {
	return Material.Grain(utils.RandInt(5) + 1) // 5 materials available not counting blank
}

func (w *World) testInit() {
	// Do not instantiate in top row, as we're not processing in the top row
	for i := range w.width * (w.height - 1) {
		if utils.RandInt(2) == 0 {
			w.area[i+w.width] = randMaterial() // Randomly set some points to sand
		} else {
			w.area[i+w.width] = Material.Blank
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	b.ReportAllocs()

	w := NewTestWorld(1920, 1080)

	for b.Loop() {
		w.UpdateWorld()
	}
}
