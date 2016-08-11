package polidor

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// retentionT - if today is 'date' and we want to retain directories
// for n 'days', do we 'keep' the directory?
type retentionT struct {
	date time.Time
	days int
	keep bool
}

// testsRetentions contains data directories and
// dates and retention periods for testing
var testsRetentions = []struct {
	base         string
	companyID    int
	deviceName   string
	deviceNumber int
	path         string
	retentions   []retentionT
}{
	{
		"/data/fb",
		1289,
		"j1_readnews_com",
		2466,
		"/data/fb/1289/j1_readnews_com/2466/2015/08/11",

		[]retentionT{
			// comparing the path date (2015/08/11) against these
			// "run dates"

			// path is tomorrow (pathological case)
			{time.Date(2015, 8, 10, 0, 0, 0, 0, time.UTC), 0, true},
			{time.Date(2015, 8, 10, 0, 0, 0, 0, time.UTC), 1, true},

			// path is today
			{time.Date(2015, 8, 11, 0, 0, 0, 0, time.UTC), 0, true},
			{time.Date(2015, 8, 11, 0, 0, 0, 0, time.UTC), 1, true},

			// path is yesterday
			{time.Date(2015, 8, 12, 0, 0, 0, 0, time.UTC), 0, false},
			{time.Date(2015, 8, 12, 0, 0, 0, 0, time.UTC), 1, true},

			// path is two days ago
			{time.Date(2015, 8, 13, 0, 0, 0, 0, time.UTC), 0, false},
			{time.Date(2015, 8, 13, 0, 0, 0, 0, time.UTC), 1, false},
			{time.Date(2015, 8, 13, 0, 0, 0, 0, time.UTC), 2, true},
		},
	},
	{
		"/var/tmp",
		123,
		"j2_readnews_com",
		2467,
		"/var/tmp/generator/123/j2_readnews_com/2467/2016/07/15",
		[]retentionT{
			{time.Date(2016, 8, 15, 14, 39, 0, 0, time.UTC), 10, false},
		},
	},
}

// TestRetainPath tests Retain()
func TestRetainPath(t *testing.T) {
	for i, test := range testsRetentions {
		d := Directory{test.base,
			test.companyID,
			test.deviceName,
			test.deviceNumber,
		}
		for j, retention := range test.retentions {
			keep, err := d.Retain(retention.date, test.path, retention.days)
			if err != nil || keep != retention.keep {
				t.Errorf("(%d:%d) expected: %v got: %v %v",
					i, j, retention.keep, keep, err)
			}
		}
	}
}

// testsPaths contains data for testing conversions between data directories
// and epochs
var testsPaths = []struct {
	base         string
	companyID    int
	deviceName   string
	deviceNumber int
	epoch        time.Time
	path         string
}{
	{
		"/data/fb",
		1289,
		"j1_readnews_com",
		2466,
		time.Unix(1470873600, 0).UTC(), // Thu, 11 Aug 2015 00:00:00 GMT
		"/data/fb/1289/j1_readnews_com/2466/2016/08/11",
	},
}

// TestToPath tests ToPath()
func TestToPath(t *testing.T) {
	for i, test := range testsPaths {
		d := Directory{test.base, test.companyID, test.deviceName, test.deviceNumber}
		result := d.ToPath(test.epoch)
		if result != test.path {
			t.Errorf("%d: failed:\nexpected: %s\nresult:   %s",
				i, test.path, result)
		}
	}
}

// TestToEpoch tests ToEpoch()
func TestToEpoch(t *testing.T) {
	for i, test := range testsPaths {
		d := Directory{test.base, test.companyID, test.deviceName, test.deviceNumber}
		result, err := d.ToEpoch(test.path)
		epochNoSeconds := test.epoch.Truncate(time.Minute)
		if err != nil || result != epochNoSeconds {
			t.Errorf("%d: failed:\nexpected: %s\nresult:   %s",
				i, epochNoSeconds, result)
		}
	}
}

var testsBuildDirectory = []struct {
	base          string
	dataDirectory string
	result        Directory
	err           error
}{
	{
		"/data/fb",
		"/data/fb/1289/j1_readnews_com/2466/2015/08/11/09/18",
		Directory{"/data/fb", 1289, "j1_readnews_com", 2466},
		nil,
	},
	{
		"/var/tmp/generator",
		"/var/tmp/generator/123/j1_readnews_com",
		Directory{},
		errors.New("unable to find date in /var/tmp/generator/123/j1_readnews_com"),
	},
	{
		"/data/fb",
		"/foo/bar",
		Directory{},
		errors.New("unable to match storage /data/fb in path /foo/bar"),
	},
}

// TestBuildDirectory tests BuildDirectory()
func TestBuildDirectory(t *testing.T) {
	for i, test := range testsBuildDirectory {
		result, err := BuildDirectory(test.base, test.dataDirectory)
		if fmt.Sprint(err) != fmt.Sprint(test.err) {
			t.Errorf("%d: failed: got err '%v' expected err '%v'", i, err, test.err)
		}
		if result != test.result {
			t.Errorf("%d: failed: got %v expected %v", i, result, test.result)
		}
	}
}

/* date parsing: blame Russ Cox.
const (
    month/day/hour/minute/second 1/2/3/4/5
	year/tz 6/7
	and 24hr is 15 ie 3pm is 15:00 hrs

    stdLongMonth      = "January"
    stdMonth          = "Jan"
    stdNumMonth       = "1"
    stdZeroMonth      = "01"
    stdLongWeekDay    = "Monday"
    stdWeekDay        = "Mon"
    stdDay            = "2"
    stdUnderDay       = "_2"
    stdZeroDay        = "02"
    stdHour           = "15"
    stdHour12         = "3"
    stdZeroHour12     = "03"
    stdMinute         = "4"
    stdZeroMinute     = "04"
    stdSecond         = "5"
    stdZeroSecond     = "05"
    stdLongYear       = "2006"
    stdYear           = "06"
    stdPM             = "PM"
    stdpm             = "pm"
    stdTZ             = "MST"
    stdISO8601TZ      = "Z0700"  // prints Z for UTC
    stdISO8601ColonTZ = "Z07:00" // prints Z for UTC
    stdNumTZ          = "-0700"  // always numeric
    stdNumShortTZ     = "-07"    // always numeric
    stdNumColonTZ     = "-07:00" // always numeric
)
*/
