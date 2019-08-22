package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func setPropertyInFile(filename string, property string, value string, tmpdir string) error {
	propsFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to access properties file %s", filename)
	}
	//noinspection GoUnhandledErrorResult
	defer propsFile.Close()

	tmpFile, err := ioutil.TempFile("", "*.properties")
	if err != nil {
		return errors.Wrap(err, "unable to create temporary file for modifications")
	}
	//noinspection GoUnhandledErrorResult
	defer os.Remove(tmpFile.Name())
	//noinspection GoUnhandledErrorResult
	defer tmpFile.Close()
	writer := bufio.NewWriter(tmpFile)

	commentsRe := regexp.MustCompile("#.*")
	propRe := regexp.MustCompile(`(.+?)\s*=\s*(.*)`)

	var modified = false
	var found = false
	scanner := bufio.NewScanner(propsFile)
	for scanner.Scan() {
		resultLine := scanner.Text()

		// strip away commented parts of line
		line := commentsRe.ReplaceAllString(resultLine, "")

		// trim surrounding whitespace
		line = strings.TrimSpace(line)

		// see if remainder is a property setting line
		if groups := propRe.FindStringSubmatch(line); groups != nil {
			// our property?
			if groups[1] == property {
				found = true
				// value differs from what we want to set?
				if groups[2] != value {
					resultLine = fmt.Sprintf("%s=%s", property, value)
					modified = true
				}
			}
		}

		_, err := fmt.Fprintln(writer, resultLine)
		if err != nil {
			return errors.Wrap(err, "failed to write to temp file")
		}
	}
	if !found {
		modified = true
		if _, err := fmt.Fprintf(writer, "%s=%s\n", property, value); err != nil {
			return errors.Wrap(err, "failed to write to temp file")
		}
	}

	if modified {
		if err := writer.Flush(); err != nil {
			return errors.Wrap(err, "failed to flush temp content")
		}

		_, err = tmpFile.Seek(0, 0)
		if err != nil {
			return errors.Wrap(err, "could not rewind temp content file")
		}
		_, err = propsFile.Seek(0, 0)
		if err != nil {
			return errors.Wrap(err, "could not rewind properties file for re-writing")
		}
		if err := propsFile.Truncate(0); err != nil {
			return errors.Wrap(err, "failed to truncate properties file")
		}

		_, err := io.Copy(propsFile, tmpFile)
		if err != nil {
			return errors.Wrap(err, "could not copy temp content to properties file")
		}
	}

	return nil
}
