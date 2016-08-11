package main

// generate test data under "dataDir"

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	p "github.com/soniah/polidor"
	"gopkg.in/yaml.v2"
)

func init() {
	// only seed random number generator once
	rand.Seed(time.Now().UTC().UnixNano())
}

const (
	// location of data storage
	// const for simplicity, could be a cmdline arg
	dataDir = "/var/tmp/generator"

	// location of config
	// const for simplicity, could be a cmdline arg
	configFile = "generator.yml"
)

type subyamlT map[int]string
type yamlT map[int]subyamlT

func main() {
	var err error

	// delete and make dataDir
	err = os.RemoveAll(dataDir)
	p.Check(err)
	err = os.Mkdir(dataDir, 0755)
	p.Check(err)

	// read in config
	dat, err := ioutil.ReadFile(configFile)
	p.Check(err)

	// unmarshal yaml
	datmap := make(yamlT)
	err = yaml.Unmarshal([]byte(dat), &datmap)
	p.Check(err)

	// generate files
	for companyID, devices := range datmap {
		for deviceID, deviceName := range devices {
			err = fillDirectories(companyID, deviceName, deviceID)
			p.Check(err)
		}

	}
}

// generate 100 random directories. Directories have path dates  between now
// and one month ago. Fill directories with files.
func fillDirectories(companyID int, deviceName string, deviceID int) error {

	directory := p.Directory{dataDir, companyID, deviceName, deviceID}
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	for i := 0; i < 100; i++ {
		path := directory.ToPath(randomTime(oneMonthAgo, now))
		err := os.MkdirAll(path, 0755)
		p.Check(err)
		fmt.Println(path)
		err = randomFiles(path)
		p.Check(err)
	}
	return nil
}

// generate 50 empty files at 'path'
func randomFiles(path string) error {
	for i := 0; i < 50; i++ {
		filename := fmt.Sprintf("%s/%d.dat", path, i)
		// TODO create larger files using /dev/urandom
		f, err := os.Create(filename)
		p.Check(err)
		err = f.Close()
		p.Check(err)
	}
	return nil
}

// generate a random time between 'min' and 'max'
func randomTime(min time.Time, max time.Time) time.Time {
	minUnix := min.Unix()
	maxUnix := max.Unix()
	return time.Unix(random(minUnix, maxUnix), 0)
}

// generate a random number between 'min' and 'max', suitable for
// treating as an epoch
func random(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}
