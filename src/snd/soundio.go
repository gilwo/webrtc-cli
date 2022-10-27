package snd

import (
	"errors"
	"strings"
	"time"
)

type Params struct {
	DeviceOrFile string
	Rate         int
	Channels     int
	FrameLength  time.Duration
}

type Batch struct {
	Data []int16
	Err  error
}

type Reader interface {
	Batches() <-chan Batch
	Stop()
}

func NewReader(params Params) (Reader, error) {
	if isWavFile(params.DeviceOrFile) {
		return NewWavReader(params)
	}
	if strings.Contains(params.DeviceOrFile, "malgo") && NewMalgoRecorder != nil {
		return NewMalgoRecorder(params)
	} else if NewPulseRecorder != nil {
		return NewPulseRecorder(params)
	} else {
		return nil, errors.New("NewReader: missing implementation for" + params.DeviceOrFile)
	}
}

type Player interface {
	Batches() chan<- []int16
	Errors() <-chan error
	Stopped() <-chan struct{}
	Stop()
}

func NewPlayer(params Params) (Player, error) {
	if strings.Contains(params.DeviceOrFile, "malgo") {
		return newMalgoPlayer(params)
	} else if NewPulsePlayer != nil {
		return NewPulsePlayer(params)
	} else {
		return nil, errors.New("NewPlayer: missing implementation for" + params.DeviceOrFile)
	}
}

var (
	NewPulseRecorder func(Params) (Reader, error)
	NewMalgoRecorder func(Params) (Reader, error)
	NewPulsePlayer   func(Params) (Player, error)
	NewMalgoPlayer   func(Params) (Player, error)
)
