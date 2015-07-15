package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/etrepat/postman/version"
	"github.com/etrepat/postman/watch"
	flag "github.com/ogier/pflag"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	wFlags, err := parseAndCheckFlags()
	if err != nil {
		printMessageAndExit(err.Error())
	}

	watch := watch.New(wFlags)
	go watch.Start()

	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then Close IMAP connection
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	close(ch)
	watch.Stop()

	fmt.Println("Have a nice day.")
}

func parseAndCheckFlags() (*watch.Flags, error) {
	wflags := watch.NewFlags()
	printVersion := false

	flag.Usage = printUsage

	flag.StringVarP(&wflags.Host, "host", "h", "imap.gmail.com", "IMAP server hostname or ip address.")
	flag.UintVarP(&wflags.Port, "port", "p", 993, "IMAP server port number. Defaults to 143 or 993 for ssl.")
	flag.BoolVar(&wflags.Ssl, "ssl", true, "Enforce a SSL connection. Defaults to true if port is 993.")
	flag.StringVarP(&wflags.Username, "user", "U", "", "IMAP login username.")
	flag.StringVarP(&wflags.Password, "password", "P", "", "IMAP login password.")
	flag.StringVarP(&wflags.Mailbox, "mailbox", "b", "INBOX", "Mailbox to monitor/idle on. Defaults to: \"INBOX\".")
	flag.StringVarP(&wflags.Mode, "mode", "m", "", fmt.Sprintf("Mode of delivery. Valid delivery modes are: %s.", strings.Join(watch.ValidDeliveryModes(), ", ")))
	flag.StringVar(&wflags.PostbackUrl, "postback-url", "", "(postback only) URL to post incoming raw email message data.")
	flag.BoolVar(&wflags.PostEncoded, "encode", false, "(postback only) POST messages as form data (x-form-urlencoded). See `parname` flag.")
	flag.StringVar(&wflags.PostParamName, "parname", "message", "(postback only) POST parameter name. Defaults to: \"message\".")
	flag.BoolVarP(&printVersion, "version", "v", false, "Outputs the version information.")
	flag.StringVarP(&wflags.RoomAuth, "auth", "a", "", "(hipchat only) room authentication token.")
	flag.StringVarP(&wflags.RoomName, "name", "n", "", "(hipchat only) room name.")
	flag.StringVarP(&wflags.RoomColor, "color", "c", "green", "(hipchat only) room color. Defaults to \"green\".")

	flag.Parse()

	if flag.NFlag() == 0 {
		return wflags, newFlagsError("No options provided.")
	}

	if printVersion {
		return wflags, newError("%s\n", version.Version())
	}

	if wflags.Host == "" {
		return wflags, newFlagsError("IMAP server host is mandatory.")
	}

	if wflags.Mode == "" {
		return wflags, newFlagsError("Delivery mode must be specified. Should be one of: %s.", strings.Join(watch.ValidDeliveryModes(), ", "))
	} else {
		if !watch.DeliveryModeValid(wflags.Mode) {
			return wflags, newFlagsError("Unknown delivery mode: \"%s\". Must be one of: %s.", wflags.Mode, strings.Join(watch.ValidDeliveryModes(), ", "))
		} else if wflags.Mode == "postback" && wflags.PostbackUrl == "" {
			return wflags, newFlagsError("On postback mode, delivery url must be specified.")
		} else if wflags.Mode == "hipchat" && (wflags.RoomAuth == "" || wflags.RoomName == "") {
			return wflags, newFlagsError("On hipchat mode, room auth and room name must be specified.")
		}
	}

	if wflags.Port == 143 && wflags.Ssl == true {
		wflags.Port = 993
	} else if wflags.Port == 993 && wflags.Ssl == false {
		wflags.Ssl = true
	}

	return wflags, nil
}

func usageMessage() string {
	var usageStr string

	usageStr = "IMAP idling daemon which delivers incoming email to a webhook.\n\n"

	usageStr += "Usage:\n"
	usageStr += fmt.Sprintf("  %s [OPTIONS]\n", version.App())

	usageStr += "\nOptions are:\n"

	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Shorthand) > 0 {
			usageStr += fmt.Sprintf("  -%s, --%s\r\t\t\t%s\n", f.Shorthand, f.Name, f.Usage)
		} else {
			usageStr += fmt.Sprintf("      --%s\r\t\t\t%s\n", f.Name, f.Usage)
		}
	})

	usageStr += "\n"
	usageStr += fmt.Sprintf("       --help\r\t\t\tThis help screen\n")
	usageStr += "\n"

	return usageStr
}

func printMessage(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func printMessageAndExit(format string, args ...interface{}) {
	printMessage(format, args...)
	os.Exit(1)
}

func printUsage() {
	printMessage(usageMessage())
}

func newError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func newFlagsError(format string, args ...interface{}) error {
	errorMessage := fmt.Sprintf(format, args...)

	return newError("%s: %s\nTry \"%s --help\" for more information.\n", version.App(), errorMessage, version.App())
}
