package zagent

import (
	"encoding/binary"
	"io"
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
func (r *Response) Interface() any {
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

// ParseResponse reads a response in the Zabbix wire format: a 5-byte header,
// the data length as a 64-bit little-endian integer, then the data.
func ParseResponse(rd io.Reader) (*Response, error) {
	res := newResponse()
	dataLength := make([]byte, 8)

	if _, err := io.ReadFull(rd, res.Header); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(rd, dataLength); err != nil {
		return nil, err
	}
	res.DataLength = binary.LittleEndian.Uint64(dataLength)

	data, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	res.Data = data

	return res, nil
}
