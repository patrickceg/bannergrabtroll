package main

// Program that attempts to send back randomized payload to any TCP connection

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// unit16Ranges takes the input string and encodes it as 2d slices of uint16
// this function is intended for lists of ports
// Each "row" (the first value) in the 2D slice is 1 or 2 elements: a 1 element
// slice is for single numbers, wheras a 2 element slice is used for ranges, where
// the first element is the minimum and the last element is the maximum
func unit16Ranges(portString string) ([][]uint16, error) {
	allLists := strings.Split(portString, ",")
	var toReturn [][]uint16
	if len(allLists) > 0 {
		for _, portSetString := range allLists {
			ports := strings.Split(portSetString, ":")
			portsLen := len(ports)
			parsePortErrorMessage := "Could not parse port %s from %s"
			if portsLen == 1 || portsLen == 2 {
				minOrSingle, errLo := strconv.ParseUint(ports[0], 10, 16)
				if errLo != nil {
					return [][]uint16{}, fmt.Errorf(parsePortErrorMessage, ports[0], ports)
				}
				if portsLen == 2 {
					max, errHi := strconv.ParseUint(ports[1], 10, 16)
					if errHi != nil {
						return [][]uint16{}, fmt.Errorf(parsePortErrorMessage, ports[1], ports)
					}
					if minOrSingle > max {
						return [][]uint16{}, fmt.Errorf("Reversed range: %d should be less than %d", minOrSingle, max)
					}
					toReturn = append(toReturn, []uint16{uint16(minOrSingle), uint16(max)})
				} else {
					toReturn = append(toReturn, []uint16{uint16(minOrSingle)})
				}
			} else {
				return [][]uint16{}, fmt.Errorf("Could not parse %s as a single port x or x:y range", ports)
			}
		}
	} else {
		return [][]uint16{}, fmt.Errorf("Error parsing port list from %s", portString)
	}
	return toReturn, nil
}

func main() {
	// Name of a file that will go in the folder this program runs from if the disclaimer is accepted
	consentFileName := "iamcrazy"
	var portString string
	flag.StringVar(&portString, "ports", "8082:8085,9543", "Ports separated by comma, colon for ranges: 300:400,505 is for ports 300 to 400 plus port 505")

	// Grab and validate arguments
	portRanges, portRangeErr := unit16Ranges(portString)
	if portRangeErr != nil {
		fmt.Fprintf(os.Stderr, "Invalid port ranges %v", portRangeErr)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	// Validate other arguments are positive
	payloadKbytes := flag.Int("payloadkbytes", 64, "How many kilobytes to send as payload to a connection")
	if *payloadKbytes <= 0 {
		fmt.Fprintf(os.Stderr, "Payload size must be postive: got %d", payloadKbytes)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	rateKbytesPerConnection := flag.Int("ratekbytesperconnection", 32, "Rate limit in kilobytes per second per connection")
	if *rateKbytesPerConnection <= 0 {
		fmt.Fprintf(os.Stderr, "Rate per connection must be postive: got %d", rateKbytesPerConnection)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	numConnections := flag.Int("numconnections", 16, "Number of simultaneous connections across any ports")
	if *numConnections <= 0 {
		fmt.Fprintf(os.Stderr, "Number of connections must be postive: got %d", numConnections)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	// If user hasn't accepted disclaimer yet, print it and have the user type in a string that says they accept
	if _, err := os.Stat(consentFileName); os.IsNotExist(err) {
		var accepted = false
		for !accepted {
			fmt.Println("I hereby acknowledge that running this program may expose my computer ")
			fmt.Println("or anything attached to the same network as my computer to additional ")
			fmt.Println("threats. Type ACCEPT followed by the enter key to continue,")
			fmt.Println("or press CTRL+C to exit.")
			var userInputString string
			fmt.Scan(userInputString)
			if strings.ToLower(strings.TrimSpace(userInputString)) == "ACCEPT" {
				// Set condition to exit the loop
				accepted = true
				// Create blank file
				_, createErr := os.Create(consentFileName)
				if createErr != nil {
					fmt.Fprintf(os.Stderr, "Could not create %s file - you will be asked for consent again", consentFileName)
					fmt.Fprintln(os.Stderr)
				}
			}
		}
	}

	// Print user selection
	fmt.Printf("Setting up traps on ports %v with payload size %d kbytes, data rate up to %d kbytes x %d", portRanges, payloadKbytes, rateKbytesPerConnection, numConnections)
	fmt.Println()

	// TODO: Start TCP listener
}
