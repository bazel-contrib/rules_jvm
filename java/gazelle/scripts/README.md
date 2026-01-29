# Profiling Scripts

This directory contains scripts for profiling and monitoring gazelle performance.

- Use `quick-profile.sh` for quick checks during development
- Use `monitor-gazelle.sh` for detailed performance analysis and benchmarking
- Multiple iterations in `monitor-gazelle.sh` help identify performance regressions or improvements

## monitor-gazelle.sh

Comprehensive monitoring script that tracks CPU and memory usage of gazelle and Java parser processes over multiple iterations. This script is ideal for detailed performance analysis and generating reports.

**Usage:**

```bash
./monitor-gazelle.sh [--iterations=N|--iterations N] [--] [gazelle-command-args...]
```

**Options:**
- `--iterations=N` or `--iterations N`: Number of times to run the command (default: 1)
- `--`: Separator to indicate start of gazelle command arguments
- `gazelle-command-args`: Command to run (default: `bazel run //:gazelle`)

**Examples:**

```bash
# Run default command once
./monitor-gazelle.sh

# Run 5 iterations
./monitor-gazelle.sh --iterations=5

# Run custom command
./monitor-gazelle.sh -- bazel run //java/gazelle:gazelle -- update

# Run custom command with multiple iterations
./monitor-gazelle.sh --iterations 3 -- bazel run //java/gazelle:gazelle
```

**Output:**

The script creates a timestamped output directory (e.g., `gazelle-profile-20240101-120000`) containing:

- `summary.txt` - Summary report with execution times, exit codes, and resource usage statistics
- `resource-usage.csv` - Time-series data with columns: Iteration, Timestamp, PID, Process, CPU%, Memory%, RSS(KB), VSZ(KB), CPU_Time
- `gazelle-output-{N}.log` - Gazelle stdout/stderr for each iteration
- `rss-chart.png` - RSS memory usage chart (if gnuplot is installed)
- `cpu-memory-chart.png` - CPU and memory percentage chart (if gnuplot is installed)

**Features:**

- Monitors both gazelle and Java parser processes
- Tracks CPU%, memory%, RSS, VSZ, and CPU time
- Supports multiple iterations for performance comparison
- Generates visual charts using gnuplot (optional, install with `brew install gnuplot`)
- Calculates min/max/average statistics across iterations
- Real-time console output showing current resource usage

**Viewing Results:**

```bash
# View CSV data in a table
cat gazelle-profile-*/resource-usage.csv | column -t -s,

# View summary report
cat gazelle-profile-*/summary.txt

# View charts (macOS)
open gazelle-profile-*/rss-chart.png
open gazelle-profile-*/cpu-memory-chart.png
```

## quick-profile.sh

Simple real-time monitoring script that displays CPU and memory usage for gazelle and Java parser processes. Ideal for quick checks during development.

**Usage:**

```bash
./quick-profile.sh [gazelle-command...]
```

**Examples:**

```bash
# Run default command
./quick-profile.sh

# Run custom command
./quick-profile.sh bazel run //java/gazelle:gazelle -- update
```

**Features:**

- Real-time display updated every 2 seconds
- Shows process tree
- Displays PID, CPU%, memory%, RSS, VSZ, CPU time, and command
- Press Ctrl+C to stop monitoring (gazelle will continue running)

**Output:**

The script displays:
- Process information (PIDs for gazelle and Java parser)
- Process tree
- Real-time resource usage that updates every 2 seconds

## Requirements

- Bash shell
- `ps` command (standard on Unix-like systems)
- Optional: `gnuplot` for chart generation in `monitor-gazelle.sh` (install with `brew install gnuplot` on macOS)
