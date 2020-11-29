package rsmt2d

import (
	"reflect"
	"testing"

	"github.com/lazyledger/nmt/namespace"
)

func TestComputeExtendedDataSquare(t *testing.T) {
	codec := codecs[RSGF8].codecType()
	result, err := ComputeExtendedDataSquare([][]byte{
		{1}, {2},
		{3}, {4},
	}, codec)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(result.square, [][][]byte{
		{{1}, {2}, {7}, {13}},
		{{3}, {4}, {13}, {31}},
		{{5}, {14}, {19}, {41}},
		{{9}, {26}, {47}, {69}},
	}) {
		t.Errorf("NewExtendedDataSquare failed for 2x2 square with chunk size 1")
	}
}

func TestComputeNamedExtendedDataSquare(t *testing.T) {
	codec := codecs[RSGF8].codecType()
	data := [][]byte{
		{1},
		{2},
		{3},
		{4},
	}
	names := []namespace.ID{
		{1},
		{2},
		{3},
		{4},
	}
	result, err := ComputeNamedExtendedDataSquare(data, names, codec)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(result.square, [][][]byte{
		{{1, 1}, {2, 2}, {255, 7}, {255, 13}},
		{{3, 3}, {4, 4}, {255, 13}, {255, 31}},
		{{255, 5}, {255, 14}, {255, 19}, {255, 41}},
		{{255, 9}, {255, 26}, {255, 47}, {255, 69}},
	}) {
		t.Errorf("NewExtendedDataSquare failed for 2x2 square with chunk size 1")
	}
}
