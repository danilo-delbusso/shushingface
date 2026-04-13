package malgo

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"

	"codeberg.org/dbus/shushingface/internal/audio"
	"github.com/gen2brain/malgo"
)

type recorder struct {
	mu         sync.Mutex
	mctx       *malgo.AllocatedContext
	device     *malgo.Device
	deviceID   string
	samples    []int16
	recording  bool
	sampleRate uint32
}

func NewRecorder(sampleRate uint32) (audio.Recorder, error) {
	mctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}

	r := &recorder{
		mctx:       mctx,
		sampleRate: sampleRate,
	}

	if err := r.initDevice(""); err != nil {
		if uErr := mctx.Uninit(); uErr != nil {
			slog.Warn("malgo context uninit during init failure", "error", uErr)
		}
		return nil, err
	}
	return r, nil
}

// initDevice creates the capture device with the given ID (empty = default).
// Caller must hold r.mu unless called from constructor.
func (r *recorder) initDevice(id string) error {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = r.sampleRate

	if id != "" {
		raw, err := base64.StdEncoding.DecodeString(id)
		if err != nil {
			return fmt.Errorf("decode device id: %w", err)
		}
		var devID malgo.DeviceID
		copy(devID[:], raw)
		deviceConfig.Capture.DeviceID = devID.Pointer()
	}

	onRecvFrames := func(_, pSampleIn []byte, _ uint32) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if !r.recording {
			return
		}
		sampleCount := len(pSampleIn) / 2
		s := make([]int16, sampleCount)
		for i := range sampleCount {
			s[i] = int16(pSampleIn[i*2]) | int16(pSampleIn[i*2+1])<<8
		}
		r.samples = append(r.samples, s...)
	}

	device, err := malgo.InitDevice(r.mctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onRecvFrames})
	if err != nil {
		return err
	}
	r.device = device
	r.deviceID = id
	return nil
}

func (r *recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples = nil
	if !r.recording {
		if err := r.device.Start(); err != nil {
			return err
		}
	}
	r.recording = true
	return nil
}

func (r *recorder) Stop() ([]int16, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.recording {
		if err := r.device.Stop(); err != nil {
			slog.Warn("malgo device stop failed", "error", err)
		}
	}
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
		if err := r.mctx.Uninit(); err != nil {
			slog.Warn("malgo context uninit failed", "error", err)
		}
	}
}

func (r *recorder) ListDevices() ([]audio.DeviceInfo, error) {
	infos, err := r.mctx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("enumerate capture devices: %w", err)
	}
	out := make([]audio.DeviceInfo, 0, len(infos))
	for _, info := range infos {
		out = append(out, audio.DeviceInfo{
			ID:        base64.StdEncoding.EncodeToString(info.ID[:]),
			Name:      info.Name(),
			IsDefault: info.IsDefault != 0,
		})
	}
	return out, nil
}

func (r *recorder) SetDevice(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.recording {
		return fmt.Errorf("cannot switch device while recording")
	}
	if id == r.deviceID && r.device != nil {
		return nil
	}
	if r.device != nil {
		r.device.Uninit()
		r.device = nil
	}
	return r.initDevice(id)
}
