package tracker

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cjbrigato/go-vtm/synth"
)

// TrackerNote represents a single note in a pattern
type TrackerNote struct {
	Note       int     // MIDI note number (-1 for rest)
	Instrument int     // Instrument number
	Volume     float64 // 0.0 to 1.0
	Effect     string  // Effect command
}

// Pattern represents a pattern of notes across channels
type Pattern struct {
	Rows     int
	Channels [][]TrackerNote // [channel][row]
}

// Instrument defines synthesis parameters
type Instrument struct {
	Name string

	// Traditional synthesis
	WaveType synth.WaveType
	Attack   float64
	Decay    float64
	Sustain  float64
	Release  float64

	// FM synthesis
	IsFM        bool
	FMPreset    string // "PIANO", "EPIANO", "CUSTOM"
	FMAlgorithm synth.FMAlgorithm
	// For custom FM instruments, we'll store parameters as strings and parse them
	FMParams map[string]string
}

// TrackerModule represents a complete tracker module
type TrackerModule struct {
	Title       string
	Tempo       int
	TicksPerRow int
	Patterns    []Pattern
	Sequence    []int // Pattern order
	Instruments []Instrument
}

// LoadVTM loads a VESAsterizer Tracker Module file
func LoadVTM(filename string) (*TrackerModule, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	module := &TrackerModule{
		Tempo:       120,
		TicksPerRow: 6,
		Patterns:    make([]Pattern, 0),
		Sequence:    make([]int, 0),
		Instruments: make([]Instrument, 0),
	}

	scanner := bufio.NewScanner(file)
	var currentPattern *Pattern
	var currentChannel int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "TITLE":
			if len(parts) > 1 {
				module.Title = strings.Join(parts[1:], " ")
			}

		case "TEMPO":
			if len(parts) > 1 {
				module.Tempo, _ = strconv.Atoi(parts[1])
			}

		case "INSTRUMENT":
			if len(parts) >= 7 {
				inst := Instrument{
					Name:     parts[1],
					WaveType: parseWaveType(parts[2]),
					IsFM:     false,
				}
				inst.Attack, _ = strconv.ParseFloat(parts[3], 64)
				inst.Decay, _ = strconv.ParseFloat(parts[4], 64)
				inst.Sustain, _ = strconv.ParseFloat(parts[5], 64)
				inst.Release, _ = strconv.ParseFloat(parts[6], 64)
				module.Instruments = append(module.Instruments, inst)
			}

		case "FMINSTRUMENT":
			// FMINSTRUMENT Name Preset
			// Where Preset can be: PIANO, EPIANO, or CUSTOM (future extension)
			if len(parts) >= 3 {
				inst := Instrument{
					Name:     parts[1],
					IsFM:     true,
					FMPreset: strings.ToUpper(parts[2]),
					FMParams: make(map[string]string),
				}
				// Parse any additional parameters after the preset name
				for i := 3; i < len(parts); i++ {
					param := parts[i]
					if kv := strings.Split(param, "="); len(kv) == 2 {
						inst.FMParams[kv[0]] = kv[1]
					}
				}
				module.Instruments = append(module.Instruments, inst)
			}

		case "PATTERN":
			if len(parts) > 2 {
				rows, _ := strconv.Atoi(parts[1])
				channels, _ := strconv.Atoi(parts[2])
				currentPattern = &Pattern{
					Rows:     rows,
					Channels: make([][]TrackerNote, channels),
				}
				for i := range currentPattern.Channels {
					currentPattern.Channels[i] = make([]TrackerNote, rows)
					// Initialize with rests
					for j := range currentPattern.Channels[i] {
						currentPattern.Channels[i][j] = TrackerNote{Note: -1, Volume: 1.0}
					}
				}
			}

		case "CH":
			// CH 0: C-4 01 F ..
			if currentPattern != nil && len(parts) >= 3 {
				// Remove colon from channel number
				chanStr := strings.TrimSuffix(parts[1], ":")
				currentChannel, _ = strconv.Atoi(chanStr)
				// Parse notes after the channel specification
				for i := 2; i < len(parts); i++ {
					row := i - 2
					if row >= currentPattern.Rows {
						break
					}
					note := parseTrackerNote(parts[i])
					if currentChannel < len(currentPattern.Channels) {
						currentPattern.Channels[currentChannel][row] = note
					}
				}
			}

		case "ENDPATTERN":
			if currentPattern != nil {
				module.Patterns = append(module.Patterns, *currentPattern)
				currentPattern = nil
			}

		case "SEQUENCE":
			for i := 1; i < len(parts); i++ {
				patNum, _ := strconv.Atoi(parts[i])
				module.Sequence = append(module.Sequence, patNum)
			}
		}
	}

	return module, scanner.Err()
}

