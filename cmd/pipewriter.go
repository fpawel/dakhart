package main

import (
	"encoding/binary"
	"net"
	"os"
)

type pipeWriter struct {
	conn net.Conn
	addr byte
	level uint32
}

func (x pipeWriter) Write(b []byte) (n int, err error) {
	defer func() {
		if err != nil {
			n, err = os.Stdout.Write(b)
		}
	}()

	// отправка - адрес
	if n, err = x.conn.Write([]byte{x.addr}); err != nil {
		return
	}

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, x.level)

	// отправка - уровень сообщения
	if n, err = x.conn.Write(bs); err != nil {
		return
	}

	bt, err := utf8ToWindows1251(b)
	if err != nil {
		return
	}
	bs = make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(bt)))
	// отправка - длина сообщения
	if _, err = x.conn.Write(bs); err != nil {
		return
	}
	// отправка - содержимое сообщения
	if _, err = x.conn.Write(bt); err != nil {
		return
	}
	return
}
