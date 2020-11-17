// Package rsmt2d implements the two dimensional Reed-Solomon merkle tree data availability scheme.
package rsmt2d

import (
	"bytes"
	"errors"
	"sync"
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

// ParallelComputeExtendedDataSquare computes the extended data square for some chunks of data in a parallel fashion.
func ParallelComputeExtendedDataSquare(data [][]byte, codec Codec, workers int) (*ExtendedDataSquare, error) {
	if len(data) > codec.maxChunks() {
		return nil, errors.New("number of chunks exceeds the maximum")
	}

	ds, err := newDataSquare(data)
	if err != nil {
		return nil, err
	}

	eds := ExtendedDataSquare{dataSquare: ds}
	err = eds.parallelExtend(workers)
	if err != nil {
		return nil, err
	}

	return &eds, nil
}

// fillRight encodes the original data in the data square horizontally. Assumes
// that the data square has been extended
func (eds *ExtendedDataSquare) fillRight(row uint, codec Codec) error {
	// Extend horizontally
	shares, err := codec.encode(eds.rowSlice(row, 0, eds.originalDataWidth))
	if err != nil {
		return err
	}
	eds.mtx.Lock()
	defer eds.mtx.Unlock()
	if err := eds.setRowSlice(row, eds.originalDataWidth, shares); err != nil {
		return err
	}
	return nil
}

// fillDown encodes the original data in the data square vertically. Assumes
// that the data square has been extended
func (eds *ExtendedDataSquare) fillDown(col uint, codec Codec) error {
	// Extend vertically
	shares, err := codec.encode(eds.columnSlice(0, col, eds.originalDataWidth))
	if err != nil {
		return err
	}
	eds.mtx.Lock()
	defer eds.mtx.Unlock()
	if err := eds.setColumnSlice(eds.originalDataWidth, col, shares); err != nil {
		return err
	}
	return nil
}

func (eds *ExtendedDataSquare) parallelExtend(workers int) error {
	// extend the datasquare with empty data
	eds.originalDataWidth = eds.width
	if err := eds.extendSquare(eds.width, bytes.Repeat([]byte{0}, int(eds.chunkSize))); err != nil {
		return err
	}

	// issue jobs for filling the datasquare phase 1
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
	wg := sync.WaitGroup{}
	errc := make(chan error, eds.originalDataWidth)
	phaseOneJobs := make(chan uint, eds.originalDataWidth)
	go func() {
		defer close(phaseOneJobs)
		for i := uint(0); i < eds.originalDataWidth; i++ {
			phaseOneJobs <- i
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)

		// pass jobs to workers and collect errors
		go func(jobs <-chan uint, errs chan<- error) {
			defer wg.Done()
			codec := NewRSGF8Codec()
			for job := range jobs {

				// encode data vertically
				err := eds.fillDown(job, codec)
				if err != nil {

					errs <- err
					return
				}
				// encode data vertically
				err = eds.fillRight(job, codec)
				if err != nil {

					errs <- err
					return
				}
				errs <- nil

			}

		}(phaseOneJobs, errc)
	}
	wg.Wait()
	close(errc)
	for err := range errc {
		if err != nil {
			return err
		}
	}

	// start phase two
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
	errc = make(chan error, eds.originalDataWidth)
	phaseTwoJobs := make(chan uint, eds.originalDataWidth)
	go func() {
		defer close(phaseTwoJobs)
		for i := eds.originalDataWidth; i < eds.width; i++ {
			phaseTwoJobs <- i
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		// pass jobs to workers and collect errors
		go func(jobs <-chan uint, errs chan<- error) {
			defer wg.Done()
			codec := NewRSGF8Codec()
			for job := range jobs {
				// encode data vertically
				err := eds.fillRight(job, codec)
				if err != nil {
					errs <- err
					return
				}
				errs <- nil
			}
		}(phaseTwoJobs, errc)
	}

	wg.Wait()
	// check for errors
	close(errc)
	for err := range errc {
		if err != nil {
			return err
		}
	}

	return nil
}

func (eds *ExtendedDataSquare) deepCopy() (ExtendedDataSquare, error) {
	eds, err := ImportExtendedDataSquare(eds.flattened(), eds.codec)
	return *eds, err
}
