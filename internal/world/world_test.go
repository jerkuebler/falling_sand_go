package World

import (
	"testing"

	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

func NewTestWorld(width, height int) *World {
	area := make([]nodeUpdate, width*height)
	next := make([]nodeUpdate, width*height)
	zero := make([]nodeUpdate, width*height)
	zeroCopy := make([]nodeUpdate, width*height)
	heldNode := Material.Sand // Default node type
	w := &World{
		area:     area,
		next:     next,
		zero:     zero,
		zeroCopy: zeroCopy,
		width:    width,
		height:   height,
		heldNode: heldNode,
		paused:   false,
	}
	w.testInit()
	return w
}

func randMaterial() Material.Node {
	return Material.Node(utils.RandInt(5) + 1) // 5 materials available not counting blank
}

func (w *World) testInit() {
	// Do not instantiate in top row, as we're not processing in the top row
	for i := range w.width * (w.height - 1) {
		if utils.RandInt(2) == 0 {
			w.area[i+w.width] = nodeUpdate{randMaterial(), 0, false} // Randomly set all points
		} else {
			w.area[i+w.width] = nodeUpdate{Material.Blank, 0, false}
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
