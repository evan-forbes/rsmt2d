package rsmt2d

import (
	"reflect"
	"testing"
)

func TestComputeExtendedDataSquare(t *testing.T) {
	codec := codecs[RSGF8].codecType()
	result, err := ComputeExtendedDataSquare([][]byte{
		{1}, {2},
		{3}, {4},
	}, codec, NewDefaultTree)
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

func TestParallelComputeExtendedDataSquare(t *testing.T) {
	codec := NewRSGF8Codec()
	result, err := ParallelComputeExtendedDataSquare([][]byte{
		{1}, {2},
		{3}, {4},
	}, codec, 16)
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
