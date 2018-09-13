/*
 * ipv6-addr2hex: convert IPv6 addresses to hex IP format
 * Author: Pawel Foremski <pjf@foremski.pl>
 */

package main

import "os"
import "bufio"
import "fmt"
import "net"
import "flag"
import "strings"

const d2x = "0123456789abcdef"

// command line arguments
var (
	opt_d = flag.String("d", "\t", "field delimiter")
	opt_f = flag.Int("f", 0, "field number")
)

func main() {
	// parse args
	flag.Parse()
	*opt_f -= 1

	// prepare
	var d []string
	var ipstr string
	s := make([]byte, 32)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		// extract field?
		if *opt_f >= 0 {
			d = strings.Split(line, *opt_d)
			if *opt_f >= len(d) {
				fmt.Println(line)
				continue
			}
			ipstr = d[*opt_f]
		} else {
			ipstr = line
		}

		// parsable?
		ip := net.ParseIP(ipstr)
		if ip == nil {
			fmt.Println(line)
			continue
		}

		// adapted from golang's net/ip.go
		s = s[:len(ip)*2]
		for i,d := range ip {
			s[i*2], s[i*2+1] = d2x[d>>4], d2x[d&0xf]
		}

		if *opt_f >= 0 {
			d[*opt_f] = string(s)
			fmt.Println(strings.Join(d, *opt_d))
		} else {
			fmt.Println(string(s))
		}
	}
}
