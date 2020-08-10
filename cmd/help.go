package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"strconv"
)

func read3(log comm.Logger, addr modbus.Addr, var3 modbus.Var) (float64, error){
	return modbus.Read3Value(log, context.Background(), commModbus, addr, var3, modbus.BCD)
}

func write32(log comm.Logger, addr modbus.Addr, cmd modbus.DevCmd, value float64) error{
	return modbus.RequestWrite32{
		Addr:      addr,
		ProtoCmd:  0x10,
		DeviceCmd: cmd,
		Format:    modbus.BCD,
		Value:     value,
	}.GetResponse(log, context.Background(), commModbus)
}



func utf8ToWindows1251(b []byte) (r []byte, err error) {
	buf := new(bytes.Buffer)
	wToWin1251 := transform.NewWriter(buf, charmap.Windows1251.NewEncoder())
	_, err = io.Copy(wToWin1251, bytes.NewReader(b))
	if err == nil {
		r = buf.Bytes()
	}
	return
}

func stringToAddr(s string) (modbus.Addr, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if v < 0 || v > 255 {
		return 0, errors.New("value must be 0..255")
	}
	return modbus.Addr(v), nil
}