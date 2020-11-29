package rsmt2d

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepairExtendedDataSquare(t *testing.T) {
	for _, codec := range codecs {
		codec := codec.codecType()

		bufferSize := 64
		ones := bytes.Repeat([]byte{1}, bufferSize)
		twos := bytes.Repeat([]byte{2}, bufferSize)
		threes := bytes.Repeat([]byte{3}, bufferSize)
		fours := bytes.Repeat([]byte{4}, bufferSize)

		original, err := ComputeExtendedDataSquare([][]byte{
			ones, twos,
			threes, fours,
		}, codec)
		if err != nil {
			panic(err)
		}
		ovcp := newDefaultVCP(original)
		rowRoots := Commitments(Row, ovcp, original.width)
		colRoots := Commitments(Column, ovcp, original.width)
		flattened := original.flattened()
		flattened[0], flattened[2], flattened[3] = nil, nil, nil
		flattened[4], flattened[5], flattened[6], flattened[7] = nil, nil, nil, nil
		flattened[8], flattened[9], flattened[10] = nil, nil, nil
		flattened[12], flattened[13] = nil, nil
		var result *ExtendedDataSquare
		result, err = RepairExtendedDataSquare(rowRoots, colRoots, flattened, codec, ovcp)
		if err != nil {
			t.Errorf("unexpected err while repairing data square: %v, codec: :%v", err, codec)
		} else {
			assert.Equal(t, result.square[0][0], ones)
			assert.Equal(t, result.square[0][1], twos)
			assert.Equal(t, result.square[1][0], threes)
			assert.Equal(t, result.square[1][1], fours)
		}

		flattened = original.flattened()
		flattened[0], flattened[2], flattened[3] = nil, nil, nil
		flattened[4], flattened[5], flattened[6], flattened[7] = nil, nil, nil, nil
		flattened[8], flattened[9], flattened[10] = nil, nil, nil
		flattened[12], flattened[13], flattened[14] = nil, nil, nil
		_, err = RepairExtendedDataSquare(rowRoots, colRoots, flattened, codec, ovcp)
		if err == nil {
			t.Errorf("did not return an error on trying to repair an unrepairable square")
		}
		var corrupted ExtendedDataSquare
		corrupted, err = original.deepCopy()
		if err != nil {
			t.Fatalf("unexpected err while copying original data: %v, codec: :%v", err, codec)
		}
		corruptChunk := bytes.Repeat([]byte{66}, bufferSize)
		corrupted.setCell(0, 0, corruptChunk)
		_, err = RepairExtendedDataSquare(rowRoots, colRoots, corrupted.flattened(), codec, ovcp)
		if err == nil {
			t.Errorf("did not return an error on trying to repair a square with bad roots")
		}

		var ok bool
		corrupted, err = original.deepCopy()
		if err != nil {
			t.Fatalf("unexpected err while copying original data: %v, codec: :%v", err, codec)
		}
		corrupted.setCell(0, 0, corruptChunk)
		// recalculate roots
		corruptVCP := newDefaultVCP(&corrupted)
		corruptRowRoots := Commitments(Row, corruptVCP, corrupted.width)
		corruptColRoots := Commitments(Column, corruptVCP, corrupted.width)
		_, err = RepairExtendedDataSquare(corruptRowRoots, corruptColRoots, corrupted.flattened(), codec, corruptVCP)
		if err, ok = err.(*ByzantineRowError); !ok {
			t.Errorf("did not return a ByzantineRowError for a bad row; got: %v", err)
		}

		corrupted, err = original.deepCopy()
		if err != nil {
			t.Fatalf("unexpected err while copying original data: %v, codec: :%v", err, codec)
		}
		corrupted.setCell(0, 3, corruptChunk)
		// recalculate roots
		corruptVCP = newDefaultVCP(&corrupted)
		corruptRowRoots = Commitments(Row, corruptVCP, corrupted.width)
		corruptColRoots = Commitments(Column, corruptVCP, corrupted.width)
		_, err = RepairExtendedDataSquare(corruptRowRoots, corruptColRoots, corrupted.flattened(), codec, corruptVCP)
		if err, ok = err.(*ByzantineRowError); !ok {
			t.Errorf("did not return a ByzantineRowError for a bad row; got %v", err)
		}

		corrupted, err = original.deepCopy()
		if err != nil {
			t.Fatalf("unexpected err while copying original data: %v, codec: :%v", err, codec)
		}
		corrupted.setCell(0, 0, corruptChunk)
		flattened = corrupted.flattened()
		flattened[1], flattened[2], flattened[3] = nil, nil, nil
		// recalculate roots
		corruptVCP = newDefaultVCP(&corrupted)
		corruptRowRoots = Commitments(Row, corruptVCP, corrupted.width)
		corruptColRoots = Commitments(Column, corruptVCP, corrupted.width)
		_, err = RepairExtendedDataSquare(corruptRowRoots, corruptColRoots, flattened, codec, corruptVCP)
		if err, ok = err.(*ByzantineColumnError); !ok {
			t.Errorf("did not return a ByzantineColumnError for a bad column; got %v", err)
		}

		corrupted, err = original.deepCopy()
		if err != nil {
			t.Fatalf("unexpected err while copying original data: %v, codec: :%v", err, codec)
		}
		corrupted.setCell(3, 0, corruptChunk)
		flattened = corrupted.flattened()
		flattened[1], flattened[2], flattened[3] = nil, nil, nil
		// recalculate roots
		corruptVCP = newDefaultVCP(&corrupted)
		corruptRowRoots = Commitments(Row, corruptVCP, corrupted.width)
		corruptColRoots = Commitments(Column, corruptVCP, corrupted.width)
		_, err = RepairExtendedDataSquare(corruptRowRoots, corruptColRoots, flattened, codec, corruptVCP)
		if err, ok = err.(*ByzantineColumnError); !ok {
			t.Errorf("did not return a ByzantineColumnError for a bad column; got %v", err)
		}
	}
}
