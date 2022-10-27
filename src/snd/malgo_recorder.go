package snd

import (
	"log"

	"github.com/gen2brain/malgo"
)

func OK(err error) {
	if err != nil {
		panic(err)
	}
}

type MalgoRecorder struct {
	batchCh chan Batch
	doneCh  chan struct{}

	ctx          *malgo.AllocatedContext
	deviceConfig malgo.DeviceConfig
	device       *malgo.Device
	buf          []byte
}

func init() {
	NewMalgoRecorder = newMalgoRecorder
}

func newMalgoRecorder(params Params) (Reader, error) {
	var err error
	m := &MalgoRecorder{
		batchCh: make(chan Batch, 1),
		doneCh:  make(chan struct{}),
	}
	m.ctx, err = malgo.InitContext(nil, malgo.ContextConfig{},
		func(message string) { log.Fatalf("LOG <%v>\n", message) })
	if err != nil {
		return nil, err
	}
	m.deviceConfig = malgo.DefaultDeviceConfig(malgo.Capture)
	m.deviceConfig.Capture.Format = malgo.FormatS16
	m.deviceConfig.Capture.Channels = uint32(params.Channels)
	m.deviceConfig.SampleRate = uint32(params.Rate)
	m.deviceConfig.Alsa.NoMMap = 1

	captureFrameSize := malgo.SampleSizeInBytes(m.deviceConfig.Capture.Format) *
		int(m.deviceConfig.Capture.Channels)
	batchFrameBytes := durationToSamples(params.FrameLength, params.Rate) *
		captureFrameSize
	m.device, err = malgo.InitDevice(m.ctx.Context, m.deviceConfig,
		malgo.DeviceCallbacks{
			Data: func(_, pInputSamples []byte, framecount uint32) {
				dataSize := framecount * uint32(captureFrameSize)
				m.buf = append(m.buf, pInputSamples[:dataSize]...)
				if len(m.buf) > batchFrameBytes {
					m.batchCh <- Batch{Data: bytesToInt16(m.buf[:batchFrameBytes])}
					m.buf = m.buf[batchFrameBytes:]
				}
			},
			Stop: func() {
				close(m.doneCh)
			},
		})
	if err != nil {
		return nil, err
	}
	err = m.device.Start()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *MalgoRecorder) Batches() <-chan Batch {
	return p.batchCh
}

func (m *MalgoRecorder) Stop() {
	if err := m.device.Stop(); err != nil {
		panic(err)
	}

	<-m.doneCh
	m.device.Uninit()
	if err := m.ctx.Uninit(); err != nil {
		panic(err)
	}
	m.ctx.Free()
}
