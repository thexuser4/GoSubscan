# GoSubscan
Validate if a sub domain exists from IP.

Runs concurrently and uses "FastHTTP" for ultra fast HTTP requests.

WOrkflow:
IP list --> DNA Lookup for domain --> see if subdomain "edc" - "edc99" is valid --> saves valid subdomains.


If you would like to compile install:
go get github.com/valyala/fasthttp
go get github.com/joeguo/tldextract


Use zmap to find IPs with port 2525 open!
