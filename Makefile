default: all
all: profiles clusters ipv6-addr2hex ipv6-hex2addr

profiles: profiles.go
	go build -o profiles ./profiles.go

clusters: clusters.go
	go build -o clusters ./clusters.go

ipv6-addr2hex: ipv6-addr2hex.go
	go build -o ipv6-addr2hex ./ipv6-addr2hex.go

ipv6-hex2addr: ipv6-hex2addr.go
	go build -o ipv6-hex2addr ./ipv6-hex2addr.go

.PHONY: clean
clean:
	-rm -f ./clusters ./profiles ./ipv6-hex2addr ./ipv6-addr2hex
