/*
 * profiles: compute entropy profiles for hex ipv6 addresses on stdin
 *
 * Copyright (C) 2018 Pawel Foremski <pjf@foremski.pl>
 * Licensed under GNU GPL v3
 */

package main

import "os"
import "bufio"
import "fmt"
import "flag"
import "sync"
import "math"
import "sort"
import "strings"

// represents data for a prefix
type Prefixd struct {
	counts    [32][16]uint64    // hex value counts for all 32 nybbles
	input     chan string       // line input
	lines     uint64            // lines read so far
	wg        *sync.WaitGroup   // waitgroup
}

// command line arguments
var (
	opt_P = flag.Int("P", 8, "prefix length")
	opt_m = flag.Uint64("m", 100, "minimum number of IPv6 addrs required in a prefix")
	opt_p = flag.Bool("p", false, "print counts")
	opt_f = flag.Int("f", 0, "use CSV column f as prefix")
)

// processor counts hex values in given prefix
func (pd *Prefixd) processor() {
	for line := range pd.input {
		pd.lines++
		for i, v := range line[*opt_P:32] {
			if v >= '0' && v <= '9' {
				pd.counts[i][v-'0']++
			} else if v >= 'a' && v <= 'f' {
				pd.counts[i][v-'a'+10]++
			} else if v >= 'A' && v <= 'F' {
				pd.counts[i][v-'A'+10]++
			}
		}
	}
	pd.wg.Done()
}

func main() {
	// parse args
	flag.Parse()

	// prepare
	prefixes := make(map[string]*Prefixd)

	// read input and distribute to processors
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] != '2' { continue }

		// determine prefix - the table key
		var prefix string
		if *opt_f > 0 {
			d := strings.Split(line, ",")
			prefix = d[*opt_f-1]
		} else {
			prefix = line[0:*opt_P]
		}

		pd, ok := prefixes[prefix]
		if !ok {
			pd = &Prefixd{}
			prefixes[prefix] = pd

			pd.input = make(chan string, 1000)
			pd.wg = new(sync.WaitGroup)
			pd.wg.Add(1)
			go pd.processor()
		}

		pd.input <- line
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}

	// notify we're done, collect prefixes
	keys := make([]string, 0, len(prefixes))
	for prefix := range prefixes {
		close(prefixes[prefix].input)
		keys = append(keys, prefix)
	}
	sort.Strings(keys)

	// print header
	fmt.Printf("#prefix,addresses,")
	for i := 0; i < 32 - *opt_P; i++ {
		if i > 0 { fmt.Printf(",") }
		fmt.Printf("ent%d", *opt_P + i + 1)
	}
	fmt.Printf("\n")

	// iterate
	for _,prefix := range keys {
		// sync
		pd := prefixes[prefix]
		pd.wg.Wait()

		// big enough?
		if pd.lines < *opt_m { continue }

		// print counts?
		if *opt_p {
			fmt.Printf("# === prefix %s ===\n", prefix)
			fmt.Printf("#  value:   ")
			for v := '0'; v <= '9'; v++ { fmt.Printf("  '%c' ", v) }
			for v := 'a'; v <= 'f'; v++ { fmt.Printf("  '%c' ", v) }
			fmt.Printf("\n")
			for i := 0; i < 32 - *opt_P; i++ {
				fmt.Printf("# char %2d: %5d\n", i+1+*opt_P, pd.counts[i])
			}
		}

		// print entropy
		fmt.Printf("%s,%d,", prefix, pd.lines)
		for i := 0; i < 32 - *opt_P; i++ {
			var entropy float64

			for v := 0; v < 16; v++ {
				if pd.counts[i][v] == 0 { continue }

				freq := float64(pd.counts[i][v]) / float64(pd.lines)
				entropy -= freq * math.Log2(freq)
			}

			// print normalized
			if i > 0 { fmt.Printf(",") }
			fmt.Printf("%.3f", entropy / 4.0)
		}
		fmt.Printf("\n")
	}
}
