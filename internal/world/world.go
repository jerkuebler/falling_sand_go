package World

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	Material "github.com/jerkuebler/falling_sand_go/internal/material"
)

type Point struct {
	x int
	y int
}

type World struct {
	area      []Material.Grain
	width     int
	height    int
	heldGrain Material.Grain
}

func NewWorld(width, height int) *World {
	area := make([]Material.Grain, width*height)
	heldGrain := Material.Sand // Default grain type
	w := &World{area: area, width: width, height: height, heldGrain: heldGrain}
	w.init()
	return w
}

func (w *World) init() {
	for i := 0; i < w.width*w.height; i++ {
		if rand.Intn(30) == 0 {
			w.area[i] = 1 // Randomly set some points to true
		} else {
			w.area[i] = 0
		}
	}
}

func (w *World) Update() {
	// Update logic for the world can be added here
	next := make([]Material.Grain, w.width*w.height)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		mouse_pos := Point{0, 0}
		mouse_pos.x, mouse_pos.y = ebiten.CursorPosition()
		next[mouse_pos.y*w.width+mouse_pos.x] = w.heldGrain
	}

	w.HandlePressedKeys()

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			switch w.area[y*w.width+x] {
			case 1:
				w.UpdateSand(x, y, next)
			case 2:
				w.UpdateWater(x, y, next)
			}
		}
	}
	w.area = next
}

func (w *World) HandlePressedKeys() {
	if ebiten.IsKeyPressed(ebiten.Key1) {
		w.heldGrain = Material.Sand
	}

	if ebiten.IsKeyPressed(ebiten.Key2) {
		w.heldGrain = Material.Water
	}
}

func (w *World) UpdateSand(x, y int, next []Material.Grain) {

	if w.IsBottomBound(y) {
		w.SetNextGrain(x, y, Hold, next) // If at the bottom edge, stay in place
		return
	}

	if w.IsGrain(x, y, Below, Material.Blank) {
		w.SetNextGrain(x, y, Below, next) // If the cell below is empty, move down
		return
	}

	firstDir := randomDirection()
	secondDir := firstDir * -1

	// If the cell below and to the left/right is empty, move diagonally
	if w.DiagonalGrainCheck(x, y, firstDir, Material.Blank) {
		w.SetDiagonalGrain(x, y, firstDir, next)
		return
	}

	if w.DiagonalGrainCheck(x, y, secondDir, Material.Blank) {
		w.SetDiagonalGrain(x, y, secondDir, next)
		return
	}

	// If no movement is possible, stay in place
	w.SetNextGrain(x, y, Hold, next)
}

func (w *World) UpdateWater(x, y int, next []Material.Grain) {
	if w.IsBottomBound(y) {
		w.SetNextGrain(x, y, Hold, next) // If at the bottom edge, stay in place
		return
	}

	if w.IsGrain(x, y, Below, Material.Blank) {
		w.SetNextGrain(x, y, Below, next) // If the cell below is empty, move down
		return
	}

	firstDir := randomDirection()
	secondDir := firstDir * -1

	// If the cell below and to the left/right is empty, move diagonally
	if w.DiagonalGrainCheck(x, y, firstDir, Material.Blank) {
		w.SetDiagonalGrain(x, y, firstDir, next)
		return
	}

	if w.DiagonalGrainCheck(x, y, secondDir, Material.Blank) {
		w.SetDiagonalGrain(x, y, secondDir, next)
		return
	}

	if w.LateralGrainCheck(x, y, firstDir, Material.Blank, next) {
		w.SetLateralGrain(x, y, firstDir, next)
		return
	}
	if w.LateralGrainCheck(x, y, firstDir, Material.Blank, next) {
		w.SetLateralGrain(x, y, firstDir, next)
		return
	}

	// If no movement is possible, stay in place
	w.SetNextGrain(x, y, Hold, next)
}

type Direction int

const (
	Above Direction = iota
	AboveLeft
	AboveRight
	Below
	BelowLeft
	BelowRight
	Left
	Right
	Hold
)

func (d Direction) Delta() (dx, dy int) {
	switch d {
	case Above:
		return 0, -1
	case AboveLeft:
		return -1, -1
	case AboveRight:
		return 1, -1
	case Below:
		return 0, 1
	case BelowLeft:
		return -1, 1
	case BelowRight:
		return 1, 1
	case Left:
		return -1, 0
	case Right:
		return 1, 0
	default:
		return 0, 0
	}
}

func randomDirection() int {
	randomOffset := 1
	if rand.Intn(2) == 0 {
		randomOffset = -1
	}
	return randomOffset
}

func (w *World) LateralGrainCheck(x, y, offset int, checkFor Material.Grain, next []Material.Grain) bool {
	if !w.InLateralBounds(x, offset) {
		return false
	}
	if offset == 1 {
		return w.IsGrain(x, y, Right, checkFor) && w.IsNextGrain(x, y, Right, checkFor, next)
	}
	return w.IsGrain(x, y, Left, checkFor) && w.IsNextGrain(x, y, Left, checkFor, next)
}

func (w *World) SetLateralGrain(x, y, offset int, next []Material.Grain) {
	if offset == 1 {
		w.SetNextGrain(x, y, Right, next)
	} else {
		w.SetNextGrain(x, y, Left, next)
	}
}

func (w *World) DiagonalGrainCheck(x, y, offset int, checkFor Material.Grain) bool {
	if !w.InLateralBounds(x, offset) {
		return false
	}
	if offset == 1 {
		return w.IsGrain(x, y, BelowRight, checkFor)
	}
	return w.IsGrain(x, y, BelowLeft, checkFor)
}

func (w *World) SetDiagonalGrain(x, y, offset int, next []Material.Grain) {
	if offset == 1 {
		w.SetNextGrain(x, y, BelowRight, next)
	} else {
		w.SetNextGrain(x, y, BelowLeft, next)
	}
}

func (w *World) IsNextGrain(x, y int, dir Direction, checkFor Material.Grain, next []Material.Grain) bool {
	dx, dy := dir.Delta()
	return next[(y+dy)*w.width+x+dx] == checkFor
}

func (w *World) IsGrain(x, y int, dir Direction, checkFor Material.Grain) bool {
	dx, dy := dir.Delta()
	return w.area[(y+dy)*w.width+x+dx] == checkFor
}

func (w *World) SetNextGrain(x, y int, dir Direction, next []Material.Grain) {
	dx, dy := dir.Delta()
	next[(y+dy)*w.width+x+dx] = w.area[y*w.width+x]
}

func (w *World) InLateralBounds(x, dir int) bool {
	return x-dir > 0 && x+dir < w.width
}

func (w *World) IsBottomBound(y int) bool {
	return y+1 >= w.height
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
