package cart

// NewMBC5 returns a new MBC5 memory controller.
func NewMBC5(data []byte) BankingController {
	return &MBC5{
		BaseMBC: BaseMBC{
			Rom:     data,
			RomBank: 1,
			Ram:     make([]byte, 0x20000),
		},
	}
}

// MBC5 is a GameBoy cartridge that supports rom and ram banking.
type MBC5 struct {
	BaseMBC
	RamBank uint32
}

// Read returns a value at a memory address in the ROM.
func (r *MBC5) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.Rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.Rom[uint32(address-0x4000)+(r.RomBank*0x4000)] // Use selected rom bank
	default:
		return r.Ram[(0x2000*r.RamBank)+uint32(address-0xA000)] // Use selected ram bank
	}
}

// WriteROM attempts to switch the ROM or RAM bank.
func (r *MBC5) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		if value&0xF == 0xA {
			r.RamEnabled = true
		} else if value&0xF == 0x0 {
			r.RamEnabled = false
		}
	case address < 0x3000:
		// ROM bank number
		r.RomBank = (r.RomBank & 0x100) | uint32(value)
	case address < 0x4000:
		// ROM/RAM banking
		r.RomBank = (r.RomBank & 0xFF) | uint32(value&0x01)<<8
	case address < 0x6000:
		r.RamBank = uint32(value & 0xF)
	}
}

// WriteRAM writes data to the ram if it is enabled.
func (r *MBC5) WriteRAM(address uint16, value byte) {
	if r.RamEnabled {
		r.Ram[(0x2000*r.RamBank)+uint32(address-0xA000)] = value
	}
}

// GetSaveData returns the save data for this banking controller.
func (r *MBC5) GetSaveData() []byte {
	data := make([]byte, len(r.Ram))
	copy(data, r.Ram)
	return data
}

// LoadSaveData loads the save data into the cartridge.
func (r *MBC5) LoadSaveData(data []byte) {
	r.Ram = data
}
