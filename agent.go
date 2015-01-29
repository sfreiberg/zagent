// The zagent package allows you to query zabbix agents running over a network.
package zagent

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

var (
	// http://localhost:6060/pkg/encoding/binary/#Uvarint
	DataLengthBufferTooSmall = errors.New("DataLength buffer too small")
	DataLengthOverflow       = errors.New("DataLength is too large")

	// This is the default timeout when contacting a Zabbix Agent.
	DefaultTimeout = time.Duration(30 * time.Second)
)

const (
	NotSupported = "ZBX_NOTSUPPORTED"
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

/*
	Run the check (key) against the Zabbix agent with the specified timeout.
	If timeout is < 1 DefaultTimeout will be used.
*/
func (a *Agent) Query(key string, timeout time.Duration) (*Response, error) {
	res := newResponse()

	if timeout < 1 {
		timeout = DefaultTimeout
	}

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

	if res.Supported() == false {
		return res, fmt.Errorf("%s is not supported", key)
	}

	return res, nil
}

/*
	Run query and return the result as a string.
*/
func (a *Agent) QueryS(key string, timeout time.Duration) (string, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return "", err
	}

	return res.DataS(), nil
}

/*
	Run query and return the result as a bool.
*/
func (a *Agent) QueryBool(key string, timeout time.Duration) (bool, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(res.DataS())
}

/*
	Call agent.hostname on the zabbix agent.
*/
func (a *Agent) AgentHostname(timeout time.Duration) (string, error) {
	return a.QueryS("agent.hostname", timeout)
}

/*
	Call agent.ping on the zabbix agent. Returns true if it
	gets the correct response ("1") and doesn't receive any
	errors in the process.
*/
func (a *Agent) AgentPing(timeout time.Duration) (bool, error) {
	res, err := a.Query("agent.ping", timeout)
	if err != nil {
		return false, err
	}

	if res.Supported() && res.DataS() == "1" {
		return true, nil
	}

	return false, nil
}

/*
	Calls agent.version on the zabbix agent and returns the version
	and/or any errors associated with the action.
*/
func (a *Agent) AgentVersion(timeout time.Duration) (string, error) {
	res, err := a.Query("agent.version", timeout)
	if err != nil {
		return "", err
	}

	return res.DataS(), nil
}
