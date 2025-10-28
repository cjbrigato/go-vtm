package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/cjbrigato/go-vtm"
)

func main() {
	musicFile := flag.String("music", "music/raster-madness.vtm", "Path to VTM music file")
	flag.Parse()

	// Load the music module
	player, err := vtm.NewVTMPlayer(*musicFile, vtm.DefaultSampleRate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading music: %v\n", err)
		os.Exit(1)
	}

	module := player.Module()

	fmt.Printf("Loaded: %s\n", module.Title)
	fmt.Printf("Tempo: %d BPM\n", module.Tempo)
	fmt.Printf("Instruments: %d\n", len(module.Instruments))
	fmt.Printf("Patterns: %d\n", len(module.Patterns))
	fmt.Printf("Sequence length: %d\n", len(module.Sequence))
	fmt.Println("\nInstruments:")
	for i, inst := range module.Instruments {
		fmt.Printf("  %d: %s (%v) ADSR: %.3f/%.3f/%.3f/%.3f\n",
			i, inst.Name, inst.WaveType, inst.Attack, inst.Decay, inst.Sustain, inst.Release)
	}

	fmt.Println("\nSequence:")
	fmt.Printf("  %v\n", module.Sequence)

	fmt.Println("\nPattern preview (first pattern):")
	if len(module.Patterns) > 0 {
		pat := module.Patterns[0]
		fmt.Printf("  Rows: %d, Channels: %d\n", pat.Rows, len(pat.Channels))
		for row := 0; row < pat.Rows && row < 8; row++ {
			fmt.Printf("  %02d:", row)
			for ch := 0; ch < len(pat.Channels); ch++ {
				note := pat.Channels[ch][row]
				if note.Note >= 0 {
					fmt.Printf(" %s", formatNote(note.Note))
				} else {
					fmt.Printf(" ---")
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\nMusic engine loaded successfully!")
	fmt.Println("Run with -music <file.vtm> to load different tracks")

	// Load and start music if specified
	player.Play()
	defer player.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Printf("Signal received, stopping music\n")
		player.Stop()
	}()

	fmt.Printf("Music playing...\n")

	for {

		if !player.IsDone() && player.IsPlaying() {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	fmt.Printf("Music stopped\n")
	runtime.KeepAlive(player)
}

func formatNote(note int) string {
	noteNames := []string{"C-", "C#", "D-", "D#", "E-", "F-", "F#", "G-", "G#", "A-", "A#", "B-"}
	octave := note / 12
	noteInOctave := note % 12
	return fmt.Sprintf("%s%d", noteNames[noteInOctave], octave)
}
