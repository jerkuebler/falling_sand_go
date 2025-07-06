package World

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

type World struct {
	area      []Material.Grain
	next      []Material.Grain
	zero      []Material.Grain
	width     int
	height    int
	heldGrain Material.Grain
	paused    bool
}

func NewWorld(width, height int) *World {
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
	w.init()
	return w
}

func (w *World) init() {
	bottomHalf := w.width * w.height / 2 // Only apply to bottom half to speed up init
	for i := range bottomHalf {
		if utils.RandInt(30) == 0 {
			w.area[bottomHalf+i] = Material.Sand // Randomly set some points to sand
		} else {
			w.area[bottomHalf+i] = Material.Blank
		}
	}
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

func (w *World) Update() {
	if !w.paused {
		w.UpdateWorld()
		fmt.Printf("Blank: %d, Sand: %d, Water: %d, Rock: %d, Lava: %d, Steam: %d\n",
			utils.CountValue(w.next, Material.Blank),
			utils.CountValue(w.next, Material.Sand),
			utils.CountValue(w.next, Material.Water),
			utils.CountValue(w.next, Material.Rock),
			utils.CountValue(w.next, Material.Lava),
			utils.CountValue(w.next, Material.Steam),
		)
	}
	w.handleInput()
}

func (w *World) UpdateWorld() {
	// Update logic for the world can be added here
	_ = copy(w.next, w.zero)

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			w.updateFuncs(x, y)
		}
	}
	w.area = w.next
}

type setterFunctions func(*World, int, int) bool

var grainFuncs = map[Material.Grain][]setterFunctions{
	Material.Sand: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).defaultHold,
	},
	Material.Water: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
	Material.Rock: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).defaultHold,
	},
	Material.Lava: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
	Material.Steam: {
		(*World).randomMove,
		(*World).defaultHold,
	},
}

func (w *World) updateFuncs(x, y int) {

	// If a material makes it to the top of the screen, ignore it so it disappears.
	if y == 0 {
		return
	}

	selfMaterial := w.getCurrentGrain(x, y)

	if selfMaterial == Material.Blank {
		return
	}

	for _, setFunc := range grainFuncs[selfMaterial] {
		if setFunc(w, x, y) {
			return
		}
	}

	panic(fmt.Sprintf("The material update function failed somehow during material# %d", selfMaterial))
}

func (w *World) handleInput() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		mouse_pos := utils.Point{X: 0, Y: 0}
		mouse_pos.X, mouse_pos.Y = ebiten.CursorPosition()
		w.next[mouse_pos.Y*w.width+mouse_pos.X] = w.heldGrain
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		w.heldGrain = Material.Sand
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		w.heldGrain = Material.Water
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		w.heldGrain = Material.Rock
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		w.heldGrain = Material.Lava
	}

	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		w.paused = !w.paused
	}
}

func (w *World) holdOrRise(x, y int) {
	selfMaterial := w.getCurrentGrain(x, y)
	// If no movement is possible and haven't already been replaced, stay in place
	if w.getNextGrain(x, y) == Material.Blank {
		w.setNextGrainToSelf(x, y, utils.Hold)
	} else {
		// For a liquid, find the first position above the invalid home position
		dy := w.nextNearestBlankAbove(x, y)
		w.setNextGrainTo(x, dy, selfMaterial)
	}
}

func (w *World) randomMove(x, y int) bool {

	randomDir, ok := utils.RandomDirection(80)

	if ok {
		_, dy := randomDir.Delta()
		if !w.inBottomBound(y + dy) {
			return false
		}
		// fmt.Printf("RMove x: %d, y: %d, dir: %d\n", x, y, randomDir)
		return w.directionalGrainCheck(x, y, randomDir)
	}
	return false

}

func (w *World) defaultHold(x, y int) bool {
	selfPhase := w.getCurrentGrain(x, y).GetPhase()
	if selfPhase != Material.Solid {
		w.holdOrRise(x, y)
		return true
	}
	w.setNextGrainToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
	return true
}

func (w *World) holdAtBottom(x, y int) bool {
	selfPhase := w.getCurrentGrain(x, y).GetPhase()
	if !w.inBottomBound(y+1) && selfPhase != Material.Solid {
		w.holdOrRise(x, y)
		return true
	}
	if !w.inBottomBound(y + 1) {
		w.setNextGrainToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
		return true
	}
	return false
}

func (w *World) trySetBelow(x, y int) bool {

	if w.directionalGrainCheck(x, y, utils.Below) {
		w.setNextGrainToSelf(x, y, utils.Below)
		return true
	}

	return false
}

func (w *World) directionalGrainCheck(x, y int, dir utils.Direction) bool {
	dx, dy := dir.Delta()

	if !w.inLateralBounds(x, dx) {
		return false
	}

	nextMat := w.getNextGrain(x+dx, y+dy)

	if nextMat != Material.Blank {
		return false
	}

	selfMat := w.getCurrentGrain(x, y)
	tgtMat := w.getCurrentGrain(x+dx, y+dy)

	// TODO: Figure out how to prevent duplications occurring during interactions.
	// Current best idea is to only allow when moving down, and to make change occur on current frame, which feels off.
	if result, ok := Material.MaterialInteractions[[2]Material.Grain{selfMat, tgtMat}]; ok {
		w.setNextGrainTo(x, y, result[0])
		w.setNextGrainTo(x+dx, y+dy, result[1])
		return true
	}

	if selfMat.GetDensity() > tgtMat.GetDensity() {
		w.setNextGrainToSelf(x, y, dir)
		return true
	}
	return false
}

func (w *World) trySetLateral(x, y int) bool {
	firstDir, secondDir := utils.RandomLateral()
	if w.directionalGrainCheck(x, y, firstDir) {
		return true
	}
	if w.directionalGrainCheck(x, y, secondDir) {
		return true
	}
	return false
}

func (w *World) trySetDiagonal(x, y int) bool {
	firstDir, secondDir := utils.RandomDownDiagonal()
	if w.directionalGrainCheck(x, y, firstDir) {
		return true
	}
	if w.directionalGrainCheck(x, y, secondDir) {
		return true
	}
	return false
}

func (w *World) nextNearestBlankAbove(x, y int) int {
	dy := y
	for !(w.getNextGrain(x, dy) == Material.Blank) && dy != 0 {
		dy -= 1
	}
	return dy
}

func (w *World) setNextGrainTo(x, y int, setTo Material.Grain) {
	w.next[(y)*w.width+x] = setTo
}

func (w *World) setNextGrainToSelf(x, y int, dir utils.Direction) {
	dx, dy := dir.Delta()
	w.next[(y+dy)*w.width+x+dx] = w.area[y*w.width+x]
}

func (w *World) getNextGrain(x, y int) Material.Grain {
	return w.next[y*w.width+x]
}

func (w *World) inLateralBounds(x, dir int) bool {
	return x+dir > 0 && x+dir < w.width
}

func (w *World) inBottomBound(y int) bool {
	return y < w.height
}

func (w *World) getCurrentGrain(x, y int) Material.Grain {
	return w.area[y*w.width+x]
}

// func (w *World) setCurrentGrain(x, y int, setTo Material.Grain) {
// 	w.area[y*w.width+x] = setTo
// }
