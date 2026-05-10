#!/bin/bash
# Quick profiling - shows real-time CPU and memory for gazelle and Java parser
# Usage: ./quick-profile.sh

set -e

echo "Starting gazelle..."
if [ $# -eq 0 ]; then
    bazel run //:gazelle &
else
    "$@" &
fi
GAZELLE_PID=$!

echo "Gazelle PID: $GAZELLE_PID"
echo "Waiting for processes to start..."
sleep 2

# Find Java parser
JAVA_PID=$(ps -o pid= --ppid $GAZELLE_PID 2>/dev/null | head -1)
if [ -z "$JAVA_PID" ]; then
    JAVA_PID=$(ps aux | grep -E "java.*Main|javaparser" | grep -v grep | awk '{print $2}' | head -1)
fi

echo ""
echo "=== Process Information ==="
echo "Gazelle PID: $GAZELLE_PID"
if [ -n "$JAVA_PID" ]; then
    echo "Java Parser PID: $JAVA_PID"
else
    echo "Java Parser: Not found (may start later)"
fi
echo ""

# Show process tree
echo "=== Process Tree ==="
ps -ef | grep -E "($GAZELLE_PID|$JAVA_PID)" | grep -v grep || echo "No processes found"
echo ""

# Monitor in real-time
echo "=== Real-time Monitoring (Ctrl+C to stop) ==="
echo ""

trap 'echo ""; echo "Stopping..."; kill $GAZELLE_PID 2>/dev/null || true; exit' INT TERM

while ps -p $GAZELLE_PID > /dev/null 2>&1; do
    clear
    echo "=== Gazelle Process Monitoring ==="
    echo "Time: $(date)"
    echo ""

    if ps -p $GAZELLE_PID > /dev/null 2>&1; then
        echo "Gazelle (PID $GAZELLE_PID):"
        ps -o pid,pcpu,pmem,rss,vsz,time,command -p $GAZELLE_PID
        echo ""
    fi

    if [ -n "$JAVA_PID" ] && ps -p $JAVA_PID > /dev/null 2>&1; then
        echo "Java Parser (PID $JAVA_PID):"
        ps -o pid,pcpu,pmem,rss,vsz,time,command -p $JAVA_PID
        echo ""
    fi

    echo "Press Ctrl+C to stop monitoring (gazelle will continue)"
    sleep 2
done

echo ""
echo "Gazelle completed."
