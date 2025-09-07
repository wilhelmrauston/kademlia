package main

import (
	"github.com/wilhelmrauston/kademlia/internal/cli"
	"github.com/wilhelmrauston/kademlia/pkg/build"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime
	cli.Execute()
}
