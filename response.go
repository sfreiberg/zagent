package zagent

import (
	"bufio"
	"encoding/binary"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// Response is the response from the zabbix agent.
// Response.Data is generally what most people care
// about. See the wire format here:
// https://www.zabbix.com/documentation/2.2/manual/appendix/items/activepassive
type Response struct {
	Header     []byte // This should always be: ZBXD\x01
	DataLength uint64 // The size of the response
	Data       []byte // The results of the query
}

// Returns true if the key is supported, false if it wasn't.
func (r *Response) Supported() bool {
	return !strings.Contains(r.String(), NotSupported)
}

// Convenience wrapper to return Response.Data as a string.
func (r *Response) String() string {
	return string(r.Data)
}

// Convenience wrapper to return Response.Data as a bool.
func (r *Response) Bool() (bool, error) {
	return strconv.ParseBool(r.String())
}

// Convenience wrapper to return Response.Data as an int.
func (r *Response) Int() (int, error) {
	return strconv.Atoi(r.String())
}

// Convenience wrapper to return Response.Data as an int64.
func (r *Response) Int64() (int64, error) {
	return strconv.ParseInt(r.String(), 10, 64)
}

// Convenience wrapper to return Response.Data as an float64.
func (r *Response) Float64() (float64, error) {
	return strconv.ParseFloat(r.String(), 64)
}

/*
	Convert Response.Data to the most appropriate type. Useful when
	you want a concrete type but don't know it ahead of time.
*/
func (r *Response) Interface() interface{} {
	// Attempt int64
	i, err := strconv.ParseInt(r.String(), 10, 64)
	if err == nil {
		return i
	}

	// Attempt float64
	f, err := strconv.ParseFloat(r.String(), 64)
	if err == nil {
		return f
	}

	// Attempt bool
	b, err := strconv.ParseBool(r.String())
	if err == nil {
		return b
	}

	return r.String()
}

// Create a new Response type
func newResponse() *Response {
	return &Response{
		// Header is always 5 bytes
		Header: make([]byte, 5),
	}
}

func ParseResponse(rd io.Reader) (*Response, error) {
	res := newResponse()
	dataLength := make([]byte, 8)

	reader := bufio.NewReader(rd)
	reader.Read(res.Header)
	reader.Read(dataLength)
	res.Data, _ = ioutil.ReadAll(reader)

	// Convert dataLength from binary to uint
	var bytesRead int
	res.DataLength, bytesRead = binary.Uvarint(dataLength)
	if bytesRead <= 0 {
		if bytesRead == 0 {
			return nil, DataLengthBufferTooSmall
		}
		return nil, DataLengthOverflow
	}

	return res, nil
}
