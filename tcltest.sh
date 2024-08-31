#!/bin/bash

# Define variables
SERVER_CMD="air"
TEST_CMD="./runtest --host 127.0.0.1 --port 7379 --tags -needs:debug --tags -cluster:skip"
SERVER_PORT=7379
SERVER_PID_FILE="server.pid"

# Start the Go server in the background
echo "Starting the server..."
$SERVER_CMD &
SERVER_PID=$!

# Save the server PID to a file for later use
echo $SERVER_PID > $SERVER_PID_FILE

# Wait for the server to start
echo "Waiting for the server to start..."
sleep 5  # Adjust this sleep duration based on how long the server takes to start

# Run the tests
echo "Running tests..."
$TEST_CMD

# Capture the exit code of the test command
TEST_EXIT_CODE=$?

# Stop the server
echo "Stopping the server..."
kill $SERVER_PID

# Remove the server PID file
rm -f $SERVER_PID_FILE

# Exit with the same code as the test command
exit $TEST_EXIT_CODE
