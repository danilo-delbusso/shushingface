package audio

// DeviceInfo describes a capture device.
type DeviceInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

// Recorder defines the interface for an audio capture device.
type Recorder interface {
	Start() error
	Stop() ([]int16, error)
	Close()
	// ListDevices returns the available capture devices.
	ListDevices() ([]DeviceInfo, error)
	// SetDevice re-initialises the underlying device with the given ID.
	// Empty string means the system default. Safe to call only while not recording.
	SetDevice(id string) error
	// Level returns a buffered channel that receives the most recent capture
	// buffer's RMS amplitude in [0,1] while recording is active. Buffer
	// capacity is 1: producers drop old values when nobody's reading, so
	// consumers always see the freshest sample without back-pressuring the
	// audio callback. The channel is alive for the lifetime of the recorder.
	Level() <-chan float32
}
