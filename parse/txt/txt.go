package txt

import (
	"bytes"
	"io"
)

func ParseLocal(data []byte) string {
	buffer := bytes.NewBuffer(data)
	txtString := buffer.String()
	return txtString
}

func ParseBody(body io.ReadCloser, fileSize int64) (string, error) {

	defer body.Close()
	buff, _ := io.ReadAll(body)
	buffer := bytes.NewBuffer(buff)
	txtString := buffer.String()

	return txtString, nil
}
