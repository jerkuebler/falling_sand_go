package Material

type Phase int

const (
	Empty Phase = iota
	Solid
	Liquid
	Gas
)

type Grain int

// Use order in enum as relative density, reordering will break functions
const (
	Blank Grain = iota
	Water
	Sand
)

var grainPhases = map[Grain]Phase{
	Blank: Empty,
	Sand:  Solid,
	Water: Liquid,
}

var grainColors = map[Grain][]byte{
	Blank: {0x00, 0x00, 0x00, 0xff},
	Sand:  {0xde, 0xbd, 0x1a, 0xff},
	Water: {0x00, 0x00, 0xff, 0xff},
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
