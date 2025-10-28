package vtm

import (
	"fmt"
	"slices"

	"github.com/cjbrigato/go-vtm/audio"
	"github.com/cjbrigato/go-vtm/tracker"
)

const DefaultSampleRate = 44100

var compatibleSampleRates = []int{DefaultSampleRate, 48000}

type VTMPlayer struct {
	module        *tracker.TrackerModule
	audioPlayback *audio.AudioPlayback
}

func (p *VTMPlayer) Title() string {
	return p.module.Title
}

func (p *VTMPlayer) Tempo() int {
	return p.module.Tempo
}

func (p *VTMPlayer) Instruments() []tracker.Instrument {
	return p.module.Instruments
}

func (p *VTMPlayer) Patterns() []tracker.Pattern {
	return p.module.Patterns
}

func (p *VTMPlayer) Sequence() []int {
	return p.module.Sequence
}

func (p *VTMPlayer) Pattern(index int) tracker.Pattern {
	return p.module.Patterns[index]
}

func (p *VTMPlayer) SequenceLength() int {
	return len(p.module.Sequence)
}

func (p *VTMPlayer) Instrument(index int) tracker.Instrument {
	return p.module.Instruments[index]
}

func (p *VTMPlayer) PatternLength(index int) int {
	return len(p.module.Patterns[index].Channels)
}

func (p *VTMPlayer) PatternRowCount(patternIndex int) int {
	return len(p.module.Patterns[patternIndex].Channels)
}

func (p *VTMPlayer) PatternChannel(patternIndex, channelIndex int) []tracker.TrackerNote {
	return p.module.Patterns[patternIndex].Channels[channelIndex]
}

func (p *VTMPlayer) PatternRow(patternIndex, channelIndex, rowIndex int) tracker.TrackerNote {
	return p.module.Patterns[patternIndex].Channels[channelIndex][rowIndex]
}

func (p *VTMPlayer) Module() *tracker.TrackerModule {
	return p.module
}

func (p *VTMPlayer) AudioPlayback() *audio.AudioPlayback {
	return p.audioPlayback
}

func NewVTMPlayer(filename string, sampleRate int) (*VTMPlayer, error) {
	module, err := tracker.LoadVTM(filename)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(compatibleSampleRates, sampleRate) {
		return nil, fmt.Errorf("unsupported sample rate: %d (supported: %v)", sampleRate, compatibleSampleRates)
	}
	audioPlayback, err := audio.NewAudioPlayback(module, sampleRate)
	if err != nil {
		return nil, err
	}
	return &VTMPlayer{module: module, audioPlayback: audioPlayback}, nil
}

func (p *VTMPlayer) Play() error {
	return p.audioPlayback.Play()
}

func (p *VTMPlayer) Stop() {
	p.audioPlayback.Stop()
}

func (p *VTMPlayer) IsDone() bool {
	return p.audioPlayback.IsDone()
}
