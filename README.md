# Multirun - Run Multiple Command Instances with Color-Coded Output

[![Test](https://github.com/ericfialkowski/multirun/actions/workflows/test.yml/badge.svg)](https://github.com/ericfialkowski/multirun/actions/workflows/test.yml)

A Go CLI tool that runs multiple instances of a command simultaneously and color-codes the output from each instance.

## Features

- Run multiple instances of any command
- Color-coded output (different color per instance)
- Customizable instance prefixes
- Proper signal handling (Ctrl+C)
- Streams both stdout and stderr
- No color mode for piping output

## Installation

```bash
go build -o multirun multirun.go
```

Or install directly:

```bash
go install github.com/ericfialkowski/multirun@latest
```

## Usage

```bash
multirun [options] <command> [args...]
```

### Options

- `-n, -count <number>`: Number of instances to run (default: 2)
- `-no-color`: Disable colored output
- `-prefix <format>`: Custom prefix format (use `{id}` for instance number)

### Examples

**Run 3 instances of ping:**
```bash
multirun -n 3 ping google.com
```

**Run 5 workers with custom prefix:**
```bash
multirun -n 5 -prefix '[Worker {id}]' ./worker.sh
```

**Run without colors (for piping):**
```bash
multirun -n 3 -no-color echo "Hello" > output.txt
```

**Run a script with arguments:**
```bash
multirun -n 4 python script.py --verbose --iterations 100
```

## How It Works

1. Spawns the specified number of command instances
2. Assigns a unique color to each instance (cycles through 36 colors)
3. Prefixes each output line with `[instance_id]` in that instance's color
4. Handles Ctrl+C gracefully by terminating all spawned processes
5. Streams output in real-time (line-buffered)

## Color Palette

The tool cycles through 36 colors in three style groups:
- **Regular**: Red, Green, Yellow, Blue, Magenta, Cyan + bright variants
- **Bold**: Bold versions of all 12 colors above
- **Underline**: Underline versions of all 12 colors above

## Example Output

```
[1] PING google.com (142.250.80.46): 56 data bytes
[2] PING google.com (142.250.80.46): 56 data bytes
[3] PING google.com (142.250.80.46): 56 data bytes
[1] 64 bytes from 142.250.80.46: icmp_seq=0 ttl=117 time=10.123 ms
[2] 64 bytes from 142.250.80.46: icmp_seq=0 ttl=117 time=10.456 ms
[3] 64 bytes from 142.250.80.46: icmp_seq=0 ttl=117 time=10.789 ms
```

(Each instance will be in a different color)

## License

MIT
