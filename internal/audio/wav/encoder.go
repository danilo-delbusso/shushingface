package wav

import (
	"io"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// Encode takes raw int16 audio samples and a sample rate, encodes them into
// a WAV format byte slice, and returns the result. It uses a temporary file
// to satisfy the io.WriteSeeker requirement of the underlying wav encoder.
func Encode(samples []int16, sampleRate uint32) ([]byte, error) {
	// We need a WriteSeeker, so we use a temp file
	tmpFile, err := os.CreateTemp("", "sussurro-encode-*.wav")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	e := wav.NewEncoder(tmpFile, int(sampleRate), 16, 1, 1)

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

	// Read back the data
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return nil, err
	}
	return io.ReadAll(tmpFile)
}
