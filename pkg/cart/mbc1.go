package cart

import (
	"encoding/binary"
	"io"
)

// NewMBC1 returns a new MBC1 memory controller.
func NewMBC1(data []byte) BankingController {
	return &MBC1{
		BaseMBC: BaseMBC{
			Rom:     data,
			RomBank: 1,
			Ram:     make([]byte, 0x8000),
		},
	}
}

// MBC1 is a GameBoy cartridge that supports rom and ram banking.
type MBC1 struct {
	BaseMBC
	RamBank    uint32
	RomBanking bool
}

// Read returns a value at a memory address in the ROM or RAM.
func (r *MBC1) Read(address uint16) byte {
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
func (r *MBC1) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		if value&0xF == 0xA {
			r.RamEnabled = true
		} else if value&0xF == 0x0 {
			r.RamEnabled = false
		}
	case address < 0x4000:
		// ROM bank number (lower 5)
		r.RomBank = (r.RomBank & 0xe0) | uint32(value&0x1f)
		r.updateRomBankIfZero()
	case address < 0x6000:
		// ROM/RAM banking
		if r.RomBanking {
			r.RomBank = (r.RomBank & 0x1F) | uint32(value&0xe0)
			r.updateRomBankIfZero()
		} else {
			r.RamBank = uint32(value & 0x3)
		}
	case address < 0x8000:
		// ROM/RAM select mode
		r.RomBanking = value&0x1 == 0x00
		if r.RomBanking {
			r.RamBank = 0
		} else {
			r.RomBank = r.RomBank & 0x1F
		}
	}
}

// Update the romBank if it is on a value which cannot be used.
func (r *MBC1) updateRomBankIfZero() {
	if r.RomBank == 0x00 || r.RomBank == 0x20 || r.RomBank == 0x40 || r.RomBank == 0x60 {
		r.RomBank++
	}
}

// WriteRAM writes data to the ram if it is enabled.
func (r *MBC1) WriteRAM(address uint16, value byte) {
	if r.RamEnabled {
		r.Ram[(0x2000*r.RamBank)+uint32(address-0xA000)] = value
	}
}

// GetSaveData returns the save data for this banking controller.
func (r *MBC1) GetSaveData() []byte {
	data := make([]byte, len(r.Ram))
	copy(data, r.Ram)
	return data
}

// LoadSaveData loads the save data into the cartridge.
func (r *MBC1) LoadSaveData(data []byte) {
	r.Ram = data
}

// SaveState saves the state of the banking controller.
func (r *MBC1) SaveState(writer io.Writer) error {
	// Write BaseMBC
	if err := r.BaseMBC.SaveState(writer); err != nil {
		return err
	}

	// Write RomBanking
	bnk := byte(0)
	if r.RomBanking {
		bnk = 1
	}
	_, err := writer.Write([]byte{byte(bnk)})
	if err != nil {
		return err
	}

	// Write rambank
	_, err = writer.Write([]byte{byte(r.RamBank)})
	return err
}

// LoadState loads the state of the banking controller.
func (r *MBC1) LoadState(reader io.Reader) error {
	// Read BaseMBC
	if err := r.BaseMBC.LoadState(reader); err != nil {
		return err
	}

	// Read RomBanking
	var bnk byte
	if err := binary.Read(reader, binary.LittleEndian, &bnk); err != nil {
		return err
	}
	r.RomBanking = bnk == 1

	// Read rambank
	var tmp byte
	if err := binary.Read(reader, binary.LittleEndian, &tmp); err != nil {
		return err
	}
	r.RamBank = uint32(tmp)
	return nil
}
