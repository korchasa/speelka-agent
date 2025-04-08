// Package utils contains various utilities for the entire project
// CustomLogFormatter - custom formatter for logs that places the file field before msg
package utils

import (
    "bytes"
    "fmt"
    "path/filepath"
    "sort"
    "strings"

    "github.com/sirupsen/logrus"
)

// CustomLogFormatter - custom formatter for logrus
// Responsibility: Formatting logs with fields in a specified order
// Features: Arranges fields in the order: time, level, file, msg, ...
type CustomLogFormatter struct {
    TimestampFormat string // time format (default "15:04:05")
    FullTimestamp   bool   // whether to display the full date
}

// Format formats the log entry according to specified rules
func (f *CustomLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    var b *bytes.Buffer
    if entry.Buffer != nil {
        b = entry.Buffer
    } else {
        b = &bytes.Buffer{}
    }

    timestampFormat := f.TimestampFormat
    if timestampFormat == "" {
        timestampFormat = "15:04:05"
    }

    // Form the basic log structure
    b.WriteString(entry.Time.Format(timestampFormat))
    b.WriteString(" " + strings.ToUpper(entry.Level.String()))

    if entry.HasCaller() {
        b.WriteString(fmt.Sprintf(" %s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line))
    }

    b.WriteString(" " + entry.Message)

    // Add remaining fields in alphabetical order
    keys := make([]string, 0, len(entry.Data))
    for k := range entry.Data {
        if k != "file" && k != "msg" { // Skip file and msg, they are already added
            keys = append(keys, k)
        }
    }
    sort.Strings(keys)

    for _, key := range keys {
        b.WriteString(" ")
        f.appendKeyValue(b, key, entry.Data[key])
    }

    b.WriteByte('\n')
    return b.Bytes(), nil
}

// appendKeyValue formats a key-value pair for the log
func (f *CustomLogFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
    b.WriteString(key)
    b.WriteByte('=')

    switch value := value.(type) {
    case string:
        if strings.ContainsAny(value, " \t\r\n\"=:") {
            fmt.Fprintf(b, "\"%s\"", value)
        } else {
            b.WriteString(value)
        }
    case error:
        fmt.Fprintf(b, "\"%s\"", value.Error())
    default:
        fmt.Fprintf(b, "%v", value)
    }
}

// NewCustomLogFormatter creates a new instance of the custom formatter
func NewCustomLogFormatter() *CustomLogFormatter {
    return &CustomLogFormatter{
        TimestampFormat: "15:04:05",
        FullTimestamp:   false,
    }
}
