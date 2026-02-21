package utils

import (
	"math/rand"
)

type Point struct {
	X int
	Y int
}

type Direction int

const (
	Above Direction = iota
	AboveLeft
	AboveRight
	Below
	BelowLeft
	BelowRight
	Left
	Right
	Hold
)

func (d Direction) Delta() (dx, dy int) {
	switch d {
	case Above:
		return 0, -1
	case AboveLeft:
		return -1, -1
	case AboveRight:
		return 1, -1
	case Below:
		return 0, 1
	case BelowLeft:
		return -1, 1
	case BelowRight:
		return 1, 1
	case Left:
		return -1, 0
	case Right:
		return 1, 0
	default:
		return 0, 0
	}
}

func CountValue[T comparable](slice []T, target T) int {
	count := 0
	for _, val := range slice {
		if val == target {
			count++
		}
	}
	return count
}

func RandomDirection(chance int) (Direction, bool) {
	// chance: chance out of 100 that the return will be a direction
	if RandInt(100) >= chance {
		return Hold, false
	}
	return Direction(RandInt(8)), true
}

func RandomUpDiagonal() (Direction, Direction) {
	r := rand.Intn(2)

	if r == 0 {
		return AboveLeft, AboveRight
	}

	return AboveRight, AboveLeft
}

func RandomDownDiagonal() (Direction, Direction) {
	r := rand.Intn(2)

	if r == 0 {
		return BelowLeft, BelowRight
	}

	return BelowRight, BelowLeft

}

func RandomLateral() (Direction, Direction) {
	r := rand.Intn(2)

	if r == 0 {
		return Left, Right
	}

	return Right, Left
}

func RandInt(r int) int {
	// in case I change the RNG method later
	return rand.Intn(r)
}
