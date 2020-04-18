# pocket2rm
`pocket2rm` is a tool to get articles from read-later platform [pocket](https://app.getpocket.com/) to the [reMarkable paper tablet](https://remarkable.com/).

- retrieve URLs for articles from pocket
- PDFs are downloaded directly
- webpages are converted to a [readable format](github.com/go-shiori/go-readability) and converted to epub
- folder with PDFs and epubs can be synced using the reMarkable desktop application

## Getting started
- create a pocket application: [https://getpocket.com/developer/apps/new](https://getpocket.com/developer/apps/new) to obtain a `consumerKey`. The application only needs the 'Retrieve' permission.
- run `go run pocket2rm setup`:
  - interactive prompt asks for `consumerKey`
  - pocket user is redirected to pocket page to accept application
  - after confirmation in prompt, the `consumerKey` and `accessToken` are written to `.pocket2mr`
- run `go run pocket2rm.go`

## Improvements
- extend with automatic push to reMarkable using [this library](https://github.com/juruen/rmapi)
- process articles concurrently for higher throughput
- check which articles are already retrieved

## Non-goals
- support other read-later services / e-reader targets
