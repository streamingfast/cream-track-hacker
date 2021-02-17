package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	eth "github.com/dfuse-io/eth-go"
)

func parseAddressesFlag(in string) (out []string, celFiler string, err error) {
	parts := strings.Split(in, ",")
	if len(parts) == 0 {
		return nil, "", errors.New("expecting at least on address, found none")
	}

	out = make([]string, len(parts))
	celParts := make([]string, len(parts))

	for i, part := range parts {
		address, err := eth.NewAddress(strings.TrimSpace(part))
		if err != nil {
			return nil, "", fmt.Errorf("invalid address %q: %w", part, err)
		}

		out[i] = address.Pretty()
		celParts[i] = "'" + strings.ToLower(out[i]) + "'"
	}

	return out, "[" + strings.Join(celParts, ",") + "]", nil
}

func setupFlag() {
	flag.CommandLine.Usage = func() {
		fmt.Print(usage())
	}
	flag.Parse()
}

func flagUsage() string {
	buf := bytes.NewBuffer(nil)
	oldOutput := flag.CommandLine.Output()
	defer func() { flag.CommandLine.SetOutput(oldOutput) }()

	flag.CommandLine.SetOutput(buf)
	flag.CommandLine.PrintDefaults()

	return buf.String()
}

func errorUsage(message string, args ...interface{}) string {
	return fmt.Sprintf(message+"\n\n"+usage(), args...)
}

func ensure(condition bool, message string, args ...interface{}) {
	if !condition {
		noError(fmt.Errorf(message, args...), "invalid arguments")
	}
}

func noError(err error, message string, args ...interface{}) {
	if err != nil {
		quit(message+": "+err.Error(), args...)
	}
}

func quit(message string, args ...interface{}) {
	printf(message+"\n", args...)
	os.Exit(1)
}

func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func println(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
