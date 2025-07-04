package World

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	Substances "github.com/jerkuebler/falling_sand_go/internal/substances"
)

type Point struct {
	x int
	y int
}

type World struct {
	area      []Substances.Grain
	width     int
	height    int
	heldGrain Substances.Grain
}

func NewWorld(width, height int) *World {
	area := make([]Substances.Grain, width*height)
	heldGrain := Substances.GrainSand // Default grain type
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
	next := make([]Substances.Grain, w.width*w.height)
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
		w.heldGrain = Substances.GrainSand
	}

	if ebiten.IsKeyPressed(ebiten.Key2) {
		w.heldGrain = Substances.GrainWater
	}
}

func (w *World) UpdateSand(x, y int, next []Substances.Grain) {
	randomOffset := 1
	if rand.Intn(2) == 0 {
		randomOffset = -1
	} else {
		randomOffset = 1
	}
	current_pos := y*w.width + x
	switch {
	case y+1 >= w.height:
		next[current_pos] = w.area[current_pos] // If at the bottom edge, stay in place
	case w.EmptyBelowCheck(x, y):
		next[(y+1)*w.width+x] = w.area[current_pos] // If the cell below is empty, move down
	// If the cell below and to the left/right is empty, move diagonally
	case w.HorizontalScreenCheck(x, randomOffset) && w.area[(y+1)*w.width+x+randomOffset] == Substances.GrainBlank:
		next[(y+1)*w.width+x+randomOffset] = w.area[current_pos]
	case w.HorizontalScreenCheck(x, randomOffset*-1) && w.area[(y+1)*w.width+x+randomOffset*-1] == Substances.GrainBlank:
		next[(y+1)*w.width+x+randomOffset*-1] = w.area[current_pos]
	// If no movement is possible, stay in place
	default:
		next[current_pos] = w.area[current_pos]
	}
}

func (w *World) UpdateWater(x, y int, next []Substances.Grain) {
	randomOffset := 1
	if rand.Intn(2) == 0 {
		randomOffset = -1
	}
	// top_check := (y - 1) >= 0
	// bottom_check := y+1 < w.height
	switch {
	case y+1 >= w.height:
		next[y*w.width+x] = Substances.GrainWater // If at the bottom edge, stay in place
	case w.area[(y+1)*w.width+x] == Substances.GrainBlank:
		next[(y+1)*w.width+x] = Substances.GrainWater // If the cell below is empty, move down
	// If the cell below and to the left/right is empty, move diagonally
	case w.HorizontalScreenCheck(x, randomOffset) && w.area[(y+1)*w.width+x+randomOffset] == Substances.GrainBlank:
		next[(y+1)*w.width+x+randomOffset] = Substances.GrainWater
	case w.HorizontalScreenCheck(x, randomOffset*-1) && w.area[(y+1)*w.width+x+randomOffset*-1] == Substances.GrainBlank:
		next[(y+1)*w.width+x+randomOffset*-1] = Substances.GrainWater
	case w.HorizontalScreenCheck(x, randomOffset) && w.area[(y)*w.width+x+randomOffset] == Substances.GrainBlank && w.area[(y)*w.width+x-randomOffset] != Substances.GrainBlank && next[(y)*w.width+x+randomOffset] == Substances.GrainBlank:
		next[(y)*w.width+x+randomOffset] = Substances.GrainWater
	case w.HorizontalScreenCheck(x, randomOffset*-1) && w.area[(y)*w.width+x+randomOffset*-1] == Substances.GrainBlank && w.area[(y)*w.width+x-randomOffset*-1] == Substances.GrainWater && next[(y)*w.width+x+randomOffset*-1] == Substances.GrainBlank:
		next[(y)*w.width+x+randomOffset*-1] = Substances.GrainWater
	default:
		next[y*w.width+x] = Substances.GrainWater
	}
}

type RelativePosition struct {
}

func (w *World) SetRelativePosition(x, y int)

func (w *World) HorizontalScreenCheck(x, dir int) bool {
	return x-dir > 0 && x+dir < w.width
}

func (w *World) EmptyBelowCheck(x, y int) bool {
	return w.area[(y+1)*w.width+x] == Substances.GrainBlank
}

func (w *World) Draw(pixels []byte) {
	for i, v := range w.area {
		if v != 0 {
			color := Substances.Grain(v).GetColor()
			for j := range 4 {
				pixels[i*4+j] = color[j]
			}
		} else {
			color := Substances.GrainBlank.GetColor()
			for j := range 4 {
				pixels[i*4+j] = color[j]
			}
		}
	}
}
