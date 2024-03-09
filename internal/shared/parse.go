package shared

//
import (
	"encoding/binary"
	"fmt"
	"io"
)

var magicLen = 8

func ParseStdout(r io.Reader) (int, []byte, error) {
	bufBytes, err := io.ReadAll(r)
	if err != nil {
		return 0, nil, err
	}

	outLen := len(bufBytes)
	if outLen < magicLen {
		return 0, nil, fmt.Errorf("expect output have at least 8 magic bytes, actual: %d", len(bufBytes))
	}
	magicStart := outLen - magicLen

	statusCode := int(binary.LittleEndian.Uint32(bufBytes[magicStart : magicStart+4]))
	bufferLen := int(binary.LittleEndian.Uint32(bufBytes[magicStart+4:]))

	if bufferLen > outLen-magicLen {
		return 0, nil, fmt.Errorf("expect buffer with len %d, available: %d", bufferLen, len(bufBytes)-magicLen)
	}
	bodyStart := magicStart - bufferLen
	fmt.Println(bodyStart)
	body := bufBytes[bodyStart : bodyStart+bufferLen]
	logs := bufBytes[:bodyStart]
	fmt.Println(string(logs))
	fmt.Println(string(body))
	return statusCode, body, nil
}
