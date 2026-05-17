package audio

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/gen2brain/malgo"
)

// S16Slice converts a byte slice to an int16 slice without copying.
func S16Slice(data []byte) []int16 {
	if len(data) == 0 {
		return nil
	}
	return (*[1 << 30]int16)(unsafe.Pointer(&data[0]))[: len(data)/2 : len(data)/2]
}

// CalculateRMS calculates the root mean square (RMS) energy of audio samples.
func CalculateRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum int64
	for _, s := range samples {
		sum += int64(s) * int64(s)
	}
	return math.Sqrt(float64(sum) / float64(len(samples)))
}

// Recorder handles audio capture using malgo.
type Recorder struct {
	mu            sync.Mutex
	buffer        []int16
	ctx           *malgo.AllocatedContext
	device        *malgo.Device
	isRecording   bool
	sampleRate    int
	speechStarted bool
	silenceCount  int
	deviceName    string
}

// NewRecorder initializes the malgo context.
func NewRecorder() (*Recorder, error) {
	// Initialize malgo context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize malgo context: %v", err)
	}

	return &Recorder{
		ctx:        ctx,
		sampleRate: 16000, // Target sample rate for ASR
	}, nil
}

// Start begins audio capture.
func (r *Recorder) Start() error {
	r.mu.Lock()
	if r.isRecording {
		r.mu.Unlock()
		return nil
	}
	r.buffer = make([]int16, 0, 16000*5) // Pre-allocate for 5 seconds of 16kHz audio
	r.isRecording = true
	r.mu.Unlock()

	fmt.Printf("Starting recorder (sampleRate: %d)...\n", r.sampleRate)

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = uint32(r.sampleRate)
	deviceConfig.Alsa.NoMMap = 1

	onData := func(pSample2, pSample []byte, frameCount uint32) {
		samples := S16Slice(pSample)
		r.mu.Lock()
		if r.isRecording {
			r.buffer = append(r.buffer, samples...)
		}
		r.mu.Unlock()
	}

	device, err := malgo.InitDevice(r.ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: onData,
	})
	if err != nil {
		r.mu.Lock()
		r.isRecording = false
		r.mu.Unlock()
		fmt.Printf("Error: Failed to initialize audio device: %v\n", err)
		return fmt.Errorf("failed to initialize device: %v", err)
	}

	r.device = device
	r.deviceName = "麦克风" // malgo 不直接提供设备名，使用固定名称
	if err := device.Start(); err != nil {
		device.Uninit()
		r.mu.Lock()
		r.isRecording = false
		r.mu.Unlock()
		fmt.Printf("Error: Failed to start audio device: %v\n", err)
		return fmt.Errorf("failed to start device: %v", err)
	}

	fmt.Println("Recorder started successfully")
	return nil
}

// Stop ends audio capture and returns the recorded PCM data.
func (r *Recorder) Stop() ([]int16, error) {
	r.mu.Lock()
	if !r.isRecording {
		r.mu.Unlock()
		return nil, nil
	}
	r.isRecording = false
	r.mu.Unlock()

	if r.device != nil {
		r.device.Stop()
		r.device.Uninit()
		r.device = nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	data := make([]int16, len(r.buffer))
	copy(data, r.buffer)
	r.buffer = nil
	return data, nil
}

// IsRecording returns the current recording status.
func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isRecording
}

// GetDeviceName returns the name of the current capture device.
func (r *Recorder) GetDeviceName() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.deviceName
}

// GetDefaultCaptureDeviceName returns the name of the default capture device.
func GetDefaultCaptureDeviceName() string {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return "无法检测设备"
	}
	defer ctx.Uninit()

	devices, err := ctx.Devices(malgo.Capture)
	if err != nil || len(devices) == 0 {
		return "未检测到设备"
	}

	// Get full info for the first (default) device
	fullInfo, err := ctx.DeviceInfo(malgo.Capture, devices[0].ID, malgo.Shared)
	if err != nil {
		return "未知设备"
	}

	deviceName := fullInfo.Name()
	if deviceName == "" {
		deviceName = "默认麦克风"
	}

	return deviceName
}

// GetDeviceInfo returns detailed device information.
func GetDeviceInfo() map[string]interface{} {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil
	}
	defer ctx.Uninit()

	devices, err := ctx.Devices(malgo.Capture)
	if err != nil || len(devices) == 0 {
		return nil
	}

	fullInfo, err := ctx.DeviceInfo(malgo.Capture, devices[0].ID, malgo.Shared)
	if err != nil {
		return nil
	}

	// IsDefault is a uint32 field, 1 means it's the default device
	isDefault := fullInfo.IsDefault == 1

	deviceName := fullInfo.Name()
	if deviceName == "" {
		deviceName = "默认麦克风"
	}
	if isDefault {
		deviceName += " (默认)"
	}

	return map[string]interface{}{
		"name":       deviceName,
		"is_default": isDefault,
		"formats":    fullInfo.Formats,
	}
}

// Close releases malgo resources.
func (r *Recorder) Close() {
	if r.device != nil {
		r.device.Uninit()
	}
	if r.ctx != nil {
		r.ctx.Uninit()
		r.ctx.Free()
	}
}
