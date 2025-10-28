package synth

// FMAlgorithm defines how operators are connected
type FMAlgorithm int

const (
	// FM2OpSimple: Op2 modulates Op1 (carrier)
	// Op2 -> Op1 -> Out
	FM2OpSimple FMAlgorithm = iota
	
	// FM2OpParallel: Both operators in parallel (additive)
	// Op1 -> Out
	// Op2 -> Out
	FM2OpParallel
	
	// FM3OpStack: Op3 -> Op2 -> Op1 -> Out
	FM3OpStack
	
	// FM4OpPiano: Classic piano algorithm
	// Op4 -> Op3 \
	//             Op1 -> Out
	// Op2 -------/
	FM4OpPiano
)

// FMInstrument implements multi-operator FM synthesis
type FMInstrument struct {
	operators       []*FMOperator
	algorithm       FMAlgorithm
	modulationIndex float64 // Controls modulation depth (brightness)
	sampleRate      float64
	baseFrequency   float64
	volume          float64
}

// NewFMInstrument creates a new FM instrument with specified number of operators
func NewFMInstrument(numOperators int, algorithm FMAlgorithm, sampleRate float64) *FMInstrument {
	operators := make([]*FMOperator, numOperators)
	for i := range operators {
		operators[i] = NewFMOperator(sampleRate)
	}
	
	return &FMInstrument{
		operators:       operators,
		algorithm:       algorithm,
		modulationIndex: 2.0,
		sampleRate:      sampleRate,
		volume:          1.0,
	}
}

// SetOperatorRatio sets the frequency ratio for an operator
func (fm *FMInstrument) SetOperatorRatio(opIndex int, ratio float64) {
	if opIndex < len(fm.operators) {
		fm.operators[opIndex].SetRatio(ratio)
	}
}

// SetOperatorEnvelope sets the ADSR envelope for an operator
func (fm *FMInstrument) SetOperatorEnvelope(opIndex int, attack, decay, sustain, release float64) {
	if opIndex < len(fm.operators) {
		fm.operators[opIndex].SetEnvelope(attack, decay, sustain, release)
	}
}

// SetOperatorLevel sets the output level for an operator
func (fm *FMInstrument) SetOperatorLevel(opIndex int, level float64) {
	if opIndex < len(fm.operators) {
		fm.operators[opIndex].SetOutputLevel(level)
	}
}

// SetModulationIndex sets the modulation depth (affects brightness)
func (fm *FMInstrument) SetModulationIndex(index float64) {
	fm.modulationIndex = index
}

// NoteOn triggers all operators with the given frequency
func (fm *FMInstrument) NoteOn(note int, velocity float64) {
	fm.baseFrequency = NoteToFrequency(note)
	fm.volume = velocity
	
	// Set frequency for all operators based on their ratios
	for _, op := range fm.operators {
		op.SetFrequency(fm.baseFrequency)
		op.Trigger()
	}
}

// NoteOff releases all operators
func (fm *FMInstrument) NoteOff() {
	for _, op := range fm.operators {
		op.Release()
	}
}

// Next generates the next audio sample based on the algorithm
func (fm *FMInstrument) Next() float64 {
	if len(fm.operators) == 0 {
		return 0.0
	}
	
	var output float64
	
	switch fm.algorithm {
	case FM2OpSimple:
		// Op2 modulates Op1
		if len(fm.operators) >= 2 {
			modulator := fm.operators[1].Next(0) * fm.modulationIndex
			carrier := fm.operators[0].Next(modulator)
			output = carrier
		}
		
	case FM2OpParallel:
		// Both operators in parallel (additive)
		if len(fm.operators) >= 2 {
			output = (fm.operators[0].Next(0) + fm.operators[1].Next(0)) * 0.5
		}
		
	case FM3OpStack:
		// Op3 -> Op2 -> Op1
		if len(fm.operators) >= 3 {
			mod2 := fm.operators[2].Next(0) * fm.modulationIndex
			mod1 := fm.operators[1].Next(mod2) * fm.modulationIndex
			carrier := fm.operators[0].Next(mod1)
			output = carrier
		}
		
	case FM4OpPiano:
		// Classic 4-operator piano algorithm
		// Op4 -> Op3 -> Op1 (modulation chain 1)
		// Op2 -> Op1 (modulation chain 2)
		if len(fm.operators) >= 4 {
			// Chain 1: Op4 modulates Op3
			mod4 := fm.operators[3].Next(0) * fm.modulationIndex * 0.5
			mod3 := fm.operators[2].Next(mod4) * fm.modulationIndex
			
			// Chain 2: Op2 modulates Op1
			mod2 := fm.operators[1].Next(0) * fm.modulationIndex * 0.3
			
			// Op1 is carrier, modulated by both chains
			carrier := fm.operators[0].Next(mod3 + mod2)
			output = carrier
		}
	}
	
	return output * fm.volume
}

// IsActive returns true if any operator is still active
func (fm *FMInstrument) IsActive() bool {
	for _, op := range fm.operators {
		if op.IsActive() {
			return true
		}
	}
	return false
}

