package testutils

import (
	"bufio"
	"io"
	"regexp"
)

// Logs is a set of logs
type Logs struct {
	logs []string
}

// NewLogs creates logs from a reader
func NewLogs(r io.Reader) (*Logs, error) {
	logs := make([]string, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	err := scanner.Err()
	if err != nil {
		return nil, err
	}
	return &Logs{logs}, nil
}

// ContainsMatch determines if some log matches the provided regexp
func (l *Logs) ContainsMatch(re *regexp.Regexp) bool {
	for _, log := range l.logs {
		if re.MatchString(log) {
			return true
		}
	}
	return false
}
