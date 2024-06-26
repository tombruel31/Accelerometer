// Package i2c provides low level control over the linux i2c bus.
//
// Before usage you should load the i2c-dev kernel module
//
//      sudo modprobe i2c-dev
//
// Each i2c bus can address 127 independent i2c devices, and most
// linux systems contain several buses.
package i2c

import (
	"fmt"
	"os"
	"syscall"
)

const (
	I2C_SLAVE = 0x20
)

// I2C represents a connection to an i2c device.
type I2C struct {
	rc *os.File
}

// New opens a connection to an i2c device.
func NewI2C(addr uint8, bus int) (*I2C, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := ioctl(f.Fd(), I2C_SLAVE, uintptr(addr)); err != nil {
		return nil, err
	}
	v := &I2C{rc: f}
	return v, nil
}

func (v *I2C) write(buf []byte) (int, error) {
	return v.rc.Write(buf)
}

// Write sends buf to the remote i2c device. The interpretation of
// the message is implementation dependant.
func (v *I2C) WriteBytes(buf []byte) (int, error) {
	return v.write(buf)
}

// func (v *I2C) WriteByte(b byte) (int, error) {
// 	return v.write([]byte{b})
// }

func (v *I2C) read(buf []byte) (int, error) {
	return v.rc.Read(buf)
}

func (v *I2C) ReadBytes(buf []byte) (int, error) {
	n, err := v.read(buf)
	return n, err
}

func (v *I2C) Close() error {
	return v.rc.Close()
}

// SMBus (System Management Bus) protocol over I2C.
// Read count of n byte's sequence from i2c device
// starting from reg address.
func (v *I2C) ReadRegBytes(reg byte, n int) ([]byte, int, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return nil, 0, err
	}
	buf := make([]byte, n)
	c, err := v.ReadBytes(buf)
	if err != nil {
		return nil, 0, err
	}
	return buf, c, nil

}

// SMBus (System Management Bus) protocol over I2C.
// Read byte from i2c device register specified in reg.
func (v *I2C) ReadRegU8(reg byte) (byte, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 1)
	_, err = v.ReadBytes(buf)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// SMBus (System Management Bus) protocol over I2C.
// Write byte to i2c device register specified in reg.
func (v *I2C) WriteRegU8(reg byte, value byte) error {
	buf := []byte{reg, value}
	_, err := v.WriteBytes(buf)
	if err != nil {
		return err
	}
	return nil
}

// SMBus (System Management Bus) protocol over I2C.
// Read unsigned big endian word (16 bits) from i2c device
// starting from address specified in reg.
func (v *I2C) ReadRegU16BE(reg byte) (uint16, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 2)
	_, err = v.ReadBytes(buf)
	if err != nil {
		return 0, err
	}
	w := uint16(buf[0])<<8 + uint16(buf[1])
	return w, nil
}

// SMBus (System Management Bus) protocol over I2C.
// Read unsigned little endian word (16 bits) from i2c device
// starting from address specified in reg.
func (v *I2C) ReadRegU16LE(reg byte) (uint16, error) {
	w, err := v.ReadRegU16BE(reg)
	if err != nil {
		return 0, err
	}
	// exchange bytes
	w = (w&0xFF)<<8 + w>>8
	return w, nil
}

// SMBus (System Management Bus) protocol over I2C.
// Read signed big endian word (16 bits) from i2c device
// starting from address specified in reg.
func (v *I2C) ReadRegS16BE(reg byte) (int16, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 2)
	_, err = v.ReadBytes(buf)
	if err != nil {
		return 0, err
	}
	w := int16(buf[0])<<8 + int16(buf[1])
	return w, nil
}

// SMBus (System Management Bus) protocol over I2C.
// Read unsigned little endian word (16 bits) from i2c device
// starting from address specified in reg.
func (v *I2C) ReadRegS16LE(reg byte) (int16, error) {
	w, err := v.ReadRegS16BE(reg)
	if err != nil {
		return 0, err
	}
	// exchange bytes
	w = (w&0xFF)<<8 + w>>8
	return w, nil

}

// SMBus (System Management Bus) protocol over I2C.
// Write unsigned big endian word (16 bits) value to i2c device
// starting from address specified in reg.
func (v *I2C) WriteRegU16BE(reg byte, value uint16) error {
	buf := []byte{reg, byte((value & 0xFF00) >> 8), byte(value & 0xFF)}
	_, err := v.WriteBytes(buf)
	if err != nil {
		return err
	}
	return nil
}

// SMBus (System Management Bus) protocol over I2C.
// Write unsigned big endian word (16 bits) value to i2c device
// starting from address specified in reg.
func (v *I2C) WriteRegU16LE(reg byte, value uint16) error {
	w := (value*0xFF00)>>8 + value<<8
	return v.WriteRegU16BE(reg, w)
}

// SMBus (System Management Bus) protocol over I2C.
// Write signed big endian word (16 bits) value to i2c device
// starting from address specified in reg.
func (v *I2C) WriteRegS16BE(reg byte, value int16) error {
	buf := []byte{reg, byte((uint16(value) & 0xFF00) >> 8), byte(value & 0xFF)}
	_, err := v.WriteBytes(buf)
	if err != nil {
		return err
	}
	return nil
}

// SMBus (System Management Bus) protocol over I2C.
// Write signed big endian word (16 bits) value to i2c device
// starting from address specified in reg.
func (v *I2C) WriteRegS16LE(reg byte, value int16) error {
	w := int16((uint16(value)*0xFF00)>>8) + value<<8
	return v.WriteRegS16BE(reg, w)
}

func ioctl(fd, cmd, arg uintptr) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if err != 0 {
		return err
	}
	return nil
}
