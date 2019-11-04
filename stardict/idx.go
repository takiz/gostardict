package stardict

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Sense has information belonging to single item position in dictionary
type Sense struct {
	Offset uint64
	Size   uint64
}

// Idx implements an in-memory index for a dictionary
type Idx struct {
	Items map[string][]*Sense
}

// NewIdx initializes idx struct
func NewIdx() *Idx {
	idx := new(Idx)
	idx.Items = make(map[string][]*Sense)
	return idx
}

// Add adds an item to in-memory index
func (idx *Idx) Add(item string, offset uint64, size uint64) {
	idx.Items[item] = append(idx.Items[item], &Sense{Offset: offset, Size: size})
}

// Get gets all translations for an item
func (idx Idx) Get(item string) []*Sense {
	return idx.Items[item]
}

// ReadIndex reads dictionary index into a memory and returns in-memory index structure
func ReadIndex(filename string, info *Info) (idx *Idx, err error) {
	reader, err := os.Open(filename)
	if err != nil {
		return
	}
	defer reader.Close()

	var r io.Reader
	if strings.HasSuffix(filename, ".gz") { // if file is compressed then read it from archive
		r, err = gzip.NewReader(reader)
	} else {
		r = reader
	}
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	//	==================
	// data, err := ioutil.ReadFile(filename)

	// // unable to read index
	// if err != nil {
	// 	return
	// }

	idx = NewIdx()
	var a [512]byte // temporary buffer
	var aIdx int
	var expect int

	var dataStr string
	var dataOffset uint64
	var dataSize uint64

	var maxIntBytes int

	if info.Is64 == true {
		maxIntBytes = 8
	} else {
		maxIntBytes = 4
	}

	for _, b := range data {
		if expect == 0 {
			a[aIdx] = b
			if b == 0 {
				dataStr = string(a[:aIdx])
				aIdx = 0
				expect++
				continue
			}
			aIdx++
		} else {
			if expect == 1 {
				a[aIdx] = b
				if aIdx == maxIntBytes-1 {
					if info.Is64 {
						dataOffset = binary.BigEndian.Uint64(a[:maxIntBytes])
					} else {
						dataOffset = uint64(binary.BigEndian.Uint32(a[:maxIntBytes]))
					}
					aIdx = 0
					expect++
					continue
				}
				aIdx++
			} else {
				a[aIdx] = b
				if aIdx == maxIntBytes-1 {
					if info.Is64 {
						dataSize = binary.BigEndian.Uint64(a[:maxIntBytes])
					} else {
						dataSize = uint64(binary.BigEndian.Uint32(a[:maxIntBytes]))
					}

					aIdx = 0
					expect = 0

					// finished with one record
					idx.Add(dataStr, dataOffset, dataSize)

					continue
				}
				aIdx++
			}
		}
	}

	return idx, err
}
