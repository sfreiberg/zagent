package zagent

import "strings"

// Response is the response from the zabbix agent.
// Response.Data is generally what most people care
// about. See the wire format here:
// https://www.zabbix.com/documentation/2.2/manual/appendix/items/activepassive
type Response struct {
	Header     []byte // This should always be: ZBXD\x01
	DataLength uint64 // The size of the response
	Data       []byte // The results of the query
	key        string
}

// Returns true if the key is supported, false if it wasn't.
// Most of the time you shouldn't need to call this as Agent.Get()
// will return an error if the key is unsupported.
func (r *Response) Supported() bool {
	return !strings.Contains(r.DataS(), NotSupported)
}

// Returns the key that was used in the query against the Zabbix agent.
func (r *Response) Key() string {
	return r.key
}

// Convenience wrapper to return Data as a string.
func (r *Response) DataS() string {
	return string(r.Data)
}

// Create a new Response type
func newResponse() *Response {
	return &Response{
		// Header is always 5 bytes
		Header: make([]byte, 5),
	}
}
