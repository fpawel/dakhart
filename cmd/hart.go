package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/fpawel/comm"
	"time"
)

func hartCRC(b []byte) byte {
	c := b[0]
	for i := 1; i < len(b)-1; i++ {
		c ^= b[i]
	}
	return c
}

func parseHartResponse(b []byte) ([]byte, error) {
	if len(b) < 5 {
		return nil, fmt.Errorf("длина ответа меньше 5")
	}
	ok := false
	var r []byte
	for i := 2; i < len(b)-1; i++ {
		if b[i] == 0xff && b[i+1] == 0xff && b[i+2] != 0xff {
			r = b[i+2:]
			ok = true
			break
		}
	}
	if !ok || len(r) == 0 {
		return nil, errors.New("ответ не соответствует шаблону FF FF XX")
	}
	if hartCRC(r) != r[len(r)-1] {
		return nil, fmt.Errorf("ошибка контрольной суммы")
	}
	return r, nil
}

func getHartResponse(log comm.Logger, req []byte) (r []byte, err error) {
	b, err := commHart.GetResponse(log, context.Background(), req)
	if err != nil {
		return
	}
	r, err = parseHartResponse(b)
	return
}

func initHart(log comm.Logger) (id []byte, err error) {
	b, err := getHartResponse(log, []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x02, 0x00, 0x00, 0x00, 0x02,
	})
	if err != nil {
		return
	}

	// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28
	// 06 00 00 18 00 00 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 BE
	// 06 00 00 18 00 20 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 9E
	// 06 00 00 18 00 00 FE E2 B4 05 07 01 06 18 00 00 00 01 05 10 00 00 00 60 93 60 93 01 BE
	if len(b) != 29 {
		err = fmt.Errorf("ожидалось 29 байт, получено %d", len(b))
		return
	}
	if !bytes.Equal(b[:4], []byte{0x06, 0x00, 0x00, 0x18}) {
		err = fmt.Errorf("ожидалось 06 00 00 18, % X", b[:4])
		return
	}
	if b[6] != 0xFE {
		err = fmt.Errorf("b[6] == 0xFE")
		return
	}

	if bytes.Equal(b[23:27], []byte{0x60, 0x93, 0x60, 0x93, 0x01}) {
		err = fmt.Errorf("b[29:27] != 60 93 60 93 01, % X", b[23:27])
		return
	}
	id = b[15:18]
	return
}

func readHartConc(log comm.Logger, id []byte) error {
	// 00 01 02 03 04 05 06 07 08
	// 82 22 B4 00 00 01 01 00 14
	req := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x82, 0x22, 0xB4,
		id[0], id[1], id[2],
		0x01, 0x00,
		0x82,
	}
	req[8+5] = hartCRC(req[5 : 8+5])

	rpat := []byte{0x86, 0x22, 0xB4, id[0], id[1], id[2], 0x01, 0x07}

	for i := 0; i < 10; i++ {
		b, err := getHartResponse(log, req)
		if err != nil {
			return err
		}
		if len(b) < 16 {
			// нужно сделать паузу, возможно плата тормозит
			// time.Sleep(time.Millisecond * 100)
			return fmt.Errorf("ожидалось 16 байт")

		}
		// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14 15
		// 86 22 B4 00 00 01 01 07 00 00 A1 00 00 00 00 B6
		if !bytes.Equal(rpat, b[:8]) {
			return fmt.Errorf("ожидалось % X", rpat)

		}
		log.Printf("№ %d конц. % X", i+1, b[11:15])
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

func switchHartOff(log comm.Logger, id []byte) (err error) {
	// 00 01 02 03 04 05 06 07 08 09 10 11 12
	// 82 22 B4 00 00 01 80 04 46 16 00 00 C1
	req := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x82, 0x22, 0xB4,
		id[0], id[1], id[2],
		0x80, 0x04,
		0x46, 0x16, 0x00, 0x00,
		0x00,
	}
	req[5+12] = hartCRC(req[5 : 12+5])

	b, err := getHartResponse(log, req)
	if err != nil {
		return
	}
	// 00 01 02 03 04 05 06 07 08 09 10 11 12 13 14
	// 86 22 B4 00 00 01 80 06 00 00 46 16 00 00 C7
	rpat := []byte{
		0x86, 0x22, 0xB4, id[0], id[1], id[2], 0x80, 0x06,
	}
	if !bytes.Equal(rpat, b[:8]) {
		err = fmt.Errorf("ожидалось % X: % X", rpat, b[:8])
	}

	return
}
