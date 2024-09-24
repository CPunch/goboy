package cart

// NewMBC2 returns a new MBC2 memory controller.
func NewMBC2(data []byte) BankingController {
	return &MBC2{
		BaseMBC{
			Rom:     data,
			RomBank: 1,
			Ram:     make([]byte, 0x2000),
		},
	}
}

// MBC2 is a basic Gameboy cartridge.
type MBC2 struct {
	BaseMBC
}

// Read returns a value at a memory address in the ROM or RAM.
func (r *MBC2) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.Rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.Rom[uint32(address-0x4000)+(r.RomBank*0x4000)] // Use selected rom bank
	default:
		return r.Ram[address-0xA000] // Use ram
	}
}

// WriteROM attempts to switch the ROM or RAM bank.
func (r *MBC2) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		if address&0x100 == 0 {
			if value&0xF == 0xA {
				r.RamEnabled = true
			} else if value&0xF == 0x0 {
				r.RamEnabled = false
			}
		}
		return
	case address < 0x4000:
		// ROM bank number (lower 4)
		if address&0x100 == 0x100 {
			r.RomBank = uint32(value & 0xF)
			if r.RomBank == 0x00 || r.RomBank == 0x20 || r.RomBank == 0x40 || r.RomBank == 0x60 {
				r.RomBank++
			}
		}
		return
	}
}

// WriteRAM writes data to the ram if it is enabled.
func (r *MBC2) WriteRAM(address uint16, value byte) {
	if r.RamEnabled {
		r.Ram[address-0xA000] = value & 0xF
	}
}

// GetSaveData returns the save data for this banking controller.
func (r *MBC2) GetSaveData() []byte {
	data := make([]byte, len(r.Ram))
	copy(data, r.Ram)
	return data
}

// LoadSaveData loads the save data into the cartridge.
func (r *MBC2) LoadSaveData(data []byte) {
	r.Ram = data
}
