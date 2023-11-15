// Checks the license key(s) of package.yml files against the list of SPDX identifiers.
// Assumes you are running from the root of the package monorepo.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

const licenseListJSONsource = "https://spdx.org/licenses/licenses.json"
const licenseListJSONfile = "common/licenses.json"

// https://mholt.github.io/json-to-go/
type licensesData struct {
	LicenseListVersion string `json:"licenseListVersion"`
	Licenses           []struct {
		Reference             string   `json:"reference"`
		IsDeprecatedLicenseID bool     `json:"isDeprecatedLicenseId"`
		DetailsURL            string   `json:"detailsUrl"`
		ReferenceNumber       int      `json:"referenceNumber"`
		Name                  string   `json:"name"`
		LicenseID             string   `json:"licenseId"`
		SeeAlso               []string `json:"seeAlso"`
		IsOsiApproved         bool     `json:"isOsiApproved"`
		IsFsfLibre            bool     `json:"isFsfLibre,omitempty"`
	} `json:"licenses"`
	ReleaseDate string `json:"releaseDate"`
}

type yamlDataSlice struct {
	License []string `yaml:"license"`
}
type yamlDataString struct {
	License string `yaml:"license"`
}

func main() {

	// If no licenses.json file exists in the current directory, download it.
	if _, err := os.Stat(licenseListJSONfile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No licenses.json file found, fetching new")
		cmd := exec.Command("wget", licenseListJSONsource, "-O", licenseListJSONfile)
		if err := cmd.Run(); err != nil {
			fmt.Println("wget download error with the following URL:")
			fmt.Println(licenseListJSONsource)
			os.Remove(licenseListJSONfile)
			log.Fatal(err)
		}
	}

	// Open the licenses.json file and un-marshall it
	JSONfileContent, err := os.ReadFile(licenseListJSONfile)
	if err != nil {
		log.Fatal("Error when opening file: ", licenseListJSONfile, err)
	}
	var licensesJSON licensesData
	err = json.Unmarshal(JSONfileContent, &licensesJSON)
	if err != nil {
		log.Fatal("JSON Un-marshall error: ", err)
	}

	// Put all the valid identifiers in a list
	var validIdentifiers []string
	for _, entry := range licensesJSON.Licenses {
		if !entry.IsDeprecatedLicenseID {
			validIdentifiers = append(validIdentifiers, entry.LicenseID)
		}
	}
	// Walk the "package" directory, passing any found package.ymls off to the checker
	err = filepath.Walk("packages", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Fatal("File walk error: ", err)
		}
		//fmt.Println(path)
		if info.Name() == "package.yml" {
			licenseCheck(path, validIdentifiers)
		}
		return nil
	})
}

// The checker functions
// Yaml is parsed twice because the license key can either contain a single string or a list of strings
func licenseCheck(yamlPath string, licenseList []string) {
	YAMLfileContent, err := os.ReadFile(yamlPath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var yamlParseString yamlDataString
	var yamlParseSlice yamlDataSlice

	// Try parsing the yaml with a single string at the license key
	err = yaml.Unmarshal([]byte(YAMLfileContent), &yamlParseString)

	// if it errors, assume it is because the license is a list
	if err != nil {
		//fmt.Println("Switching to multi-license parsing")
		err = yaml.Unmarshal([]byte(YAMLfileContent), &yamlParseSlice)
		if err !=nil {
			log.Fatal("Error parsing multiline: ", err)
		}
		for _, id := range yamlParseSlice.License {
			if slices.Contains(licenseList, id) {
				//fmt.Println("ok: ", yamlPath, " ", id)
			} else {
				fmt.Println("BAD: ", yamlPath, " ", id)
				
			}
		}
		return // return early for yamls with license list
	}
	if slices.Contains(licenseList, yamlParseString.License) {
		//fmt.Println("ok: ", yamlPath, " ", yamlParseString.License)
	} else {
		fmt.Println("BAD: ", yamlPath, " ", yamlParseString.License)
	}
}
