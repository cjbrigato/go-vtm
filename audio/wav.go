package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/cjbrigato/go-vtm/tracker"
)

// WAVWriter handles writing audio samples to a WAV file
type WAVWriter struct {
	file       *os.File
	sampleRate int
	numSamples int
}

// NewWAVWriter creates a new WAV file writer
func NewWAVWriter(filename string, sampleRate int) (*WAVWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &WAVWriter{
		file:       file,
		sampleRate: sampleRate,
		numSamples: 0,
	}, nil
}

// WriteHeader writes the WAV file header
func (w *WAVWriter) writeHeader() error {
	// RIFF header
	if err := binary.Write(w.file, binary.LittleEndian, []byte("RIFF")); err != nil {
		return err
	}

	// File size (placeholder, will be updated later)
	fileSize := uint32(36 + w.numSamples*2*2) // 2 channels, 2 bytes per sample
	if err := binary.Write(w.file, binary.LittleEndian, fileSize); err != nil {
		return err
	}

	// WAVE header
	if err := binary.Write(w.file, binary.LittleEndian, []byte("WAVE")); err != nil {
		return err
	}

	// fmt chunk
	if err := binary.Write(w.file, binary.LittleEndian, []byte("fmt ")); err != nil {
		return err
	}

	// fmt chunk size (16 for PCM)
	if err := binary.Write(w.file, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}

	// Audio format (1 = PCM)
	if err := binary.Write(w.file, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}

	// Number of channels (2 = stereo)
	if err := binary.Write(w.file, binary.LittleEndian, uint16(2)); err != nil {
		return err
	}

	// Sample rate
	if err := binary.Write(w.file, binary.LittleEndian, uint32(w.sampleRate)); err != nil {
		return err
	}

	// Byte rate (sample rate * channels * bytes per sample)
	byteRate := uint32(w.sampleRate * 2 * 2)
	if err := binary.Write(w.file, binary.LittleEndian, byteRate); err != nil {
		return err
	}

	// Block align (channels * bytes per sample)
	if err := binary.Write(w.file, binary.LittleEndian, uint16(4)); err != nil {
		return err
	}

	// Bits per sample
	if err := binary.Write(w.file, binary.LittleEndian, uint16(16)); err != nil {
		return err
	}

	// data chunk
	if err := binary.Write(w.file, binary.LittleEndian, []byte("data")); err != nil {
		return err
	}

	// data chunk size
	dataSize := uint32(w.numSamples * 2 * 2) // 2 channels, 2 bytes per sample
	if err := binary.Write(w.file, binary.LittleEndian, dataSize); err != nil {
		return err
	}

	return nil
}

// WriteSample writes a single stereo sample to the WAV file (expects sample in range -1.0 to 1.0)
func (w *WAVWriter) WriteSample(left, right float64) error {
	// Convert float64 [-1.0, 1.0] to int16
	leftSample := int16(left * 32767.0)
	rightSample := int16(right * 32767.0)

	if err := binary.Write(w.file, binary.LittleEndian, leftSample); err != nil {
		return err
	}

	if err := binary.Write(w.file, binary.LittleEndian, rightSample); err != nil {
		return err
	}

	w.numSamples++
	return nil
}

// Close finalizes and closes the WAV file
func (w *WAVWriter) Close() error {
	// Update file size in header
	if _, err := w.file.Seek(4, io.SeekStart); err != nil {
		return err
	}
	fileSize := uint32(36 + w.numSamples*2*2)
	if err := binary.Write(w.file, binary.LittleEndian, fileSize); err != nil {
		return err
	}

	// Update data chunk size
	if _, err := w.file.Seek(40, io.SeekStart); err != nil {
		return err
	}
	dataSize := uint32(w.numSamples * 2 * 2)
	if err := binary.Write(w.file, binary.LittleEndian, dataSize); err != nil {
		return err
	}

	return w.file.Close()
}

// RenderToWAV renders an entire tracker module to a WAV file
func RenderToWAV(module *tracker.TrackerModule, sampleRate int, filename string) error {
	// Create player
	player := NewPlayer(module, float64(sampleRate))

	// Create WAV writer with temporary header
	wavWriter, err := NewWAVWriter(filename, sampleRate)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %v", err)
	}

	// Write placeholder header (will be updated at the end)
	if err := wavWriter.writeHeader(); err != nil {
		wavWriter.file.Close()
		return fmt.Errorf("failed to write WAV header: %v", err)
	}

	// Generate and write samples
	for !player.IsDone() {
		sample := player.Next()
		
		// Clamp sample to prevent clipping
		sample = math.Max(-1.0, math.Min(1.0, sample))
		
		// Write stereo (same sample for both channels for now)
		if err := wavWriter.WriteSample(sample, sample); err != nil {
			wavWriter.file.Close()
			return fmt.Errorf("failed to write sample: %v", err)
		}
	}

	// Close and finalize
	if err := wavWriter.Close(); err != nil {
		return fmt.Errorf("failed to close WAV file: %v", err)
	}

	return nil
}

