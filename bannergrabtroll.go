package main

import (
	"flag"
)

func main() {
	var portString string
	flag.StringVar(&portString, "ports", "8082:8085,9543", "Ports separated by comma, colon for ranges: 300:400,505 is for ports 300 to 400 plus port 505")
	payloadKbytes := flag.Int("payloadkbytes", 64, "How many kilobytes to send as payload to a connection")
	rateKbytesPerConnection := flag.Int("ratekbytesperconnection", 32, "Rate limit in kilobytes per second per connection")
	numConnections := flag.Int("numconnections", 16, "Number of simultaneous connections across any ports")

	// TODO: Validate ports list

	// TODO: Validate other arguments are positive

	// TODO: If user hasn't accepted disclaimer yet, print it and have the user type in a string that says they accept

	// TODO: Print user selection

	// TODO: Start TCP listener
}
