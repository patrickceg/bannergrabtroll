package main

// Program that attempts to send back randomized payload to any TCP connection

import (
	"crypto/rand"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Loops indefinitely trying to start a connection on a port,
// and if a connection is established, it hands that to handleConnection to
// send garbage to that connection
// The channel doesn't actually use values: it instead is just used as a
// semaphore. We have run out of connections when sendersChannel's buffer is
// full, and each time handleConnection finishes / fails, it takes from the buffer
// The garbagePool is the 10k of pseudorandom data senders get to choose from
func startConnectionListener(payloadKbytes int, rateKbytesPerConnection int,
	sendersChannel chan bool, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err == nil {
			fmt.Printf("Detected connection from %v to %v", conn.LocalAddr(), conn.RemoteAddr())
			fmt.Println()
			// if a successful connection is formed, queue it up for the buffer by pushing to the channel
			sendersChannel <- true
			// when resources are available, start sending garbage to the connection
			go handleConnection(conn, payloadKbytes, rateKbytesPerConnection, sendersChannel)
		}
	}
}

// Sends garbage to the given TCP connection until either the
// payloadkbytes limit or the channel fails. It then signals back to the
// 'done' channel.
// (to only have up to 2k sent at a time)
// (The channel is essentially used like a wait-notify rather than passing of data)
// Do NOT reprogram this to use a UDP connection, or you've just built yourself a
// distributed denial of service (DDoS) node!
func handleConnection(tcpConnection net.Conn,
	payloadKbytes int, rateKbytesPerConnection int, sendersChannel <-chan bool) {
	// using 64-bit int because someone may be especially evil and troll with > 4 GiB of data
	var currentBytes uint64 //= 0
	maxBytes := uint64(payloadKbytes) * 1024
	tenthOfSecond := time.Duration(100000)
	bytesPerTenthOfSecond := uint32(rateKbytesPerConnection) / 10
	// do the sending
	var sendError error // = nil
	for currentBytes < maxBytes && sendError == nil {
		// start the timer
		sendStartTime := time.Now()
		// generate garbage to send
		c := bytesPerTenthOfSecond
		garbageSlice := make([]byte, c)
		_, sendError := rand.Read(garbageSlice)
		if sendError == nil {
			// send the garbage if it generated properly
			_, sendError = tcpConnection.Write(garbageSlice)
			// check if we need to wait to throttle the sending
			sendEndTime := time.Now()
			timeSpentSending := sendEndTime.Sub(sendStartTime)
			timeToSleep := tenthOfSecond - timeSpentSending
			currentBytes = currentBytes + uint64(bytesPerTenthOfSecond)
			if timeToSleep > 0 {
				time.Sleep(timeToSleep)
			}
		}
	}
	// Log what garbage was sent
	fmt.Printf("Sent %d bytes from %v to %v", currentBytes, tcpConnection.LocalAddr(), tcpConnection.RemoteAddr())
	// free up resource by pulling from the channel
	<-sendersChannel
}

// Add TCP listener of that port to the map
// Prints a warning if the port is a duplicate or otherwise could not get opened
func addConnectionToMap(port uint16, toAdd map[uint16]net.Listener) {
	// Don't add duplicate connections
	_, ok := toAdd[port]
	if !ok {
		portString := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", portString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open port %d: %v", port, err)
			fmt.Fprintln(os.Stderr)
		}
		toAdd[port] = ln
	}
}

// unit16Ranges takes the input string and encodes it as 2d slices of uint16
// this function is intended for lists of ports
// Each "row" (the first value) in the 2D slice is 1 or 2 elements: a 1 element
// slice is for single numbers, wheras a 2 element slice is used for ranges,
// where the first element is the minimum and the last element is the maximum
func unit16Ranges(portString string) ([][]uint16, error) {
	allLists := strings.Split(portString, ",")
	var toReturn [][]uint16
	if len(allLists) > 0 {
		for _, portSetString := range allLists {
			ports := strings.Split(portSetString, "-")
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

// Processes the disclaimer for the program
func doDisclaimer() {
	consentFileName := "iamcrazy"
	if _, err := os.Stat(consentFileName); os.IsNotExist(err) {
		var accepted = false
		for !accepted {
			fmt.Println("I hereby acknowledge that running this program may expose my computer ")
			fmt.Println("or anything attached to the same network as my computer to additional ")
			fmt.Println("threats. Type ACCEPT followed by the enter key to continue,")
			fmt.Println("or press CTRL+C to exit.")
			var userInputString string
			fmt.Scanln(&userInputString)
			if strings.ToLower(strings.TrimSpace(userInputString)) == "accept" {
				// Set condition to exit the loop
				accepted = true
				// Create blank file
				_, createErr := os.Create(consentFileName)
				if createErr != nil {
					fmt.Fprintf(os.Stderr, "Could not create %s file - you will be asked for consent again", consentFileName)
					fmt.Fprintln(os.Stderr)
				}
			} else {
				fmt.Printf("Invalid input '%s'", userInputString)
				fmt.Println()
			}
		}
	}
}

func main() {
	// Name of a file that will go in the folder this program runs from if the disclaimer is accepted
	notSpecifiedString := "NOT_SPECIFIED"
	var portString string
	flag.StringVar(&portString, "p", notSpecifiedString, "Ports separated by comma, colon for ranges: 300:400,505 is for ports 300 to 400 plus port 505")
	rateKbytesPerConnection := flag.Int("r", 32, "Rate limit in kilobytes per second per connection")
	payloadKbytes := flag.Int("s", 64, "How many kilobytes to send as payload to a connection")
	maxConnections := flag.Int("n", 16, "Maximum number of simultaneous connections across any ports")
	flag.Parse()

	if portString == notSpecifiedString {
		fmt.Fprintln(os.Stderr, "Specify ports to listen on")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Get port ranges
	portRanges, portRangeErr := unit16Ranges(portString)
	if portRangeErr != nil {
		fmt.Fprintf(os.Stderr, "Invalid port ranges %v", portRangeErr)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	// Validate other arguments are positive
	if *payloadKbytes <= 0 {
		fmt.Fprintf(os.Stderr, "Payload size must be postive: got %d", *payloadKbytes)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	if *rateKbytesPerConnection <= 0 {
		fmt.Fprintf(os.Stderr, "Rate per connection must be postive: got %d", *rateKbytesPerConnection)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	if *maxConnections <= 0 {
		fmt.Fprintf(os.Stderr, "Number of connections must be postive: got %d", *maxConnections)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	// If user hasn't accepted disclaimer yet, print it and have the user type in a string that says they accept
	doDisclaimer()

	// Print user selection
	fmt.Printf("Setting up traps on ports %v with payload size %d kbytes, data rate up to %d kbytes x %d", portRanges, *payloadKbytes, *rateKbytesPerConnection, *maxConnections)
	fmt.Println()

	var listenerMap = make(map[uint16]net.Listener)
	for _, portRange := range portRanges {
		if len(portRange) == 1 {
			addConnectionToMap(portRange[0], listenerMap)
		} else {
			for portI := portRange[0]; portI <= portRange[1]; portI++ {
				addConnectionToMap(portI, listenerMap)
			}
		}
	}

	sendersChannel := make(chan bool, *maxConnections)
	for _, connection := range listenerMap {
		go startConnectionListener(*payloadKbytes, *rateKbytesPerConnection, sendersChannel, connection)
	}

	for {
		time.Sleep(time.Duration(1000000))
	}

}
