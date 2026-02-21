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
	width        int
	height       int
	heldNodeType Material.NodeType
	paused       bool
}

func NewWorld(width, height int) *World {
	area := make([]Material.Node, width*height)
	next := make([]Material.Node, width*height)
	heldNodeType := Material.SandType
	w := &World{
		area:         area,
		next:         next,
		width:        width,
		height:       height,
		heldNodeType: heldNodeType,
		paused:       false,
	}
	w.init()
	return w
}

func (w *World) init() {
	bottomHalf := w.width * w.height / 2
	for i := range bottomHalf {
		if utils.RandInt(30) == 0 {
			w.area[bottomHalf+i] = Material.MakeNode(Material.SandType)
		}
	}
}

func (w *World) Draw(pixels []byte) {
	var color []byte
	blankColor := Material.BlankType.GetColor()
	for i, v := range w.area {
		if v.NodeType != Material.BlankType {
			color = v.NodeType.GetColor()
		} else {
			color = blankColor
		}
		pixels[i*4] = color[0]
		pixels[i*4+1] = color[1]
		pixels[i*4+2] = color[2]
		pixels[i*4+3] = color[3]
	}
}

func (w *World) Update() {
	if !w.paused {
		w.UpdateWorld()
		// fmt.Printf("Blank: %d, Sand: %d, Water: %d, Rock: %d, Lava: %d, Steam: %d\n",
		// 	utils.CountValue(w.next, Material.Node{NodeType: Material.BlankType, Dirty: false}),
		// 	utils.CountValue(w.next, Material.Node{NodeType: Material.SandType, Dirty: false}),
		// 	utils.CountValue(w.next, Material.Node{NodeType: Material.WaterType, Dirty: false}),
		// 	utils.CountValue(w.next, Material.Node{NodeType: Material.RockType, Dirty: false}),
		// 	utils.CountValue(w.next, Material.Node{NodeType: Material.LavaType, Dirty: false}),
		// 	utils.CountValue(w.next, Material.Node{NodeType: Material.SteamType, Dirty: false}),
		// )
	}
	w.handleInput()
}

func (w *World) UpdateWorld() {

	for y := 0; y < w.height; y++ {
		for x := 0; x < w.width; x++ {
			w.updateFuncs(x, y)
		}
	}
	copy(w.area, w.next)
	clear(w.next)
}

type setterFunctions func(*World, int, int) bool

var nodeFuncs = map[Material.NodeType][]setterFunctions{
	Material.SandType: {
		(*World).holdAtVerticalEdge,
		(*World).trySetBelow,
		(*World).trySetDownDiagonal,
		(*World).defaultHold,
	},
	Material.WaterType: {
		(*World).holdAtVerticalEdge,
		(*World).trySetBelow,
		(*World).trySetDownDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
	Material.RockType: {
		(*World).holdAtVerticalEdge,
		(*World).trySetBelow,
		(*World).trySetDownDiagonal,
		(*World).defaultHold,
	},
	Material.LavaType: {
		(*World).holdAtVerticalEdge,
		(*World).trySetBelow,
		(*World).trySetDownDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
	Material.SteamType: {
		(*World).holdAtVerticalEdge,
		(*World).trySetAbove,
		(*World).trySetUpDiagonal,
		(*World).trySetLateral,
		(*World).defaultHold,
	},
}

func (w *World) updateFuncs(x, y int) {

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

func (w *World) holdOrDisplace(x, y int, node Material.Node) {

	// If no movement is possible and haven't already been replaced, stay in place
	if w.getNextNode(x, y).NodeType == Material.BlankType {
		w.setNextNodeTo(x, y, 0, 0, node)
	} else {
		// For a liquid, find the first position above the invalid home position
		dy, ok := w.nearestBlank(x, y)
		if ok {
			w.setNextNodeTo(x, y, 0, dy, node)
		}
		// if we fail to find a valid position, ignore the particle and let it disappear
	}
}

func (w *World) randomMove(x, y int) bool {

	randomDir, ok := utils.RandomDirection(80)

	if ok {
		// _, dy := randomDir.Delta()
		if !w.inVerticalBounds(y) {
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
		w.holdOrDisplace(x, y, tgtNode)
		w.setNextNodeToSelf(x, y, utils.Hold)
		return true
	}
	w.holdOrDisplace(x, y, selfNode) // If at the bottom edge, stay in place
	return true
}

func (w *World) holdAtVerticalEdge(x, y int) bool {

	if w.inVerticalBounds(y) {
		return false
	}

	selfNode := w.getCurrentNode(x, y)
	if selfNode.NodeType.GetPhase() != Material.Solid {
		w.holdOrDisplace(x, y, selfNode)
		return true
	}

	w.setNextNodeToSelf(x, y, utils.Hold) // If at the bottom edge, stay in place
	return true

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
		w.holdOrDisplace(x, y, Material.MakeNode(result[0]))
		w.holdOrDisplace(x+dx, y+dy, Material.MakeNode(result[1]))
		return true
	}

	if selfMat.NodeType.GetDensity() > tgtMat.NodeType.GetDensity() {
		w.setNextNodeToSelf(x, y, dir)
		return true
	}
	return false
}

func (w *World) trySetBelow(x, y int) bool {
	return w.directionalNodeCheck(x, y, utils.Below)
}

func (w *World) trySetAbove(x, y int) bool {
	return w.directionalNodeCheck(x, y, utils.Above)
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

func (w *World) trySetUpDiagonal(x, y int) bool {
	firstDir, secondDir := utils.RandomUpDiagonal()
	if w.directionalNodeCheck(x, y, firstDir) {
		return true
	}
	if w.directionalNodeCheck(x, y, secondDir) {
		return true
	}
	return false
}

func (w *World) trySetDownDiagonal(x, y int) bool {
	firstDir, secondDir := utils.RandomDownDiagonal()
	if w.directionalNodeCheck(x, y, firstDir) {
		return true
	}
	if w.directionalNodeCheck(x, y, secondDir) {
		return true
	}
	return false
}

func (w *World) nearestBlank(x, y int) (int, bool) {
	dy := 0

	for !(w.getNextNode(x, y+dy).NodeType == Material.BlankType) && y+dy >= 0 {
		dy -= 1
	}

	if dy < 0 {
		return dy, true
	}

	dy = 0
	for !(w.getNextNode(x, y+dy).NodeType == Material.BlankType) && y+dy <= w.height {
		dy += 1
	}

	if dy > w.height {
		return 0, false
	}

	return dy, true
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

func (w *World) inVerticalBounds(y int) bool {
	return y+1 < w.height && y-1 > 0
}

func (w *World) getCurrentNode(x, y int) Material.Node {
	return w.area[y*w.width+x]
}
