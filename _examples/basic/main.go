package main

import (
	"fmt"
	"time"

	"github.com/majicloud/maji-config/config-go"
)

func main() {
	c, err := config.NewClient("-KifDKTgeS_4O_dFQMHB")
	if err != nil {
		panic(err)
	}

	c.Host = "http://localhost:4000"

	if err := c.Start(); err != nil {
		panic(err)
	}

	for {
		b, ok := c.GetBoolean("boolean")
		if !ok {
			fmt.Print("no key `boolean` set")
		} else {
			fmt.Printf("boolean: %v\n", b)
		}

		s, ok := c.GetString("string")
		if !ok {
			fmt.Print("no key `string` set")
		} else {
			fmt.Printf("string: %v\n", s)
		}

		time.Sleep(time.Second)
	}
}
