package toxics

import (
	"fmt"
	"io"

	"github.com/Shopify/toxiproxy/v2/stream"
)

type MitmCallback struct {
	WriteBack             bool
	OverwriteWrittenBytes int
}

type Mitm interface {
	attack([]byte) MitmCallback
}

const (
	MITM_BUFFER_SIZE = 32 * 1024
)

func MitmPipe(stub *ToxicStub, mitm Mitm) {
	buf := make([]byte, MITM_BUFFER_SIZE)
	writer := stream.NewChanWriter(stub.Output)
	reader := stream.NewChanReader(stub.Input)
	reader.SetInterrupt(stub.Interrupt)

	for {
		n, err := reader.Read(buf)
		if err == stream.ErrInterrupted {
			fmt.Printf("LOG PREFIX ErrInterrupted\n")
			writer.Write(buf[:n])
			return
		} else if err == io.EOF {
			fmt.Printf("LOG PREFIX EOF\n")
			stub.Close()
			return
		} else if err != nil {
			fmt.Printf("LOG PREFIX Got Error %+v\n", err)
		}

		fmt.Printf("DEBUG PREFIX attacking w/ buf: %+v\n", buf)
		ret := mitm.attack(buf)
		fmt.Printf("DEBUG PREFIX ret: %+v\n", ret)
		fmt.Printf("DEBUG PREFIX new buf: %+v\n", buf)
		if ret.WriteBack {
			writeBytes := n
			if ret.OverwriteWrittenBytes > 0 {
				writeBytes = ret.OverwriteWrittenBytes
			}

			fmt.Printf("DEBUG PREFIX writeBytes %+v\n", writeBytes)
			fmt.Printf("DEBUG PREFIX writing buf %+v\n", string(buf[5:writeBytes]))

			writer.Write(buf[:writeBytes])
		}
	}
}

type MitmToxic struct{}

func (t *MitmToxic) attack(_ []byte) MitmCallback {
	return MitmCallback{
		WriteBack: true,
	}
}

func (t *MitmToxic) Pipe(stub *ToxicStub) {
}

func init() {
	Register("mitm", new(MitmToxic))
}
