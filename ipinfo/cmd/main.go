package main

import (
	"context"
	"fmt"
	"time"

	"github.com/polite-007/toolbox/ipinfo"
)

func main() {
	client, err := ipinfo.NewIpinfoClient(
		ipinfo.WithTimeout(10 * time.Second),
		// ipinfo.WithProxy("socks5://127.0.0.1:1080"),
	)
	if err != nil {
		fmt.Println("create failed:", err)
		return
	}

	info, err := client.Lookup(context.Background(), "45.136.15.66")
	if err != nil {
		fmt.Println("lookup failed:", err)
		return
	}
	fmt.Println("IP Info:", info)
}
