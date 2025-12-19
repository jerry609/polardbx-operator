#!/bin/bash
# Test script for all-in-one UI backend deployment
# This script builds the image, runs it locally, and tests the endpoints

set -e

IMAGE_NAME="polardbx-ui-all-in-one"
IMAGE_TAG="dev"
CONTAINER_NAME="polardbx-ui-test"
PORT=8080

echo "=== Building all-in-one Docker image ==="
docker build -f tools/ui-all-in-one.Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .

echo ""
echo "=== Stopping existing container if any ==="
docker stop ${CONTAINER_NAME} 2>/dev/null || true
docker rm ${CONTAINER_NAME} 2>/dev/null || true

echo ""
echo "=== Starting container ==="
docker run -d \
  --name ${CONTAINER_NAME} \
  -p ${PORT}:8080 \
  -e UI_STATIC_DIR=/app/ui \
  ${IMAGE_NAME}:${IMAGE_TAG}

echo ""
echo "=== Waiting for container to be ready ==="
sleep 5

echo ""
echo "=== Testing endpoints ==="

# Test health endpoint
echo "1. Testing /api/v1/health"
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/api/v1/health || echo "000")
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

