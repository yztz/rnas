# windows: 
# 	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build .

.PHONY: example

example:
	go build -o rnas ./example

clean:
	- rm rnas
	
