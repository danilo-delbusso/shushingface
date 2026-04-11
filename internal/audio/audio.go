package audio

// Recorder defines the interface for an audio capture device.
type Recorder interface {
	// Start begins audio recording.
	Start() error
	// Stop ends audio recording and returns the captured samples.
	Stop() ([]int16, error)
	// Close uninitializes the audio device and context.
	Close()
}
