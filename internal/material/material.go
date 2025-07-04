package Material

type Grain int

const (
	Blank Grain = iota
	Sand
	Water
)

func (g Grain) GetColor() []byte {
	switch g {
	case Sand:
		return []byte{0xde, 0xbd, 0x1a, 0xff} // RGBA for sand
	case Water:
		return []byte{0x00, 0x00, 0xff, 0xff} // RGBA for water
	default:
		return []byte{0x00, 0x00, 0x00, 0xff} // RGBA for blank
	}
}
