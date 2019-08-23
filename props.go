package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type bulkEntry struct {
	Env      string
	Mappings map[string]string
	Allowed  []string
}

type bulkConfig map[string]*bulkEntry

func setPropertiesInFile(filename string, bulkConfig bulkConfig, tmpdir string) error {

	// Open the properties file
	// ...for reading and writing, since we'll read it first time through and then possibly re-write it
	// ...also open with create flag to create the file, if absent
	propsFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to access properties file %s", filename)
	}
	//noinspection GoUnhandledErrorResult
	defer propsFile.Close()

	// Setup a temp file for making changes as we go
	tmpFile, err := ioutil.TempFile(tmpdir, "*.properties")
	if err != nil {
		return errors.Wrap(err, "unable to create temporary file for modifications")
	}
	// ...remove temp file after we leave here
	//noinspection GoUnhandledErrorResult
	defer os.Remove(tmpFile.Name())
	//noinspection GoUnhandledErrorResult
	defer tmpFile.Close()
	writer := bufio.NewWriter(tmpFile)

	commentsRe := regexp.MustCompile("#.*")
	propRe := regexp.MustCompile(`(.+?)\s*=\s*(.*)`)

	// Go through each line of the existing properties file
	var modified = false
	scanner := bufio.NewScanner(propsFile)
	for scanner.Scan() {
		resultLine := scanner.Text()

		// strip away commented parts of line
		line := commentsRe.ReplaceAllString(resultLine, "")

		// trim surrounding whitespace
		line = strings.TrimSpace(line)

		// see if remainder is a property setting line
		if groups := propRe.FindStringSubmatch(line); groups != nil {

			property := groups[1]
			if entry, entryExists := bulkConfig[property]; entryExists {
				delete(bulkConfig, property)

				value, err := resolveValue(property, entry)
				if err != nil {
					return err
				}

				if value != "" && value != groups[2] {
					modified = true
					log.Printf("Setting %s to %s in %s\n", property, value, filename)
					resultLine = fmt.Sprintf("%s=%s", property, value)
				}
			}
		}

		// write the newly modified or existing, if not a match, line to temp file
		_, err := fmt.Fprintln(writer, resultLine)
		if err != nil {
			return errors.Wrap(err, "failed to write to temp file")
		}
	}
	// Process properties that weren't in existing file
	for property, entry := range bulkConfig {
		modified = true
		value, err := resolveValue(property, entry)
		if err != nil {
			return err
		}
		if value != "" {
			log.Printf("Setting %s to %s in %s\n", property, value, filename)
			_, err = fmt.Fprintf(writer, "%s=%s\n", property, value)
			if err != nil {
				return errors.Wrap(err, "failed to write to temp file")
			}
		}
	}

	// If modification was needed
	if modified {
		// ...flush temp content to disk
		if err := writer.Flush(); err != nil {
			return errors.Wrap(err, "failed to flush temp content")
		}

		return copyOverTempFile(tmpFile, propsFile)
	}

	return nil
}

func resolveValue(property string, entry *bulkEntry) (string, error) {
	value := os.Getenv(entry.Env)
	if value != "" {
		if entry.Mappings != nil {
			if mappedValue, mappingExists := entry.Mappings[value]; mappingExists {
				value = mappedValue
			}
		}
		if isAllowed(entry.Allowed, value) {
			return value, nil
		} else {
			return "", errors.Errorf("Value '%s' for property %s is in allowed list: %v", value, property, entry.Allowed)
		}
	} else {
		return "", nil
	}
}

func copyOverTempFile(tmpFile *os.File, propsFile *os.File) error {
	// ...rewind the files to the beginning
	_, err := tmpFile.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "could not rewind temp content file")
	}
	_, err = propsFile.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "could not rewind properties file for re-writing")
	}
	// ...truncate original properties file since we're overwriting it
	if err := propsFile.Truncate(0); err != nil {
		return errors.Wrap(err, "failed to truncate properties file")
	}
	// ...copy modified, temp file content over into what was the original properties file
	_, err = io.Copy(propsFile, tmpFile)
	if err != nil {
		return errors.Wrap(err, "could not copy temp content to properties file")
	}
	return nil
}

func setBulkProperties(filename string, bulkDefinitionsFilename string, tmpdir string) error {
	bulkDefinitionsFile, err := os.Open(bulkDefinitionsFilename)
	if err != nil {
		return errors.Wrap(err, "unable to open bulk definitions file")
	}
	//noinspection GoUnhandledErrorResult
	defer bulkDefinitionsFile.Close()

	decoder := json.NewDecoder(bulkDefinitionsFile)

	var bulkConfig bulkConfig
	err = decoder.Decode(&bulkConfig)
	if err != nil {
		return errors.Wrap(err, "unable to decode bulk definitions")
	}

	return setPropertiesInFile(filename, bulkConfig, tmpdir)
}

func setSingleProperty(filename string, property string, envVar string, mappings map[string]string, allowed []string, tmpdir string) error {

	bulkConfig := make(bulkConfig)
	bulkConfig[property] = &bulkEntry{
		Env:      envVar,
		Mappings: mappings,
		Allowed:  allowed,
	}

	return setPropertiesInFile(filename, bulkConfig, tmpdir)
}

func isAllowed(allowed []string, value string) bool {
	if allowed != nil && len(allowed) > 0 {
		for _, v := range allowed {
			if value == v {
				return true
			}
		}

		return false
	} else {
		return true
	}
}
