#!/bin/bash

# Simple test script that outputs messages with delays
# This demonstrates how multirun handles multiple instances

INSTANCE=${1:-"unknown"}
COUNT=${2:-5}

echo "Starting instance $INSTANCE"

for i in $(seq 1 $COUNT); do
    echo "Instance $INSTANCE - Message $i"
    sleep 1
done

echo "Instance $INSTANCE completed"
