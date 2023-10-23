package toxics

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

type SnifferToxic struct {
	Path string `json:"path"`
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
	stopWriting := false
	file, openErr := getOutputFile("/tmp/eof-" + epochNowString() + t.Path + ".txt")

	if openErr != nil {
		fmt.Printf("LOG PREFIX Failed to open file")
		stopWriting = true
	}

	if !stopWriting {
		_, writeErr := file.Write([]byte(hex.Dump(buf)))
		if writeErr != nil {
			fmt.Printf("LOG PREFIX Error writing %+v\n", writeErr)
			stopWriting = true
			_, writeErr = file.Write([]byte("FAILED TO WRITE FULL FILE"))
			if writeErr != nil {
				fmt.Printf("LOG PREFIX Failed to write full file")
			}
		}
	}

	return MitmCallback{
		WriteBack: true,
	}
}

func (t *SnifferToxic) Pipe(stub *ToxicStub) {
	MitmPipe(stub, t)
}

func init() {
	Register("sniffer", new(SnifferToxic))
}
