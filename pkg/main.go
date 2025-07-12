package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	World "github.com/jerkuebler/falling_sand_go/internal/world"
)

const (
	screenWidth  = 320
	screenHeight = 240
	screenScale  = 3
)

var (
	uiFaceSource *text.GoTextFaceSource
)

type Game struct {
	world  *World.World
	pixels []byte
}

func (g *Game) Update() error {
	g.world.Update()

	// log.Printf("TPS: %.2f, FPS: %.2f", ebiten.ActualTPS(), ebiten.ActualFPS())

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, screenWidth*screenHeight*4)
	}
	g.world.Draw(g.pixels)
	screen.WritePixels(g.pixels)

	msg := fmt.Sprintf("TPS: %0.1f, FPS: %0.1f", ebiten.ActualTPS(), ebiten.ActualFPS())
	op := &text.DrawOptions{}
	op.GeoM.Translate(20, 20)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, msg, &text.GoTextFace{
		Source: uiFaceSource,
		Size:   12,
	}, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	uiFaceSource = s

	ebiten.SetWindowSize(screenWidth*screenScale, screenHeight*screenScale)
	ebiten.SetWindowTitle("Falling Sand")
	g := &Game{
		world:  World.NewWorld(screenWidth, screenHeight),
		pixels: make([]byte, screenWidth*screenHeight*4),
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
