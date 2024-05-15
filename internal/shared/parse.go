package shared

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var magicLen = 4

// ParseStdout returns logs, body, status, err, body is bytes of pb.HTTPResponse
func ParseStdout(r io.Reader) (logs []byte, body []byte, err error) {
	var bufBytes []byte
	bufBytes, err = io.ReadAll(r)
	if err != nil {
		return
	}

	outLen := len(bufBytes)
	if outLen < magicLen {
		err = fmt.Errorf("expect output have at least 8 magic bytes, actual: %d", len(bufBytes))
		return
	}
	magicStart := outLen - magicLen
	bufferLen := int(binary.LittleEndian.Uint16(bufBytes[magicStart:]))

	if bufferLen > outLen-magicLen {
		err = fmt.Errorf("expect buffer with len %d, available: %d", bufferLen, outLen-magicLen)
		return
	}

	bodyStart := magicStart - bufferLen
	body = bufBytes[bodyStart : bodyStart+bufferLen] // body here is pb.HTTPResponse
	logs = bufBytes[:bodyStart]
	return
}

func read(buf []byte) <-chan string {
	responseCh := make(chan string)
	go func() {
		bufReader := bytes.NewBufferString(string(buf))
		for {
			line, err := bufReader.ReadString('\n')
			if err != nil {
				break
			}
			if len(line) == 0 {
				continue
			}
			responseCh <- line[:len(line)-1]
		}
		close(responseCh)
	}()
	return responseCh
}

func ParseLog(log []byte) ([]string, error) {
	logCh := read(log)
	var res []string
	for line := range logCh {
		res = append(res, line)
	}

	return res, nil
}
