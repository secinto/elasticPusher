package pusher

import (
	"github.com/sirupsen/logrus"
)

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *Hook) Fire(entry *logrus.Entry) error {
	line, err := entry.Bytes()
	if err != nil {
		return err
	}
	err = hook.Pusher.PushString(string(line))
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *Hook) Levels() []logrus.Level {
	return hook.LogLevels
}
