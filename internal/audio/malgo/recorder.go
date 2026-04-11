package malgo

import (
	"sync"

	"codeberg.org/dbus/sussurro/internal/audio"
	"github.com/gen2brain/malgo"
)

// recorder implements the audio.Recorder interface using the malgo library.
type recorder struct {
	mu         sync.Mutex
	mctx       *malgo.AllocatedContext
	device     *malgo.Device
	samples    []int16
	recording  bool
	sampleRate uint32
}

// NewRecorder creates a new Malgo recorder instance that satisfies the audio.Recorder interface.
func NewRecorder(sampleRate uint32) (audio.Recorder, error) {
	mctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}

	r := &recorder{
		mctx:       mctx,
		sampleRate: sampleRate,
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = sampleRate

	onRecvFrames := func(pSampleOut, pSampleIn []byte, framecount uint32) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.recording {
			sampleCount := len(pSampleIn) / 2
			s := make([]int16, sampleCount)
			for i := range sampleCount {
				s[i] = int16(pSampleIn[i*2]) | int16(pSampleIn[i*2+1])<<8
			}
			r.samples = append(r.samples, s...)
		}
	}

	device, err := malgo.InitDevice(mctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onRecvFrames})
	if err != nil {
		mctx.Uninit()
		return nil, err
	}
	r.device = device

	if err := r.device.Start(); err != nil {
		r.device.Uninit()
		r.mctx.Uninit()
		return nil, err
	}

	return r, nil
}

func (r *recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples = nil
	r.recording = true
	return nil
}

func (r *recorder) Stop() ([]int16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recording = false
	samples := make([]int16, len(r.samples))
	copy(samples, r.samples)
	r.samples = nil
	return samples, nil
}

func (r *recorder) Close() {
	if r.device != nil {
		r.device.Uninit()
	}
	if r.mctx != nil {
		r.mctx.Uninit()
	}
}
