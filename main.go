package main

import (
	"flag"
	"fmt"
	"github.com/itzg/go-flagsfiller"
	"log"
	"os"
)

var (
	version string
	commit  string
	date    string
)

type args struct {
	Version bool `usage:"Print version information"`

	File     string            `aliases:"f" usage:"The properties [file] to update"`
	Property string            `aliases:"p" usage:"The [name] of the property to set"`
	EnvVar   string            `aliases:"e" usage:"The [name] of the environment variable to map from"`
	Allowed  []string          `aliases:"a" usage:"If specified, limits the [values] that can be passed from ENVVAR"`
	Mapping  map[string]string `aliases:"m" usage:"Defines mappings of a given value into property value to set. Syntax is [from=to]"`
	Bulk     string            `usage:"The name of a bulk definition JSON [file]"`
}

func main() {
	var args args

	err := flagsfiller.Parse(&args)
	if err != nil {
		log.Fatalf("Failed to parse flags: %s", err.Error())
	}

	if args.Version {
		fmt.Printf("set-property %s (%s @ %s)", version, commit, date)
		os.Exit(0)
	}

	if args.File == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Need to specify a properties file\n")
		flag.Usage()
		os.Exit(1)
	}

	if args.Bulk != "" {
		err := setBulkProperties(args.File, args.Bulk, "")
		if err != nil {
			log.Fatalf("Failed to bulk-set properties in file: %s", err.Error())
		}
	} else if args.Property != "" && args.EnvVar != "" {
		err := setSingleProperty(args.File, args.Property, args.EnvVar, args.Mapping, args.Allowed, "")
		if err != nil {
			log.Fatalf("Failed to set property in file: %s", err.Error())
		}
	} else {
		fmt.Println("Need to pass single property definition or bulk file")
		flag.Usage()
		os.Exit(1)
	}
}
