package zagent

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// You must set and export the shell variable ZABBIX_HOST in
// order to specify which host to test against. If you leave
// this unset it will default to localhost. At least on the
// Mac. This test is pretty rudimentary.
func TestAgent(t *testing.T) {
	zabbixHost := os.Getenv("ZABBIX_HOST")
	fmt.Println("Testing against:", zabbixHost)

	agent := NewAgent(zabbixHost)

	res, err := agent.Query("agent.ping", 0)
	if err != nil {
		t.Fatal(err)
	}

	if res.DataS() != "1" {
		t.Fatal("agent.ping results are incorrect.")
	}
}

func TestAgentPing(t *testing.T) {
	zabbixHost := os.Getenv("ZABBIX_HOST")

	agent := NewAgent(zabbixHost)

	pingable, err := agent.AgentPing(0)
	if err != nil {
		t.Fatal("Ping test failed with error:", err)
	}

	if !pingable {
		t.Fatal("Zabbix agent wasn't pingable")
	}
}

func TestAgentHostname(t *testing.T) {
	zabbixHost := os.Getenv("ZABBIX_HOST")

	agent := NewAgent(zabbixHost)

	hostname, err := agent.AgentHostname(0)
	if err != nil {
		t.Fatal("Hostname test failed with error:", err)
	}

	if hostname == "" {
		t.Fatal("Zabbix hostname was empty")
	}

	fmt.Println("Zabbix hostname:", hostname)
}

func TestAgentVersion(t *testing.T) {
	zabbixHost := os.Getenv("ZABBIX_HOST")

	agent := NewAgent(zabbixHost)

	version, err := agent.AgentVersion(0)
	if err != nil {
		t.Fatal("Version test failed with error:", err)
	}

	if version == "" {
		t.Fatal("Zabbix version was empty")
	}

	fmt.Println("Zabbix version:", version)
}

func TestAgentUnsupported(t *testing.T) {
	zabbixHost := os.Getenv("ZABBIX_HOST")
	fmt.Println("Testing against:", zabbixHost)

	agent := NewAgent(zabbixHost)

	res, err := agent.Query("Supercalifragilisticexpialidocious", 0)
	if err == nil {
		t.Fatal("An error isn't thrown when calling an unknown key")
	}

	if err != nil && !strings.HasSuffix(err.Error(), " is not supported") {
		t.Fatal(err)
	}

	if res.Supported() {
		t.Fatal("Response.Supported() reports true")
	}
}

func TestQueryInterface(t *testing.T) {
	zabbixHost := os.Getenv("ZABBIX_HOST")

	agent := NewAgent(zabbixHost)

	// Should return a string
	hostname, err := agent.QueryInterface("agent.hostname", 0)
	if err != nil {
		t.Fatal("Couldn't get hostname: " + err.Error())
	}

	switch hostname.(type) {
	default:
		t.Fatal("agent.hostname didn't get converted to a string")
	case string:
		fmt.Println("agent.hostname returned a string")
	}

	// Should return a float64
	load, err := agent.QueryInterface("system.cpu.load", 0)
	if err != nil {
		t.Fatal("Couldn't get cpu load: " + err.Error())
	}

	switch load.(type) {
	default:
		t.Fatal("system.cpu.load didn't get converted to a float64")
	case float64:
		fmt.Println("system.cpu.load returned a float64")
	}

	// Should return an int64
	cpus, err := agent.QueryInterface("system.cpu.num", 0)
	if err != nil {
		t.Fatal("Couldn't get number of cpus: " + err.Error())
	}

	switch cpus.(type) {
	default:
		t.Fatal("system.cpu.num didn't get converted to an int64")
	case int64:
		fmt.Println("system.cpu.num was converted to int64")
	}
}
