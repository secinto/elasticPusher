package pusher

import (
	"github.com/sirupsen/logrus"
)

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *Hook) Fire(entry *logrus.Entry) error {
	bytes, err := hook.Formatter.Format(entry)

	if err != nil {
		return err
	}
	err = hook.Pusher.PushLog(string(bytes), entry.Level.String())
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *Hook) Levels() []logrus.Level { return logrus.AllLevels[:hook.LogLevel+1] }
