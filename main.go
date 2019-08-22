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
	Property string     `arg:"-p,required"`
	EnvVar   string     `arg:"-e,required"`
	File     string     `arg:"-f,required"`
	Allowed  []string   `help:"If specified, limits the values that can be passed from ENVVAR"`
	Map      []keyValue `arg:"-m" help:"Defines mappings of a given value into property value to set. Syntax is from=to"`
}

func (args) Version() string {
	return fmt.Sprintf("set-property %s (%s @ %s)", version, commit, date)
}

func (args) Description() string {
	return "Conditionally updates a given properties file when the given environment variable is set"
}

func main() {
	var args args
	arg.MustParse(&args)

	value, exists := os.LookupEnv(args.EnvVar)
	if !exists {
		// not set, so quickly get out of here
		return
	}

	value = mapValue(args, value)

	if !isAllowed(args, value) {
		log.Fatalf("Value '%s' from %s is not in allowed list: %v", value, args.EnvVar, args.Allowed)
	}

	err := setPropertyInFile(args.File, args.Property, value, "")
	if err != nil {
		log.Fatalf("Failed to set property in file: %s", err.Error())
	}
}

func mapValue(args args, value string) string {
	if len(args.Map) == 0 {
		return value
	}

	for _, entry := range args.Map {
		if entry.From == value {
			return entry.To
		}
	}

	return value
}

func isAllowed(args args, value string) bool {
	if len(args.Allowed) == 0 {
		return true
	}

	for _, v := range args.Allowed {
		if value == v {
			return true
		}
	}

	return false
}
