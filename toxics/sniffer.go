package toxics

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

type SnifferToxic struct {
	Path        string `json:"path"`
	file        *os.File
	stopWriting bool
}

func epochNowString() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

func getOutputFile(path string) (*os.File, error) {
	if path == "" {
		path = "/tmp/" + epochNowString() + ".txt"
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("LOG PREFIX Failed to open file %+v\n", err)
		return nil, err
	}

	return f, nil
}

func (t *SnifferToxic) attack(buf []byte) MitmCallback {
	if !t.stopWriting {
		_, writeErr := t.file.Write([]byte(hex.Dump(buf)))
		if writeErr != nil {
			fmt.Printf("LOG PREFIX Error writing %+v\n", writeErr)
			t.stopWriting = true
		}
	}

	return MitmCallback{
		WriteBack: true,
	}
}

func (t *SnifferToxic) Pipe(stub *ToxicStub) {
	t.stopWriting = false
	file, openErr := getOutputFile(t.Path)
	if openErr != nil {
		fmt.Printf("LOG PREFIX Failed to open file %s. Continuing with NOOP MITM\n", t.Path)
		MitmPipe(stub, new(MitmToxic))
		return
	}

	t.file = file
	MitmPipe(stub, t)
}

func init() {
	Register("sniffer", new(SnifferToxic))
}
