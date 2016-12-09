package scheduler

import (
	"flag"
	"os"
)

func main() {
	fs := flag.NewFlagSet("scheduler", flag.ExitOnError)

	config := new(configuration)
	config.initialize(fs)

	fs.Parse(os.Args[1:])

	scheduler := NewScheduler(config)
}
