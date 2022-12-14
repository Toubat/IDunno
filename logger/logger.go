package logger

import (
	"fmt"
	"mp4/api"

	"os"
	"path/filepath"
	"time"
)

type LogStats struct {
	StartTime    time.Time
	BytesWritten int
	BytesRead    int
	NumFailures  int
	NumPings     int
}

func NewLogStats() *LogStats {
	return &LogStats{
		StartTime:    time.Now(),
		BytesWritten: 0,
		BytesRead:    0,
		NumFailures:  0,
		NumPings:     0,
	}
}

// log types
const (
	READ     = "READ"
	WRITE    = "WRITE"
	JOIN     = "JOIN"
	PING     = "PING"
	LEAVE    = "LEAVE"
	UPDATE   = "UPDATE"
	FAILURE  = "FAILURE"
	DELETE   = "DELETE"
	INFO     = "INFO"
	ERROR    = "ERROR"
	PUT      = "PUT"
	GET      = "GET"
	REMOVE   = "REMOVE"
	LOOKUP   = "LOOKUP"
	TRANSFER = "TRANSFER"
	NEW_JOB  = "NEW_JOB"
	SCHEDULE = "SCHEDULE"
	QUERY    = "QUERY"
)

// logger global states
var STAT *LogStats = nil
var LOG_FILE string = ""
var IGNORED_TYPES = make([]string, 0)

func Init(name string, ignoredTypes []string) {
	LOG_FILE = name + ".log"

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	LOG_FILE = filepath.Join(exPath, LOG_FILE)

	// delete old machine.1.log file
	os.Remove(LOG_FILE)
	// create new machine.1.log file
	f, err := os.Create(LOG_FILE)
	if err != nil {
		panic(err)
	}

	f.Close()
	STAT = NewLogStats()
	IGNORED_TYPES = ignoredTypes
}

func Error(message string) {
	appendLog(ERROR, message)
}

func Info(message string) {
	appendLog(INFO, message)
}

func StatServerError(message string) {
	appendLog(ERROR, "IDunnoStatServer - "+message)
}

func StatServerInfo(message string) {
	appendLog(INFO, "IDunnoStatServer - "+message)
}

func Read(bytes int) {
	STAT.BytesRead += bytes
	appendLog(READ, fmt.Sprintf("Read %d bytes from UDP", bytes))
}

func Write(bytes int) {
	STAT.BytesWritten += bytes
	appendLog(WRITE, fmt.Sprintf("Write %d bytes to UDP", bytes))
}

func Join(process *api.Process) {
	appendLog(JOIN, formatServiceMessage(process, "joined"))
}

func Ping(process *api.Process) {
	STAT.NumPings++
	appendLog(PING, formatServiceMessage(process, "pinged"))
}

func Leave(process *api.Process) {
	appendLog(LEAVE, formatServiceMessage(process, "leaved"))
}

func Update(process *api.Process) {
	appendLog(UPDATE, formatServiceMessage(process, "updated"))
}

func Failure(process *api.Process) {
	STAT.NumFailures++
	appendLog(FAILURE, formatServiceMessage(process, "failed"))
}

func Delete(process *api.Process) {
	appendLog(DELETE, formatServiceMessage(process, "deleted"))
}

func Get(filename string, version int) {
	appendLog(GET, fmt.Sprintf("Reading file %v with version %d...", filename, version))
}

func Put(filename string) {
	appendLog(PUT, fmt.Sprintf("Writing file %v...", filename))
}

func Remove(filename string) {
	appendLog(REMOVE, fmt.Sprintf("Removing file %v...", filename))
}

func Lookup(filename string) {
	appendLog(LOOKUP, fmt.Sprintf("Lookup file %v...", filename))
}

func Transfer(num_file int) {
	appendLog(TRANSFER, fmt.Sprintf("Transfer %d files...", num_file))
}

func BulkRead(filename string, numBersion int) {
	appendLog(LOOKUP, fmt.Sprintf("Bulk read file %v for version %d ...", filename, numBersion))
}

func Schedule(jobId string, workerAddr string) {
	appendLog(SCHEDULE, fmt.Sprintf("Job %v is scheduled to worker %v", jobId, workerAddr))
}

func NewJob(job *api.Job) {
	appendLog(NEW_JOB, fmt.Sprintf("Created job %v - total queries: %v, batch input size: %v", job.Id, job.TotalQueries, job.BatchSize))
}

func Query(message string) {
	appendLog(QUERY, message)
}

func Stats() {
	elapsedTime := time.Since(STAT.StartTime)
	fmt.Println("Elapsed time: ", elapsedTime)
	fmt.Println("Bytes written: ", STAT.BytesWritten)
	fmt.Println("Bytes read: ", STAT.BytesRead)
	fmt.Println("Number of failures: ", STAT.NumFailures)
	fmt.Println("Number of pings: ", STAT.NumPings)
	fmt.Println("Bps write: ", float64(STAT.BytesWritten)/elapsedTime.Seconds())
	fmt.Println("Bps read: ", float64(STAT.BytesRead)/elapsedTime.Seconds())
}

func appendLog(logType string, message string) {
	// ignore log types
	for _, t := range IGNORED_TYPES {
		if t == logType {
			return
		}
	}

	f, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	time := time.Now().Format("2006-01-02 15:04:05")
	_, err = f.WriteString("[" + time + "]" + "[" + logType + "] " + message + "\n")
	if err != nil {
		return
	}
}

func formatServiceMessage(process *api.Process, action string) string {
	addr := process.Address()
	status := process.Status.String()
	lastUpdateTime := process.LastUpdateTime.AsTime().Format("2006-01-02 15:04:05")
	joinTime := process.JoinTime.AsTime().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("Process %s %s. Status: %s, LastUpdateTime: %s, JoinTime: %s", addr, action, status, lastUpdateTime, joinTime)
}
