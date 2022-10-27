package snd

import (
	"log"
	"sync"

	"github.com/gen2brain/malgo"
)

type MalgoPlayer struct {
	ctx          *malgo.AllocatedContext
	deviceConfig malgo.DeviceConfig
	device       *malgo.Device
	buf          []byte
	bufLock      sync.Mutex

	initCh   chan error
	dataCh   chan []int16
	errCh    chan error
	cancelCh chan struct{}
	doneCh   chan struct{}
	bufCh    chan []byte
}

func init() {
	NewMalgoPlayer = newMalgoPlayer
}

func newMalgoPlayer(params Params) (Player, error) {
	var err error
	m := &MalgoPlayer{
		initCh:   make(chan error, 1),
		dataCh:   make(chan []int16),
		errCh:    make(chan error, 1),
		cancelCh: make(chan struct{}),
		doneCh:   make(chan struct{}),
		bufCh:    make(chan []byte),
	}
	m.ctx, err = malgo.InitContext(nil, malgo.ContextConfig{},
		func(message string) { log.Fatalf("LOG <%v>\n", message) })
	if err != nil {
		return nil, err
	}
	m.deviceConfig = malgo.DefaultDeviceConfig(malgo.Playback)
	m.deviceConfig.Playback.Format = malgo.FormatS16
	m.deviceConfig.Playback.Channels = uint32(params.Channels)
	m.deviceConfig.SampleRate = uint32(params.Rate)
	m.deviceConfig.Alsa.NoMMap = 1

	playbackFrameSize := malgo.SampleSizeInBytes(m.deviceConfig.Playback.Format) *
		int(m.deviceConfig.Playback.Channels)
	m.device, err = malgo.InitDevice(m.ctx.Context, m.deviceConfig,
		malgo.DeviceCallbacks{
			Data: func(pOutputSample, _ []byte, framecount uint32) {
				dataSize := framecount * uint32(playbackFrameSize)

				var chunk []byte
				select {
				default:
				case chunk = <-m.bufCh:
				}
				m.buf = append(m.buf, chunk...)

				act := len(m.buf)
				if act > int(dataSize) {
					act = int(dataSize)
				}
				copy(pOutputSample, m.buf[:act])
				m.buf = m.buf[act:]
			},
			Stop: func() {
				close(m.doneCh)
			},
		})
	go func() {
		for {
			var data []int16

			select {
			default:
			case <-m.cancelCh:
				return
			}

			select {
			case data = <-m.dataCh:
			case <-m.cancelCh:
				return
			}

			if len(data) == 0 {
				continue
			}
			m.bufCh <- int16ToBytes(data)
		}
	}()
	if err != nil {
		return nil, err
	}
	err = m.device.Start()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MalgoPlayer) Batches() chan<- []int16 {
	return m.dataCh
}

func (m *MalgoPlayer) Errors() <-chan error {
	return m.errCh
}

func (m *MalgoPlayer) Stopped() <-chan struct{} {
	return m.cancelCh
}

func (m *MalgoPlayer) Stop() {
	close(m.cancelCh)
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
