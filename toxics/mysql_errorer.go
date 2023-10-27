package toxics

import (
	"fmt"
	"strconv"
)

const (
	COM_QUERY_INDEX        = 4
	COM_QUERY              = 0x03
	SEQUENCE_INDEX         = 3
	SKIP_SEQUENCE_INDICIES = 1
)

type MySQLErrorerToxic struct {
	ErrNo                int    `json:"errno"`
	ErrMsg               string `json:"errmsg"`
	query                bool
	multipleQueryPackets bool
}

func (t *MySQLErrorerToxic) buildSignal() string {
	return "SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '" + t.ErrMsg + "', MYSQL_ERRNO = " + strconv.Itoa(t.ErrNo)
	// return "SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'foobar', MYSQL_ERRNO = 1290"
}

// int to little endian byte array
// func itoleba(num int) []byte {
// 	buf := make([]byte, 4)
// 	binary.LittleEndian.PutUint16(buf, uint16(num))
//
// 	return buf[:3]
// }

func (t *MySQLErrorerToxic) attack(buf []byte) MitmCallback {
	t.query = buf[COM_QUERY_INDEX] == COM_QUERY

	// We do not want to block any of the packets that are being sent that are not COM_QUERY
	if !t.query {
		return MitmCallback{WriteBack: true}
	}

	sequenceNotZero := int(buf[SEQUENCE_INDEX]) >= SKIP_SEQUENCE_INDICIES
	t.multipleQueryPackets = t.query && sequenceNotZero

	// If we see a COM_QUERY, and sequence > 0, we do not want to write the rest of the upstream packets to MySQL
	// server, this is because when we see a COM_QUERY, we re-write it as a
	// """
	//     SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = <ERRMSG>, MYSQL_ERRNO = <ERRNO>
	// """
	// and as such, sending the rest of the payload in the upcoming packets, we will be sending invalid data
	if t.multipleQueryPackets {
		return MitmCallback{WriteBack: false}
	}

	errpacketStr := t.buildSignal()
	fmt.Printf("DEBUG PREFIX errpacket %+v\n", errpacketStr)
	errpacket := []byte(errpacketStr)
	fmt.Printf("DEBUG PREFIX errpacket %+v\n", string(errpacket))
	errpacketSize := len(errpacket) + 1 // the +1 here is because packet size is the payload + the command type
	// errpacketSizePacket := itoleba(errpacketSize)

	// At this point, we have the packet we want to rewrite
	// buf[0], [1], and [2] are the packet size, in little endian, errpacketSizePacket
	// buf[3] is the sequency byte, this remains unchanged @ 0x00
	// buf[4] is the COM_QUERY byte, this remains unchanged @ 0x03
	// buf[5+] is the payload, this is what we want to rewrite
	// if our errpacketSize < count, we need to tell the writer to only write the first errpacketSize bytes
	// int 0000 0000 - 0000 0000 - 0000 0000 - 0000 0000
	//         D           C           B           A
	// we want A B C
	buf[0] = byte(errpacketSize)       // errpacketSizePacket[0]
	buf[1] = byte(errpacketSize >> 8)  // errpacketSizePacket[1]
	buf[2] = byte(errpacketSize >> 16) // errpacketSizePacket[2]

	// sz sz sz sq cm 5
	writeBytes := 0
	for writeBytes < (errpacketSize - 1) /* the -1 here is due to the +1 above */ {
		buf[writeBytes+5] = errpacket[writeBytes]
		writeBytes = writeBytes + 1
	}

	return MitmCallback{
		WriteBack:             true,
		OverwriteWrittenBytes: errpacketSize + 4,
	}
}

func (t *MySQLErrorerToxic) Pipe(stub *ToxicStub) {
	t.query = false
	t.multipleQueryPackets = false

	MitmPipe(stub, t)
}

func init() {
	Register("mysql_errorer", new(MySQLErrorerToxic))
}
