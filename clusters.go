/*
 * clusters: find clusters of entropy profiles
 *
 * Copyright (C) 2018 Pawel Foremski <pjf@foremski.pl>
 * Licensed under GNU GPL v3
 */

package main

import "os"
import "bufio"
import "fmt"
import "flag"
import "strings"
import "strconv"
import "sort"
import "github.com/pforemski/gouda/point"
import "github.com/pforemski/gouda/dbscan"
import "github.com/pforemski/gouda/kmeans"
import "github.com/fatih/color"

// command line arguments
var (
	opt_eps        = flag.Float64("eps", 0.1, "DBSCAN eps parameter")
	opt_min_points = flag.Int("min-points", 2, "DBSCAN min points parameter")
	opt_F          = flag.Int("F", 0, "ignore first F hex characters")
	opt_L          = flag.Int("L", 0, "ignore last L hex characters")
	opt_N          = flag.Bool("N", false, "show noise (cluster 0)")
	opt_C          = flag.Bool("C", false, "print entropy values in color")
	opt_P          = flag.String("P", "", "use given prefix2as mapping file")
	opt_S          = flag.Bool("S", false, "print means & std deviations")

	opt_kmeans     = flag.Bool("kmeans", false, "use k-means clustering instead of DBSCAN")
	opt_k          = flag.Int("k", 10, "number of k-means clusters to look for")
	opt_max_iter   = flag.Int("max-iter", 50, "max. number of k-means iterations")
	opt_min_change = flag.Float64("min-change", 0.01, "min. change in k-means centers to keep trying")
	opt_maxdiff    = flag.Bool("maxdiff", false, "use Maxdiff distance for k-means")
)

// rewrite a slice of strings in color, using color_fslice()
func color_slice(str []string) []string {
	vals := make([]float64, len(str))
	for i,s := range str { vals[i],_ = strconv.ParseFloat(s, 64) }
	return color_fslice(vals, 1.0)
}

// rewrite a slice of float64s in color, according to their values 0.0-1.0*fact
func color_fslice(vals []float64, fact float64) []string {
	ret := make([]string, len(vals))

	for i,val := range vals {
		fval := val * fact
		switch {
		case fval > 0.9: ret[i] = color.RedString("%.3f", val)
		case fval > 0.5: ret[i] = color.YellowString("%.3f", val)
		case fval > 0.3: ret[i] = color.MagentaString("%.3f", val)
		case fval > 0.1: ret[i] = color.CyanString("%.3f", val)
		case fval > 0.025: ret[i] = color.BlueString("%.3f", val)
		default: ret[i] = color.GreenString("%.3f", val)
		}
	}

	return ret
}

// percentage pretty-printer
func pcnt_pp(pcnt float64) string {
	str := make([]byte, 0, 100)
	for i := 0.0; i < pcnt; i += 0.01 {
		str = append(str, '#')
	}
	return string(str[:]) + fmt.Sprintf(" %.3g%%", pcnt*100.0)
}

// read a prefix2as file
func read_prefix2as(path string, p2a map[string]string) error {
	fh, err := os.Open(path)
	if err != nil { return err }

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] != '2' { continue }
		d := strings.Split(line, "\t")
		if len(d) != 3 { continue }

		prefix := d[0]
		plen,_ := strconv.Atoi(d[1])
		plen /= 4
		asn := d[2]

		// sanitize asn
		if len(asn) > 5 { asn = asn[0:5] }
		if i := strings.Index(asn, "_"); i > 0 { asn = asn[0:i] }

		if _,ok := p2a[prefix[0:8]]; !ok { p2a[prefix[0:8]] = asn }
		if _,ok := p2a[prefix[0:plen]]; !ok { p2a[prefix[0:plen]] = asn }
	}

	return nil
}

