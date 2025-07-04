package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	World "github.com/jerkuebler/falling_sand_go/internal/world"
)

const (
	screenWidth  = 320
	screenHeight = 240
	screenScale  = 4
)

type Game struct {
	world  *World.World
	pixels []byte
}

func (g *Game) Update() error {
	g.world.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, screenWidth*screenHeight*4)
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
		world: World.NewWorld(screenWidth, screenHeight),
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
