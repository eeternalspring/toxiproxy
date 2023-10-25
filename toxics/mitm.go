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

		ret := mitm.attack(buf)
		if ret.WriteBack {
			writeBytes := n
			if ret.OverwriteWrittenBytes > 0 {
				writeBytes = ret.OverwriteWrittenBytes
			}

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