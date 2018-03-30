package main

import (
	"flag"
	"fmt"
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
	var portString string
	flag.StringVar(&portString, "ports", "8082:8085,9543", "Ports separated by comma, colon for ranges: 300:400,505 is for ports 300 to 400 plus port 505")

	// portRanges, err := unit16Ranges(portString)

	//payloadKbytes := flag.Int("payloadkbytes", 64, "How many kilobytes to send as payload to a connection")
	//rateKbytesPerConnection := flag.Int("ratekbytesperconnection", 32, "Rate limit in kilobytes per second per connection")
	//numConnections := flag.Int("numconnections", 16, "Number of simultaneous connections across any ports")

	// TODO: Validate ports list

	// TODO: Validate other arguments are positive

	// TODO: If user hasn't accepted disclaimer yet, print it and have the user type in a string that says they accept

	// TODO: Print user selection

	// TODO: Start TCP listener
}
