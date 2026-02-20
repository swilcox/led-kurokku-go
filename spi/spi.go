package spi

import (
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// Device wraps a periph.io SPI connection.
type Device struct {
	port spi.PortCloser
	conn spi.Conn
}

// Open initializes the host drivers and opens the named SPI bus.
// Pass "" for the default bus (typically /dev/spidev0.0).
func Open(busName string, speedHz int64) (*Device, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	port, err := spireg.Open(busName)
	if err != nil {
		return nil, err
	}

	conn, err := port.Connect(physic.Frequency(speedHz)*physic.Hertz, spi.Mode0, 8)
	if err != nil {
		port.Close()
		return nil, err
	}

	return &Device{port: port, conn: conn}, nil
}

// Tx sends data over SPI.
func (d *Device) Tx(w []byte) error {
	return d.conn.Tx(w, nil)
}

// Close releases the SPI port.
func (d *Device) Close() error {
	return d.port.Close()
}
