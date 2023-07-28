package main

type tilejson struct {
	tilejson      string
	name          string
	description   string
	schema        string
	tiles         []string
	minzoom       int
	maxzoom       int
	bounds        []float32
	vector_layers []vector_layer
}

type vector_layer struct {
	Id     string
	Fields map[string]string
}
