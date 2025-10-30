#!/bin/bash

# 使用 hostPath 方式部署简单的文件服务器用于备份测试
# 这个方案适合单节点 minikube 环境

set -e

NAMESPACE="default"

echo "🚀 部署本地文件服务器用于备份测试..."

# 1. 在 minikube 节点上创建备份目录
echo "📁 创建备份存储目录..."
minikube ssh "sudo mkdir -p /data/polardbx-backups && sudo chmod 777 /data/polardbx-backups"

# 2. 部署简单的 SFTP 服务器
cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: sftp-users
  namespace: $NAMESPACE
data:
  users.conf: |
    polardbx:polardbx123:1000:1000:/backups
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sftp-server
  namespace: $NAMESPACE
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sftp-server
  template:
    metadata:
      labels:
        app: sftp-server
    spec:
      containers:
      - name: sftp
        image: atmoz/sftp:latest
        ports:
        - containerPort: 22
          name: ssh
        volumeMounts:
        - name: backups
          mountPath: /home/polardbx/backups
        - name: users
          mountPath: /etc/sftp/users.conf
          subPath: users.conf
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: backups
        hostPath:
          path: /data/polardbx-backups
          type: DirectoryOrCreate
      - name: users
        configMap:
          name: sftp-users
---
apiVersion: v1
kind: Service
metadata:
  name: sftp-server
  namespace: $NAMESPACE
spec:
  type: NodePort
  ports:
  - port: 22
    targetPort: 22
    nodePort: 30022
    name: ssh
  selector:
    app: sftp-server
EOF

echo "⏳ 等待 SFTP 服务器就绪..."
kubectl wait --for=condition=Ready pod -l app=sftp-server -n $NAMESPACE --timeout=180s || {
    echo "⚠️  SFTP Pod 启动超时，检查状态..."
    kubectl get pod -l app=sftp-server -n $NAMESPACE
    kubectl describe pod -l app=sftp-server -n $NAMESPACE | tail -30
}

SFTP_POD=$(kubectl get pod -l app=sftp-server -n $NAMESPACE -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [ -z "$SFTP_POD" ]; then
    echo "❌ SFTP Pod 未能启动"
    exit 1
fi

echo "✅ SFTP 服务器部署完成！"
echo ""
echo "📝 SFTP 连接信息："
echo "   主机: sftp-server.$NAMESPACE.svc.cluster.local (集群内)"
echo "   主机: $(minikube ip):30022 (集群外)"
echo "   端口: 22"
echo "   用户名: polardbx"
echo "   密码: polardbx123"
echo "   根目录: /backups"
echo ""

# 3. 更新 HPFS 配置
echo "📋 更新 HPFS 配置..."

MINIKUBE_IP=$(minikube ip)

kubectl patch configmap polardbx-hpfs-config -n polardbx-operator-system --type merge -p "$(cat <<EOF
{
  "data": {
    "config.yaml": "sinks:\n  - host: sftp-server.default.svc.cluster.local\n    name: sftp\n    password: polardbx123\n    port: 22\n    rootPath: /backups\n    type: sftp\n    user: polardbx\nbackupBinlogConfig:\n  rootDirectories:\n    - /data/xstore\n    - /data-log/xstore\n"
  }
}
EOF
)"

echo "✅ HPFS 配置已更新"
echo ""

# 4. 重启 HPFS Pod 以应用新配置
echo "🔄 重启 HPFS Pods..."
kubectl delete pod -l app.kubernetes.io/name=polardbx-hpfs -n polardbx-operator-system

echo "⏳ 等待 HPFS Pods 重启..."
sleep 5
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=polardbx-hpfs -n polardbx-operator-system --timeout=60s || true

echo ""
echo "✅ 部署完成！"
echo ""
echo "🧪 测试 SFTP 连接："
echo "   kubectl exec -n $NAMESPACE $SFTP_POD -- ls -la /home/polardbx/backups"
echo ""
echo "🔧 现在可以运行备份测试："
echo "   ./test-backup-with-sftp.sh"
echo ""
