package World

import (
	"testing"

	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

func randMaterial() Material.Node {
	return Material.Node{
		NodeType: Material.NodeType(utils.RandInt(5) + 1),
		Dirty:    false,
	} // 5 materials available not counting blank
}

func (w *World) testInit() {
	// Do not instantiate in top row, as we're not processing in the top row
	for i := range w.width * (w.height - 1) {
		if utils.RandInt(2) == 0 {
			w.area[i+w.width] = randMaterial()
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	b.ReportAllocs()

	w := NewWorld(1920, 1080)
	w.testInit()

	for b.Loop() {
		w.UpdateWorld()
	}
}
