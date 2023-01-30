all:
	go build -o webtool cmd/main.go

clean:
	rm -f webtool
