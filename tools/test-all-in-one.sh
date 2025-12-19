#!/bin/bash
# Test script for all-in-one UI backend deployment
# This script builds the image, runs it locally, and tests the endpoints

set -e

IMAGE_NAME="polardbx-dashboard"
IMAGE_TAG="test"
CONTAINER_NAME="polardbx-dashboard-test"
PORT=8080

echo "=== Building all-in-one Docker image ==="
docker build -f tools/dashboard.Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .

echo ""
echo "=== Stopping existing container if any ==="
docker stop ${CONTAINER_NAME} 2>/dev/null || true
docker rm ${CONTAINER_NAME} 2>/dev/null || true

echo ""
echo "=== Starting container ==="
# Note: For local development with minikube, we need to:
# 1. Use host network mode to access minikube API server
# 2. Mount kubeconfig file
# 3. Mount .minikube directory (contains certificate files referenced by kubeconfig)
docker run -d \
  --name ${CONTAINER_NAME} \
  --network host \
  -v ${HOME}/.kube/config:/etc/kube/kubeconfig:ro \
  -v ${HOME}/.minikube:${HOME}/.minikube:ro \
  -e UI_STATIC_DIR=/app/ui \
  -e KUBECONFIG=/etc/kube/kubeconfig \
  ${IMAGE_NAME}:${IMAGE_TAG}

echo ""
echo "=== Waiting for container to be ready ==="
sleep 5

echo ""
echo "=== Testing endpoints ==="

# Test health endpoint
echo "1. Testing /health"
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/health || echo "000")
if [ "$HEALTH_RESPONSE" = "200" ]; then
  echo "   ✓ Health check passed (200)"
else
  echo "   ✗ Health check failed (got $HEALTH_RESPONSE)"
fi

# Test root (should serve index.html)
echo "2. Testing / (root path)"
ROOT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/ || echo "000")
if [ "$ROOT_RESPONSE" = "200" ]; then
  echo "   ✓ Root path serves UI (200)"
else
  echo "   ✗ Root path failed (got $ROOT_RESPONSE)"
fi

# Test API endpoint
echo "3. Testing /api/v1/system/context"
API_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/api/v1/system/context || echo "000")
if [ "$API_RESPONSE" = "200" ] || [ "$API_RESPONSE" = "401" ] || [ "$API_RESPONSE" = "500" ]; then
  echo "   ✓ API endpoint responds (got $API_RESPONSE, expected 200/401/500)"
else
  echo "   ✗ API endpoint failed (got $API_RESPONSE)"
fi

echo ""
echo "=== Container logs (last 20 lines) ==="
docker logs --tail 20 ${CONTAINER_NAME}

echo ""
echo "=== Summary ==="
echo "Container: ${CONTAINER_NAME}"
echo "Image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "Port: http://localhost:${PORT}"
echo ""
echo "To stop the container: docker stop ${CONTAINER_NAME}"
echo "To view logs: docker logs -f ${CONTAINER_NAME}"

