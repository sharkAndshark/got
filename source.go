package main

import (
	_ "github.com/lib/pq"
)

type source interface {
	Tilejson() tilejson
	Tile(z int, x int, y int) ([]byte, error)
}
