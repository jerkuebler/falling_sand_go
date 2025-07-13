package Material

type Node struct {
	NodeType NodeType
	Dirty    bool
}

func MakeNode(nodeType NodeType) Node {
	return Node{
		NodeType: nodeType,
		Dirty:    false,
	}
}

type NodeType int

// Use order in enum as relative density, reordering will break functions
const (
	BlankType NodeType = iota
	WaterType
	SandType
	RockType
	LavaType
	SteamType
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

func (n NodeType) GetDensity() Density {
	return nodeInfo[n].density
}

func (n NodeType) GetColor() []byte {
	return nodeInfo[n].color
}

func (n NodeType) GetPhase() Phase {
	return nodeInfo[n].phase
}

var MaterialInteractions = map[[2]NodeType][2]NodeType{
	{LavaType, WaterType}:  {RockType, SteamType},
	{WaterType, LavaType}:  {SteamType, RockType},
	{SteamType, SteamType}: {WaterType, WaterType},
}