// NewPianoFMInstrument creates a preset FM instrument configured for piano sounds
func NewPianoFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(4, FM4OpPiano, sampleRate)
	
	// Operator 1: Carrier - fundamental frequency
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.001, 0.3, 0.0, 0.2) // Fast attack, medium decay, no sustain
	fm.SetOperatorLevel(0, 1.0)
	
	// Operator 2: Modulator - adds brightness
	fm.SetOperatorRatio(1, 2.0) // One octave up
	fm.SetOperatorEnvelope(1, 0.001, 0.15, 0.0, 0.1) // Shorter envelope for brightness
	fm.SetOperatorLevel(1, 0.7)
	
	// Operator 3: Modulator - adds harmonic complexity
	fm.SetOperatorRatio(2, 3.5) // Slightly inharmonic
	fm.SetOperatorEnvelope(2, 0.001, 0.2, 0.0, 0.15)
	fm.SetOperatorLevel(2, 0.5)
	
	// Operator 4: Modulator - adds attack transient
	fm.SetOperatorRatio(3, 5.0) // High harmonic
	fm.SetOperatorEnvelope(3, 0.0005, 0.05, 0.0, 0.05) // Very fast envelope
	fm.SetOperatorLevel(3, 0.4)
	
	// Moderate modulation index for piano-like timbre
	fm.SetModulationIndex(1.5)
	
	return fm
}

// NewElectricPianoFMInstrument creates an electric piano preset
func NewElectricPianoFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(2, FM2OpSimple, sampleRate)
	
	// Classic DX7 E.Piano algorithm
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.002, 0.4, 0.3, 0.3)
	fm.SetOperatorLevel(0, 1.0)
	
	fm.SetOperatorRatio(1, 14.0) // High ratio for bell-like tone
	fm.SetOperatorEnvelope(1, 0.001, 0.2, 0.0, 0.2)
	fm.SetOperatorLevel(1, 0.8)
	
	fm.SetModulationIndex(3.0) // Higher modulation for brightness
	
	return fm
}

// NewFMBassFMInstrument creates a powerful FM bass preset
func NewFMBassFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(3, FM3OpStack, sampleRate)
	
	// Operator 1: Carrier - sub bass fundamental
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.001, 0.2, 0.7, 0.15)
	fm.SetOperatorLevel(0, 1.0)
	
	// Operator 2: Modulator - adds growl
	fm.SetOperatorRatio(1, 2.01) // Slightly detuned for fatness
	fm.SetOperatorEnvelope(1, 0.001, 0.15, 0.5, 0.1)
	fm.SetOperatorLevel(1, 0.9)
	
	// Operator 3: Modulator - attack bite
	fm.SetOperatorRatio(2, 4.0)
	fm.SetOperatorEnvelope(2, 0.0005, 0.08, 0.2, 0.05)
	fm.SetOperatorLevel(2, 0.7)
	
	fm.SetModulationIndex(3.5) // High modulation for aggressive bass
	
	return fm
}

// NewFMLeadFMInstrument creates a cutting lead synth preset
func NewFMLeadFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(2, FM2OpSimple, sampleRate)
	
	// Operator 1: Carrier
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.005, 0.1, 0.8, 0.15)
	fm.SetOperatorLevel(0, 1.0)
	
	// Operator 2: Modulator - bright harmonics
	fm.SetOperatorRatio(1, 3.0) // Perfect fifth harmonic
	fm.SetOperatorEnvelope(1, 0.002, 0.08, 0.6, 0.1)
	fm.SetOperatorLevel(1, 1.0)
	
	fm.SetModulationIndex(5.0) // Very high for cutting lead
	
	return fm
}

// NewFMBrassFMInstrument creates a brass-like preset
func NewFMBrassFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(3, FM3OpStack, sampleRate)
	
	// Operator 1: Carrier
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.02, 0.15, 0.7, 0.2)
	fm.SetOperatorLevel(0, 1.0)
	
	// Operator 2: Modulator - brass harmonics
	fm.SetOperatorRatio(1, 1.5) // Fifth
	fm.SetOperatorEnvelope(1, 0.015, 0.12, 0.6, 0.15)
	fm.SetOperatorLevel(1, 0.8)
	
	// Operator 3: Modulator - brightness
	fm.SetOperatorRatio(2, 3.0)
	fm.SetOperatorEnvelope(2, 0.01, 0.1, 0.5, 0.12)
	fm.SetOperatorLevel(2, 0.6)
	
	fm.SetModulationIndex(2.0)
	
	return fm
}

// NewFMBellFMInstrument creates a metallic bell preset
func NewFMBellFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(2, FM2OpSimple, sampleRate)
	
	// Operator 1: Carrier
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.001, 0.5, 0.2, 0.4)
	fm.SetOperatorLevel(0, 1.0)
	
	// Operator 2: Modulator - inharmonic for metallic sound
	fm.SetOperatorRatio(1, 11.0) // High, inharmonic ratio
	fm.SetOperatorEnvelope(1, 0.001, 0.3, 0.0, 0.3)
	fm.SetOperatorLevel(1, 0.9)
	
	fm.SetModulationIndex(4.0)
	
	return fm
}

// NewFMArpFMInstrument creates a fast arpeggio/chiptune preset
func NewFMArpFMInstrument(sampleRate float64) *FMInstrument {
	fm := NewFMInstrument(2, FM2OpSimple, sampleRate)
	
	// Operator 1: Carrier
	fm.SetOperatorRatio(0, 1.0)
	fm.SetOperatorEnvelope(0, 0.001, 0.05, 0.3, 0.05)
	fm.SetOperatorLevel(0, 1.0)
	
	// Operator 2: Modulator - chippy harmonics
	fm.SetOperatorRatio(1, 2.0)
	fm.SetOperatorEnvelope(1, 0.001, 0.03, 0.1, 0.03)
	fm.SetOperatorLevel(1, 0.85)
	
	fm.SetModulationIndex(2.5) // Medium-high for chip sound
	
	return fm
}

