# My First Go Program

Mostly an exercise to see if Go is usable (for me) as a Ruby+Sinatra replacement. Hence, lots of bad go. Idioms? What idioms?

## installing (on Windows)

* download the latest stable golang version from https://golang.org/doc/install
* clone this repo
* ensure you have your $GOPATH set up ok (in a parent folder of your project folder)
* get and install the 3rd party import (which ends up in $GOPATH/src/... and $GOPATH/pkg/...)
   * $ go get     github.com/joho/godotenv
   * $ go install github.com/joho/godotenv
* create a .env file and add a valid FT Search API key
   * SAPI_KEY=...

## building and running

* $ go build web-server.go
* $ ./web-server.exe

## Notes

The article data is taken from the Financial Times' Search API. Its the most recent 100 (max) articles which contain the user-supplied phrase.