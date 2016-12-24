# negroni-cache

[![Go Report Card](https://goreportcard.com/badge/github.com/trumanw/negroni-cache)](https://goreportcard.com/report/github.com/trumanw/negroni-cache)
[![Build Status](https://travis-ci.org/trumanw/negroni-cache.svg?branch=master)](https://travis-ci.org/trumanw/negroni-cache)
[![Coverage Status](https://coveralls.io/repos/github/trumanw/negroni-cache/badge.svg?branch=master)](https://coveralls.io/github/trumanw/negroni-cache?branch=master)

A standard compatible([RFC 7234](http://www.rfc-base.org/rfc-7234.html)) HTTP Cache middleware for [negroni](https://github.com/urfave/negroni).

# Usage

~~~ go
package main

import (
    "fmt"
    "net/http"

    "github.com/urfave/negroni"
    cah "github.com/trumanw/negroni-cache"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
          fmt.Fprintf(w, "Welcome to the home page!")
    })

    n := negroni.Classic()
    n.Use(cah.NewMiddleware(cah.NewMemoryCache()))
    n.UseHandler(mux)
    n.Run(":3000")
}

~~~
