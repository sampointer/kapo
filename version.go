package main

//Name is the name of the software
const name string = "kapo"

//Version is the version of the software. It should be passed in at build time
//with '-ldflags "-X main.version=$VERSION"'. Commonly this is the release tag
//of the form 1.2.3 or, for development builds, the current git SHA.
var version string
