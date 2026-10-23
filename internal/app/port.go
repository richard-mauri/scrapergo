package app

import (
 "flag"
 "fmt"
 "io"
 "strconv"
)

func ResolvePort(args []string, envPort string, output io.Writer) (string, error) {
	fs := flag.NewFlagSet("scrapergo", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.Usage = func() {
		fmt.Fprintln(output, "Usage: scrapergo [--port PORT] [--help]")
		fmt.Fprintln(output, "\nOptions:")
		fs.PrintDefaults()
		fmt.Fprintln(output, "\nExamples:")
		fmt.Fprintln(output, "  scrapergo --port 8081")
		fmt.Fprintln(output, "  ./dist/scrapergo-darwin-arm64 --port 8081")
		fmt.Fprintln(output, "  PORT=8081 scrapergo")
		fmt.Fprintln(output, "\n--port overrides the PORT environment variable. Default: 8080.")
	}
	cliPort := fs.String("port", "", "HTTP listening port (1-65535)")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() != 0 {
		return "", fmt.Errorf("unexpected argument %q (use --help)", fs.Arg(0))
	}
	port := "8080"
	if envPort != "" {
		port = envPort
	}
	if *cliPort != "" {
		port = *cliPort
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid port %q: specify a number from 1 to 65535 (example: --port 8081)", port)
	}
	return strconv.Itoa(n), nil
}

