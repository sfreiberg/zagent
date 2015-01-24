zagent
======

Zagent is a small/simple library for getting values from a zabbix agent running on a remote machine.

License
=======

Zagent is licensed under the MIT license.

Installation
============
`go get github.com/sfreiberg/zagent`

Documentation
=============
[GoDoc](http://godoc.org/github.com/sfreiberg/zagent)

Example
=======

```
package main

import (
	"github.com/sfreiberg/zagent"

	"fmt"
)

func main() {
	agent := zagent.NewAgent("127.0.0.1")

	res, err := agent.Get("agent.ping")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", res.Data)
}

```
