#!/bin/bash

BACKEND_URL="http://localhost:8080"
CLUSTER_NAME="test-pitr"
BACKUP_NAME="test-backup-$(date +%Y%m%d-%H%M%S)"

# Get kubeconfig
KUBECONFIG_B64=$(cat ~/.kube/config | base64 -w 0)

# Create backup
echo "Creating backup: $BACKUP_NAME"
curl -X POST "$BACKEND_URL/api/v1/clusters/default/$CLUSTER_NAME/backups" \
  -H "Content-Type: application/json" \
  -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
  -d '{
    "name": "'$BACKUP_NAME'",
    "cluster": {
      "name": "'$CLUSTER_NAME'"
    },
    "storageProvider": {
      "storageName": "s3",
      "sink": "s3"
    }
  }' | jq .

echo "Checking backup status..."
sleep 5
kubectl get polardbxbackup $BACKUP_NAME
