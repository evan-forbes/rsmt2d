// Package rsmt2d implements the two dimensional Reed-Solomon merkle tree data availability scheme.
package rsmt2d

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/lazyledger/nmt/namespace"
)

// ExtendedDataSquare represents an extended piece of data.
type ExtendedDataSquare struct {
	*dataSquare
	originalDataWidth uint
	codec             CodecType
}

// ComputeExtendedDataSquare computes the extended data square for some chunks of data.
func ComputeExtendedDataSquare(data [][]byte, codecType CodecType) (*ExtendedDataSquare, error) {
	if codec, ok := codecs[codecType]; !ok {
		return nil, errors.New("unsupported codecType")
	} else {
		if len(data) > codec.maxChunks() {
			return nil, errors.New("number of chunks exceeds the maximum")
		}
	}

	ds, err := newDataSquare(data)
	if err != nil {
		return nil, err
	}

	eds := ExtendedDataSquare{dataSquare: ds, codec: codecType}
	err = eds.erasureExtendSquare()
	if err != nil {
		return nil, err
	}

	return &eds, nil
}

// ComputeNamedExtendedDataSquare computes the extended data square for some chunks of namespaced data.
func ComputeNamedExtendedDataSquare(data [][]byte, names []namespace.ID, codecType CodecType) (*ExtendedDataSquare, error) {
	eds, err := ComputeExtendedDataSquare(data, codecType)
	if err != nil {
		return nil, err
	}
	// prefix extended data square with namespace IDs
	err = eds.addNameAsPrefix(names)
	if err != nil {
		return nil, err
	}
	return eds, nil
}

// ImportExtendedDataSquare imports an extended data square, represented as flattened chunks of data.
func ImportExtendedDataSquare(data [][]byte, codecType CodecType) (*ExtendedDataSquare, error) {
	if codec, ok := codecs[codecType]; !ok {
		return nil, errors.New("unsupported codecType")
	} else {
		if len(data) > 4*codec.maxChunks() {
			return nil, errors.New("number of chunks exceeds the maximum")
		}
	}
	ds, err := newDataSquare(data)
	if err != nil {
		return nil, err
	}

	eds := ExtendedDataSquare{dataSquare: ds, codec: codecType}
	if eds.width%2 != 0 {
		return nil, errors.New("square width must be even")
	}

	eds.originalDataWidth = eds.width / 2

	return &eds, nil
}

func (eds *ExtendedDataSquare) erasureExtendSquare() error {
	eds.originalDataWidth = eds.width
	if err := eds.extendSquare(eds.width, bytes.Repeat([]byte{0}, int(eds.chunkSize))); err != nil {
		return err
	}

	var shares [][]byte
	var err error

	// Extend original square horizontally and vertically
	//  ------- -------
	// |       |       |
	// |   O → |   E   |
	// |   ↓   |       |
	//  ------- -------
	// |       |
	// |   E   |
	// |       |
	//  -------
	for i := uint(0); i < eds.originalDataWidth; i++ {
		// Extend horizontally
		shares, err = Encode(eds.rowSlice(i, 0, eds.originalDataWidth), eds.codec)
		if err != nil {
			return err
		}
		if err := eds.setRowSlice(i, eds.originalDataWidth, shares); err != nil {
			return err
		}

		// Extend vertically
		shares, err = Encode(eds.columnSlice(0, i, eds.originalDataWidth), eds.codec)
		if err != nil {
			return err
		}
		if err := eds.setColumnSlice(eds.originalDataWidth, i, shares); err != nil {
			return err
		}
	}

	// Extend extended square horizontally
	//  ------- -------
	// |       |       |
	// |   O   |   E   |
	// |       |       |
	//  ------- -------
	// |       |       |
	// |   E → |   E   |
	// |       |       |
	//  ------- -------
	for i := eds.originalDataWidth; i < eds.width; i++ {
		// Extend horizontally
		shares, err = Encode(eds.rowSlice(i, 0, eds.originalDataWidth), eds.codec)
		if err != nil {
			return err
		}
		if err := eds.setRowSlice(i, eds.originalDataWidth, shares); err != nil {
			return err
		}
	}

	return nil
}

// addNameAsPrefix adds namespaces IDs to the extended datasquare. Assumes that
// data has already been erasured
func (eds *ExtendedDataSquare) addNameAsPrefix(ids []namespace.ID) error {
	// ensure appropriate length
	count := eds.originalDataWidth * eds.originalDataWidth
	if uint(len(ids)) != count {
		return fmt.Errorf("unexpected number of IDs: wanted %d got %d", count, len(ids))
	}
	// create the appropriate sized parity data
	parityNamespace := genParityNameSpaceID(int8(ids[0].Size()))
	// add names to each row
	for r := uint(0); r < eds.width; r++ {
		// naming Q0 and Q1
		if r < eds.originalDataWidth {
			// name Q0 portion of row
			err := eds.nameRow(r, 0, ids[r*eds.originalDataWidth:(r+1)*eds.originalDataWidth])
			if err != nil {
				return err
			}
			// name Q1 portion of row
			eds.uniformNameRow(r, eds.originalDataWidth, parityNamespace)
			continue
		}
		// naming Q2 and Q3
		eds.uniformNameRow(r, 0, parityNamespace)
	}
	return nil
}

// genParityNameSpaceIDs creates filler namespace.ID s for parity data
func genParityNameSpaceID(size int8) namespace.ID {
	var parityID namespace.ID
	for i := int8(0); i < size; i++ {
		parityID = append(parityID, 0xFF)
	}
	return parityID
}

func (eds *ExtendedDataSquare) deepCopy() (ExtendedDataSquare, error) {
	eds, err := ImportExtendedDataSquare(eds.flattened(), eds.codec)
	return *eds, err
}
