package main

import (
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"gopkg.in/natefinch/npipe.v2"
	"os"
	"time"

	"fmt"
)

const usage = "usage: dakhart1.exe [MODBUS] [HART] [ADDR STEND] [ADDR 1]...[ADDR N]"

func main() {

	pipeListener, err := npipe.Listen(`\\.\pipe\$TestHart$`)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.ErrIfFail(pipeListener.Close)
	}()

	log.Println("pipeRunner: ожидается")
	pipeConn, err = pipeListener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.ErrIfFail(pipeConn.Close)
	}()

	log.Println("pipeRunner: соединение установлено")

	log := logPipe(log, 2, 0)
	log.Print(os.Args)

	if len(os.Args) < 5 {
		log.Fatalf("должно быть не менее пяти аргументов: %v, usage: %v", os.Args, usage)
	}

	if stendAddr, err = stringToAddr(os.Args[3]); err != nil {
		log.Fatalf("bad stend addr argument %s: %v, args: %v, usage: %v", os.Args[3], err, os.Args, usage)
	}

	var addresses []modbus.Addr

	for i, s := range os.Args[4:] {
		v, err := stringToAddr(s)
		if err != nil {
			log.Fatalf("bad address argument %d: %v: %v", i, s, v)
		}
		addresses = append(addresses, v)
	}
	comportModbus := comport.NewPort(comport.Config{
		Name:        os.Args[1],
		Baud:        9600,
		ReadTimeout: time.Millisecond,
	})
	comportHart := comport.NewPort(comport.Config{
		Name:        os.Args[2],
		Baud:        1200,
		ReadTimeout: time.Microsecond,
		Parity:      comport.ParityOdd,
		StopBits:    comport.Stop1,
	})
	
	defer func() {
		log.ErrIfFail(comportModbus.Close)
		log.ErrIfFail(comportHart.Close)
	}()

	commModbus = comm.New(comportModbus, comm.Config{
		TimeoutGetResponse: 2 * time.Second,
		TimeoutEndResponse: time.Millisecond * 50,
		MaxAttemptsRead:    2,
	})
	commHart = comm.New(comportHart, comm.Config{
		TimeoutGetResponse: 2 * time.Second,
		TimeoutEndResponse: time.Millisecond * 100,
		MaxAttemptsRead:    3,
	})

	var oks, errs []modbus.Addr

	for _, addr := range addresses {
		log := logAppendPrefixKeys(log, "адрес", addr)

		err = processAddr(log, addr)
		if err == nil {
			logPipe(log, 1, addr).Print("OK")
			oks = append(oks, addr)
		} else {
			logPipe(log, 0, addr).PrintErr(err)
			errs = append(errs, addr)
		}
	}

	if len(errs) > 0 {
		logPipe(log, 0, 0).PrintErr(fmt.Errorf("приборы, не прошедщие проверку HART протокола: %02d", errs))
	}

	if len(oks) > 0 {
		log.Printf("приборы, успешно прошедщие проверку HART протокола: %02d", oks)
	}
}

func processAddr(log comm.Logger, addr modbus.Addr) error {
	if stendAddr == 0 {
		log.Print("переключение канала стенда: пропуск операции, адресс стенда 0?")
	}
	log.Print("переключение канала стенда")
	if _,err := read3(log, stendAddr, (modbus.Var(addr) - 1) * 2); err != nil {
		return fmt.Errorf("не удалось переключить стенд: %v", err)
	}
	log.Print("включение HART протокола")

	if err := write32(log, addr, 0x80, 1000); err == nil {
		log.Print("включение HART протокола: ОК")
	} else {
		log.PrintErr( fmt.Printf("включение HART протокола: %v. Включен ранее?", err) )
	}

	log.Print("инициализация HART")
	id, err := initHart(log)
	if err != nil {
		return fmt.Errorf("инициализация HART: %v", err)
	}
	log.Print("инициализация HART: ОК")

	log.Print("запрос концентрации HART")
	err = readHartConc(log, id)
	if err != nil {
		return fmt.Errorf("запрос концентрации HART: %v", err)
	}
	log.Print("запрос концентрации HART: ОК")

	log.Print("выключение HART протокола")
	err = switchHartOff(log, id)
	if err != nil {
		return fmt.Errorf("выключение HART протокола: %v", err)
	}
	log.Print("выключение: ОК")

	log.Print("запрос концентрации MODBUS")
	if _, err = read3(log, addr, 0); err != nil {
		return fmt.Errorf("запрос концентрации MODBUS: %v", err)
	}
	log.Print("запрос концентрации MODBUS: ОК")
	return nil
}

//func testRunClientPipeApp() {
//	cmd := exec.Command(os.Getenv("GOPATH") + "/src/fpawel/hart6015/test_client/Project1.exe")
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	err := cmd.Start()
//	if err != nil {
//		log.Fatal(err)
//	}
//}




