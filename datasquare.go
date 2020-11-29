package rsmt2d

import (
	"errors"
	"math"

	"github.com/lazyledger/nmt/namespace"
)

type dataSquare struct {
	square    [][][]byte
	width     uint
	chunkSize uint
}

func newDataSquare(data [][]byte) (*dataSquare, error) {
	width := int(math.Ceil(math.Sqrt(float64(len(data)))))
	if int(math.Pow(float64(width), 2)) != len(data) {
		return nil, errors.New("number of chunks must be a square number")
	}

	square := make([][][]byte, width)
	chunkSize := len(data[0])
	for i := 0; i < width; i++ {
		square[i] = data[i*width : i*width+width]

		for j := 0; j < width; j++ {
			if len(square[i][j]) != chunkSize {
				return nil, errors.New("all chunks must be of equal size")
			}
		}
	}

	return &dataSquare{
		square:    square,
		width:     uint(width),
		chunkSize: uint(chunkSize),
	}, nil
}

func (ds *dataSquare) extendSquare(extendedWidth uint, fillerChunk []byte) error {
	if uint(len(fillerChunk)) != ds.chunkSize {
		return errors.New("filler chunk size does not match data square chunk size")
	}

	newWidth := ds.width + extendedWidth
	newSquare := make([][][]byte, newWidth)

	fillerExtendedRow := make([][]byte, extendedWidth)
	for i := uint(0); i < extendedWidth; i++ {
		fillerExtendedRow[i] = fillerChunk
	}

	fillerRow := make([][]byte, newWidth)
	for i := uint(0); i < newWidth; i++ {
		fillerRow[i] = fillerChunk
	}

	row := make([][]byte, ds.width)
	for i := uint(0); i < ds.width; i++ {
		copy(row, ds.square[i])
		newSquare[i] = append(row, fillerExtendedRow...)
	}

	for i := ds.width; i < newWidth; i++ {
		newSquare[i] = make([][]byte, newWidth)
		copy(newSquare[i], fillerRow)
	}

	ds.square = newSquare
	ds.width = newWidth

	return nil
}

func (ds *dataSquare) rowSlice(x uint, y uint, length uint) [][]byte {
	return ds.square[x][y : y+length]
}

// Row returns the data in a row.
func (ds *dataSquare) Row(x uint) [][]byte {
	return ds.rowSlice(x, 0, ds.width)
}

func (ds *dataSquare) setRowSlice(x uint, y uint, newRow [][]byte) error {
	for i := uint(0); i < uint(len(newRow)); i++ {
		if len(newRow[i]) != int(ds.chunkSize) {
			return errors.New("invalid chunk size")
		}
	}

	for i := uint(0); i < uint(len(newRow)); i++ {
		ds.square[x][y+i] = newRow[i]
	}

	return nil
}

func (ds *dataSquare) nameRow(x, y uint, names []namespace.ID) error {
	// ensure uniform namespace size
	for _, id := range names {
		if id.Size() != names[0].Size() {
			return errors.New("variable size of namespace.IDs")
		}
	}

	for i := uint(0); i < uint(len(names)); i++ {
		ds.square[x][y+i] = append(names[i], ds.square[x][y+i]...)
	}
	return nil
}

func (ds *dataSquare) uniformNameRow(x, y uint, name namespace.ID) {
	for i := uint(0); i < uint(len(ds.square[x]))-y; i++ {
		prefix := make([]byte, name.Size())
		copy(prefix, name)
		ds.square[x][y+i] = append(prefix, ds.square[x][y+i]...)
	}
}

func (ds *dataSquare) columnSlice(x uint, y uint, length uint) [][]byte {
	columnSlice := make([][]byte, length)
	for i := uint(0); i < length; i++ {
		columnSlice[i] = ds.square[x+i][y]
	}

	return columnSlice
}

// Column returns the data in a column.
func (ds *dataSquare) Column(y uint) [][]byte {
	return ds.columnSlice(0, y, ds.width)
}

func (ds *dataSquare) setColumnSlice(x uint, y uint, newColumn [][]byte) error {
	for i := uint(0); i < uint(len(newColumn)); i++ {
		if len(newColumn[i]) != int(ds.chunkSize) {
			return errors.New("invalid chunk size")
		}
	}

	for i := uint(0); i < uint(len(newColumn)); i++ {
		ds.square[x+i][y] = newColumn[i]
	}

	return nil
}

// Cell returns a single chunk at a specific cell.
func (ds *dataSquare) Cell(x uint, y uint) []byte {
	cell := make([]byte, ds.chunkSize)
	copy(cell, ds.square[x][y])
	return cell
}

func (ds *dataSquare) setCell(x uint, y uint, newChunk []byte) {
	ds.square[x][y] = newChunk
}

func (ds *dataSquare) flattened() [][]byte {
	flattened := [][]byte(nil)
	for _, data := range ds.square {
		flattened = append(flattened, data...)
	}

	return flattened
}

// Width returns the width of the square.
func (ds *dataSquare) Width() uint {
	return ds.width
}
