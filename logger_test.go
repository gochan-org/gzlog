package gzlog

import (
	"os"
	"strings"
	"testing"
)

const (
	// used for testing gcl.rotate to make sure it properly splits up log files before writing if maxSize > 0
	writeStr = `What you're referring to as Linux, is in fact, GNU/Linux, or as I've recently taken to calling it, GNU plus Linux. Linux is not an operating system unto itself, but rather another free component of a fully functioning GNU system made useful by the GNU corelibs, shell utilities and vital system components comprising a full OS as defined by POSIX.
Many computer users run a modified version of the GNU system every day, without realizing it. Through a peculiar turn of events, the version of GNU which is widely used today is often called "Linux", and many of its users are not aware that it is basically the GNU system, developed by the GNU Project.
There really is a Linux, and these people are using it, but it is just a part of the system they use. Linux is the kernel: the program in the system that allocates the machine's resources to the other programs that you run. The kernel is an essential part of an operating system, but useless by itself; it can only function in the context of a complete operating system. Linux is normally used in combination with the GNU operating system: the whole system is basically GNU with Linux added, or GNU/Linux. All the so-called "Linux" distributions are really distributions of GNU/Linux.
- Richard Stallman`
	maxSize = 1 * 1000 // 1 kb
)

// yes I know I'm technically not supposed to rely on I/O for unit tests, but who cares

func populateLog(fn string, text string, t *testing.T) {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
		return
	}
	if _, err = f.WriteString(text); err != nil {
		t.Fatal(err)
	}
}

func generateLogs(t *testing.T) {
	err := os.Mkdir("logs", 0644)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
		return
	}
	populateLog("logs/gzlog.log", "blah blah blah blah blah", t)
	populateLog("logs/gzlog.1.log", "blah blah blah blah blah", t)
	populateLog("logs/gzlog.2.log", "blah blah blah blah blah", t)
}

func logTxt(gcl *GzLog, str string, t *testing.T) {
	_, err := gcl.Println(str)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestContinueLog(t *testing.T) {
	generateLogs(t)
	populateLog("logs/gzlog.3.log", "good", t)
	fn, err := getSuitableFile("logs", "gzlog", 5, 0644)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using %s", fn)
	if fn != "logs/gzlog.3.log" {
		t.Fatal("logs/gzlog.3.log should be short enough to be usable")
	}
}

func TestCreateNewLog(t *testing.T) {
	generateLogs(t)
	populateLog("logs/gzlog.3.log", "too big", t)
	fn, err := getSuitableFile("logs", "gzlog", 5, 0644)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using %s", fn)
	if fn != "logs/gzlog.4.log" {
		t.Fatal("we should be using logs/gzlog.4.log here")
	}
}

func TestOpenLog(t *testing.T) {
	dir := "logs"
	gcl, err := OpenFile(dir, "gzlog-newlog", maxSize, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	arr := strings.Split(writeStr, ".")
	for _, sentence := range arr {
		logTxt(gcl, sentence, t)
	}
	_, err = gcl.ReadAllString()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s size: %d", gcl.filename, gcl.Size())
}

func TestMaxSize(t *testing.T) {
	gcl, err := OpenFile("logs", "gzlog-nomax", 0, 0644)
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
	gcl, err := UseFile(os.Stdout, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	arr := strings.Split(writeStr, ".")
	for _, sentence := range arr {
		logTxt(gcl, sentence, t)
	}
}

func TestStderr(t *testing.T) {
	gcl, err := UseFile(os.Stderr, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer gcl.Close()
	arr := strings.Split(writeStr, ".")
	for _, sentence := range arr {
		logTxt(gcl, sentence, t)
	}
}
