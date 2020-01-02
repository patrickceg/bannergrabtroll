package main

// Program that attempts to send back randomized payload to any TCP connection

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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
// This also uses a map to store the most recent access by any given address
func startConnectionListener(payloadKbytes int, rateKbytesPerConnection int,
	sendersChannel chan bool, listener net.Listener, abuseipdbKey string, addrMap map[string]time.Time) {
	for {
		conn, err := listener.Accept()
		if err == nil {
			// func SplitHostPort(hostport string) (host, port string, err error)
			_, port, _ := net.SplitHostPort(conn.LocalAddr().String())
			remoteAddr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			attackTime := time.Now()
			// Check if an IP already had an entry
			previousTime, ok := addrMap[remoteAddr]
			var shouldReport bool
			if !ok {
				fmt.Printf("%v: Detected connection on port %s from %s - First attack", time.Now(), port, remoteAddr)
				fmt.Println()
				shouldReport = true
			} else if attackTime.Sub(previousTime) > time.Hour*12 {
				fmt.Printf("%v: Detected connection on port %s from %s - Already reported %v, a while ago", time.Now(), port, remoteAddr, previousTime)
				fmt.Println()
				shouldReport = true
			} else {
				fmt.Printf("%v: Detected connection on port %s from %s - Already reported %v, recently", time.Now(), port, remoteAddr, previousTime)
				fmt.Println()
				shouldReport = false
			}
			if shouldReport {
				// update the time of this attack for throttling of reports
				addrMap[remoteAddr] = attackTime
				// report connection to abuseipdb if key is present
				if abuseipdbKey != "" {
					go reportAbuseipdb(remoteAddr, port, abuseipdbKey)
				}
			}
			// if a successful connection is formed, queue it up for the buffer by pushing to the channel
			sendersChannel <- true
			// when resources are available, start sending garbage to the connection
			go handleConnection(conn, payloadKbytes, rateKbytesPerConnection, sendersChannel)
		}
	}
}

// Reports the connection to abuseipdb.com
func reportAbuseipdb(remoteAddr string, port string, abuseipdbKey string) {
	client := &http.Client{}
	// Create request body
	comment := fmt.Sprintf("TCP port %s: Scan and connection", port)
	category := "14" // port scan
	params := fmt.Sprintf("ip=%s&comment=%s&categories=%s", url.QueryEscape(remoteAddr), url.QueryEscape(comment), category)
	fmt.Printf("Query %s", params) // TODO remove
	req, reqErr := http.NewRequest("POST", "https://api.abuseipdb.com/api/v2/report?"+params, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Key", abuseipdbKey)
	if reqErr != nil {
		fmt.Fprintln(os.Stderr, "Error creating API query for abuseipdb", reqErr)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error posting to abuseipdb", err)
		return
	}
	// TODO: Check if we need to check resp being nil, or if an error being not nil means
	// body will be there
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		if body != nil {
			fmt.Printf("Posted to AbuseIPDB, response: %q", body)
			fmt.Println()
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
	bytesPerTenthOfSecond := uint32(rateKbytesPerConnection) * 1024 / 10
	// do the sending
	var sendError error // = nil
	for currentBytes < maxBytes && sendError == nil {
		// start the timer
		sendStartTime := time.Now()
		// generate garbage to send
		c := bytesPerTenthOfSecond
		garbageSlice := make([]byte, c)
		_, sendErrorRand := rand.Read(garbageSlice)
		sendError = sendErrorRand
		if sendError == nil {
			// drop the connection if it goes away for a few seconds
			tcpConnection.SetWriteDeadline(time.Now().Add(5 * time.Second))
			// send the garbage if it generated properly
			_, sendErrorTCPWrite := tcpConnection.Write(garbageSlice)
			// try to read equivalent bytes from the connection
			// TODO: make reads count against some limit as well
			readSlice := make([]byte, bytesPerTenthOfSecond)
			tcpConnection.SetReadDeadline(time.Now().Add(25 * time.Millisecond))
			tcpConnection.Read(readSlice)
			sendError = sendErrorTCPWrite
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
	tcpConnection.Close()
	// Log what garbage was sent
	fmt.Printf("%v: Sent %d bytes from %v to %v with error %v", time.Now(), currentBytes, tcpConnection.LocalAddr(),
		tcpConnection.RemoteAddr(), sendError)
	fmt.Println()
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

	// get abuseipdb key from environment variable if it exists
	abuseipdbKey := os.Getenv("ABUSEIPDB_KEY")
	if abuseipdbKey != "" {
		fmt.Println("Enabled using abuseipdb to report detections via HTTP POST")
	}

	var addrMap = make(map[string]time.Time)

	sendersChannel := make(chan bool, *maxConnections)
	for _, connection := range listenerMap {
		go startConnectionListener(*payloadKbytes, *rateKbytesPerConnection, sendersChannel, connection,
			abuseipdbKey, addrMap)
	}

	for {
		time.Sleep(time.Duration(1000000))
	}

}
