package cart

import (
	"encoding/binary"
	"io"
)

// NewMBC3 returns a new MBC3 memory controller.
func NewMBC3(data []byte) BankingController {
	return &MBC3{
		BaseMBC: BaseMBC{
			Rom:     data,
			RomBank: 1,
			Ram:     make([]byte, 0x8000),
		},
		Rtc:        make([]byte, 0x10),
		LatchedRtc: make([]byte, 0x10),
	}
}

// MBC3 is a GameBoy cartridge that supports rom and ram banking and possibly
// a real time clock (RTC).
type MBC3 struct {
	BaseMBC
	RamBank uint32

	Rtc        []byte
	LatchedRtc []byte
	Latched    bool
}

// Read returns a value at a memory address in the ROM.
func (r *MBC3) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.Rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.Rom[uint32(address-0x4000)+(r.RomBank*0x4000)] // Use selected rom bank
	default:
		if r.RamBank >= 0x4 {
			if r.Latched {
				return r.LatchedRtc[r.RamBank]
			}
			return r.Rtc[r.RamBank]
		}
		return r.Ram[(0x2000*r.RamBank)+uint32(address-0xA000)] // Use selected ram bank
	}
}

// WriteROM attempts to switch the ROM or RAM bank.
func (r *MBC3) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		r.RamEnabled = (value & 0xA) != 0
	case address < 0x4000:
		// ROM bank number (lower 5)
		r.RomBank = uint32(value & 0x7F)
		if r.RomBank == 0x00 {
			r.RomBank++
		}
	case address < 0x6000:
		r.RamBank = uint32(value)
	case address < 0x8000:
		if value == 0x1 {
			r.Latched = false
		} else if value == 0x0 {
			r.Latched = true
			copy(r.Rtc, r.LatchedRtc)
		}
	}
}

// WriteRAM writes data to the ram or RTC if it is enabled.
func (r *MBC3) WriteRAM(address uint16, value byte) {
	if r.RamEnabled {
		if r.RamBank >= 0x4 {
			r.Rtc[r.RamBank] = value
		} else {
			r.Ram[(0x2000*r.RamBank)+uint32(address-0xA000)] = value
		}
	}
}

// GetSaveData returns the save data for this banking controller.
func (r *MBC3) GetSaveData() []byte {
	data := make([]byte, len(r.Ram))
	copy(data, r.Ram)
	return data
}

// LoadSaveData loads the save data into the cartridge.
func (r *MBC3) LoadSaveData(data []byte) {
	r.Ram = data
}

// SaveState saves the state of the banking controller.
func (r *MBC3) SaveState(writer io.Writer) error {
	// Write BaseMBC
	if err := r.BaseMBC.SaveState(writer); err != nil {
		return err
	}

	// Write rambank
	_, err := writer.Write([]byte{byte(r.RamBank)})
	if err != nil {
		return err
	}

	// Write rtc
	_, err = writer.Write(r.Rtc)
	if err != nil {
		return err
	}

	// Write latched rtc
	_, err = writer.Write(r.LatchedRtc)
	if err != nil {
		return err
	}

	// Write latched
	ltch := byte(0)
	if r.Latched {
		ltch = 1
	}
	_, err = writer.Write([]byte{byte(ltch)})
	return err
}

// LoadState loads the state of the banking controller.
func (r *MBC3) LoadState(reader io.Reader) error {
	// Read BaseMBC
	if err := r.BaseMBC.LoadState(reader); err != nil {
		return err
	}

	// Read rambank
	var tmp byte
	if err := binary.Read(reader, binary.LittleEndian, &tmp); err != nil {
		return err
	}
	r.RamBank = uint32(tmp)

	// Read rtc
	_, err := reader.Read(r.Rtc)
	if err != nil {
		return err
	}

	// Read latched rtc
	_, err = reader.Read(r.LatchedRtc)
	if err != nil {
		return err
	}

	// Read latched
	var ltch byte
	if err := binary.Read(reader, binary.LittleEndian, &ltch); err != nil {
		return err
	}
	r.Latched = ltch == 1
	return nil
}
