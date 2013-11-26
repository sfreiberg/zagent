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
	opts := zagent.NewOptions()
	opts.Host = "www.example.com"

	res, err := zagent.Get("agent.ping", opts)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("%s\n", res.Data)
}
```