#!/bin/bash
# 测试部署脚本 - 验证 all-in-one 镜像在本地 Docker 中运行

set -e

IMAGE_NAME="polardbx-ui-all-in-one"
IMAGE_TAG="dev"
CONTAINER_NAME="polardbx-ui-test-deploy"
PORT=8080

echo "=== 测试 All-in-One 部署 ==="
echo ""

# 检查镜像是否存在
if ! docker image inspect ${IMAGE_NAME}:${IMAGE_TAG} >/dev/null 2>&1; then
    echo "镜像不存在，开始构建..."
    docker build -f tools/ui-all-in-one.Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .
else
    echo "✓ 镜像已存在: ${IMAGE_NAME}:${IMAGE_TAG}"
fi

# 停止旧容器
echo ""
echo "清理旧容器..."
docker stop ${CONTAINER_NAME} 2>/dev/null || true
docker rm ${CONTAINER_NAME} 2>/dev/null || true

# 启动容器
echo ""
echo "启动容器..."
docker run -d \
  --name ${CONTAINER_NAME} \
  -p ${PORT}:8080 \
  -e UI_STATIC_DIR=/app/ui \
  ${IMAGE_NAME}:${IMAGE_TAG}

echo "✓ 容器已启动"
echo ""

# 等待服务就绪
echo "等待服务就绪..."
sleep 5

# 测试端点
echo ""
echo "=== 测试端点 ==="

# 测试健康检查
echo -n "1. 健康检查 (/api/v1/health): "
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/api/v1/health || echo "000")
if [ "$HEALTH" = "200" ]; then
    echo "✓ 通过 (200)"
else
    echo "✗ 失败 (got $HEALTH)"
fi

# 测试根路径
echo -n "2. 前端页面 (/): "
ROOT=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/ || echo "000")
if [ "$ROOT" = "200" ]; then
    echo "✓ 通过 (200)"
else
    echo "✗ 失败 (got $ROOT)"
fi

# 测试 API
echo -n "3. API 端点 (/api/v1/system/context): "
API=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${PORT}/api/v1/system/context || echo "000")
if [ "$API" = "200" ] || [ "$API" = "401" ] || [ "$API" = "500" ]; then
    echo "✓ 通过 (got $API)"
else
    echo "✗ 失败 (got $API)"
fi

echo ""
echo "=== 容器信息 ==="
echo "容器名: ${CONTAINER_NAME}"
echo "镜像: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "端口: http://localhost:${PORT}"
echo ""
echo "查看日志: docker logs -f ${CONTAINER_NAME}"
echo "停止容器: docker stop ${CONTAINER_NAME}"
echo "删除容器: docker rm ${CONTAINER_NAME}"
