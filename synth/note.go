package synth

import "math"

// Note represents a musical note
type Note int

const (
	C0 Note = iota
	CS0
	D0
	DS0
	E0
	F0
	FS0
	G0
	GS0
	A0
	AS0
	B0
)

// NoteToFrequency converts a note number to frequency in Hz
// Note 0 = C0, Note 12 = C1, etc.
func NoteToFrequency(note int) float64 {
	// A4 (note 57) = 440 Hz
	// f = 440 * 2^((n-57)/12)
	return 440.0 * math.Pow(2.0, float64(note-57)/12.0)
}

// ParseNote parses a note string like "C-4", "C#5", "A-3"
func ParseNote(noteStr string) int {
	if len(noteStr) < 3 {
		return -1 // rest/empty
	}

	noteMap := map[byte]int{
		'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11,
	}

	noteName := noteStr[0]
	sharp := noteStr[1] == '#'
	octave := int(noteStr[2] - '0')

	if octave < 0 || octave > 8 {
		return -1
	}

	noteValue, ok := noteMap[noteName]
	if !ok {
		return -1
	}

	if sharp {
		noteValue++
	}

	return octave*12 + noteValue
}

// FormatNote formats a note number as a string like "C-4"
func FormatNote(note int) string {
	if note < 0 {
		return "---"
	}

	noteNames := []string{"C-", "C#", "D-", "D#", "E-", "F-", "F#", "G-", "G#", "A-", "A#", "B-"}
	octave := note / 12
	noteInOctave := note % 12

	return noteNames[noteInOctave] + string(rune('0'+octave))
}
