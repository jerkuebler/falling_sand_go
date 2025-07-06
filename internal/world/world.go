package World

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

type World struct {
	area     []Material.Node
	next     []Material.Node
	zero     []Material.Node
	width    int
	height   int
	heldNode Material.Node
	paused   bool
}

func NewWorld(width, height int) *World {
	area := make([]Material.Node, width*height)
	next := make([]Material.Node, width*height)
	zero := make([]Material.Node, width*height)
	heldNode := Material.Sand // Default Node type
	w := &World{
		area:     area,
		next:     next,
		zero:     zero,
		width:    width,
		height:   height,
		heldNode: heldNode,
		paused:   false,
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
			color := Material.Node(v).GetColor()
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

var nodeFuncs = map[Material.Node][]setterFunctions{
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

	selfMaterial := w.getCurrentNode(x, y)

	if selfMaterial == Material.Blank {
		return
	}

	for _, setFunc := range nodeFuncs[selfMaterial] {
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
		w.next[mouse_pos.Y*w.width+mouse_pos.X] = w.heldNode
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		w.heldNode = Material.Sand
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		w.heldNode = Material.Water
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		w.heldNode = Material.Rock
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		w.heldNode = Material.Lava
	}

	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		w.paused = !w.paused
	}
}

func (w *World) holdOrRise(x, y int) {
	selfMaterial := w.getCurrentNode(x, y)
	// If no movement is possible and haven't already been replaced, stay in place
	if w.getNextNode(x, y) == Material.Blank {
		w.setNextNodeToSelf(x, y, utils.Hold)
	} else {
		// For a liquid, find the first position above the invalid home position
		dy := w.nextNearestBlankAbove(x, y)
		w.setNextNodeTo(x, dy, selfMaterial)
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
		return w.directionalNodeCheck(x, y, randomDir)
	}
	return false

}

func (w *World) defaultHold(x, y int) bool {
	selfPhase := w.getCurrentNode(x, y).GetPhase()
	if selfPhase != Material.Solid {
		w.holdOrRise(x, y)
		return true
	}
	w.setNextNodeToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
	return true
}

func (w *World) holdAtBottom(x, y int) bool {
	selfPhase := w.getCurrentNode(x, y).GetPhase()
	if !w.inBottomBound(y+1) && selfPhase != Material.Solid {
		w.holdOrRise(x, y)
		return true
	}
	if !w.inBottomBound(y + 1) {
		w.setNextNodeToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
		return true
	}
	return false
}

func (w *World) trySetBelow(x, y int) bool {

	if w.directionalNodeCheck(x, y, utils.Below) {
		w.setNextNodeToSelf(x, y, utils.Below)
		return true
	}

	return false
}

func (w *World) directionalNodeCheck(x, y int, dir utils.Direction) bool {
	dx, dy := dir.Delta()

	if !w.inLateralBounds(x, dx) {
		return false
	}

	nextMat := w.getNextNode(x+dx, y+dy)

	if nextMat != Material.Blank {
		return false
	}

	selfMat := w.getCurrentNode(x, y)
	tgtMat := w.getCurrentNode(x+dx, y+dy)

	// TODO: Figure out how to prevent duplications occurring during interactions.
	// Current best idea is to only allow when moving down, and to make change occur on current frame, which feels off.
	if result, ok := Material.MaterialInteractions[[2]Material.Node{selfMat, tgtMat}]; ok {
		w.setNextNodeTo(x, y, result[0])
		w.setNextNodeTo(x+dx, y+dy, result[1])
		return true
	}

	if selfMat.GetDensity() > tgtMat.GetDensity() {
		w.setNextNodeToSelf(x, y, dir)
		return true
	}
	return false
}

func (w *World) trySetLateral(x, y int) bool {
	firstDir, secondDir := utils.RandomLateral()
	if w.directionalNodeCheck(x, y, firstDir) {
		return true
	}
	if w.directionalNodeCheck(x, y, secondDir) {
		return true
	}
	return false
}

func (w *World) trySetDiagonal(x, y int) bool {
	firstDir, secondDir := utils.RandomDownDiagonal()
	if w.directionalNodeCheck(x, y, firstDir) {
		return true
	}
	if w.directionalNodeCheck(x, y, secondDir) {
		return true
	}
	return false
}

func (w *World) nextNearestBlankAbove(x, y int) int {
	dy := y
	for !(w.getNextNode(x, dy) == Material.Blank) && dy != 0 {
		dy -= 1
	}
	return dy
}

func (w *World) setNextNodeTo(x, y int, setTo Material.Node) {
	w.next[(y)*w.width+x] = setTo
}

func (w *World) setNextNodeToSelf(x, y int, dir utils.Direction) {
	dx, dy := dir.Delta()
	w.next[(y+dy)*w.width+x+dx] = w.area[y*w.width+x]
}

func (w *World) getNextNode(x, y int) Material.Node {
	return w.next[y*w.width+x]
}

func (w *World) inLateralBounds(x, dir int) bool {
	return x+dir > 0 && x+dir < w.width
}

func (w *World) inBottomBound(y int) bool {
	return y < w.height
}

func (w *World) getCurrentNode(x, y int) Material.Node {
	return w.area[y*w.width+x]
}

// func (w *World) setCurrentNode(x, y int, setTo Material.Node) {
// 	w.area[y*w.width+x] = setTo
// }
