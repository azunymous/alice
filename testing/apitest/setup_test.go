package apitest

import "gopkg.in/h2non/baloo.v3"

const apiURL = "http://localhost:8080"

// test stores the HTTP testing client preconfigured
var test = baloo.New(apiURL)
