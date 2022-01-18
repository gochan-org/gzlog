package gzlog

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const (
	fnFmt      = "%s.%d.log"
	fileFlags  = os.O_CREATE | os.O_APPEND
	logTimeFmt = "2006/01/02 15:04:05 "
)

var (
	ErrInvalidSize = errors.New("log size must be a positive number")
)

func getPrefix() string {
	prefix := time.Now().Format(logTimeFmt)
	return prefix
}

func gzipFile(fn string, mode fs.FileMode) error {
	gzPath := fn + ".gz"
	file, err := os.OpenFile(gzPath, os.O_WRONLY|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	ba, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	zw := gzip.NewWriter(file)
	defer zw.Close()

	_, err = zw.Write(ba)
	return err
}

func exists(fn string) bool {
	_, err := os.Stat(fn)
	return err == nil || !os.IsNotExist(err)
}

func getSuitableFile(logdir string, basename string, maxSize int, mode fs.FileMode) (string, error) {
	if maxSize < 0 {
		return "", ErrInvalidSize
	}
	num := 0
	fn := basename + ".log"
	absPath := path.Join(logdir, fn)
	for {
		absPath = path.Join(logdir, fn)
		fi, err := os.Stat(absPath)
		if err != nil {
			// file doesn't exist, use it.
			return absPath, nil
		}

		if strings.HasSuffix(fn, ".gz") {
			// file is an archive, moving on
			continue
		}
		if fi.Size() < int64(maxSize) || maxSize == 0 {
			// file isn't too big, use this
			break
		}
		if exists(absPath + ".gz") {
			// file is already archived
		} else {
			// file is too big but hasn't been archived yet. Archive it and move on
			gzipFile(absPath, mode)
		}
		num++
		fn = fmt.Sprintf(fnFmt, basename, num)
	}
	return absPath, nil
}

// GzLog is a logging tool for writing text to log files and automatically compressing and rotating them
// if the current log file is larger than the maxSize before writing to it to avoid having huge log files
// that can be tedious to use for debugging
type GzLog struct {
	basename     string
	filename     string
	dir          string
	file         *os.File
	stat         os.FileInfo
	maxSize      int64
	externalFile bool
}

// Close cleans up the log file
func (gl *GzLog) Close() error {
	if gl.file == nil || gl.file == os.Stdout || gl.file == os.Stderr {
		return nil
	}
	return gl.file.Close()
}

// Filename returns the filename of the current log file
func (gl *GzLog) Filename() string {
	return gl.filename
}

func (gl *GzLog) IsExternalFile() bool {
	return gl.externalFile
}

// MaxSize returns the maximum size set when GzLog was created. If maxSize == 0, the log
// will never be rotated (i.e. it will essentially have no maximum file size). This defeats
// the purpose of this package, but I figured I may as well include it anyway
func (gl *GzLog) MaxSize() int64 {
	return gl.maxSize
}

// ReadAllString reads the contents of the current log file and returns the string and any errors
func (gl *GzLog) ReadAllString() (string, error) {
	ba, err := gl.ReadAll()
	return string(ba), err
}

// ReadAll reads the contents of the current log file into a byte array and returns any errors
func (gl *GzLog) ReadAll() ([]byte, error) {
	size := gl.stat.Size()
	ba := make([]byte, size)
	_, err := gl.file.ReadAt(ba, 0)
	return ba, err
}

// Print behaves similarly to fmt.Print and log.Print
func (gl *GzLog) Print(a ...interface{}) (string, error) {
	str := fmt.Sprint(a...)
	err := gl.writeStr(str, true)
	return str, err
}

// Printf behaves similarly to fmt.Printf and log.Printf
func (gl *GzLog) Printf(format string, a ...interface{}) (string, error) {
	str := fmt.Sprintf(format, a...)
	err := gl.writeStr(str, true)
	return str, err
}

// Println behaves similarly to fmt.Println and log.Println
func (gl *GzLog) Println(a ...interface{}) (string, error) {
	str := fmt.Sprintln(a...)
	err := gl.writeStr(str, true)
	return str, err
}

func (gl *GzLog) writeStr(str string, rotate bool) error {
	var err error
	if rotate {
		err = gl.rotate()
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}
	if _, err = gl.file.WriteString(getPrefix() + str + "\n"); err != nil {
		return err
	}
	return err
}

func (gl *GzLog) resetStat() error {
	var err error
	gl.stat, err = gl.file.Stat()
	return err
}

// Size returns the file size of the current log file in bytes
func (gl *GzLog) Size() int64 {
	if gl.file == os.Stdout || gl.file == os.Stderr {
		return 0
	}
	gl.resetStat()
	return gl.stat.Size()
}

// FileMode returns the UNIX file mode (e.g. 0644)
func (gl *GzLog) FileMode() fs.FileMode {
	gl.resetStat()
	return gl.stat.Mode()
}

// GZip compresses the log file in gz format and returns any errors
func (gl *GzLog) GZip() error {
	if gl.file == os.Stdout || gl.file == os.Stderr {
		return nil
	}
	gl.resetStat()
	return gzipFile(gl.filename, gl.stat.Mode())
}

// rotate checks to see if the file is too big and should be archived. If it is, it archives it
// and opens a new one
func (gl *GzLog) rotate() error {
	if gl.file == os.Stdout || gl.file == os.Stderr {
		return nil
	}
	var err error
	if gl.Size() < gl.maxSize || gl.maxSize == 0 {
		return nil
	}
	mode := gl.FileMode()
	gl.Close()
	gl.filename, err = getSuitableFile(gl.dir, gl.basename, int(gl.maxSize), mode)
	if err != nil {
		return err
	}
	gl.file, err = os.OpenFile(gl.filename, fileFlags, mode)
	if err != nil {
		return err
	}
	gl.stat, err = gl.file.Stat()
	return err
}

// OpenFile opens the log in the specified log directory and basename, creating logdir/ if it doesn't
// exist, and creating or rotating a new log file as necessary. Do not include the extension in
// basename, gzlog will create logdir/basename.log or logdir/basename.#.log and automatically
// compress the log if it exceeds maxSize.
//
// If maxSize == 0, the log will never be rotated (i.e. it will essentially have no maximum file size).
// This defeats the purpose of this package, but I figured I may as well include it anyway
func OpenFile(logdir string, basename string, maxSize int, fileMode fs.FileMode) (*GzLog, error) {
	if maxSize < 0 {
		return nil, ErrInvalidSize
	}
	err := os.Mkdir(logdir, fileMode)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	filename, err := getSuitableFile(logdir, basename, maxSize, fileMode)
	if err != nil {
		return nil, err
	}
	gl := &GzLog{
		basename:     basename,
		filename:     filename,
		dir:          logdir,
		maxSize:      int64(maxSize),
		externalFile: false,
	}

	gl.file, err = os.OpenFile(filename, fileFlags, fileMode)
	if err != nil {
		return gl, err
	}
	gl.stat, err = gl.file.Stat()
	return gl, err
}

// UseFile is similar to OpenFile, but it can use an already opened *os.File instead of
// loading it in this package, including os.Stdout and os.Stderr. If Stdout or Stderr
// are used as files, the log won't be rotated or compressed
func UseFile(file *os.File, basename string, maxSize int) (*GzLog, error) {
	if file == nil {
		return nil, os.ErrClosed
	}
	if maxSize < 0 {
		return nil, ErrInvalidSize
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	var logDir string
	var filename string
	if file == os.Stdout || file == os.Stderr {
		maxSize = 0
		basename = ""
	} else {
		logDir = path.Dir(file.Name())
		filename, err = getSuitableFile(logDir, basename, maxSize, fi.Mode())
		if err != nil {
			return nil, err
		}
	}
	gl := &GzLog{
		basename:     basename,
		filename:     filename,
		dir:          logDir,
		file:         file,
		maxSize:      int64(maxSize),
		stat:         fi,
		externalFile: true,
	}
	return gl, nil
}
