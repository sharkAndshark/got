package main

import (
	_ "github.com/lib/pq"
)

type source interface {
	Metadata() metadata
	Tile(z int, x int, y int) ([]byte, error)
}
