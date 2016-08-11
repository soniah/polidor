package main

// cleaner cleans up old directories under the directory 'storageDirectory',
// using the configuration file 'retentions.yml' for directory
// retention periods

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	p "github.com/soniah/polidor"
	yaml "gopkg.in/yaml.v2"
)

const (
	// location of data storage
	// const for simplicity, could be a cmdline arg
	storageDirectory = "/var/tmp/generator"

	// location of config
	// const for simplicity, could be a cmdline arg
	configFile = "retentions.yml"

	// timeout
	// const for simplicity, could be a cmdline arg
	timeoutPeriod = 1 * time.Second
)

func main() {
	var err error

	// read in config
	dat, err := ioutil.ReadFile(configFile)
	p.Check(err)
	retentionPeriods := make(p.RetentionT)
	err = yaml.Unmarshal([]byte(dat), &retentionPeriods)
	p.Check(err)

	walker(storageDirectory, retentionPeriods)
}

// walker uses filepath.Walk() to do a depth-first walk down the
// 'storage' directory, removing any directories that have passed their
// retention period
func walker(storage string, retentionPeriods p.RetentionT) {
	var err error
	timeout := time.After(time.Duration(timeoutPeriod))

	// visit is the callback function for filepath.Walk()
	var visit = func(path string, f os.FileInfo, starterr error) error {
		// ignore any 'starterrs', they're caused by disappearing directories
		// from call to os.RemoveAll()

		// are we done?
		select {
		case <-timeout:
			return errors.New("timed out")
		default:
			// non blocking ie keep going
		}

		fmt.Printf("checking path: %s\n", path)
		path = filepath.Dir(path) // remove trailing filename

		directory, err := p.BuildDirectory(storage, path)
		if err != nil {
			return nil // not interested in this directory, keep going
		}

		rPeriod := p.FindRetentionPeriod(retentionPeriods, directory.CompanyID)
		keep, err := directory.Retain(time.Now(), path, rPeriod)
		// HACK: err returned by Retain() should return a custom error type
		// that can be checked more easily
		if strings.Contains(fmt.Sprintf("%s", err), "cannot parse") {
			return nil // not interested in this directory, keep going
		}
		p.Check(err)

		if keep {
			// filepath doco:
			// If the function returns SkipDir when invoked on a directory, Walk
			// skips the directory's contents entirely.
			return filepath.SkipDir
		}
		err = os.RemoveAll(path)
		p.Check(err)
		return nil
	}

	err = filepath.Walk(storageDirectory, visit)

	if fmt.Sprintf("%s", err) == "timed out" {
		fmt.Println()
		fmt.Println("==== timed out ====")
		fmt.Println()
	} else {
		p.Check(err)
	}
}
