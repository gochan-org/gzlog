package gzlog

import (
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
)

const (
	// used for testing gcl.rotate to make sure it properly splits up log files before
	// writing if maxSize > 0
	writeStr = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua`
	maxSize  = 400 // size in bytes
)

// yes I know I'm technically not supposed to rely on I/O for unit tests, but who cares

func populateLog(fn string, text string, t *testing.T) {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
		return
	}
	if _, err = f.WriteString(text); err != nil {
		t.Fatal(err)
	}
}

func generateLogs(basename string, t *testing.T) {
	populateLog(basename+".log", "blah blah blah blah blah", t)
	populateLog(basename+".1.log", "blah blah blah blah blah", t)
	populateLog(basename+".2.log", "blah blah blah blah blah", t)
}

func logTxt(gcl *GzLog, str string, t *testing.T) {
	_, err := gcl.Println(str)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func splitTxt(text string) []string {
	re := regexp.MustCompile(`[\.\s]`)
	return re.Split(text, -1)
}

func TestContinueLog(t *testing.T) {
	// write crap to the log file to make sure it uses a file that isn't too big
	basename := path.Join(t.TempDir(), "gzlog-"+t.Name())
	generateLogs(basename, t)
	populateLog(basename+".3.log", "good", t)
	fn, err := getSuitableFile(basename, 5, 0644)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using %s", path.Base(fn))

	expected := basename + ".3.log"
	if fn != expected {
		t.Fatal(expected, "should be short enough to be usable")
	}
}

func TestCreateNewLog(t *testing.T) {
	// write crap to the log file to make sure it uses a file that isn't too big
	// and creates a new one
	basename := path.Join(t.TempDir(), "gzlog-"+t.Name())
	generateLogs(basename, t)
	populateLog(basename+".3.log", "this is still too big", t)
	fn, err := getSuitableFile(basename, 5, 0644)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using %s", path.Base(fn))
	if fn != basename+".4.log" {
		t.Fatalf("we should be using %s.4.log here", path.Base(fn))
	}
}

func TestOpenLog(t *testing.T) {
	// actually test opening a log file and working with it
	basename := path.Join(t.TempDir(), "gzlog-"+t.Name())
	gcl, err := OpenFile(basename, maxSize, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	arr := splitTxt(writeStr)
	for _, sentence := range arr {
		logTxt(gcl, sentence, t)
	}
	// _, err = gcl.ReadAllString()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s size: %d", gcl.Filename(true), gcl.Size())
}

func TestMaxSize(t *testing.T) {
	basename := path.Join(t.TempDir(), "gzlog-"+t.Name())
	gcl, err := OpenFile(basename, 0, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	arr := strings.Split(writeStr, ".")
	for _, sentence := range arr {
		logTxt(gcl, sentence, t)
	}
}

func TestStdout(t *testing.T) {
	gcl, err := ImportFile(os.Stdout, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	// arr := splitTxt(writeStr)
	// for _, sentence := range arr {
	// 	logTxt(gcl, sentence, t)
	// }
}

func TestStderr(t *testing.T) {
	gcl, err := ImportFile(os.Stderr, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	// arr := splitTxt(writeStr)
	// for _, sentence := range arr {
	// 	logTxt(gcl, sentence, t)
	// }
}
