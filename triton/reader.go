package triton

import "io"

type Reader interface {
	ReadRecord() (rec map[string]interface{}, err error)
}

// A SerialReader let's us read from multiple readers, in sequence
type SerialReader struct {
	readers []Reader
	r_idx   int
}

func (sr *SerialReader) ReadRecord() (rec map[string]interface{}, err error) {
	for sr.r_idx < len(sr.readers) {
		rec, err := sr.readers[sr.r_idx].ReadRecord()
		if err != nil {
			if err == io.EOF {
				sr.r_idx += 1
			}
		} else {
			return rec, nil
		}
	}

	return nil, io.EOF
}

func NewSerialReader(readers []Reader) Reader {
	return &SerialReader{readers, 0}
}
