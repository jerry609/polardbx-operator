#!/bin/bash

# 快速部署 MinIO 用于本地 PolarDB-X 备份测试
# MinIO 是 S3 兼容的对象存储，适合本地测试

set -e

NAMESPACE="default"
MINIO_ROOT_USER="minioadmin"
MINIO_ROOT_PASSWORD="minioadmin123"
MINIO_BUCKET="polardbx-backups"

echo "🚀 部署 MinIO 对象存储服务..."

# 1. 创建 MinIO Deployment 和 Service
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: minio-pvc
  namespace: $NAMESPACE
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: $NAMESPACE
spec:
  selector:
    matchLabels:
      app: minio
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        imagePullPolicy: Never
        args:
        - server
        - /data
        - --console-address
        - ":9001"
        env:
        - name: MINIO_ROOT_USER
          value: "$MINIO_ROOT_USER"
        - name: MINIO_ROOT_PASSWORD
          value: "$MINIO_ROOT_PASSWORD"
        ports:
        - containerPort: 9000
          name: api
        - containerPort: 9001
          name: console
        volumeMounts:
        - name: storage
          mountPath: /data
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
      volumes:
      - name: storage
        persistentVolumeClaim:
          claimName: minio-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: $NAMESPACE
spec:
  type: NodePort
  ports:
  - port: 9000
    targetPort: 9000
    nodePort: 30900
    name: api
  - port: 9001
    targetPort: 9001
    nodePort: 30901
    name: console
  selector:
    app: minio
EOF

echo "⏳ 等待 MinIO Pod 就绪..."
kubectl wait --for=condition=Ready pod -l app=minio -n $NAMESPACE --timeout=300s

echo "✅ MinIO 部署完成！"
echo ""
echo "📝 连接信息："
echo "   API 端点: http://$(minikube ip):30900"
echo "   控制台: http://$(minikube ip):30901"
echo "   用户名: $MINIO_ROOT_USER"
echo "   密码: $MINIO_ROOT_PASSWORD"
echo ""

# 2. 创建 bucket
echo "📦 创建 MinIO bucket: $MINIO_BUCKET"

MINIO_POD=$(kubectl get pod -l app=minio -n $NAMESPACE -o jsonpath='{.items[0].metadata.name}')

# 安装 mc 客户端并创建 bucket
kubectl exec -n $NAMESPACE $MINIO_POD -- sh -c "
  # 下载 mc 客户端（如果不存在）
  if [ ! -f /usr/local/bin/mc ]; then
    wget -q https://dl.min.io/client/mc/release/linux-amd64/mc -O /usr/local/bin/mc
    chmod +x /usr/local/bin/mc
  fi
  
  # 配置 mc
  mc alias set local http://localhost:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD
  
  # 创建 bucket
  mc mb local/$MINIO_BUCKET --ignore-existing
  
  # 列出 buckets
  mc ls local/
"

echo ""
echo "✅ MinIO bucket 创建完成！"
echo ""

# 3. 创建 HPFS Sink 配置
echo "📋 创建 HPFS Sink 配置..."

MINIO_ENDPOINT="http://minio.$NAMESPACE.svc.cluster.local:9000"

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: hpfs-minio-sink
  namespace: $NAMESPACE
data:
  sink.yaml: |
    name: s3
    type: s3
    config:
      endpoint: "$MINIO_ENDPOINT"
      region: "us-east-1"
      bucket: "$MINIO_BUCKET"
      access_key: "$MINIO_ROOT_USER"
      secret_key: "$MINIO_ROOT_PASSWORD"
      use_ssl: false
      path_style: true
EOF

echo ""
echo "✅ 配置创建完成！"
echo ""
echo "🧪 测试配置："
echo ""
echo "export MINIO_ENDPOINT=\"http://$(minikube ip):30900\""
echo "export MINIO_ACCESS_KEY=\"$MINIO_ROOT_USER\""
echo "export MINIO_SECRET_KEY=\"$MINIO_ROOT_PASSWORD\""
echo "export MINIO_BUCKET=\"$MINIO_BUCKET\""
echo ""
echo "现在可以运行备份测试脚本："
echo "  ./test-backup-pitr.sh s3"
echo ""
echo "或者在浏览器中访问 MinIO 控制台："
echo "  http://$(minikube ip):30901"
echo ""
