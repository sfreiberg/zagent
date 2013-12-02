package zagent

import (
	"fmt"
	"os"
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

	res, err := agent.Get("agent.ping")
	if err != nil {
		t.Fatal(err)
	}

	if !res.Supported() {
		t.Fatal("agent.ping is not supported. Very strange.")
	}

	if string(res.Data) != "1" {
		t.Fatal("agent.ping results are incorrect.")
	}
}
