all : bin/hs110-exporter

bin/hs110-exporter :
	go build -o $@ .

