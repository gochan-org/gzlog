# gzlog
A logging package that compresses log files when opened or written to if the size exceeds the set max size

## Usage
### Creating a new log file
```Go
basename := "logs/logfile"
maxSize := 5 * 1000 * 1000 // 5 MB
fileMode := 0644
// do not include the extension in basename, gzlog will create basename.log or
// basename.#.log as necessary if it does not exist, or append new lines to it
// if its size < maxSize
gcl, err := gzlog.OpenFile(basename, maxSize, fileMode)
if err != nil {
	panic(err)
}
defer gzlog.Close()
str, err := gcl.Println("Receiving new HTTP request from", "8.8.8.8")
if err != nil {
	panic(err)
}
fmt.Println(str) // prints "Receiving new HTTP request from 8.8.8.8" (without quotes)
logText, err := gcl.ReadAllString() // contains all the contents of the current log file, including prefixes
if err != nil {
	panic(err)
}
doSomethingWith(logText)
```

### Importing an already existing log file
```Go
gcl, err := ImportFile(os.Stdout, "", 0)
if err != nil {
	panic(err)
}
```