package wav

import (
	"errors"
	"io"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// memWriteSeeker is an in-memory io.WriteSeeker backed by a byte slice.
type memWriteSeeker struct {
	buf []byte
	pos int
}

func (m *memWriteSeeker) Write(p []byte) (int, error) {
	minCap := m.pos + len(p)
	if minCap > cap(m.buf) {
		buf2 := make([]byte, len(m.buf), minCap+len(p))
		copy(buf2, m.buf)
		m.buf = buf2
	}
	if minCap > len(m.buf) {
		m.buf = m.buf[:minCap]
	}
	copy(m.buf[m.pos:], p)
	m.pos += len(p)
	return len(p), nil
}

func (m *memWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	var newPos int
	switch whence {
	case io.SeekStart:
		newPos = int(offset)
	case io.SeekCurrent:
		newPos = m.pos + int(offset)
	case io.SeekEnd:
		newPos = len(m.buf) + int(offset)
	}
	if newPos < 0 {
		return 0, errors.New("negative seek position")
	}
	m.pos = newPos
	return int64(newPos), nil
}

// Encode takes raw int16 audio samples and a sample rate, encodes them into
// a WAV format byte slice, and returns the result.
func Encode(samples []int16, sampleRate uint32) ([]byte, error) {
	ws := &memWriteSeeker{}
	e := wav.NewEncoder(ws, int(sampleRate), 16, 1, 1)

	intSamples := make([]int, len(samples))
	for i, v := range samples {
		intSamples[i] = int(v)
	}

	audioBuf := &audio.IntBuffer{
		Data: intSamples,
		Format: &audio.Format{
			SampleRate:  int(sampleRate),
			NumChannels: 1,
		},
	}

	if err := e.Write(audioBuf); err != nil {
		return nil, err
	}
	if err := e.Close(); err != nil {
		return nil, err
	}

	return ws.buf, nil
}
