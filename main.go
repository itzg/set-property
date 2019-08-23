package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"log"
	"os"
	"strings"
)

var (
	version string
	commit  string
	date    string
)

type keyValue struct {
	From, To string
}

func (kv *keyValue) UnmarshalText(b []byte) error {
	s := string(b)
	parts := strings.SplitN(s, "=", 2)

	if len(parts) != 2 {
		return fmt.Errorf("missing = in %s", s)
	}

	kv.From = parts[0]
	kv.To = parts[1]

	return nil
}

type args struct {
	File     string     `arg:"-f,required"`
	Property string     `arg:"-p"`
	EnvVar   string     `arg:"-e"`
	Allowed  []string   `arg:"-a" help:"If specified, limits the values that can be passed from ENVVAR"`
	Map      []keyValue `arg:"-m" help:"Defines mappings of a given value into property value to set. Syntax is from=to"`
	Bulk     string     `help:"The name of a bulk definition JSON file"`
}

func (args) Version() string {
	return fmt.Sprintf("set-property %s (%s @ %s)", version, commit, date)
}

func (args) Description() string {
	return "Conditionally updates a given properties file when the given environment variable is set"
}

func main() {
	var args args
	argsParser := arg.MustParse(&args)

	if args.Bulk != "" {
		err := setBulkProperties(args.File, args.Bulk, "")
		if err != nil {
			log.Fatalf("Failed to bulk-set properties in file: %s", err.Error())
		}
	} else if args.Property != "" && args.EnvVar != "" {
		mappings := make(map[string]string)
		for _, entry := range args.Map {
			mappings[entry.From] = entry.To
		}
		err := setSingleProperty(args.File, args.Property, args.EnvVar, mappings, args.Allowed, "")
		if err != nil {
			log.Fatalf("Failed to set property in file: %s", err.Error())
		}
	} else {
		fmt.Println("Need to pass single property definition or bulk file")
		argsParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
}
