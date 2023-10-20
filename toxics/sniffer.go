package toxics

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Shopify/toxiproxy/v2/stream"
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

func (t *SnifferToxic) Pipe(stub *ToxicStub) {
	buf := make([]byte, 32*1024)
	writer := stream.NewChanWriter(stub.Output)
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)
	stopWriting := false
	file, openErr := getOutputFile("/tmp/eof-" + epochNowString() + t.Path + ".txt")
	if openErr != nil {
		fmt.Printf("LOG PREFIX Failed to open file")
		stopWriting = true
	}

	for {
		n, err := reader.Read(buf)
		if err == stream.ErrInterrupted {
			// fmt.Printf("LOG PREFIX ErrInterrupted\n")
			writer.Write(buf[:n])
			return
		} else if err == io.EOF {
			// fmt.Printf("LOG PREFIX EOF\n")
			stub.Close()
			return
		} else if err != nil {
			fmt.Printf("LOG PREFIX Got Error %+v\n", err)
		}

		// fmt.Printf("LOG PREFIX Writing to buffer: %+v\n", hex.Dump(buf[:n]))
		writer.Write(buf[:n])

		if !stopWriting {
			_, writeErr := file.Write([]byte(hex.Dump(buf[:n])))
			if writeErr != nil {
				fmt.Printf("LOG PREFIX Error writing %+v\n", writeErr)
				stopWriting = true
				_, writeErr = file.Write([]byte("FAILED TO WRITE FULL FILE"))
				if writeErr != nil {
					fmt.Printf("LOG PREFIX Failed to write full file")
				}
			}
		}
	}

	// fmt.Printf("LOG PREFIX out of loop\n")
}

func init() {
	Register("sniffer", new(SnifferToxic))
}
