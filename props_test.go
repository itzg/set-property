package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

//noinspection GoUnhandledErrorResult
func TestSetPropertyInFile_NewFile(t *testing.T) {

	var tests = []struct {
		Name            string
		ExistingContent string
		Property        string
		Value           string
		ExpectedContent string
	}{
		{
			Name:            "NewFile",
			Property:        "prop1",
			Value:           "value1",
			ExpectedContent: "prop1=value1\n",
		},
		{
			Name:            "ExistingNoChange",
			Property:        "prop1",
			Value:           "value1",
			ExistingContent: "prop1 = value1 # shouldn't touch this line\n",
			ExpectedContent: "prop1 = value1 # shouldn't touch this line\n",
		},
		{
			Name:            "ExistingNeedChange",
			Property:        "prop1",
			Value:           "value2",
			ExistingContent: "# just a comment line\nprop1 = value1 # touch this line\nprop2 = valueB",
			// Tests a few things:
			// - leaves non property lines alone
			// - leaves non matching lines alone
			// - adds a final newline
			// - modifies matching property even in the middle of file
			ExpectedContent: "# just a comment line\nprop1=value2\nprop2 = valueB\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mainDir, err := ioutil.TempDir("", "set-property-main-")
			require.NoError(t, err)
			defer os.RemoveAll(mainDir)
			tmpdir, err := ioutil.TempDir("", "set-property-temp-")
			require.NoError(t, err)
			defer os.RemoveAll(mainDir)

			propFilename := path.Join(mainDir, fmt.Sprintf("TestSetPropertyInFile_%s.properties", tt.Name))

			if tt.ExistingContent != "" {
				createdFile, err := os.Create(propFilename)
				require.NoError(t, err)

				_, err = createdFile.Write([]byte(tt.ExistingContent))
				createdFile.Close()
				require.NoError(t, err)
			}

			varName := "TestSetPropertyInFile"
			os.Setenv(varName, tt.Value)
			err = setSingleProperty(propFilename,
				tt.Property, varName, nil, nil, tmpdir)
			require.NoError(t, err)

			propFile, err := os.Open(propFilename)
			require.NoError(t, err)
			defer propFile.Close()
			propContent, err := ioutil.ReadAll(propFile)
			require.NoError(t, err)

			assert.Equal(t, tt.ExpectedContent, string(propContent))

			infos, err := ioutil.ReadDir(tmpdir)
			require.NoError(t, err)
			assert.Empty(t, infos)
		})
	}
}
