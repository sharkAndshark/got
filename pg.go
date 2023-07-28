package main

import "fmt"

type pg struct {
	Name         string
	Description  string
	Schema       string
	MinZoom      int
	MaxZoom      int
	Bounds       []float32
	VectorLayers []vector_layer `yaml:"vectro_layers"`
	Sqls         map[string]string
}

func (p pg) Tilejson() tilejson {
	return tilejson{
		tilejson:      "3.0.0",
		name:          p.Name,
		description:   p.Description,
		schema:        p.Schema,
		tiles:         []string{},
		minzoom:       p.MinZoom,
		maxzoom:       p.MaxZoom,
		bounds:        p.Bounds,
		vector_layers: p.VectorLayers,
	}
}

func (p pg) Tile(z int, x int, y int) ([]byte, error) {
	//TODO implement me
	//todo context
	query := p.findSql(z)

	rows, err := DB.NamedQuery(query, map[string]interface{}{
		"z": z,
		"x": x, "y": y,
	})

	var mvtTile []byte
	// what should I do to get mvt tile by using the rows?
	for rows.Next() {
		err := rows.Scan(&mvtTile)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return mvtTile, nil
}

func (p pg) findSql(z int) string {
	_, ok := p.Sqls["z*"]
	if ok {
		return p.Sqls["z*"]
	} else {
		zoomKey := fmt.Sprintf("z%d", z)
		_, ok := p.Sqls[zoomKey]
		if ok {
			return p.Sqls[""]
		} else {
			return ""
		}
	}
}
