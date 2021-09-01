package fixedwidth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileparser(t *testing.T) {
	type TestFileparserStruct struct {
		Foo   string `txn:"record=1,start=0,end=0"`
		Lorem string `txn:"record=1,start=2,end=6,literal=LOREM"`
		Bar   bool   `txn:"record=1,start=7,end=7"`
		Bang  string `txn:"record=2,start=5,end=22"`
		Wiz   string `txn:"record=3,start=2,end=3"`
	}

	// Marshal
	lines, err := Marshal(&TestFileparserStruct{
		Foo: "Y",
		Bar: true,
		Wiz: "AB",
	})
	assert.NoError(t, err)
	assert.Equal(t, string(lines[0]), "Y LOREM1")

	// Unmarshal
	result := TestFileparserStruct{}
	err = Unmarshal(lines, &result)
	assert.NoError(t, err)
	assert.Equal(t, TestFileparserStruct{Foo: "Y", Lorem: "LOREM", Bar: true, Wiz: "AB"}, result)
}
