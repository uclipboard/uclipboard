package nanos

import (
	"encoding/binary"
	"io"
)

// Hdr 结构体定义
type Hdr struct {
	FullSz uint32
	HdrSz  uint16
}

type NanoS struct {
	Hdr
	Data []byte
}

func Load(r io.Reader) (*NanoS, error) {
	var s NanoS
	// 读取 FullSz 和 HdrSz
	if err := binary.Read(r, binary.BigEndian, &s.FullSz); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &s.HdrSz); err != nil {
		return nil, err
	}

	dataLen := int(s.FullSz) - binary.Size(s.FullSz) - binary.Size(s.HdrSz)

	if dataLen < 0 {
		return nil, io.ErrUnexpectedEOF
	}
	s.Data = make([]byte, dataLen)
	if _, err := io.ReadFull(r, s.Data); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *NanoS) Write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.FullSz); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.HdrSz); err != nil {
		return err
	}
	if _, err := w.Write(s.Data); err != nil {
		return err
	}
	return nil
}

func New(b []byte) (*NanoS, error) {
	var n NanoS

	n.HdrSz = uint16(binary.Size(n.FullSz) + binary.Size(n.HdrSz))
	n.Data = b
	n.FullSz = uint32(len(n.Data) + binary.Size(n.Hdr))
	return &n, nil
}
