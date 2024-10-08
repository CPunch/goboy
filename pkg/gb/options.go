package gb

import "io"

// GameboyOption is an option for the Gameboy execution.
type GameboyOption func(o *gameboyOptions)

type gameboyOptions struct {
	sound   bool
	cgbMode bool
	saver   io.ReadWriter // Save location

	// Callback when the serial port is written to
	transferFunction func(byte)
}

// DebugFlags are flags which can be set to alter the execution of the Gameboy.
type DebugFlags struct {
	// HideSprites turns off rendering of sprites to the display.
	HideSprites bool

	// HideBackground turns off rendering of background tiles to the display.
	HideBackground bool

	// OutputOpcodes will log the current opcode to the console on each tick.
	// This will slow down execution massively so is only used for debugging
	// issues with the emulation.
	OutputOpcodes bool
}

func (flags *DebugFlags) toggleBackGround() {
	flags.HideBackground = !flags.HideBackground
}

func (flags *DebugFlags) toggleSprites() {
	flags.HideSprites = !flags.HideSprites
}

func (flags *DebugFlags) toggleOutputOpCode() {
	flags.OutputOpcodes = !flags.OutputOpcodes
}

// WithCGBEnabled runs the Gameboy with cgb mode enabled.
func WithCGBEnabled() GameboyOption {
	return func(o *gameboyOptions) {
		o.cgbMode = true
	}
}

// WithSound runs the Gameboy with sound output.
func WithSound() GameboyOption {
	return func(o *gameboyOptions) {
		o.sound = true
	}
}

func WithSaveFile(saver io.ReadWriter) GameboyOption {
	return func(o *gameboyOptions) {
		o.saver = saver
	}
}

// WithTransferFunction provides a function to callback on when the serial transfer
// address is written to.
func WithTransferFunction(transfer func(byte)) GameboyOption {
	return func(o *gameboyOptions) {
		o.transferFunction = transfer
	}
}