func parseWaveType(s string) synth.WaveType {
	switch strings.ToUpper(s) {
	case "SQUARE":
		return synth.Square
	case "SAW":
		return synth.Saw
	case "TRIANGLE":
		return synth.Triangle
	case "SINE":
		return synth.Sine
	case "NOISE":
		return synth.Noise
	default:
		return synth.Square
	}
}

func parseTrackerNote(s string) TrackerNote {
	// Format: "C-4" or "---" for rest or "===" for note-off
	if s == "---" || s == ".." || s == "..." {
		return TrackerNote{Note: -1, Instrument: 0, Volume: 1.0}
	}

	// Note-off command
	if s == "===" || s == "OFF" || s == "off" {
		return TrackerNote{Note: -2, Instrument: 0, Volume: 1.0}
	}

	note := TrackerNote{
		Note:       synth.ParseNote(s),
		Instrument: 0,
		Volume:     1.0,
	}

	return note
}

// SaveVTM saves a module to VTM format (for creating files)
func SaveVTM(filename string, module *TrackerModule) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# VESAsterizer Tracker Module\n")
	fmt.Fprintf(file, "TITLE %s\n", module.Title)
	fmt.Fprintf(file, "TEMPO %d\n\n", module.Tempo)

	// Write instruments
	for _, inst := range module.Instruments {
		if inst.IsFM {
			fmt.Fprintf(file, "FMINSTRUMENT %s %s",
				inst.Name,
				inst.FMPreset)
			// Write any custom parameters
			for k, v := range inst.FMParams {
				fmt.Fprintf(file, " %s=%s", k, v)
			}
			fmt.Fprintf(file, "\n")
		} else {
			fmt.Fprintf(file, "INSTRUMENT %s %s %.3f %.3f %.3f %.3f\n",
				inst.Name,
				waveTypeToString(inst.WaveType),
				inst.Attack, inst.Decay, inst.Sustain, inst.Release)
		}
	}
	fmt.Fprintf(file, "\n")

	// Write patterns
	for patIdx, pattern := range module.Patterns {
		fmt.Fprintf(file, "PATTERN %d %d\n", pattern.Rows, len(pattern.Channels))
		for ch := 0; ch < len(pattern.Channels); ch++ {
			fmt.Fprintf(file, "CH %d:", ch)
			for row := 0; row < pattern.Rows; row++ {
				note := pattern.Channels[ch][row]
				if note.Note < 0 {
					fmt.Fprintf(file, " ---")
				} else {
					fmt.Fprintf(file, " %s", synth.FormatNote(note.Note))
				}
			}
			fmt.Fprintf(file, "\n")
		}
		fmt.Fprintf(file, "ENDPATTERN\n\n")
		_ = patIdx
	}

	// Write sequence
	fmt.Fprintf(file, "SEQUENCE")
	for _, pat := range module.Sequence {
		fmt.Fprintf(file, " %d", pat)
	}
	fmt.Fprintf(file, "\n")

	return nil
}

func waveTypeToString(wt synth.WaveType) string {
	switch wt {
	case synth.Square:
		return "SQUARE"
	case synth.Saw:
		return "SAW"
	case synth.Triangle:
		return "TRIANGLE"
	case synth.Sine:
		return "SINE"
	case synth.Noise:
		return "NOISE"
	default:
		return "SQUARE"
	}
}
