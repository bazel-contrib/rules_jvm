#!/bin/bash
# Monitor gazelle and Java parser processes for CPU and memory usage
# Usage: ./monitor-gazelle.sh [--iterations=N|--iterations N] [--] [gazelle-command-args...]
#   --iterations=N or --iterations N: Number of times to run the command (default: 1)
#   --: Separator to indicate start of gazelle command arguments
#   gazelle-command-args: Command to run (default: bazel run //:gazelle)

set -e

# Parse arguments for iterations and command
ITERATIONS=1
GAZELLE_CMD_ARGS=()
IN_CMD_ARGS=0
SKIP_NEXT=0

for i in $(seq 1 $#); do
    arg="${!i}"

    if [ $SKIP_NEXT -eq 1 ]; then
        SKIP_NEXT=0
        continue
    fi

    if [ $IN_CMD_ARGS -eq 1 ]; then
        GAZELLE_CMD_ARGS+=("$arg")
    elif [ "$arg" = "--" ]; then
        IN_CMD_ARGS=1
    elif [[ "$arg" =~ ^--iterations=([0-9]+)$ ]]; then
        ITERATIONS="${BASH_REMATCH[1]}"
    elif [ "$arg" = "--iterations" ]; then
        # Next argument should be the number
        next_idx=$((i + 1))
        if [ $next_idx -le $# ]; then
            next_arg="${!next_idx}"
            if [[ "$next_arg" =~ ^[0-9]+$ ]]; then
                ITERATIONS="$next_arg"
                SKIP_NEXT=1
            fi
        fi
    else
        GAZELLE_CMD_ARGS+=("$arg")
    fi
done

# If no command args provided, use default
if [ ${#GAZELLE_CMD_ARGS[@]} -eq 0 ]; then
    GAZELLE_CMD_ARGS=("bazel" "run" "//:gazelle")
fi

OUTPUT_DIR="gazelle-profile-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUTPUT_DIR"

echo "Starting gazelle profiling..."
echo "Output directory: $OUTPUT_DIR"
echo "Iterations: $ITERATIONS"
echo "Command: ${GAZELLE_CMD_ARGS[*]}"
echo ""

# Function to find Java parser PID
find_java_parser() {
    local parent_pid=$1
    # Look for Java processes that are children of gazelle
    ps -o pid= --ppid "$parent_pid" 2>/dev/null | while read pid; do
        if [ -n "$pid" ]; then
            # Check if it's a Java process
            if ps -p "$pid" -o command= 2>/dev/null | grep -q java; then
                echo "$pid"
                return 0
            fi
        fi
    done
    # Also try to find by name pattern
    ps aux | grep -E "java.*Main|javaparser" | grep -v grep | awk '{print $2}' | head -1
}

# Function to format elapsed time
format_elapsed_time() {
    local seconds=$1
    local minutes=$((seconds / 60))
    local seconds_remainder=$((seconds % 60))
    local hours=$((minutes / 60))
    local minutes_remainder=$((minutes % 60))

    if [ $hours -gt 0 ]; then
        echo "${hours}h ${minutes_remainder}m ${seconds_remainder}s"
    elif [ $minutes -gt 0 ]; then
        echo "${minutes}m ${seconds_remainder}s"
    else
        echo "${seconds}s"
    fi
}

# Initialize CSV file with iteration column
echo "Iteration,Timestamp,PID,Process,CPU%,Memory%,RSS(KB),VSZ(KB),CPU_Time" > "$OUTPUT_DIR/resource-usage.csv"

# Arrays to store per-iteration statistics
declare -a ITERATION_EXIT_CODES
declare -a ITERATION_ELAPSED_SECONDS
declare -a ITERATION_START_TIMES
declare -a ITERATION_END_TIMES

# Overall start time
OVERALL_START_TIME=$(date +%s)
OVERALL_START_TIME_READABLE=$(date)

# Run iterations
for iteration in $(seq 1 $ITERATIONS); do
    echo ""
    echo "=== Iteration $iteration of $ITERATIONS ==="
    echo ""

    ITER_START_TIME=$(date +%s)
    ITER_START_TIME_READABLE=$(date)

    # Start gazelle in background
    echo "Starting gazelle..."
    "${GAZELLE_CMD_ARGS[@]}" > "$OUTPUT_DIR/gazelle-output-${iteration}.log" 2>&1 &
    GAZELLE_PID=$!

    echo "Gazelle PID: $GAZELLE_PID"
    echo "Waiting for Java parser to start..."

    # Wait for Java parser to start (check every 0.5 seconds, up to 10 seconds)
    JAVA_PID=""
    for i in {1..20}; do
        sleep 0.5
        JAVA_PID=$(find_java_parser $GAZELLE_PID)
        if [ -n "$JAVA_PID" ]; then
            echo "Java Parser PID: $JAVA_PID"
            break
        fi
    done

    if [ -z "$JAVA_PID" ]; then
        echo "Warning: Could not find Java parser process (may start later or not be used)"
    fi

    # Start monitoring
    echo ""
    echo "Monitoring processes (press Ctrl+C to stop early)..."
    echo ""

    (
        while ps -p $GAZELLE_PID > /dev/null 2>&1; do
            timestamp=$(date +%Y-%m-%d\ %H:%M:%S)

            # Gazelle stats
            if ps -p $GAZELLE_PID > /dev/null 2>&1; then
                stats=$(ps -o pcpu=,pmem=,rss=,vsz=,time= -p $GAZELLE_PID 2>/dev/null | tr -s ' ' | sed 's/^ *//;s/ *$//' | tr ' ' ',')
                if [ -n "$stats" ]; then
                    echo "$iteration,$timestamp,$GAZELLE_PID,Gazelle,$stats" >> "$OUTPUT_DIR/resource-usage.csv"
                fi
            fi

            # Java parser stats
            if [ -n "$JAVA_PID" ] && ps -p $JAVA_PID > /dev/null 2>&1; then
                stats=$(ps -o pcpu=,pmem=,rss=,vsz=,time= -p $JAVA_PID 2>/dev/null | tr -s ' ' | sed 's/^ *//;s/ *$//' | tr ' ' ',')
                if [ -n "$stats" ]; then
                    echo "$iteration,$timestamp,$JAVA_PID,JavaParser,$stats" >> "$OUTPUT_DIR/resource-usage.csv"
                fi
            else
                # Try to find it again (in case it restarted)
                JAVA_PID=$(find_java_parser $GAZELLE_PID)
            fi

            # Print current stats to console
            printf "\r[Iteration %d] [%s] " "$iteration" "$timestamp"
            if ps -p $GAZELLE_PID > /dev/null 2>&1; then
                gazelle_stats=$(ps -o pcpu=,pmem=,rss= -p $GAZELLE_PID 2>/dev/null | tr -s ' ')
                printf "Gazelle: CPU=%s%% Mem=%s%% RSS=%sKB  " $gazelle_stats
            fi
            if [ -n "$JAVA_PID" ] && ps -p $JAVA_PID > /dev/null 2>&1; then
                java_stats=$(ps -o pcpu=,pmem=,rss= -p $JAVA_PID 2>/dev/null | tr -s ' ')
                printf "JavaParser: CPU=%s%% Mem=%s%% RSS=%sKB" $java_stats
            fi

            sleep 1
        done
        echo ""  # New line after monitoring
    ) &
    MONITOR_PID=$!

    # Wait for gazelle to complete
    # Capture exit code before || true to preserve the actual exit code
    set +e  # Temporarily disable exit on error to capture exit code
    wait $GAZELLE_PID
    GAZELLE_EXIT=$?
    set -e  # Re-enable exit on error

    # Stop monitoring
    kill $MONITOR_PID 2>/dev/null || true
    wait $MONITOR_PID 2>/dev/null || true

    ITER_END_TIME=$(date +%s)
    ITER_END_TIME_READABLE=$(date)
    ITER_ELAPSED=$((ITER_END_TIME - ITER_START_TIME))

    # Store iteration statistics
    ITERATION_EXIT_CODES+=($GAZELLE_EXIT)
    ITERATION_ELAPSED_SECONDS+=($ITER_ELAPSED)
    ITERATION_START_TIMES+=("$ITER_START_TIME_READABLE")
    ITERATION_END_TIMES+=("$ITER_END_TIME_READABLE")

    ITER_ELAPSED_FORMATTED=$(format_elapsed_time $ITER_ELAPSED)
    echo ""
    echo "Iteration $iteration completed: Exit code=$GAZELLE_EXIT, Elapsed=$ITER_ELAPSED_FORMATTED"
done

OVERALL_END_TIME=$(date +%s)
OVERALL_END_TIME_READABLE=$(date)
OVERALL_ELAPSED=$((OVERALL_END_TIME - OVERALL_START_TIME))
OVERALL_ELAPSED_FORMATTED=$(format_elapsed_time $OVERALL_ELAPSED)

# Generate charts with gnuplot if available
generate_charts() {
    local csv_file="$1"
    local output_dir="$2"
    local num_iterations=$3

    if ! command -v gnuplot >/dev/null 2>&1; then
        echo "gnuplot not found - skipping chart generation"
        echo "  Install with: brew install gnuplot"
        return 1
    fi

    if [ ! -f "$csv_file" ] || [ ! -s "$csv_file" ]; then
        echo "No CSV data available for chart generation"
        return 1
    fi

    echo "Generating charts with gnuplot..."

    # Color palette for iterations (up to 10 iterations)
    local colors=('#1f77b4' '#ff7f0e' '#2ca02c' '#d62728' '#9467bd' '#8c564b' '#e377c2' '#7f7f7f' '#bcbd22' '#17becf')

    # CSV format: Iteration,Timestamp,PID,Process,CPU%,Memory%,RSS(KB),VSZ(KB),CPU_Time
    # Check if we have data
    local has_gazelle=0
    local has_javaparser=0

    if tail -n +2 "$csv_file" | awk -F',' '$4=="Gazelle" && $7 != "" && $7 != "0" {found=1; exit} END {exit !found}'; then
        has_gazelle=1
    fi
    if tail -n +2 "$csv_file" | awk -F',' '$4=="JavaParser" && $7 != "" && $7 != "0" {found=1; exit} END {exit !found}'; then
        has_javaparser=1
    fi

    if [ $has_gazelle -eq 0 ] && [ $has_javaparser -eq 0 ]; then
        echo "No valid data available for chart generation"
        return 1
    fi

    # Create gnuplot script for RSS chart
    local rss_script="$output_dir/rss-plot.gp"
    {
        cat <<GPEOF
set terminal pngcairo size 1200,600 enhanced font 'Arial,10'
set output '$output_dir/rss-chart.png'
set title 'Resident Set Size (RSS) Over Time'
set xlabel 'Sample Number'
set ylabel 'RSS (MB)'
set grid
set key top left
set datafile separator ","

plot \\
GPEOF
        local plot_items=()

        for iter in $(seq 1 $num_iterations); do
            local iter_data_gazelle="$output_dir/iter-${iter}-gazelle.dat"
            local iter_data_javaparser="$output_dir/iter-${iter}-javaparser.dat"

            # Extract data for this iteration
            # CSV: Iteration,Timestamp,PID,Process,CPU%,Memory%,RSS(KB),VSZ(KB),CPU_Time
            # Output: sample_num,RSS(KB),CPU%,Memory%
            awk -F',' -v iter=$iter '
                BEGIN { sample=0 }
                NR>1 && $1==iter && $4=="Gazelle" && $7 != "" && $7 != "0" {
                    sample++
                    print sample "," $7 "," $5 "," $6
                }
            ' "$csv_file" > "$iter_data_gazelle"

            awk -F',' -v iter=$iter '
                BEGIN { sample=0 }
                NR>1 && $1==iter && $4=="JavaParser" && $7 != "" && $7 != "0" {
                    sample++
                    print sample "," $7 "," $5 "," $6
                }
            ' "$csv_file" > "$iter_data_javaparser"

            local color_idx=$(( (iter - 1) % ${#colors[@]} ))
            local color="${colors[$color_idx]}"

            if [ -s "$iter_data_gazelle" ] && grep -qE '^[0-9]+,[0-9]+' "$iter_data_gazelle" 2>/dev/null; then
                if [ ${#plot_items[@]} -gt 0 ]; then
                    echo -n ", \\" >> "$rss_script"
                    echo "" >> "$rss_script"
                fi
                echo -n "     '$iter_data_gazelle' using 1:(\$2/1024) with lines title 'Gazelle Iter $iter' lw 2 lc rgb '$color' dashtype 2" >> "$rss_script"
                plot_items+=("gazelle-$iter")
            fi

            if [ -s "$iter_data_javaparser" ] && grep -qE '^[0-9]+,[0-9]+' "$iter_data_javaparser" 2>/dev/null; then
                if [ ${#plot_items[@]} -gt 0 ]; then
                    echo -n ", \\" >> "$rss_script"
                    echo "" >> "$rss_script"
                fi
                echo -n "     '$iter_data_javaparser' using 1:(\$2/1024) with lines title 'JavaParser Iter $iter' lw 2 lc rgb '$color' dashtype 1" >> "$rss_script"
                plot_items+=("javaparser-$iter")
            fi
        done
        echo "" >> "$rss_script"
    } > "$rss_script"

    # Create gnuplot script for CPU/Memory chart
    local cpu_mem_script="$output_dir/cpu-memory-plot.gp"
    {
        cat <<GPEOF
set terminal pngcairo size 1200,600 enhanced font 'Arial,10'
set output '$output_dir/cpu-memory-chart.png'
set title 'CPU and Memory Usage Over Time'
set xlabel 'Sample Number'
set ylabel 'Percentage (%)'
set grid
set key top left
set datafile separator ","

plot \\
GPEOF
        local plot_items=()

        for iter in $(seq 1 $num_iterations); do
            local iter_data_gazelle="$output_dir/iter-${iter}-gazelle.dat"
            local iter_data_javaparser="$output_dir/iter-${iter}-javaparser.dat"
            local color_idx=$(( (iter - 1) % ${#colors[@]} ))
            local color="${colors[$color_idx]}"

            if [ -s "$iter_data_gazelle" ] && grep -qE '^[0-9]+,[0-9]+' "$iter_data_gazelle" 2>/dev/null; then
                if [ ${#plot_items[@]} -gt 0 ]; then
                    echo -n ", \\" >> "$cpu_mem_script"
                    echo "" >> "$cpu_mem_script"
                fi
                echo -n "     '$iter_data_gazelle' using 1:3 with lines title 'Gazelle CPU% Iter $iter' lw 2 lc rgb '$color' dashtype 2" >> "$cpu_mem_script"
                plot_items+=("gazelle-cpu-$iter")

                echo -n ", \\" >> "$cpu_mem_script"
                echo "" >> "$cpu_mem_script"
                echo -n "     '$iter_data_gazelle' using 1:4 with lines title 'Gazelle Memory% Iter $iter' lw 2 lc rgb '$color' dashtype 1" >> "$cpu_mem_script"
                plot_items+=("gazelle-mem-$iter")
            fi

            if [ -s "$iter_data_javaparser" ] && grep -qE '^[0-9]+,[0-9]+' "$iter_data_javaparser" 2>/dev/null; then
                if [ ${#plot_items[@]} -gt 0 ]; then
                    echo -n ", \\" >> "$cpu_mem_script"
                    echo "" >> "$cpu_mem_script"
                fi
                echo -n "     '$iter_data_javaparser' using 1:3 with lines title 'JavaParser CPU% Iter $iter' lw 2 lc rgb '$color' dashtype 4" >> "$cpu_mem_script"
                plot_items+=("javaparser-cpu-$iter")

                echo -n ", \\" >> "$cpu_mem_script"
                echo "" >> "$cpu_mem_script"
                echo -n "     '$iter_data_javaparser' using 1:4 with lines title 'JavaParser Memory% Iter $iter' lw 2 lc rgb '$color' dashtype 3" >> "$cpu_mem_script"
                plot_items+=("javaparser-mem-$iter")
            fi
        done
        echo "" >> "$cpu_mem_script"
    } > "$cpu_mem_script"

    # Generate charts
    gnuplot "$rss_script"
    gnuplot "$cpu_mem_script"

    # Clean up temporary files
    rm -f "$rss_script" "$cpu_mem_script"
    rm -f "$output_dir"/iter-*-*.dat

    echo "Charts generated successfully"
    return 0
}

# Generate charts if CSV exists
if [ -f "$OUTPUT_DIR/resource-usage.csv" ]; then
    generate_charts "$OUTPUT_DIR/resource-usage.csv" "$OUTPUT_DIR" "$ITERATIONS"
fi

# Generate summary report
echo ""
echo "=== Generating Summary Report ==="
echo ""

cat > "$OUTPUT_DIR/summary.txt" <<EOF
Gazelle Profiling Summary
=========================
Date: $(date)
Iterations: $ITERATIONS
Command: ${GAZELLE_CMD_ARGS[*]}
Output Directory: $OUTPUT_DIR

Overall Execution Time:
  Start Time: $OVERALL_START_TIME_READABLE
  End Time:   $OVERALL_END_TIME_READABLE
  Elapsed:    $OVERALL_ELAPSED_FORMATTED ($OVERALL_ELAPSED seconds)

Per-Iteration Details:
EOF

# Add per-iteration details
for i in $(seq 0 $((ITERATIONS - 1))); do
    iter_num=$((i + 1))
    exit_code=${ITERATION_EXIT_CODES[$i]}
    elapsed=${ITERATION_ELAPSED_SECONDS[$i]}
    elapsed_formatted=$(format_elapsed_time $elapsed)
    start_time="${ITERATION_START_TIMES[$i]}"
    end_time="${ITERATION_END_TIMES[$i]}"

    echo "" >> "$OUTPUT_DIR/summary.txt"
    echo "Iteration $iter_num:" >> "$OUTPUT_DIR/summary.txt"
    echo "  Start Time: $start_time" >> "$OUTPUT_DIR/summary.txt"
    echo "  End Time:   $end_time" >> "$OUTPUT_DIR/summary.txt"
    echo "  Elapsed:    $elapsed_formatted ($elapsed seconds)" >> "$OUTPUT_DIR/summary.txt"
    echo "  Exit Code:  $exit_code" >> "$OUTPUT_DIR/summary.txt"
done

# Calculate min/max/average for elapsed time
if [ $ITERATIONS -gt 1 ]; then
    echo "" >> "$OUTPUT_DIR/summary.txt"
    echo "=== Elapsed Time Statistics (across all iterations) ===" >> "$OUTPUT_DIR/summary.txt"

    # Calculate statistics using awk
    elapsed_stats=$(printf '%s\n' "${ITERATION_ELAPSED_SECONDS[@]}" | awk '
        {
            sum += $1
            if (NR == 1 || $1 < min) min = $1
            if (NR == 1 || $1 > max) max = $1
        }
        END {
            avg = sum / NR
            printf "  Min: %d seconds (%s)\n", min, format_time(min)
            printf "  Max: %d seconds (%s)\n", max, format_time(max)
            printf "  Avg: %.2f seconds (%s)\n", avg, format_time(avg)
        }
        function format_time(sec) {
            h = int(sec / 3600)
            m = int((sec % 3600) / 60)
            s = sec % 60
            if (h > 0) return sprintf("%dh %dm %ds", h, m, s)
            if (m > 0) return sprintf("%dm %ds", m, s)
            return sprintf("%ds", s)
        }
    ')
    echo "$elapsed_stats" >> "$OUTPUT_DIR/summary.txt"

    # Calculate exit code statistics
    echo "" >> "$OUTPUT_DIR/summary.txt"
    echo "=== Exit Code Statistics ===" >> "$OUTPUT_DIR/summary.txt"
    exit_code_summary=$(printf '%s\n' "${ITERATION_EXIT_CODES[@]}" | sort -n | uniq -c | awk '{printf "  Exit code %d: %d time(s)\n", $2, $1}')
    echo "$exit_code_summary" >> "$OUTPUT_DIR/summary.txt"
fi

echo "" >> "$OUTPUT_DIR/summary.txt"
echo "=== Resource Usage Statistics ===" >> "$OUTPUT_DIR/summary.txt"
if [ -f "$OUTPUT_DIR/resource-usage.csv" ]; then
    # CSV format: Iteration,Timestamp,PID,Process,CPU%,Memory%,RSS(KB),VSZ(KB),CPU_Time
    echo "" >> "$OUTPUT_DIR/summary.txt"
    echo "Peak CPU Usage (Gazelle):" >> "$OUTPUT_DIR/summary.txt"
    tail -n +2 "$OUTPUT_DIR/resource-usage.csv" | awk -F',' '$4=="Gazelle" && $5 != "" {print "  Iteration " $1 ": " $5 "%"}' | sort -t: -k2 -rn | uniq | head -5 >> "$OUTPUT_DIR/summary.txt"

    echo "" >> "$OUTPUT_DIR/summary.txt"
    echo "Peak Memory Usage (Gazelle):" >> "$OUTPUT_DIR/summary.txt"
    tail -n +2 "$OUTPUT_DIR/resource-usage.csv" | awk -F',' '$4=="Gazelle" && $6 != "" {print "  Iteration " $1 ": " $6 "%"}' | sort -t: -k2 -rn | uniq | head -5 >> "$OUTPUT_DIR/summary.txt"

    echo "" >> "$OUTPUT_DIR/summary.txt"
    echo "Peak RSS (Gazelle):" >> "$OUTPUT_DIR/summary.txt"
    tail -n +2 "$OUTPUT_DIR/resource-usage.csv" | awk -F',' '$4=="Gazelle" && $7 != "" && $7 != "0" {printf "  Iteration %s: %.2f MB\n", $1, $7/1024}' | sort -t: -k2 -rn | uniq | head -5 >> "$OUTPUT_DIR/summary.txt"

    if [ $ITERATIONS -gt 1 ]; then
        echo "" >> "$OUTPUT_DIR/summary.txt"
        echo "=== Aggregated Resource Statistics (across all iterations) ===" >> "$OUTPUT_DIR/summary.txt"

        # Calculate min/max/avg for CPU%, Memory%, RSS
        for metric in "CPU%" "Memory%" "RSS"; do
            case $metric in
                "CPU%")
                    col=5
                    label="CPU%"
                    format="%.2f%%"
                    ;;
                "Memory%")
                    col=6
                    label="Memory%"
                    format="%.2f%%"
                    ;;
                "RSS")
                    col=7
                    label="RSS (MB)"
                    format="%.2f"
                    ;;
            esac

            echo "" >> "$OUTPUT_DIR/summary.txt"
            echo "Gazelle $label:" >> "$OUTPUT_DIR/summary.txt"
            stats=$(tail -n +2 "$OUTPUT_DIR/resource-usage.csv" | awk -F',' -v col=$col -v format="$format" '
                $4=="Gazelle" && $col != "" && $col != "0" {
                    val = $col
                    if (col == 7) val = val / 1024  # Convert KB to MB for RSS
                    sum += val
                    count++
                    if (count == 1 || val < min) min = val
                    if (count == 1 || val > max) max = val
                }
                END {
                    if (count > 0) {
                        avg = sum / count
                        printf "  Min: " format "\n", min
                        printf "  Max: " format "\n", max
                        printf "  Avg: " format "\n", avg
                    }
                }
            ')
            echo "$stats" >> "$OUTPUT_DIR/summary.txt"
        done
    fi
fi

# Display summary
cat "$OUTPUT_DIR/summary.txt"

echo ""
echo "=== Files Generated ==="
echo "  - $OUTPUT_DIR/summary.txt - Summary report"
echo "  - $OUTPUT_DIR/resource-usage.csv - Time-series resource data (all iterations)"
for i in $(seq 1 $ITERATIONS); do
    if [ -f "$OUTPUT_DIR/gazelle-output-${i}.log" ]; then
        echo "  - $OUTPUT_DIR/gazelle-output-${i}.log - Gazelle stdout/stderr (iteration $i)"
    fi
done
if [ -f "$OUTPUT_DIR/rss-chart.png" ]; then
    echo "  - $OUTPUT_DIR/rss-chart.png - RSS memory usage chart (all iterations)"
fi
if [ -f "$OUTPUT_DIR/cpu-memory-chart.png" ]; then
    echo "  - $OUTPUT_DIR/cpu-memory-chart.png - CPU and memory percentage chart (all iterations)"
fi
echo ""
echo "To view resource usage over time:"
echo "  cat $OUTPUT_DIR/resource-usage.csv | column -t -s,"
if [ -f "$OUTPUT_DIR/rss-chart.png" ]; then
    echo ""
    echo "To view charts:"
    echo "  open $OUTPUT_DIR/rss-chart.png"
    echo "  open $OUTPUT_DIR/cpu-memory-chart.png"
fi
echo ""
echo "Profiling complete!"
