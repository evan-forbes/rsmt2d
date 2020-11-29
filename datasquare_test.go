package rsmt2d

import (
	"reflect"
	"testing"

	"github.com/lazyledger/nmt/namespace"
)

func TestNewDataSquare(t *testing.T) {
	result, err := newDataSquare([][]byte{{1, 2}})
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(result.square, [][][]byte{{{1, 2}}}) {
		t.Errorf("newDataSquare failed for 1x1 square")
	}

	result, err = newDataSquare([][]byte{{1, 2}, {3, 4}, {5, 6}, {7, 8}})
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(result.square, [][][]byte{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}) {
		t.Errorf("newDataSquare failed for 2x2 square")
	}

	_, err = newDataSquare([][]byte{{1, 2}, {3, 4}, {5, 6}})
	if err == nil {
		t.Errorf("newDataSquare failed; inconsistent number of chunks accepted")
	}

	_, err = newDataSquare([][]byte{{1, 2}, {3, 4}, {5, 6}, {7}})
	if err == nil {
		t.Errorf("newDataSquare failed; chunks of unequal size accepted")
	}
}

func TestExtendSquare(t *testing.T) {
	ds, err := newDataSquare([][]byte{{1, 2}})
	if err != nil {
		panic(err)
	}
	err = ds.extendSquare(1, []byte{0})
	if err == nil {
		t.Errorf("extendSquare failed; error not returned when filler chunk size does not match data square chunk size")
	}

	ds, err = newDataSquare([][]byte{{1, 2}})
	if err != nil {
		panic(err)
	}
	err = ds.extendSquare(1, []byte{0, 0})
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(ds.square, [][][]byte{{{1, 2}, {0, 0}}, {{0, 0}, {0, 0}}}) {
		t.Errorf("extendSquare failed; unexpected result when extending 1x1 square to 2x2 square")
	}
}

func TestRoots(t *testing.T) {
	result, err := newDataSquare([][]byte{{1, 2}})
	if err != nil {
		panic(err)
	}
	vcp := newDefaultVCP(&ExtendedDataSquare{dataSquare: result, originalDataWidth: 1, codec: RSGF8})
	if !reflect.DeepEqual(vcp.Commitment(Row, 0), vcp.Commitment(Column, 0)) {
		t.Errorf("computing roots failed; expecting row and column roots for 1x1 square to be equal")
	}
}

func TestProofs(t *testing.T) {
	result, err := newDataSquare([][]byte{{1, 2}, {3, 4}, {5, 6}, {7, 8}})
	if err != nil {
		panic(err)
	}
	proof, err := newDefaultVCP(&ExtendedDataSquare{dataSquare: result}).Prove(Row, 1)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if len(proof.Set) != 2 {
		t.Errorf("computing row proof for (1, 1) in 2x2 square failed; expecting proof set of length 2")
	}
	if proof.Index != 1 {
		t.Errorf("computing row proof for (1, 1) in 2x2 square failed; expecting proof index of 1")
	}
	if proof.Leaves != 2 {
		t.Errorf("computing row proof for (1, 1) in 2x2 square failed; expecting number of leaves to be 2")
	}

	result, err = newDataSquare([][]byte{{1, 2}, {3, 4}, {5, 6}, {7, 8}})
	if err != nil {
		panic(err)
	}
	proof, err = newDefaultVCP(&ExtendedDataSquare{dataSquare: result}).Prove(Column, 1)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	if len(proof.Set) != 2 {
		t.Errorf("computing column proof for (1, 1) in 2x2 square failed; expecting proof set of length 2")
	}
	if proof.Index != 1 {
		t.Errorf("computing column proof for (1, 1) in 2x2 square failed; expecting proof index of 1")
	}
	if proof.Leaves != 2 {
		t.Errorf("computing column proof for (1, 1) in 2x2 square failed; expecting number of leaves to be 2")
	}
}

func Test_nameRow(t *testing.T) {
	raw := [][]byte{{1}, {3}, {5}, {7}}
	result, err := newDataSquare(raw)
	if err != nil {
		panic(err)
	}
	names := []namespace.ID{
		namespace.ID([]byte{byte(int8(1))}),
		namespace.ID([]byte{byte(int8(1))}),
	}
	result.nameRow(1, 0, names)
	if !reflect.DeepEqual(result.square, [][][]byte{
		{{1}, {3}},
		{{1, 5}, {1, 7}},
	}) {
		t.Errorf("NewExtendedDataSquare failed for 2x2 square with chunk size 1")
	}
}
