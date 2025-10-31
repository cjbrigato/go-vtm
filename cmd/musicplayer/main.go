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
	"github.com/cjbrigato/go-vtm/audio"
	"github.com/cjbrigato/go-vtm/tracker"
)

func main() {
	musicFile := flag.String("music", "music/raster-madness.vtm", "Path to VTM music file")
	wavOutput := flag.String("wav", "", "Output to WAV file instead of playing (e.g., output.wav)")
	flag.Parse()

	// Load the module
	module, err := tracker.LoadVTM(*musicFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading music: %v\n", err)
		os.Exit(1)
	}

	// Display module information
	fmt.Printf("\n🎵 %s\n", module.Title)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Tempo:     %d BPM\n", module.Tempo)
	fmt.Printf("Patterns:  %d\n", len(module.Patterns))
	fmt.Printf("Sequence:  %d steps\n", len(module.Sequence))

	// Show instruments
	fmt.Printf("\n📻 Instruments:\n")
	for i, inst := range module.Instruments {
		if inst.IsFM {
			fmt.Printf("  [%d] %s - FM Synthesis (%s)\n", i, inst.Name, inst.FMPreset)
		} else {
			fmt.Printf("  [%d] %s - %v (ADSR: %.2f/%.2f/%.2f/%.2f)\n",
				i, inst.Name, inst.WaveType, inst.Attack, inst.Decay, inst.Sustain, inst.Release)
		}
	}

	// Calculate approximate duration
	totalRows := 0
	for _, patIdx := range module.Sequence {
		if patIdx < len(module.Patterns) {
			totalRows += module.Patterns[patIdx].Rows
		}
	}
	secondsPerRow := 60.0 / float64(module.Tempo) / 4.0
	durationSeconds := float64(totalRows) * secondsPerRow
	minutes := int(durationSeconds / 60)
	seconds := int(durationSeconds) % 60

	fmt.Printf("\n⏱️  Duration:  ~%d:%02d\n", minutes, seconds)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Check if we should output to WAV instead of playing
	if *wavOutput != "" {
		fmt.Printf("\n💾 Rendering to WAV file: %s\n", *wavOutput)
		
		startTime := time.Now()
		err := audio.RenderToWAV(module, vtm.DefaultSampleRate, *wavOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering WAV: %v\n", err)
			os.Exit(1)
		}
		
		elapsed := time.Since(startTime)
		fmt.Printf("✅ WAV file created successfully in %.2f seconds\n", elapsed.Seconds())
		return
	}

	// Create player for real-time playback
	player, err := vtm.NewVTMPlayer(*musicFile, vtm.DefaultSampleRate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating player: %v\n", err)
		os.Exit(1)
	}

	// Load and start music if specified
	player.Play()
	defer func() {
		if player.IsPlaying() {
			player.Stop()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Printf("Signal received, stopping music\n")
		player.Stop()
	}()

	fmt.Printf("\n▶️  Playing... (Press Ctrl+C to stop)\n\n")

	// Progress indicator
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			if player.IsPlaying() && !player.IsDone() {
				elapsed := time.Since(startTime)
				fmt.Printf("\r⏱️  %02d:%02d ", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
			}
		default:
			if !player.IsDone() && player.IsPlaying() {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			// Done
			elapsed := time.Since(startTime)
			fmt.Printf("\r✅ Finished - %02d:%02d                    \n", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
			return
		}
	}
	runtime.KeepAlive(player)
}

func formatNote(note int) string {
	noteNames := []string{"C-", "C#", "D-", "D#", "E-", "F-", "F#", "G-", "G#", "A-", "A#", "B-"}
	octave := note / 12
	noteInOctave := note % 12
	return fmt.Sprintf("%s%d", noteNames[noteInOctave], octave)
}
