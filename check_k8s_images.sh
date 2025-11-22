#!/bin/bash

# This script checks if Docker images specified in all Kubernetes Deployments and StatefulSets
# within a given namespace exist in Docker repositories by attempting to pull them.

set -e

# Function to display usage information
usage() {
    echo "Usage: $0 <namespace>"
    echo "Example: $0 default"
    exit 1
}

# Check if exactly one argument (namespace) is provided
if [ "$#" -ne 1 ]; then
    echo "Error: Namespace argument is required."
    usage
fi

NAMESPACE="$1"

# Verify that the specified namespace exists
if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "Error: Namespace '$NAMESPACE' does not exist."
    exit 1
fi

# Function to extract images from Kubernetes resources
get_images() {
    local resource_type=$1
    kubectl get "$resource_type" -n "$NAMESPACE" -o jsonpath='{range .items[*]}{.kind}{" "}{.metadata.name}{" "}{range .spec.template.spec.initContainers[*]}{.name}{" "}{.image}{"\n"}{end}{range .spec.template.spec.containers[*]}{.name}{" "}{.image}{"\n"}{end}{end}'
}

# Get images from Deployments and StatefulSets
DEPLOYMENT_IMAGES=$(get_images deployments)
STATEFULSET_IMAGES=$(get_images statefulsets)

# Combine and remove empty lines
ALL_IMAGES=$(printf "%s\n%s" "$DEPLOYMENT_IMAGES" "$STATEFULSET_IMAGES" | sed '/^\s*$/d')

# Remove duplicates
ALL_IMAGES=$(echo "$ALL_IMAGES" | sort | uniq)

echo "Checking images used in Kubernetes Deployments and StatefulSets in namespace '$NAMESPACE'..."
echo

# Iterate over each image and attempt to pull it
while read -r line; do
    KIND=$(echo "$line" | awk '{print $1}')
    RESOURCE_NAME=$(echo "$line" | awk '{print $2}')
    CONTAINER_NAME=$(echo "$line" | awk '{print $3}')
    IMAGE=$(echo "$line" | awk '{print $4}')

    if [ -z "$IMAGE" ]; then
        continue
    fi

    echo "Resource: $KIND/$RESOURCE_NAME"
    echo "Container: $CONTAINER_NAME"
    echo "Image: $IMAGE"

    # Attempt to pull the image
    if docker pull "$IMAGE" >/dev/null 2>&1; then
        echo "Result: ✅ Image exists and can be pulled."
    else
        echo "Result: ❌ Image does not exist or cannot be pulled."
    fi
    echo "----------------------------------------"
done <<< "$ALL_IMAGES"
