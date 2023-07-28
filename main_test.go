package main

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestCan_read_source_correctly(t *testing.T) {
	os.Setenv("GOT_RECIPEDIR", "./test/recipes")
	readConfs()
	// act
	readSources()

	// assert
	assert.Equal(t, len(SOURCES), 1)
	s, ok := SOURCES["planet_osm_point"]
	assert.Equal(t, ok, true)
	pg := s.(pg)
	assert.Equal(t, pg.MinZoom, 7)
	assert.Equal(t, pg.MaxZoom, 14)
	assert.Equal(t, len(pg.VectorLayers[0].Fields), 2)
	assert.Equal(t, len(pg.VectorLayers[0].Fields), 2)
	assert.Equal(t, pg.VectorLayers[0].Fields["name"], "text")
	assert.Equal(t, pg.VectorLayers[0].Fields["railway"], "text")
	assert.Equal(t, len(pg.Sqls), 1)
	_, ok1 := pg.Sqls["z*"]
	assert.Equal(t, ok1, true)
}

func TestCan_get_tile_with_named_paramert_sql(t *testing.T) {
	os.Setenv("GOT_RECIPEDIR", "./test/recipes")
	os.Setenv("GOT_DBURL", "postgres://postgis:postgis@127.0.0.1:5432/spatial?sslmode=disable")
	readConfs()
	initDbPool()
	// act
	readSources()
	tempPg := SOURCES["planet_osm_point"]
	//act
	tile, err := tempPg.Tile(13, 6464, 3362)
	//assert
	assert.Equal(t, true, err == nil)
	assert.Equal(t, true, len(tile) > 0)
}
