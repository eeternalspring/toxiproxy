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

}

func init() {
	Register("sniffer", new(SnifferToxic))
}
