// The zagent package allows you to query zabbix agents running over a network.
package zagent

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
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

// Filesystem respresents a Zabbix filesystem as presented by vfs.fs.discovery
type Filesystem struct {
	Name string
	Type string
}

// NetworkInterface represents a Zabbix network interface as presented by net.if.discovery
type NetworkInterface struct {
	Name string
}

// CPU represents a Zabbix cpu as presented by system.cpu.discovery
type CPU struct {
	Number float64
	Status string
}

// Agent represents a remote zabbix agent
type Agent struct {
	Host string
	Port int
}

// Creates a new Agent with a default port of 10050
func NewAgent(host string) *Agent {
	return &Agent{Host: host, Port: 10050}
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

// Run query and return the result (Response.Data) as a string.
func (a *Agent) QueryS(key string, timeout time.Duration) (string, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return "", err
	}

	return res.DataS(), nil
}

// Run query and return the result (Response.Data) as a bool.
func (a *Agent) QueryBool(key string, timeout time.Duration) (bool, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(res.DataS())
}

// Run query and return the result (Response.Data) as an int.
func (a *Agent) QueryInt(key string, timeout time.Duration) (int, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(res.DataS())
}

// Run query and return the result (Response.Data) as an int64.
func (a *Agent) QueryInt64(key string, timeout time.Duration) (int64, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(res.DataS(), 10, 64)
}

// Run query and return the result (Response.Data) as an float64.
func (a *Agent) QueryFloat64(key string, timeout time.Duration) (float64, error) {
	res, err := a.Query(key, timeout)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(res.DataS(), 64)
}

/*
	Run query and convert the result to the most appropriate type. Useful when
	you want a concrete type but don't know it ahead of time.
*/
func (a *Agent) QueryInterface(key string, timeout time.Duration) (interface{}, error) {
	res, err := a.QueryS(key, timeout)
	if err != nil {
		return nil, err
	}

	// Attempt int64
	i, err := strconv.ParseInt(res, 10, 64)
	if err == nil {
		return i, nil
	}

	// Attempt float64
	f, err := strconv.ParseFloat(res, 64)
	if err == nil {
		return f, nil
	}

	// Attempt bool
	b, err := strconv.ParseBool(res)
	if err == nil {
		return b, nil
	}

	return res, nil
}

/*
	Run query and convert the JSON to a map[string][]map[string]interface{}.
	This is a raw version of the query and most people are expected to use
	the Discover* methods.
*/
func (a *Agent) queryJSON(key string, timeout time.Duration) (map[string][]map[string]interface{}, error) {
	data := make(map[string][]map[string]interface{})

	res, err := a.Query(key, timeout)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Data, &data)
	return data, err
}

// Return an array of Filesystem structs.
func (a *Agent) DiscoverFilesystems(timeout time.Duration) ([]*Filesystem, error) {
	fs := []*Filesystem{}

	data, err := a.queryJSON("vfs.fs.discovery", timeout)
	if err != nil {
		return nil, err
	}

	for _, f := range data["data"] {
		filesystem := &Filesystem{
			Name: f["{#FSNAME}"].(string),
			Type: f["{#FSTYPE}"].(string),
		}

		fs = append(fs, filesystem)
	}

	return fs, err
}

// Return an array of NetworkInterface structs.
func (a *Agent) DiscoverNetworkInterfaces(timeout time.Duration) ([]*NetworkInterface, error) {
	in := []*NetworkInterface{}

	data, err := a.queryJSON("net.if.discovery", timeout)
	if err != nil {
		return nil, err
	}

	for _, i := range data["data"] {
		networkIface := &NetworkInterface{
			Name: i["{#IFNAME}"].(string),
		}

		in = append(in, networkIface)
	}

	return in, err
}

// Return an array of CPUs.
func (a *Agent) DiscoverCPUs(timeout time.Duration) ([]*CPU, error) {
	cpus := []*CPU{}

	data, err := a.queryJSON("system.cpu.discovery", timeout)
	if err != nil {
		return nil, err
	}

	for _, i := range data["data"] {
		cpu := &CPU{
			Number: i["{#CPU.NUMBER}"].(float64),
			Status: i["{#CPU.STATUS}"].(string),
		}

		cpus = append(cpus, cpu)
	}

	return cpus, err
}

// Call agent.hostname on the zabbix agent.
func (a *Agent) AgentHostname(timeout time.Duration) (string, error) {
	return a.QueryS("agent.hostname", timeout)
}

/*
	Call agent.ping on the zabbix agent. Returns true if it
	gets the correct response ("1") and doesn't receive any
	errors in the process.
*/
func (a *Agent) AgentPing(timeout time.Duration) (bool, error) {
	return a.QueryBool("agent.ping", timeout)
}

/*
	Calls agent.version on the zabbix agent and returns the version
	and/or any errors associated with the action.
*/
func (a *Agent) AgentVersion(timeout time.Duration) (string, error) {
	return a.QueryS("agent.version", timeout)
}
