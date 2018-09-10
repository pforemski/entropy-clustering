default: all
all: profiles clusters

profiles: profiles.go
	go build -o profiles ./profiles.go

clusters: clusters.go
	go build -o clusters ./clusters.go

.PHONY: clean
clean:
	-rm -f ./clusters ./profiles
