package World

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	Material "github.com/jerkuebler/falling_sand_go/internal/material"
	utils "github.com/jerkuebler/falling_sand_go/internal/utils"
)

type World struct {
	area         []Material.Node
	next         []Material.Node
	zero         []Material.Node
	width        int
	height       int
	heldNodeType Material.NodeType
	paused       bool
}

func NewWorld(width, height int) *World {
	area := make([]Material.Node, width*height)
	next := make([]Material.Node, width*height)
	zero := make([]Material.Node, width*height)
	heldNodeType := Material.SandType // Default Node type
	w := &World{
		area:         area,
		next:         next,
		zero:         zero,
		width:        width,
		height:       height,
		heldNodeType: heldNodeType,
		paused:       false,
	}
	w.init()
	return w
}

func (w *World) init() {
	bottomHalf := w.width * w.height / 2 // Only apply to bottom half to speed up init
	for i := range bottomHalf {
		if utils.RandInt(30) == 0 {
			w.area[bottomHalf+i] = Material.MakeNode(Material.SandType) // Randomly set some points to sand
		} else {
			w.area[bottomHalf+i] = Material.MakeNode(Material.BlankType)
		}
	}
}

func (w *World) Draw(pixels []byte) {
	for i, v := range w.area {
		if v.NodeType != Material.BlankType {
			color := v.NodeType.GetColor()
			for j := range 4 {
				pixels[i*4+j] = color[j]
			}
		} else {
			color := Material.BlankType.GetColor()
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
			utils.CountValue(w.next, Material.Node{NodeType: Material.BlankType, Dirty: false}),
			utils.CountValue(w.next, Material.Node{NodeType: Material.SandType, Dirty: false}),
			utils.CountValue(w.next, Material.Node{NodeType: Material.WaterType, Dirty: false}),
			utils.CountValue(w.next, Material.Node{NodeType: Material.RockType, Dirty: false}),
			utils.CountValue(w.next, Material.Node{NodeType: Material.LavaType, Dirty: false}),
			utils.CountValue(w.next, Material.Node{NodeType: Material.SteamType, Dirty: false}),
		)
	}
	w.handleInput()
}

func (w *World) UpdateWorld() {
	// Update logic for the world can be added here
	_ = copy(w.next, w.zero)
	// w.next = make([]Material.Node, w.width*w.height)

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			w.updateFuncs(x, y)
			// w.area[y*w.width+x].Dirty = true
		}
	}
	_ = copy(w.area, w.next)
	// w.area = w.next
}

type setterFunctions func(*World, int, int) bool

var nodeFuncs = map[Material.NodeType][]setterFunctions{
	Material.SandType: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).defaultHold,
	},
	Material.WaterType: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
	Material.RockType: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).defaultHold,
	},
	Material.LavaType: {
		(*World).holdAtBottom,
		(*World).trySetBelow,
		(*World).trySetDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
	Material.SteamType: {
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

	if selfMaterial.NodeType == Material.BlankType {
		return
	}

	if selfMaterial.Dirty {
		return
	}

	for _, setFunc := range nodeFuncs[selfMaterial.NodeType] {
		if setFunc(w, x, y) {
			return
		}
	}

	panic(fmt.Sprintf("The material update function failed somehow during material# %d", selfMaterial.NodeType))
}

func (w *World) handleInput() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		mouse_pos := utils.Point{X: 0, Y: 0}
		mouse_pos.X, mouse_pos.Y = ebiten.CursorPosition()
		w.area[mouse_pos.Y*w.width+mouse_pos.X] = Material.MakeNode(w.heldNodeType)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		w.heldNodeType = Material.SandType
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		w.heldNodeType = Material.WaterType
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		w.heldNodeType = Material.RockType
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		w.heldNodeType = Material.LavaType
	}

	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		w.paused = !w.paused
	}
}

func (w *World) holdOrRise(x, y int, node Material.Node) {
	// If no movement is possible and haven't already been replaced, stay in place
	if w.getNextNode(x, y).NodeType == Material.BlankType {
		w.setNextNodeTo(x, y, 0, 0, node)
	} else {
		// For a liquid, find the first position above the invalid home position
		dy := w.nextNearestBlankAbove(x, y)
		w.setNextNodeTo(x, y, 0, dy, node)
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
	selfNode := w.getCurrentNode(x, y)
	tgtNode := w.getNextNode(x, y)

	if tgtNode.NodeType == Material.BlankType {
		w.setNextNodeToSelf(x, y, utils.Hold)
	}

	if selfNode.NodeType.GetDensity() > tgtNode.NodeType.GetDensity() {
		w.holdOrRise(x, y, tgtNode)
		w.setNextNodeToSelf(x, y, utils.Hold)
		return true
	}
	w.holdOrRise(x, y, selfNode) // If at the bottom edge, stay in place
	return true
}

func (w *World) holdAtBottom(x, y int) bool {
	selfNode := w.getCurrentNode(x, y)
	if !w.inBottomBound(y+1) && selfNode.NodeType.GetPhase() != Material.Solid {
		w.holdOrRise(x, y, selfNode)
		return true
	}
	if !w.inBottomBound(y + 1) {
		w.setNextNodeToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
		return true
	}
	return false
}

func (w *World) trySetBelow(x, y int) bool {
	return w.directionalNodeCheck(x, y, utils.Below)
}

func (w *World) directionalNodeCheck(x, y int, dir utils.Direction) bool {
	dx, dy := dir.Delta()

	if !w.inLateralBounds(x, dx) {
		return false
	}

	nextMat := w.getNextNode(x+dx, y+dy).NodeType

	if nextMat != Material.BlankType {
		return false
	}

	selfMat := w.getCurrentNode(x, y)
	tgtMat := w.getCurrentNode(x+dx, y+dy)

	result, ok := Material.MaterialInteractions[[2]Material.NodeType{selfMat.NodeType, tgtMat.NodeType}]
	if ok && !tgtMat.Dirty {
		// I have no idea why swapping the results makes the correct transmutations occur
		w.holdOrRise(x, y, Material.MakeNode(result[1]))
		w.holdOrRise(x+dx, y+dy, Material.MakeNode(result[0]))
		return true
	}

	if selfMat.NodeType.GetDensity() > tgtMat.NodeType.GetDensity() {
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
	dy := 0
	for !(w.getNextNode(x, y+dy).NodeType == Material.BlankType) && y+dy != 0 {
		dy -= 1
	}
	return dy
}

func (w *World) setNextNodeTo(x, y, dx, dy int, setTo Material.Node) {
	w.area[y*w.width+x].Dirty = true
	setTo.Dirty = false
	w.next[(y+dy)*w.width+x+dx] = setTo
}

func (w *World) setNextNodeToSelf(x, y int, dir utils.Direction) {
	w.area[y*w.width+x].Dirty = true
	setTo := w.area[y*w.width+x]
	setTo.Dirty = false
	dx, dy := dir.Delta()
	w.next[(y+dy)*w.width+x+dx] = setTo
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
