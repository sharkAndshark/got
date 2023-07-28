package main

import (
	"github.com/magiconair/properties/assert"
	"os"
	"testing"
)

func Testcan_can_read_source_correctly(t *testing.T) {

	os.Setenv("GOT_RECIPEDIR", "./test/")
	readConfs()
	// act
	readSources()

	// assert
	assert.Equal(t, len(SOURCES), 1)
	s, ok := SOURCES["table1"]
	assert.Equal(t, ok, true)
	pg := s.(pg)
	assert.Equal(t, pg.MinZoom, 7)
	assert.Equal(t, pg.MaxZoom, 14)
	assert.Equal(t, len(pg.Fields), 2)
	assert.Equal(t, len(pg.Fields), 2)
	assert.Equal(t, pg.Fields["field1"], "desc1")
	assert.Equal(t, pg.Fields["field2"], "desc2")
	assert.Equal(t, len(pg.Sqls), 1)
	_, ok1 := pg.Sqls["z*"]
	assert.Equal(t, ok1, true)
}
