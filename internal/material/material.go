package Material

type Grain int

// Use order in enum as relative density, reordering will break functions
const (
	Blank Grain = iota
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

type grainData struct {
	phase   Phase
	density Density
	color   []byte
}

var grainInfo = []grainData{
	{phase: Empty, density: Zero, color: []byte{0x00, 0x00, 0x00, 0xff}},         // Blank
	{phase: Liquid, density: LightLiquid, color: []byte{0x00, 0x00, 0xff, 0xff}}, // Water
	{phase: Solid, density: LightSolid, color: []byte{0xde, 0xbd, 0x1a, 0xff}},   // Sand
	{phase: Solid, density: LightSolid, color: []byte{0x80, 0x85, 0x88, 0xff}},   // Rock
	{phase: Liquid, density: HeavyLiquid, color: []byte{0xff, 0x68, 0x51, 0xff}}, // Lava
	{phase: Gas, density: LightGas, color: []byte{0xad, 0xb7, 0xc7, 0xa8}},       // Steam
}

func (g Grain) GetDensity() Density {
	return grainInfo[g].density
}

func (g Grain) GetColor() []byte {
	return grainInfo[g].color
}

func (g Grain) GetPhase() Phase {
	return grainInfo[g].phase
}

var MaterialInteractions = map[[2]Grain][2]Grain{
	{Lava, Water}:  {Rock, Steam},
	{Water, Lava}:  {Steam, Rock},
	{Steam, Steam}: {Water, Water},
}
