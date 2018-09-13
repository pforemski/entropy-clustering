/*
 * ipv6-hex2addr: convert IPv6 addresses from hex IP format
 * Author: Pawel Foremski <pjf@foremski.pl>
 */

package main

import "os"
import "bufio"
import "fmt"

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		l := scanner.Text()
		fmt.Printf("%s:%s:%s:%s:%s:%s:%s:%s\n",
			l[0:4], l[4:8], l[8:12], l[12:16],
			l[16:20], l[20:24], l[24:28], l[28:])
	}
}
