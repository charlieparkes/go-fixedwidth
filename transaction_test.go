package fixedwidth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionUnmarshal(t *testing.T) {
	type TestTransactionUnmarshal struct {
		Foo string `txn:"record=1,start=0,end=0"`
	}

	lines := [][]byte{
		[]byte("Y"),
	}

	ds := TestTransactionUnmarshal{}
	err := Unmarshal(lines, &ds)
	assert.NoError(t, err)
	assert.Equal(t, "Y", ds.Foo)
}

func TestTransactionMarshal(t *testing.T) {
	type TestTransactionMarshal struct {
		Foo string `txn:"record=1,start=0,end=0"`
	}

	ds := TestTransactionMarshal{
		Foo: "Y",
	}

	lines, err := Marshal(ds)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Y"), lines[0])
}

func TestTransactionMarshalMissingRecords(t *testing.T) {
	type TestTransactionMarshalMissingRecords struct {
		Foo string `txn:"record=1,start=0,end=0"`
		Bar string `txn:"record=3,start=0,end=0"`
	}

	ds := TestTransactionMarshalMissingRecords{
		Foo: "Y",
		Bar: "Z",
	}

	lines, err := Marshal(ds)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Y"), lines[0])
	assert.Equal(t, []byte("Z"), lines[2])
}

func TestTransactionMissingFields(t *testing.T) {
	type TestTransactionMissingFieldsStruct struct {
		Foo   string `txn:"record=1,start=0,end=0"`
		Lorem string `txn:"record=1,start=2,end=6,literal=LOREM"`
		Bar   bool   `txn:"record=1,start=7,end=7"`
		Wiz   string `txn:"record=2,start=2,end=3"`
	}

	ds := TestTransactionMissingFieldsStruct{}

	txn, err := NewTransaction(&ds)
	assert.NoError(t, err)

	lines := [][]byte{
		[]byte("X"),
	}

	err = txn.UnmarshalLines(lines, &ds)
	assert.NoError(t, err)

	assert.Equal(t, TestTransactionMissingFieldsStruct{Foo: "X", Lorem: "LOREM", Bar: false, Wiz: ""}, ds)
}

func TestTransactionTruncatedField(t *testing.T) {
	type TestTransactionTruncatedFieldStruct struct {
		Name string `txn:"record=1,start=0,end=9"`
	}

	ds := TestTransactionTruncatedFieldStruct{}

	txn, err := NewTransaction(&ds)
	assert.NoError(t, err)

	lines := [][]byte{
		[]byte("John Doe"),
	}

	err = txn.UnmarshalLines(lines, &ds)
	assert.NoError(t, err)

	assert.Equal(t, "John Doe", ds.Name)
}

func TestTransactionTime(t *testing.T) {
	type TestTransactionTimeStruct struct {
		T1      time.Time  `txn:"record=1,start=0,end=7,time=01022006"`
		T2      *time.Time `txn:"record=2,start=0,end=7"`
		T3Empty time.Time  `txn:"record=3,start=0,end=7"`
		T4Empty *time.Time `txn:"record=4,start=0,end=7"`
		T5Empty time.Time  `txn:"record=3,start=0,end=7"`
		T6Empty *time.Time `txn:"record=4,start=0,end=7"`
	}

	txn, err := NewTransaction(&TestTransactionTimeStruct{})
	assert.NoError(t, err)

	lines := [][]byte{
		[]byte("03171994"),
		[]byte("04181995"),
		[]byte(""),
		[]byte(""),
		[]byte("        "),
		[]byte("        "),
	}

	ds := TestTransactionTimeStruct{}
	err = txn.UnmarshalLines(lines, &ds)
	assert.NoError(t, err)

	t1, _ := time.Parse("01022006", "03171994")
	t2, _ := time.Parse("01022006", "04181995")
	t3 := time.Time{}
	var t4 *time.Time
	assert.Equal(t, t1, ds.T1)
	assert.Equal(t, &t2, ds.T2)
	assert.Equal(t, t3, ds.T3Empty)
	assert.Equal(t, t4, ds.T4Empty)
	assert.Equal(t, t3, ds.T5Empty)
	assert.Equal(t, t4, ds.T6Empty)
	assert.True(t, ds.T3Empty.IsZero())
	assert.True(t, ds.T5Empty.IsZero())
}
