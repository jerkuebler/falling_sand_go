package main

import (
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Point struct {
	x int
	y int
}

type Grain int

const (
	GrainBlank Grain = iota
	GrainSand
	GrainWater
)

func (g Grain) GetColor() []byte {
	switch g {
	case GrainSand:
		return []byte{0xde, 0xbd, 0x1a, 0xff} // RGBA for sand
	case GrainWater:
		return []byte{0x00, 0x00, 0xff, 0xff} // RGBA for water
	default:
		return []byte{0x00, 0x00, 0x00, 0xff} // RGBA for blank
	}
}

type World struct {
	area      []Grain
	width     int
	height    int
	heldGrain Grain
}

func NewWorld(width, height int) *World {
	area := make([]Grain, width*height)
	heldGrain := GrainSand // Default grain type
	w := &World{area: area, width: width, height: height, heldGrain: heldGrain}
	w.init()
	return w
}

func (w *World) UpdateSand(x, y int, next []Grain) {
	randomOffset := 1
	if rand.Intn(2) == 0 {
		randomOffset = -1
	} else {
		randomOffset = 1
	}
	switch {
	case y+1 >= w.height:
		next[y*w.width+x] = GrainSand // If at the bottom edge, stay in place
	case w.area[(y+1)*w.width+x] == GrainBlank:
		next[(y+1)*w.width+x] = GrainSand // If the cell below is empty, move down
	// If the cell below and to the left/right is empty, move diagonally
	case 0 < x+randomOffset && x+randomOffset < w.width && w.area[(y+1)*w.width+x+randomOffset] == GrainBlank:
		next[(y+1)*w.width+x+randomOffset] = GrainSand
	case 0 < x+randomOffset*-1 && x+randomOffset*-1 < w.width && w.area[(y+1)*w.width+x+randomOffset*-1] == GrainBlank:
		next[(y+1)*w.width+x+randomOffset*-1] = GrainSand
	// If no movement is possible, stay in place
	default:
		next[y*w.width+x] = GrainSand
	}
}

func (w *World) Update() {
	// Update logic for the world can be added here
	next := make([]Grain, w.width*w.height)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		mouse_pos := Point{0, 0}
		mouse_pos.x, mouse_pos.y = ebiten.CursorPosition()
		next[mouse_pos.y*w.width+mouse_pos.x] = w.heldGrain
	}

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			if w.area[y*w.width+x] == 1 {
				w.UpdateSand(x, y, next)
			}
		}
	}
	w.area = next
}

func (w *World) Draw(pixels []byte) {
	for i, v := range w.area {
		if v != 0 {
			color := Grain(v).GetColor()
			for j := range 4 {
				pixels[i*(screenScale*screenScale)+j] = color[j]
			}
		} else {
			color := GrainBlank.GetColor()
			for j := range 4 {
				pixels[i*(screenScale*screenScale)+j] = color[j]
			}
		}
	}
}

func (w *World) init() {
	for i := 0; i < w.width*w.height; i++ {
		if rand.Intn(10) == 0 {
			w.area[i] = 1 // Randomly set some points to true
		} else {
			w.area[i] = 0
		}
	}
}

const (
	screenWidth  = 320
	screenHeight = 240
	screenScale  = 2
)

type Game struct {
	world  *World
	pixels []byte
}

func (g *Game) Update() error {
	g.world.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, screenWidth*screenHeight*(screenScale*screenScale))
	}
	g.world.Draw(g.pixels)
	screen.WritePixels(g.pixels)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*screenScale, screenHeight*screenScale)
	ebiten.SetWindowTitle("Falling Sand")
	g := &Game{
		world: NewWorld(screenWidth, screenHeight),
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
