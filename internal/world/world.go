package World

import (
	"fmt"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

type World struct {
	area      []Material.Grain
	next      []Material.Grain
	width     int
	height    int
	heldGrain Material.Grain
}

func NewWorld(width, height int) *World {
	area := make([]Material.Grain, width*height)
	next := make([]Material.Grain, width*height)
	heldGrain := Material.Sand // Default grain type
	w := &World{area: area, next: next, width: width, height: height, heldGrain: heldGrain}
	w.init()
	return w
}

func (w *World) init() {
	bottomHalf := w.width * w.height / 2 // Only apply to bottom half to speed up init
	for i := range bottomHalf {
		if rand.Intn(30) == 0 {
			w.area[bottomHalf+i] = 1 // Randomly set some points to true
		} else {
			w.area[bottomHalf+i] = 0
		}
	}
}

func (w *World) DebugUpdate() {
	if ebiten.IsKeyPressed(ebiten.Key5) {
		w.Update()
		fmt.Printf("Blank: %d, Sand: %d, Water: %d\n", utils.CountValue(w.next, 0), utils.CountValue(w.next, 1), utils.CountValue(w.next, 2))
	}
}

func (w *World) Update() {
	// Update logic for the world can be added here
	w.next = make([]Material.Grain, w.width*w.height)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		mouse_pos := utils.Point{X: 0, Y: 0}
		mouse_pos.X, mouse_pos.Y = ebiten.CursorPosition()
		w.next[mouse_pos.Y*w.width+mouse_pos.X] = w.heldGrain
	}

	w.handlePressedKeys()

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			switch w.area[y*w.width+x] {
			case 1:
				w.updateSand(x, y)
			case 2:
				w.updateWater(x, y)
			}
		}
	}
	w.area = w.next
}

func (w *World) Draw(pixels []byte) {
	for i, v := range w.area {
		if v != 0 {
			color := Material.Grain(v).GetColor()
			for j := range 4 {
				pixels[i*4+j] = color[j]
			}
		} else {
			color := Material.Blank.GetColor()
			for j := range 4 {
				pixels[i*4+j] = color[j]
			}
		}
	}
}

func (w *World) handlePressedKeys() {
	if ebiten.IsKeyPressed(ebiten.Key1) {
		w.heldGrain = Material.Sand
	}

	if ebiten.IsKeyPressed(ebiten.Key2) {
		w.heldGrain = Material.Water
	}
}

func (w *World) updateSand(x, y int) {

	if w.isBottomBound(y) {
		w.setNextGrain(x, y, utils.Hold) // If at the bottom edge, stay in place
		return
	}

	if w.directionalGrainCheck(x, y, utils.Below, Material.Blank) {
		w.setNextGrain(x, y, utils.Below) // If the cell below is empty, move down
		return
	}

	if w.isGrain(x, y, utils.Below, Material.Water) {
		w.setNextGrain(x, y, utils.Below)
		return
	}

	// If the cell below and to the left/right is empty, move diagonally
	if w.diagonalGrainCheck(x, y, Material.Blank) {
		return
	}

	// If no movement is possible, stay in place
	w.setNextGrain(x, y, utils.Hold)
}

func (w *World) updateWater(x, y int) {
	if w.isBottomBound(y) {
		w.setNextGrain(x, y, utils.Hold) // If at the bottom edge, stay in place
		return
	}

	// If the cell below is empty, move down
	if w.trySetDirectional(x, y, utils.Below, Material.Blank) {
		return
	}

	// If the cell below and to the left/right is empty, move diagonally
	if w.diagonalGrainCheck(x, y, Material.Blank) {
		return
	}

	if w.lateralGrainCheck(x, y, Material.Blank) {
		return
	}

	w.holdOrRise(x, y)

}

func (w *World) holdOrRise(x, y int) {
	selfMaterial := w.getCurrentGrain(x, y)
	// If no movement is possible and haven't already been replaced, stay in place
	if w.isNextGrain(x, y, utils.Hold, Material.Blank) {
		w.setNextGrain(x, y, utils.Hold)
	} else {
		// For a liquid, find the first position above the invalid home position
		dy := w.nextNearestBlankAbove(x, y)
		w.setNextGrainTo(x, dy, selfMaterial)
	}
}

func (w *World) trySetDirectional(x, y int, dir utils.Direction, checkFor Material.Grain) bool {
	if w.directionalGrainCheck(x, y, dir, checkFor) {
		w.setNextGrain(x, y, dir)
		return true
	}
	return false
}

func (w *World) directionalGrainCheck(x, y int, dir utils.Direction, checkFor Material.Grain) bool {
	return w.isGrain(x, y, dir, checkFor) && w.isNextGrain(x, y, dir, checkFor)
}

func (w *World) lateralGrainCheck(x, y int, checkFor Material.Grain) bool {
	firstDir, secondDir := utils.RandomLateral()
	dx, _ := firstDir.Delta()
	if w.inLateralBounds(x, dx) && w.directionalGrainCheck(x, y, firstDir, checkFor) {
		w.setNextGrain(x, y, firstDir)
		return true
	}
	dx, _ = secondDir.Delta()
	if w.inLateralBounds(x, dx) && w.directionalGrainCheck(x, y, secondDir, checkFor) {
		w.setNextGrain(x, y, secondDir)
		return true
	}
	return false
}

func (w *World) diagonalGrainCheck(x, y int, checkFor Material.Grain) bool {
	firstDir, secondDir := utils.RandomDownDiagonal()
	dx, _ := firstDir.Delta()
	if w.inLateralBounds(x, dx) && w.directionalGrainCheck(x, y, firstDir, checkFor) {
		w.setNextGrain(x, y, firstDir)
		return true
	}
	dx, _ = secondDir.Delta()
	if w.inLateralBounds(x, dx) && w.directionalGrainCheck(x, y, secondDir, checkFor) {
		w.setNextGrain(x, y, secondDir)
		return true
	}
	return false
}

func (w *World) nextNearestBlankAbove(x, y int) int {
	dy := y
	for !w.isNextGrain(x, dy, utils.Hold, Material.Blank) && dy != 0 {
		dy -= 1
	}
	return dy
}

// TODO: Fold other checking functions into one
// func (w *World) isValidMove(x, y int, dir Direction, checkFor Material.Grain) {}

func (w *World) isNextGrain(x, y int, dir utils.Direction, checkFor Material.Grain) bool {
	dx, dy := dir.Delta()
	return w.next[(y+dy)*w.width+x+dx] == checkFor
}

func (w *World) isGrain(x, y int, dir utils.Direction, checkFor Material.Grain) bool {
	dx, dy := dir.Delta()
	return w.area[(y+dy)*w.width+x+dx] == checkFor
}

func (w *World) setNextGrainTo(x, y int, setTo Material.Grain) {
	w.next[(y)*w.width+x] = setTo
}

func (w *World) setNextGrain(x, y int, dir utils.Direction) {
	dx, dy := dir.Delta()
	w.next[(y+dy)*w.width+x+dx] = w.area[y*w.width+x]
}

func (w *World) inLateralBounds(x, dir int) bool {
	return x-dir > 0 && x+dir < w.width
}

func (w *World) isBottomBound(y int) bool {
	return y+1 >= w.height
}

func (w *World) getCurrentGrain(x, y int) Material.Grain {
	return w.area[y*w.width+x]
}
