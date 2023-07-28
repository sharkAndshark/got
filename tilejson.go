package main

type metadata struct {
	tilejson    string
	name        string
	description string
	schema      string
	tiles       []string
	minzoom     int
	maxzoom     int
	bounds      []float64
	fillzoom    int
	vector_layers
}
