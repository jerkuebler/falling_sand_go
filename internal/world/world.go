package World

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

type nodeUpdate struct {
	nodeType   Material.Node
	position   int
	changeType Material.Change
	moveTo     int
}

type World struct {
	area      []int
	next      []int
	zero      []int
	zeroCopy  []int
	nodes     map[int]nodeUpdate
	nodeIndex int
	width     int
	height    int
	heldNode  Material.Node
	paused    bool
}

func NewWorld(width, height int) *World {
	area := make([]int, width*height)
	next := make([]int, width*height)
	zero := make([]int, width*height)
	zeroCopy := make([]int, width*height)
	nodes := make(map[int]nodeUpdate, width*height)
	heldNode := Material.Sand // Default Node type
	w := &World{
		area:      area,
		next:      next,
		zero:      zero,
		zeroCopy:  zeroCopy,
		nodes:     nodes,
		nodeIndex: 0,
		width:     width,
		height:    height,
		heldNode:  heldNode,
		paused:    false,
	}
	w.init()
	return w
}

func (w *World) addNode(nodeType Material.Node, pos int) int {
	w.nodes[w.nodeIndex] = nodeUpdate{nodeType, pos, Material.NoChange, 0}
	w.nodeIndex += 1
	return w.nodeIndex - 1
}

func (w *World) init() {
	bottomHalf := w.width * w.height / 2 // Only apply to bottom half to speed up init
	for i := range bottomHalf {
		if utils.RandInt(30) == 0 {
			w.addNode(Material.Sand, bottomHalf+i)
		}
	}
}

func (w *World) Draw(pixels []byte) {
	for i, v := range w.area {
		if w.nodes[v].nodeType != 0 {
			color := Material.Node(w.nodes[v].nodeType).GetColor()
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
		// fmt.Printf("Blank: %d, Sand: %d, Water: %d, Rock: %d, Lava: %d, Steam: %d\n",
		// 	utils.CountValue(w.next, nodeUpdate{Material.Blank, 0, false}),
		// 	utils.CountValue(w.next, nodeUpdate{Material.Sand, 0, false}),
		// 	utils.CountValue(w.next, nodeUpdate{Material.Water, 0, false}),
		// 	utils.CountValue(w.next, nodeUpdate{Material.Rock, 0, false}),
		// 	utils.CountValue(w.next, nodeUpdate{Material.Lava, 0, false}),
		// 	utils.CountValue(w.next, nodeUpdate{Material.Steam, 0, false}),
		// )
	}
	w.handleInput()
}

func (w *World) UpdateWorld() {
	// Update logic for the world can be added here
	// _ = copy(w.next, w.zero)
	w.next = w.zero

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			w.updateFuncs(x, y)
		}
	}
	// _ = copy(w.area, w.next)
	w.area, w.zero = w.next, w.area
	_ = copy(w.zero, w.zeroCopy)
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

	selfMaterial := w.getCurrentNodeUpdate(x, y).nodeType

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
		mouseIndex := mouse_pos.Y*w.width + mouse_pos.X
		nodeIndex := w.addNode(w.heldNode, mouseIndex)
		w.area[mouseIndex] = nodeIndex
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
	selfMaterial := w.getCurrentNodeUpdate(x, y).nodeType
	// If no movement is possible and haven't already been replaced, stay in place
	if w.getNextNode(x, y) == Material.Blank {
		w.setNextNodeToSelf(x, y, utils.Hold)
	} else {
		// For a liquid, find the first position above the invalid home position
		dy := w.distanceToNearestBlankAbove(x, y)
		w.setNextNodeTo(x, y, 0, dy, selfMaterial)
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
	selfPhase := w.getCurrentNodeUpdate(x, y).nodeType.GetPhase()
	if selfPhase != Material.Solid {
		w.holdOrRise(x, y)
		return true
	}
	w.setNextNodeToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
	return true
}

func (w *World) holdAtBottom(x, y int) bool {
	selfPhase := w.getCurrentNodeUpdate(x, y).nodeType.GetPhase()
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

	selfMat := w.getCurrentNodeUpdate(x, y)
	tgtMat := w.getCurrentNodeUpdate(x+dx, y+dy)

	// TODO: Optimize here
	result, ok := Material.MaterialInteractions[[2]Material.Node{selfMat.nodeType, tgtMat.nodeType}]
	if ok {
		w.setNextNodeTo(x, y, 0, 0, result[0])
		w.area[y*w.width+x] = nodeUpdate{0, 0, false}
		if tgtMat.target != 0 {
			w.next[tgtMat.target] = nodeUpdate{0, 0, false}
		}
		w.area[(y+dy)*w.width+x+dx] = nodeUpdate{0, 0, false}
		// fmt.Printf("%d at %d, %d from %d\n", result[1], x+dx, y+dy, tgtMat.nodeType)
		w.setNextNodeTo(x, y, dx, dy, result[1])
		return true
	}

	if selfMat.nodeType.GetDensity() > tgtMat.nodeType.GetDensity() {
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

func (w *World) distanceToNearestBlankAbove(x, y int) int {
	dy := 0
	for !(w.getNextNode(x, y+dy) == Material.Blank) && y+dy != 0 {
		dy -= 1
	}
	return dy
}

func (w *World) setNextNodeTo(x, y, dx, dy int, setTo int) {
	newPos := (y+dy)*w.width + x + dx
	w.next[newPos] = setTo
}

func (w *World) setNextNodeToSelf(x, y int, dir utils.Direction) {
	dx, dy := dir.Delta()
	newPos := (y+dy)*w.width + x + dx
	currPos := y*w.width + x
	w.next[newPos] = w.area[currPos]

}

func (w *World) getNextNode(x, y int) Material.Node {
	return w.nodes[w.next[y*w.width+x]].nodeType
}

func (w *World) inLateralBounds(x, dir int) bool {
	return x+dir > 0 && x+dir < w.width
}

func (w *World) inBottomBound(y int) bool {
	return y < w.height
}

func (w *World) getCurrentNodeUpdate(x, y int) nodeUpdate {
	return w.nodes[w.area[y*w.width+x]]
}

// func (w *World) setCurrentNode(x, y int, setTo Material.Node) {
// 	w.area[y*w.width+x] = setTo
// }
