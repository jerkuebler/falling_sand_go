package Substances

type Grain int

const (
	GrainBlank Grain = iota
	GrainSand
	GrainWater
)

func (g Grain) GetColor() []byte {
	switch g {
	case GrainSand:
		return []byte{0xde, 0xbd, 0x1a, 0xff} // RGBA for sand
	case GrainWater:
		return []byte{0x00, 0x00, 0xff, 0xff} // RGBA for water
	default:
		return []byte{0x00, 0x00, 0x00, 0xff} // RGBA for blank
	}
}
