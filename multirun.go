package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// ANSI color codes
var colors = []string{
	"\033[31m",   // Red
	"\033[32m",   // Green
	"\033[33m",   // Yellow
	"\033[34m",   // Blue
	"\033[35m",   // Magenta
	"\033[36m",   // Cyan
	"\033[91m",   // Bright Red
	"\033[92m",   // Bright Green
	"\033[93m",   // Bright Yellow
	"\033[94m",   // Bright Blue
	"\033[95m",   // Bright Magenta
	"\033[96m",   // Bright Cyan
	"\033[1;31m", // Bold Red
	"\033[1;32m", // Bold Green
	"\033[1;33m", // Bold Yellow
	"\033[1;34m", // Bold Blue
	"\033[1;35m", // Bold Magenta
	"\033[1;36m", // Bold Cyan
	"\033[1;91m", // Bold Bright Red
	"\033[1;92m", // Bold Bright Green
	"\033[1;93m", // Bold Bright Yellow
	"\033[1;94m", // Bold Bright Blue
	"\033[1;95m", // Bold Bright Magenta
	"\033[1;96m", // Bold Bright Cyan
	"\033[4;31m", // Underline Red
	"\033[4;32m", // Underline Green
	"\033[4;33m", // Underline Yellow
	"\033[4;34m", // Underline Blue
	"\033[4;35m", // Underline Magenta
	"\033[4;36m", // Underline Cyan
	"\033[4;91m", // Underline Bright Red
	"\033[4;92m", // Underline Bright Green
	"\033[4;93m", // Underline Bright Yellow
	"\033[4;94m", // Underline Bright Blue
	"\033[4;95m", // Underline Bright Magenta
	"\033[4;96m", // Underline Bright Cyan
}

const resetColor = "\033[0m"

type instance struct {
	id     int
	cmd    *exec.Cmd
	color  string
	prefix string
}

func main() {
	var (
		count   int
		noColor bool
		prefix  string
	)

	flag.IntVar(&count, "n", 2, "Number of instances to run")
	flag.IntVar(&count, "count", 2, "Number of instances to run")
	flag.BoolVar(&noColor, "no-color", false, "Disable colored output")
	flag.StringVar(&prefix, "prefix", "", "Custom prefix format (use {id} for instance number)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: multirun [options] <command> [args...]")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExample:")
		fmt.Fprintln(os.Stderr, "  multirun -n 3 ping google.com")
		fmt.Fprintln(os.Stderr, "  multirun -n 5 -prefix '[Worker {id}]' ./worker.sh")
		os.Exit(1)
	}

	if count < 1 {
		fmt.Fprintln(os.Stderr, "Error: count must be at least 1")
		os.Exit(1)
	}

	cmdName := flag.Arg(0)
	cmdArgs := flag.Args()[1:]

	// Create instances
	instances := make([]*instance, count)
	var wg sync.WaitGroup

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal, stopping all instances...")
		for _, inst := range instances {
			if inst != nil && inst.cmd != nil && inst.cmd.Process != nil {
				inst.cmd.Process.Signal(syscall.SIGTERM)
			}
		}
	}()

	// Determine ID format based on instance count
	idFmt := "%d"
	if count >= 100 {
		idFmt = "%03d"
	} else if count >= 10 {
		idFmt = "%02d"
	}

	// Start all instances
	for i := 0; i < count; i++ {
		inst := &instance{
			id:  i + 1,
			cmd: exec.Command(cmdName, cmdArgs...),
		}

		// Assign color
		if !noColor {
			inst.color = colors[i%len(colors)]
		}

		// Set prefix
		idStr := fmt.Sprintf(idFmt, inst.id)
		if prefix != "" {
			inst.prefix = strings.ReplaceAll(prefix, "{id}", idStr)
		} else {
			inst.prefix = fmt.Sprintf("[%s]", idStr)
		}

		instances[i] = inst

		// Get stdout and stderr pipes
		stdout, err := inst.cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating stdout pipe for instance %d: %v\n", inst.id, err)
			continue
		}

		stderr, err := inst.cmd.StderrPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating stderr pipe for instance %d: %v\n", inst.id, err)
			continue
		}

		// Start the command
		if err := inst.cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting instance %d: %v\n", inst.id, err)
			continue
		}

		wg.Add(2)

		// Stream stdout
		go streamOutput(inst, stdout, os.Stdout, &wg)

		// Stream stderr
		go streamOutput(inst, stderr, os.Stderr, &wg)

		// Wait for command completion
		go func(inst *instance) {
			err := inst.cmd.Wait()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s%s%s Command exited with error: %v\n",
					inst.color, inst.prefix, resetColor, err)
			}
		}(inst)
	}

	// Wait for all output to be processed
	wg.Wait()
}

func streamOutput(inst *instance, reader io.Reader, writer io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(writer, "%s%s %s%s\n", inst.color, inst.prefix, line, resetColor)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading output from instance %d: %v\n", inst.id, err)
	}
}