func main() {
	// parse args
	flag.Parse()

	// use opts
	if *opt_C { color.NoColor = false } // force

	p2a := make(map[string]string)
	if *opt_P != "" {
		err := read_prefix2as(*opt_P, p2a)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading prefix2as:", err)
			os.Exit(1)
		}
	}

	// prepare
	profiles := make(point.Points, 0)

	// read input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 || line[0] == '#' { continue }

		// parse the profile
		d := strings.Split(line, ",")
		if len(d) < 3 { continue }
		point := &point.Point{}
		point.D = d
		point.V = make([]float64, len(d)-2-*opt_F-*opt_L)
		for i,s := range d[2+*opt_F:len(d)-*opt_L] {
			point.V[i],_ = strconv.ParseFloat(s, 64)
		}

		// append
		profiles = append(profiles, point)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}

	if len(profiles) == 0 {
		fmt.Fprintln(os.Stderr, "no profiles read from stdin, exiting")
		os.Exit(1)
	}

	// find clusters
	var eps []float64
	var clusters []point.Points
	if *opt_kmeans == true {
		// find k clusters using k-means
		if *opt_maxdiff {
			clusters = kmeans.SearchDist(profiles, *opt_k, *opt_max_iter, *opt_min_change, point.Maxdiff)
		} else {
			clusters = kmeans.SearchDist(profiles, *opt_k, *opt_max_iter, *opt_min_change, point.Euclidean)
		}

		// sort clusters by number of prefixes
		sort.Slice(clusters, func(i,j int) bool {
			return len(clusters[i]) > len(clusters[j])
		})

	} else {
		eps = make([]float64, len(profiles[0].V))
		for i := range eps { eps[i] = *opt_eps }
		clusters = dbscan.Search(profiles, eps, *opt_min_points)

		// sort clusters by number of prefixes
		sort.Slice(clusters, func(i,j int) bool {
			if i == 0 { return true }
			if j == 0 { return false }
			return len(clusters[i]) > len(clusters[j])
		})
	}

	// print clusters
	for cluster,points := range clusters {
		// skip noise?
		if *opt_kmeans == false {
			if cluster == 0 && *opt_N == false { continue }
		}

		// sort points in cluster by prefix
		sort.Slice(points, func(i,j int) bool {
			return points[i].D.([]string)[0] < points[j].D.([]string)[0]
		})

		fmt.Printf("=== cluster #%d ===\n", cluster)
		for i := range points {
			d := points[i].D.([]string)

			// print prefix
			prefix := d[0]
			if len(prefix) > 8 && prefix[len(prefix)-3] == '/' {
				fmt.Printf("%20s ", prefix)
			} else {
				fmt.Printf("%s ", prefix)
			}
			d = d[2:]

			// print AS?
			if *opt_P != "" {
				as,ok := p2a[prefix[0:8]]
				if !ok {
					for l := len(prefix); l >= 4; l-- {
						if as,ok = p2a[prefix[0:l]]; ok { break }
					}
				}
				fmt.Printf("[%5s] ", as)
			}

			// print ignored nybbles (first)
			if *opt_F > 0 { fmt.Printf("(%s) ", strings.Join(d[0:*opt_F], " ")); d = d[*opt_F:] }

			// print nybbles used for clustering
			e := d[0:len(d)-*opt_L]
			if *opt_C { e = color_slice(e) }
			fmt.Printf("%s", strings.Join(e, " "))

			// print ignored nybbles (last)
			if *opt_L > 0 { fmt.Printf(" (%s)", strings.Join(d[len(d)-*opt_L:], " ")) }

			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	// print stats
	all := len(profiles)
	fmt.Printf("SUMMARY\n")
	fmt.Printf("-------\n")
	fmt.Printf("Analyzed prefixes: %d\n", all)

	fmt.Printf("Cluster summaries:\n")
	sse := 0.0
	for i := range clusters {
		// skip noise?
		if *opt_kmeans == false && i == 0 && *opt_N == false { continue }

		mean := clusters[i].Mean()
		sd := clusters[i].Stddev(mean)
		median := clusters[i].Median()
		pcnt := float64(len(clusters[i])) / float64(all)

		// cluster quality
		wss := clusters[i].Errors(mean).Sum()
		sse += wss

		fmt.Printf("cluster %2d: ", i)
		if *opt_C {
			fmt.Printf("%s %s (%.3g)\n", color_fslice(median.V, 1.0), pcnt_pp(pcnt), wss)
			if *opt_S {
				fmt.Printf("      mean: %s\n", color_fslice(mean.V, 1.0))
				fmt.Printf("    stddev: %s\n\n", color_fslice(sd.V, 4.0))
			}
		} else {
			fmt.Printf("%.3f %.3f %.3f\n", median.V, pcnt*100.0, wss)
			if *opt_S {
				fmt.Printf("      mean: %.3f\n", mean.V)
				fmt.Printf("    stddev: %.3f\n\n", sd.V)
			}
		}
	}
	fmt.Printf("%.2f\n", sse)

	if *opt_kmeans == false {
		noise := len(clusters[0])
		incluster := all - noise

		fmt.Printf("DBSCAN parameters: eps=%g, min_points=%d\n", *opt_eps, *opt_min_points)
		fmt.Printf("Analyzed prefixes: %d\n", all)
		fmt.Printf("Clusters found: %d\n", len(clusters))
		fmt.Printf("  prefixes in clusters: %d (%.1f%%)\n",
			incluster, float64(incluster) / float64(all) * 100.0)
		fmt.Printf("  noise: %d (%.1f%%)\n", noise, float64(noise) / float64(all) * 100.0)
	}
}
