# Alignment and Haiku-detector

For a more polished write up of what this is all about, see this FT Labs blog post: http://labs.ft.com/2016/07/finding-hidden-haiku/

Initially an exercise to see if Go is usable (for me) as a Ruby+Sinatra replacement. Hence, lots of bad go. Idioms? What idioms?

Uses the CMUDict to parse article bodies and titles in to phonemes and syllables (see the rhyme folder) and looks for poetic-y snippets in FT Articles.

Currently (as of 07/02/2016) [deployed live](https://ftlabs-alignment.herokuapp.com/) for playing with.

## installing (on Windows)

* download the latest stable golang version from https://golang.org/doc/install
* clone this repo
* ensure you have your $GOPATH set up ok, in particular with this project sitting in $GOPATH/src/github.com/railsagainstignorance/alignment
* get and install the 3rd party imports (which ends up in $GOPATH/src/... etc)
   * $ go get     github.com/joho/godotenv
   * $ go install github.com/joho/godotenv
   * ditto for github.com/Financial-Times/ft-s3o-go/s3o
* create a .env file and add a valid FT Search API key
   * SAPI_KEY=...

## building and running

* $ go install github.com/railsagainstignorance/alignment
* $ $GOPATH/bin/alignment.exe

## deploying to heroku

* first get godep installed (follow [Heroku's instructions](https://devcenter.heroku.com/articles/deploying-go))
* invoke godep 
   * $ godep save -r ./...
* add/commit the new folder+files or changes thereto
* push to heroku

## Notes

* The article data is taken from the Financial Times' Search API and Content API.
* Error checking? Nope, not much.
* The /ontology route is restricted by s3o, Staff Single Sign On, requiring signing in using FT Staff credentials. This restriction may be lifted sometime.
