package datafile

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/nireo/bitcask/encoder"
	"github.com/nireo/bitcask/keydir"
)

var (
	ErrWrongByteCount = errors.New("wrote wrong amount of bytes to file.")
)

type DatafileManager struct {
	datafiles map[uint32]*Datafile
	sync.RWMutex
}

type Datafile struct {
	file *os.File
	id   uint32 // id is the unix timestamp in uint32 form

	// we need this such that we can easily create key metadata.
	offset int64
}

// readOffset reads valueSize amount of bytes starting from offset in the datafile.
func (df *Datafile) ReadOffset(offset int64, valueSize uint32) ([]byte, error) {
	// create a buffer of size valueSize and read that data starting from 'offset'
	buffer := make([]byte, valueSize)
	df.file.Seek(offset, 0)

	if _, err := df.file.Read(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

// write writes a key-value pair in to a datafile. It also returns key-metadata such that it is
// easier to then append this key into the key-dir.
func (df *Datafile) Write(key, value []byte) (*keydir.MemEntry, error) {
	// construct the entry data
	timestamp := uint32(time.Now().Unix())
	asBytes := encoder.EncodeEntry(
		key, value, timestamp,
	)

	nBytes, err := df.file.Write(asBytes)
	if err != nil {
		return nil, err
	}

	sz := len(asBytes)
	if sz != nBytes {
		return nil, ErrWrongByteCount
	}
	df.offset += int64(sz)

	return &keydir.MemEntry{
		Timestamp: timestamp,
		ValOffset: df.offset,
		ValSize:   uint32(len(value)),
		FileID:    df.id,
	}, nil
}

func (df *Datafile) Close() {
	df.file.Close()
}