package pusher

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/goflags"
	folderutil "github.com/projectdiscovery/utils/folder"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var (
	defaultConfigLocation = filepath.Join(folderutil.HomeDirOrDefault("."), ".config/elasticPusher/settings.yaml")
)

type Options struct {
	ConfigFile string
	InputFile  string
	Project    string
	Index      string
	Host       string
	Type       string
	Silent     bool
	Version    bool
	NoColor    bool
	Verbose    bool
}

// ParseOptions parses the command line flags provided by a user
func ParseOptions() *Options {
	options := &Options{}
	var err error
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(fmt.Sprintf("elasticPusher %s - Push data to the ELK stack from command line", VERSION))

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.InputFile, "file", "f", "", "input file containing data to be stored"),
		flagSet.StringVarP(&options.Index, "index", "i", "", "index under which the data should be stored"),
		flagSet.StringVarP(&options.Type, "type", "t", "json", "input is in JSONL(ines) format"),
		flagSet.StringVarP(&options.Project, "project", "p", "", "project name for metadata addition"),
		flagSet.StringVarP(&options.Host, "host", "h", "", "host name for metadata addition"),
	)

	flagSet.CreateGroup("config", "Config",
		flagSet.StringVarP(&options.ConfigFile, "config", "c", defaultConfigLocation, "flag configuration file"),
	)

	flagSet.CreateGroup("debug", "Debug",
		flagSet.BoolVar(&options.Silent, "silent", false, "show only results in output"),
		flagSet.BoolVar(&options.Version, "version", false, "show version of the project"),
		flagSet.BoolVar(&options.Verbose, "v", false, "show verbose output"),
		flagSet.BoolVarP(&options.NoColor, "no-color", "nc", false, "disable colors in output"),
	)

	if err := flagSet.Parse(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	options.configureOutput()

	if options.Version {
		fmt.Printf("Current Version: %s\n", VERSION)
		os.Exit(0)
	}

	// Validate the options passed by the user and if any
	// invalid options have been used, exit.
	err = options.validateOptions()
	if err != nil {
		log.Fatalf("Program exiting: %v\n", err)
	}

	return options
}

func (options *Options) configureOutput() {
	if options.Verbose {
		log.SetLevel(logrus.TraceLevel)
	}

	if options.NoColor {
		log.SetFormatter(&logrus.TextFormatter{
			PadLevelText:     true,
			ForceColors:      false,
			DisableTimestamp: true,
		})
	}

	if options.Silent {
		log.SetLevel(logrus.PanicLevel)
	}
}

// validateOptions validates the configuration options passed
func (options *Options) validateOptions() error {

	// Both verbose and silent flags were used
	if options.Verbose && options.Silent {
		return errors.New("both verbose and silent mode specified")
	}

	return nil
}
