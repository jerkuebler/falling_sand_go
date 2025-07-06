package Material

type Node int

// Use order in enum as relative density, reordering will break functions
const (
	Blank Node = iota
	Water
	Sand
	Rock
	Lava
	Steam
)

type Phase int

const (
	Empty Phase = iota
	Solid
	Liquid
	Gas
)

type Density int

const (
	Zero Density = iota
	LightGas
	LightLiquid
	HeavyLiquid
	LightSolid
)

type nodeData struct {
	phase   Phase
	density Density
	color   []byte
}

var nodeInfo = []nodeData{
	{phase: Empty, density: Zero, color: []byte{0x00, 0x00, 0x00, 0xff}},         // Blank
	{phase: Liquid, density: LightLiquid, color: []byte{0x00, 0x00, 0xff, 0xff}}, // Water
	{phase: Solid, density: LightSolid, color: []byte{0xde, 0xbd, 0x1a, 0xff}},   // Sand
	{phase: Solid, density: LightSolid, color: []byte{0x80, 0x85, 0x88, 0xff}},   // Rock
	{phase: Liquid, density: HeavyLiquid, color: []byte{0xff, 0x68, 0x51, 0xff}}, // Lava
	{phase: Gas, density: LightGas, color: []byte{0xad, 0xb7, 0xc7, 0xa8}},       // Steam
}

func (n Node) GetDensity() Density {
	return nodeInfo[n].density
}

func (n Node) GetColor() []byte {
	return nodeInfo[n].color
}

func (n Node) GetPhase() Phase {
	return nodeInfo[n].phase
}

var MaterialInteractions = map[[2]Node][2]Node{
	{Lava, Water}:  {Rock, Steam},
	{Water, Lava}:  {Steam, Rock},
	{Steam, Steam}: {Water, Water},
}
