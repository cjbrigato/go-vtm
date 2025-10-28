package main

import (
	"fmt"
	"time"

	"github.com/cjbrigato/go-vtm/audio"
	"github.com/cjbrigato/go-vtm/tracker"
)

func main() {
	fmt.Println("=== Direct Harmony Control Example ===\n")

	fmt.Println("Example 1: Accessing Channel Voices")
	fmt.Println("------------------------------------")

	// In actual use, you'd get the player from the audioPlayback
	// For demonstration, let's show the API structure:

	exampleModule := &tracker.TrackerModule{
		Title: "Direct Control Demo",
		Tempo: 120,
		Instruments: []tracker.Instrument{
			{
				Name:     "Piano",
				IsFM:     true,
				FMPreset: "PIANO",
			},
		},
	}

	// Create a player directly for demonstration
	demoPlayer := audio.NewPlayer(exampleModule, 44100)

	// Access channel 0's voice allocator
	channel0 := demoPlayer.VoiceAllocators[0]

	fmt.Printf("Channel 0 has %d voices available\n", channel0.GetMaxVoices())
	fmt.Printf("Currently active: %d voices\n\n", channel0.GetActiveVoiceCount())

	fmt.Println("Example 2: Playing a Chord Programmatically")
	fmt.Println("-------------------------------------------")

	// Play a C major chord (C, E, G, C)
	chord := []int{60, 64, 67, 72} // C4, E4, G4, C5
	channel0.PlayChord(chord, 0.8)

	fmt.Printf("Played chord: C major\n")
	fmt.Printf("Active notes: %v\n\n", channel0.GetActiveNotes())

	fmt.Println("Example 3: Controlling Individual Voices")
	fmt.Println("----------------------------------------")

	// Direct control of each voice in the harmony
	voices := channel0.GetVoices()
	fmt.Printf("Total voices in channel: %d\n", len(voices))

	for i, voice := range voices {
		if voice.IsActive() {
			fmt.Printf("  Voice %d: Active\n", i)
		} else {
			fmt.Printf("  Voice %d: Inactive\n", i)
		}
	}

	fmt.Println("\n" + "Example 4: Advanced Voice Control")
	fmt.Println("----------------------------------")

	// Set specific notes on specific voices (bypass allocation)
	fmt.Println("Manually assigning notes to voices:")
	channel0.SetVoiceNote(0, 60, 0.9) // Voice 0 = C4, loud
	fmt.Println("  Voice 0 → C4 (loud)")

	channel0.SetVoiceNote(1, 64, 0.7) // Voice 1 = E4, medium
	fmt.Println("  Voice 1 → E4 (medium)")

	channel0.SetVoiceNote(2, 67, 0.6) // Voice 2 = G4, soft
	fmt.Println("  Voice 2 → G4 (soft)")

	channel0.SetVoiceNote(3, 72, 0.5) // Voice 3 = C5, softer
	fmt.Println("  Voice 3 → C5 (softer)")

	fmt.Println("\n" + "Example 5: Querying Voice States")
	fmt.Println("---------------------------------")

	// Check which notes are playing
	activeNotes := channel0.GetActiveNotes()
	fmt.Printf("Active notes: %v\n", activeNotes)
	fmt.Printf("Active voice count: %d\n", channel0.GetActiveVoiceCount())

	// Find which voice is playing a specific note
	note := 64 // E4
	voice := channel0.GetVoiceForNote(note)
	if voice != nil {
		fmt.Printf("Note %d (E4) is assigned to a voice\n", note)
	}

	fmt.Println("\n" + "Example 6: Releasing Specific Voices")
	fmt.Println("------------------------------------")

	// Release individual voices
	fmt.Println("Releasing voice 2...")
	channel0.ReleaseVoice(2)

	// Or release specific notes
	fmt.Println("Releasing note 72 (C5)...")
	channel0.NoteOff(72)

	// Or release entire chord
	fmt.Println("Releasing remaining chord notes...")
	channel0.ReleaseChord([]int{60, 64})

	fmt.Println("\n" + "Example 7: Multi-Channel Harmony")
	fmt.Println("---------------------------------")

	// Access multiple channels for rich harmonies
	for ch := 0; ch < demoPlayer.GetChannelCount(); ch++ {
		channelVoices := demoPlayer.GetChannelVoices(ch)
		if channelVoices != nil {
			fmt.Printf("Channel %d: %d voices, %d active\n",
				ch,
				channelVoices.GetMaxVoices(),
				channelVoices.GetActiveVoiceCount())
		}
	}

	fmt.Println("\n" + "API Summary")
	fmt.Println("===========")
	fmt.Println("\nVoice Access:")
	fmt.Println("  player.VoiceAllocators[ch]        - Direct channel access")
	fmt.Println("  player.GetChannelVoices(ch)       - Safe channel access")
	fmt.Println("  allocator.GetVoices()             - All voices in channel")
	fmt.Println("  allocator.GetVoice(index)         - Specific voice")

	fmt.Println("\nQuery Methods:")
	fmt.Println("  allocator.GetActiveNotes()        - What's playing")
	fmt.Println("  allocator.GetActiveVoiceCount()   - How many active")
	fmt.Println("  allocator.GetVoiceForNote(note)   - Find voice for note")
	fmt.Println("  allocator.GetMaxVoices()          - Max polyphony")

	fmt.Println("\nControl Methods:")
	fmt.Println("  allocator.NoteOn(note, vel)       - Smart allocation")
	fmt.Println("  allocator.NoteOff(note)           - Release specific note")
	fmt.Println("  allocator.PlayChord(notes, vel)   - Play multiple notes")
	fmt.Println("  allocator.ReleaseChord(notes)     - Release multiple notes")
	fmt.Println("  allocator.SetVoiceNote(i,n,v)     - Direct voice control")
	fmt.Println("  allocator.ReleaseVoice(index)     - Release specific voice")
	fmt.Println("  allocator.AllNotesOff()           - Panic/reset")

	// Small delay to let notes ring (in real use)
	time.Sleep(100 * time.Millisecond)
}
