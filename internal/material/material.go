package Material

type Grain int

// Use order in enum as relative density, reordering will break functions
const (
	Blank Grain = iota
	Water
	Sand
	Rock
	Lava
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

// type grainStruct struct {
// 	phase   Phase
// 	color   []byte
// 	density Density
// }

var grainPhases = map[Grain]Phase{
	Blank: Empty,
	Sand:  Solid,
	Water: Liquid,
	Rock:  Solid,
	Lava:  Liquid,
}

var grainColors = map[Grain][]byte{
	Blank: {0x00, 0x00, 0x00, 0xff},
	Sand:  {0xde, 0xbd, 0x1a, 0xff},
	Water: {0x00, 0x00, 0xff, 0xff},
	Rock:  {0x80, 0x85, 0x88, 0xff},
	Lava:  {0xff, 0x68, 0x51, 0xff},
}

var grainDensity = map[Grain]Density{
	Blank: Zero,
	Water: LightLiquid,
	Sand:  LightSolid,
	Rock:  LightSolid,
	Lava:  HeavyLiquid,
}

func (g Grain) GetDensity() Density {
	if dens, ok := grainDensity[g]; ok {
		return dens
	}
	return Zero
}

func (g Grain) GetColor() []byte {
	if color, ok := grainColors[g]; ok {
		return color
	}
	return grainColors[Blank]
}

func (g Grain) GetPhase() Phase {

	if phase, ok := grainPhases[g]; ok {
		return phase
	}
	return Empty
}
