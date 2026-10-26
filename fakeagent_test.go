package zagent

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"
)

// frame encodes data using the Zabbix wire format: "ZBXD\x01", an 8-byte
// little-endian length, then the data.
func frame(data string) []byte {
	var b bytes.Buffer
	b.WriteString("ZBXD\x01")
	binary.Write(&b, binary.LittleEndian, uint64(len(data)))
	b.WriteString(data)
	return b.Bytes()
}

// newFakeAgent starts a minimal passive Zabbix agent on localhost that
// answers keys from responses (ZBX_NOTSUPPORTED otherwise) and returns an
// Agent pointing at it.
func newFakeAgent(t *testing.T, responses map[string]string) *Agent {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				buf := make([]byte, 1024)
				n, _ := conn.Read(buf)
				resp, ok := responses[strings.TrimSpace(string(buf[:n]))]
				if !ok {
					resp = NotSupported
				}
				conn.Write(frame(resp))
			}(conn)
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	return &Agent{Host: addr.IP.String(), Port: addr.Port}
}

func TestParseResponse(t *testing.T) {
	res, err := ParseResponse(bytes.NewReader(frame("2.2.2")))
	if err != nil {
		t.Fatal(err)
	}
	if string(res.Header) != "ZBXD\x01" {
		t.Errorf("Header = %q", res.Header)
	}
	if res.DataLength != 5 || res.String() != "2.2.2" {
		t.Errorf("DataLength = %d, Data = %q", res.DataLength, res.Data)
	}
}

func TestResponseConversions(t *testing.T) {
	r := &Response{Data: []byte("42")}
	if v, err := r.Int(); err != nil || v != 42 {
		t.Errorf("Int() = %v, %v", v, err)
	}
	if v, err := r.Int64(); err != nil || v != 42 {
		t.Errorf("Int64() = %v, %v", v, err)
	}
	if v, err := r.Float64(); err != nil || v != 42 {
		t.Errorf("Float64() = %v, %v", v, err)
	}
	if _, err := (&Response{Data: []byte("x")}).Int(); err == nil {
		t.Error("Int() on non-numeric data should fail")
	}
	if v, err := (&Response{Data: []byte("1")}).Bool(); err != nil || !v {
		t.Errorf("Bool() = %v, %v", v, err)
	}
	if !r.Supported() {
		t.Error("Supported() should be true")
	}
	if (&Response{Data: []byte(NotSupported)}).Supported() {
		t.Error("Supported() should be false for ZBX_NOTSUPPORTED")
	}
}

func TestResponseInterface(t *testing.T) {
	tests := []struct {
		data string
		want any
	}{
		{"12", int64(12)},
		{"-3", int64(-3)},
		{"0.25", 0.25},
		{"true", true},
		{"web01", "web01"},
	}
	for _, tt := range tests {
		got := (&Response{Data: []byte(tt.data)}).Interface()
		if got != tt.want {
			t.Errorf("Interface(%q) = %#v, want %#v", tt.data, got, tt.want)
		}
	}
}

func TestFakeAgentQueries(t *testing.T) {
	a := newFakeAgent(t, map[string]string{
		"agent.ping":     "1",
		"agent.hostname": "web01",
		"agent.version":  "7.0.0",
	})

	if ok, err := a.AgentPing(time.Second); err != nil || !ok {
		t.Fatalf("AgentPing() = %v, %v", ok, err)
	}
	if h, err := a.AgentHostname(time.Second); err != nil || h != "web01" {
		t.Fatalf("AgentHostname() = %q, %v", h, err)
	}
	if v, err := a.AgentVersion(time.Second); err != nil || v != "7.0.0" {
		t.Fatalf("AgentVersion() = %q, %v", v, err)
	}

	res, err := a.Query("nope", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if res.Supported() {
		t.Fatal("unknown key should not be supported")
	}
}

func TestDiscover(t *testing.T) {
	a := newFakeAgent(t, map[string]string{
		"vfs.fs.discovery":     `{"data":[{"{#FSNAME}":"/","{#FSTYPE}":"ext4"}]}`,
		"net.if.discovery":     `{"data":[{"{#IFNAME}":"lo"},{"{#IFNAME}":"eth0"}]}`,
		"system.cpu.discovery": `{"data":[{"{#CPU.NUMBER}":0,"{#CPU.STATUS}":"online"}]}`,
	})

	fs, err := a.DiscoverFilesystems(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 || fs[0].Name != "/" || fs[0].Type != "ext4" {
		t.Fatalf("filesystems = %+v", fs)
	}

	ifaces, err := a.DiscoverNetworkInterfaces(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(ifaces) != 2 || ifaces[1].Name != "eth0" {
		t.Fatalf("interfaces = %+v", ifaces)
	}

	cpus, err := a.DiscoverCPUs(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(cpus) != 1 || cpus[0].Number != 0 || cpus[0].Status != "online" {
		t.Fatalf("cpus = %+v", cpus)
	}
}
