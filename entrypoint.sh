# entrypoint.sh

#!/bin/bash
set -e

# Start Docker daemon in the background
#echo "Starting Docker daemon..."
#dockerd &

# Wait for Docker daemon to be ready
#echo "Waiting for Docker daemon to be ready..."
#until docker info >/dev/null 2>&1; do
#  sleep 1
#done

# Execute the main script
exec "$@"
