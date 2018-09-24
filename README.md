# ENTROPY CLUSTERING

Implementation of the Entropy Clustering algorithm for IPv6 networks, introduced in the ACM IMC'18 conference paper:

> [Clusters in the Expanse: Understanding and Unbiasing IPv6 Hitlists](https://ipv6hitlist.github.io/), *Oliver Gasser, Quirin Scheitle, Paweł Foremski, Qasim Lone, Maciej Korczyński, Stephen D. Strowes, Luuk Hendriks, Georg Carle*, ACM Internet Measurement Conference 2018, Boston, MA, USA

See [ipv6hitlist.github.io](https://ipv6hitlist.github.io/) for more details and output examples, as the one below:

![entropy clustering example](https://ipv6hitlist.github.io/eip/clusters-all-full-slash32-crop.png)

# PREREQUISITES

Install [Go](https://www.golang.org/) and the required packages:
```
go get github.com/pforemski/gouda/...
go get github.com/fatih/color
```

Also, if you want to plot the results using this code, install Matplotlib, e.g.
```
sudo apt-get install python-matplotlib
```

# USAGE

1. Compile.
```
make
```
2. Convert list of IPv6 addresses into a list of entropy profiles:
```
cat ips.txt | ./profiles > profiles.txt
```
3. Find entropy clusters using k-means, e.g. for k=6:
```
cat profiles.txt | ./clusters -kmeans -k 6 > clusters.txt
```
4. Finally, plot the results:
```
cat clusters.txt | ./plot-clusters.py
```

# AUTHOR
Written by Paweł Foremski, [@pforemski](https://twitter.com/pforemski), 2018.
