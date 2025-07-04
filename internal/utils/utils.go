package utils

import (
	"math/rand"
)

func CountValue[T comparable](slice []T, target T) int {
	count := 0
	for _, val := range slice {
		if val == target {
			count++
		}
	}
	return count
}

func RandomDirection() int {
	randomOffset := 1
	if rand.Intn(2) == 0 {
		randomOffset = -1
	}
	return randomOffset
}
