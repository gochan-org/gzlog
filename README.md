# gzlog
A logging package that compresses log files when opened or written to if the size exceeds the set max size

## Usage
```Go
dir := "./logs" // creates the directory if it doesn't already exist
logBasename := "logfile"
maxSize := 5 * 1000 * 1000 // 5 MB
fileMode := 0644
// do not include the extension in logBasename, gzlog will create ./logs/logfile.log
// or ./logs/logfile.#.log as necessary if it does not exist, or append new lines
// to it if its size < maxSize
gcl, err := gzlog.OpenFile(dir, logBasename, maxSize, fileMode)
if err != nil {
	panic(err)
}
defer gzlog.Close()
str, err := gcl.Println("Receiving new HTTP request")
if err != nil {
	panic(err)
}
fmt.Println(str) // prints "Receiving new HTTP request" (without quotes)
logText, err := gcl.ReadAllString() // contains all the contents of the current log file
if err != nil {
	panic(err)
}
doSomethingWith(logText)
```