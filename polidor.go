package polidor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var _ = bytes.MinRead // dummy

// dateFormat is the structure of dates in directory layouts eg "2015/08/11" in
// "/data/fb/1289/j1_readnews_com/2466/2015/08/11/09/18"
const dateFormat = "2006/01/02"

// Directory represents a directory for storage eg
// "/data/fb/1289/j1_readnews_com"
// TODO rename Directory to Database
type Directory struct {
	// Base directory for storage eg /data/fb
	Base string

	// CompanyID
	CompanyID int

	// Device name eg j1_readnews_com
	DeviceName string

	// Device Number eg 2466
	DeviceNumber int
}

// RetentionT is used for storing retention settings
type RetentionT map[int]SubRetentionT

// SubRetentionT is used for storing retention settings
type SubRetentionT map[string]int

// BuildDirectory is a constructor for the Directory struct. It takes a
// 'storage' like "/data/fb" and a 'path' like
// "/data/fb/1289/j1_readnews_com/2466/2015/08/11/09/18" and returns a
// 'Directory' struct
func BuildDirectory(storage string, path string) (Directory, error) {
	location := strings.Index(path, storage)
	if location == -1 {
		return Directory{}, fmt.Errorf("unable to match storage %s in path %s", storage, path)
	}

	location += len(storage)
	tail := path[location:]

	splits := strings.Split(tail, string(os.PathSeparator))
	if len(splits) < 4 {
		return Directory{}, fmt.Errorf("unable to find date in %s", path)
	}

	// here
	companyID, err := strconv.Atoi(splits[1])
	Check(err)
	deviceNumber, err := strconv.Atoi(splits[3])
	Check(err)

	d := Directory{storage, companyID, splits[2], deviceNumber}
	return d, nil
}

// StripStart strips directories, for example
// "/data/fb/1289/j1_readnews_com/2466/" from the start of a path eg
// "/data/fb/1289/j1_readnews_com/2466/2015/08/11/09/18" giving the date part of the path eg
// "2015/08/11/09/18"
func (d Directory) StripStart(path string) string {
	sep := string(os.PathSeparator)
	stripPath := filepath.Join(d.DeviceName, strconv.Itoa(d.DeviceNumber), sep) + sep
	location := strings.LastIndex(path, stripPath) + len(stripPath)
	return path[location:]
}

// Retain returns true if, on a given 'date', the 'path' should be
// kept if the retention period is 'period'
func (d Directory) Retain(date time.Time, path string, period int) (bool, error) {
	dateUTC := date.UTC() // belts and braces

	dateDir, err := d.ToEpoch(path)
	dateDirPeriod := dateDir.Add(time.Duration(period*24) * time.Hour)
	if err != nil {
		return false, err
	}
	if dateDirPeriod.Before(dateUTC) {
		return false, nil
	}

	return true, nil
}

// FindRetentionPeriod finds the retention period for a companyID eg 30 days,
// or returns the default retention period
func FindRetentionPeriod(retentionPeriods RetentionT, companyID int) int {
	defaultRetention := retentionPeriods[0]["retention"]
	retention := retentionPeriods[companyID]["retention"]
	if retention == 0 {
		return defaultRetention
	}
	return retention
}

// ToPath takes an epoch time eg 1470907113 and
// produces a path eg "/data/fb/1289/j1_readnews_com/2466/2015/08/11"
// ie discarding hours and minutes
func (d Directory) ToPath(epoch time.Time) string {
	return filepath.Join(
		d.Base,
		strconv.Itoa(d.CompanyID),
		d.DeviceName,
		strconv.Itoa(d.DeviceNumber),
		epoch.UTC().Format(dateFormat))
}

// ToEpoch taks a path eg "/data/fb/1289/j1_readnews_com/2466/2015/08/11"
// and returns an epoch time eg 1470907113
func (d Directory) ToEpoch(path string) (time.Time, error) {
	stripped := d.StripStart(path)
	result, err := time.Parse(dateFormat, stripped)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return result.UTC(), nil
}

// Check makes checking errors easy, so they actually get a minimal check
func Check(err error) {
	if err != nil {
		log.Fatalf("Check: %v", err)
	}
}
