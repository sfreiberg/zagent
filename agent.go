// The zagent package allows you to query zabbix agents running over a network.
package zagent

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

var (
	// http://localhost:6060/pkg/encoding/binary/#Uvarint
	DataLengthBufferTooSmall = errors.New("DataLength buffer too small")
	DataLengthOverflow       = errors.New("DataLength is too large")

	// This is the default timeout when contacting a Zabbix Agent.
	DefaultTimeout = time.Duration(30 * time.Second)
)

// Creates a new Agent with a default port of 10050
func NewAgent(host string) *Agent {
	return &Agent{Host: host, Port: 10050}
}

// Agent represents a remote zabbix agent
type Agent struct {
	Host string
	Port int
}

// Returns a string with the host and port concatenated to host:port
func (a *Agent) hostPort() string {
	portS := fmt.Sprintf("%v", a.Port)
	return net.JoinHostPort(a.Host, portS)
}

// Run the check (key) against the Zabbix agent with the default timeout
func (a *Agent) Get(key string) (*Response, error) {
	return a.GetTimeout(key, DefaultTimeout)
}

// Run the check (key) against the Zabbix agent with the specified timeout
func (a *Agent) GetTimeout(key string, timeout time.Duration) (*Response, error) {
	res := newResponse()

	conn, err := net.DialTimeout("tcp", a.hostPort(), timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, key)
	if err != nil {
		return nil, err
	}

	dataLength := make([]byte, 8)

	reader := bufio.NewReader(conn)
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

// Response is the response from the zabbix agent.
// Response.Data is generally what most people care
// about. See the wire format here:
// https://www.zabbix.com/documentation/2.2/manual/appendix/items/activepassive
type Response struct {
	Header     []byte // This should always be: ZBXD\x01
	DataLength uint64 // I assume this should match the length of Data but not really tested
	Data       []byte // The results of the query
	key        string
}

// Returns true if the command is supported, false if it wasn't
func (r *Response) Supported() bool {
	return string(r.Data) != "ZBX_NOTSUPPORTED"
}

// Returns the key that was used in the query against the Zabbix agent.
func (r *Response) Key() string {
	return r.key
}

// Create a new Response type
func newResponse() *Response {
	return &Response{
		// Header is always 5 bytes
		Header: make([]byte, 5),
	}
}
