# API 文档

PolarDB-X UI 后端 API 接口文档。

## 基础信息

- **基础 URL**: `http://localhost:8080`
- **API 版本**: v1
- **内容类型**: `application/json`
- **认证方式**: Kubeconfig Base64 编码

## 认证

大部分 API 需要通过 `X-Kubeconfig-B64` 请求头提供 Kubeconfig 认证信息。

```bash
# 将 kubeconfig 文件编码为 base64
cat ~/.kube/config | base64

# 在请求中使用
curl -H "X-Kubeconfig-B64: <base64_encoded_kubeconfig>" \
     http://localhost:8080/api/v1/clusters
```

注意事项：
- 除 `/ping` 与 `/api/v1/connect` 以外的端点，均需在请求头中携带 `X-Kubeconfig-B64`。
- 受反向代理/网关及服务器 Header 大小限制影响，请确保 kubeconfig Base64 编码后不要过大；如存在超限风险，建议通过后端会话或短期令牌方式替代头传递方案。

## API 端点

### 健康检查

#### GET /ping

检查服务器状态。

**请求示例**:
```bash
curl -X GET http://localhost:8080/ping
```

**响应示例**:
```json
{
  "message": "pong"
}
```

**响应状态码**:
- `200`: 服务正常

---

### 连接验证

#### POST /api/v1/connect

验证 Kubeconfig 配置的有效性。

**请求头**:
- `Content-Type: text/plain`

**请求体**:
原始的 kubeconfig 文件内容

**请求示例**:
```bash
curl -X POST \
     -H "Content-Type: text/plain" \
     --data-binary @~/.kube/config \
     http://localhost:8080/api/v1/connect
```

**响应示例**:
```json
{
  "status": "success",
  "message": "Successfully connected to Kubernetes cluster"
}
```

**响应状态码**:
- `200`: 连接成功
- `401`: 认证失败
- `400`: 请求格式错误
- `500`: 服务器内部错误

---

### 集群管理
### 恢复管理

#### POST /api/v1/clusters/:namespace/:name/restore
从备份集发起集群恢复。

请求头：
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`
- `Content-Type: application/json`

请求体（示例）：
```json
{
  "spec": {
    "backupSet": "backup-2025-08-01",
    "storageProvider": {"type": "OSS", "bucket": "my-bucket", "endpoint": "oss-cn-example.aliyuncs.com"}
  }
}
```

响应：
- `202 Accepted`：恢复任务已创建
- `400`：参数错误
- `401`：认证失败
- `404`：目标集群不存在
- `500`：服务器内部错误

#### POST /api/v1/clusters/:namespace/:name/pitr
基于时间点发起集群恢复（PITR）。

请求头：同上

请求体（示例）：
```json
{
  "spec": {
    "time": "2025-08-01T10:00:00Z",
    "storageProvider": {"type": "OSS", "bucket": "my-bucket"}
  }
}
```

响应：
- `202 Accepted`
- `400`、`401`、`404`、`500`

#### GET /api/v1/clusters/:namespace/:name/restore-status
查询指定集群的恢复状态。

请求头：
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`

响应：
- `200 OK`：返回当前恢复阶段与进度
- `401`、`404`、`500`

#### GET /api/v1/restore-jobs
恢复任务列表。

请求头：同上

响应：
- `200 OK`：返回任务数组
- `401`、`500`

#### GET /api/v1/restore-jobs/:namespace/:name
恢复任务详情。

请求头：同上

响应：
- `200 OK`
- `401`、`404`、`500`

#### DELETE /api/v1/restore-jobs/:namespace/:name
取消（或清理）恢复任务。

请求头：同上

响应：
- `200 OK` 或 `204 No Content`
- `401`、`404`、`500`

---

### XStoreBackup 管理

#### GET /api/v1/xstore-backups
列出 XStoreBackup。

请求头：`X-Kubeconfig-B64`

响应：`200`、`401`、`500`

#### POST /api/v1/xstore-backups
创建 XStoreBackup。

请求头：`X-Kubeconfig-B64`，`Content-Type: application/json`

请求体（示例）：
```json
{
  "metadata": {"name": "xstore-backup-1", "namespace": "default"},
  "spec": {"type": "Full", "storageProvider": {"type": "OSS"}}
}
```

响应：`201`、`400`、`401`、`409`、`500`

#### GET /api/v1/xstore-backups/:namespace/:name
获取 XStoreBackup 详情。

请求头：`X-Kubeconfig-B64`

响应：`200`、`401`、`404`、`500`

#### PUT /api/v1/xstore-backups/:namespace/:name
更新 XStoreBackup。

请求头：`X-Kubeconfig-B64`，`Content-Type: application/json`

响应：`200`、`400`、`401`、`404`、`500`

#### DELETE /api/v1/xstore-backups/:namespace/:name
删除 XStoreBackup。

请求头：`X-Kubeconfig-B64`

响应：`200` 或 `204`、`401`、`404`、`500`

---

### XStoreFollower 管理

#### GET /api/v1/xstore-followers
列出 XStoreFollower。

请求头：`X-Kubeconfig-B64`

响应：`200`、`401`、`500`

#### POST /api/v1/xstore-followers
创建 XStoreFollower。

请求头：`X-Kubeconfig-B64`，`Content-Type: application/json`

请求体（示例）：
```json
{
  "metadata": {"name": "follower-1", "namespace": "default"},
  "spec": {"sourceXStore": "xstore-1"}
}
```

响应：`201`、`400`、`401`、`409`、`500`

#### GET /api/v1/xstore-followers/:namespace/:name
获取 XStoreFollower 详情。

请求头：`X-Kubeconfig-B64`

响应：`200`、`401`、`404`、`500`

#### PUT /api/v1/xstore-followers/:namespace/:name
更新 XStoreFollower。

请求头：`X-Kubeconfig-B64`，`Content-Type: application/json`

响应：`200`、`400`、`401`、`404`、`500`

#### DELETE /api/v1/xstore-followers/:namespace/:name
删除 XStoreFollower。

请求头：`X-Kubeconfig-B64`

响应：`200` 或 `204`、`401`、`404`、`500`

---

### BackupBinlog 管理

#### GET /api/v1/backup-binlogs
列出 Binlog 备份。

请求头：`X-Kubeconfig-B64`

响应：`200`、`401`、`500`

#### POST /api/v1/backup-binlogs
创建 Binlog 备份配置。

请求头：`X-Kubeconfig-B64`，`Content-Type: application/json`

请求体（示例）：
```json
{
  "metadata": {"name": "binlog-1", "namespace": "default"},
  "spec": {"provider": {"type": "OSS"}, "retention": {"days": 7}}
}
```

响应：`201`、`400`、`401`、`409`、`500`

#### GET /api/v1/backup-binlogs/:namespace/:name
获取 Binlog 备份详情。

请求头：`X-Kubeconfig-B64`

响应：`200`、`401`、`404`、`500`

#### PUT /api/v1/backup-binlogs/:namespace/:name
更新 Binlog 备份配置。

请求头：`X-Kubeconfig-B64`，`Content-Type: application/json`

响应：`200`、`400`、`401`、`404`、`500`

#### DELETE /api/v1/backup-binlogs/:namespace/:name
删除 Binlog 备份配置。

请求头：`X-Kubeconfig-B64`

响应：`200` 或 `204`、`401`、`404`、`500`

#### GET /api/v1/clusters

获取所有 PolarDB-X 集群列表。

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`

**请求示例**:
```bash
curl -X GET \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     http://localhost:8080/api/v1/clusters
```

**响应示例**:
```json
{
  "clusters": [
    {
      "metadata": {
        "name": "polardbx-cluster-1",
        "namespace": "default",
        "creationTimestamp": "2024-01-15T10:30:00Z",
        "labels": {
          "app": "polardbx"
        }
      },
      "spec": {
        "topology": {
          "nodes": {
            "cn": {
              "replicas": 2
            },
            "dn": {
              "replicas": 3
            },
            "gms": {
              "replicas": 1
            }
          }
        }
      },
      "status": {
        "phase": "Running",
        "conditions": [
          {
            "type": "Ready",
            "status": "True",
            "lastTransitionTime": "2024-01-15T10:35:00Z"
          }
        ]
      }
    }
  ]
}
```

**响应状态码**:
- `200`: 成功获取集群列表
- `401`: 认证失败
- `500`: 服务器内部错误

#### POST /api/v1/clusters

创建新的 PolarDB-X 集群。

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`
- `Content-Type: application/json`

**请求体**:
```json
{
  "name": "my-polardbx-cluster",
  "namespace": "default",
  "spec": {
    "topology": {
      "nodes": {
        "cn": {
          "replicas": 2,
          "resources": {
            "requests": {
              "cpu": "1",
              "memory": "2Gi"
            },
            "limits": {
              "cpu": "2",
              "memory": "4Gi"
            }
          }
        },
        "dn": {
          "replicas": 3,
          "resources": {
            "requests": {
              "cpu": "1",
              "memory": "2Gi"
            },
            "limits": {
              "cpu": "2",
              "memory": "4Gi"
            }
          }
        },
        "gms": {
          "replicas": 1,
          "resources": {
            "requests": {
              "cpu": "500m",
              "memory": "1Gi"
            },
            "limits": {
              "cpu": "1",
              "memory": "2Gi"
            }
          }
        }
      }
    }
  }
}
```

**请求示例**:
```bash
curl -X POST \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     -H "Content-Type: application/json" \
     -d @cluster-config.json \
     http://localhost:8080/api/v1/clusters
```

**响应示例**:
```json
{
  "status": "success",
  "message": "Cluster created successfully",
  "cluster": {
    "name": "my-polardbx-cluster",
    "namespace": "default",
    "status": "Creating"
  }
}
```

**响应状态码**:
- `201`: 集群创建成功
- `400`: 请求参数错误
- `401`: 认证失败
- `409`: 集群已存在
- `500`: 服务器内部错误

#### GET /api/v1/clusters/:name

获取指定集群的详细信息。

**路径参数**:
- `name`: 集群名称

**查询参数**:
- `namespace`: 命名空间（可选，默认为 default）

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`

**请求示例**:
```bash
curl -X GET \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     "http://localhost:8080/api/v1/clusters/my-cluster?namespace=default"
```

**响应示例**:
```json
{
  "cluster": {
    "metadata": {
      "name": "my-cluster",
      "namespace": "default",
      "creationTimestamp": "2024-01-15T10:30:00Z",
      "labels": {
        "app": "polardbx"
      }
    },
    "spec": {
      "topology": {
        "nodes": {
          "cn": {
            "replicas": 2
          },
          "dn": {
            "replicas": 3
          },
          "gms": {
            "replicas": 1
          }
        }
      }
    },
    "status": {
      "phase": "Running",
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "lastTransitionTime": "2024-01-15T10:35:00Z"
        }
      ],
      "nodes": {
        "cn": {
          "ready": 2,
          "total": 2
        },
        "dn": {
          "ready": 3,
          "total": 3
        },
        "gms": {
          "ready": 1,
          "total": 1
        }
      }
    }
  }
}
```

**响应状态码**:
- `200`: 成功获取集群信息
- `401`: 认证失败
- `404`: 集群不存在
- `500`: 服务器内部错误

#### PUT /api/v1/clusters/:name

更新指定集群的配置。

**路径参数**:
- `name`: 集群名称

**查询参数**:
- `namespace`: 命名空间（可选，默认为 default）

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`
- `Content-Type: application/json`

**请求体**:
```json
{
  "spec": {
    "topology": {
      "nodes": {
        "cn": {
          "replicas": 3
        },
        "dn": {
          "replicas": 5
        }
      }
    }
  }
}
```

**请求示例**:
```bash
curl -X PUT \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     -H "Content-Type: application/json" \
     -d @update-config.json \
     "http://localhost:8080/api/v1/clusters/my-cluster?namespace=default"
```

**响应示例**:
```json
{
  "status": "success",
  "message": "Cluster updated successfully",
  "cluster": {
    "name": "my-cluster",
    "namespace": "default",
    "status": "Updating"
  }
}
```

**响应状态码**:
- `200`: 集群更新成功
- `400`: 请求参数错误
- `401`: 认证失败
- `404`: 集群不存在
- `500`: 服务器内部错误

#### DELETE /api/v1/clusters/:name

删除指定集群。

**路径参数**:
- `name`: 集群名称

**查询参数**:
- `namespace`: 命名空间（可选，默认为 default）

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`

**请求示例**:
```bash
curl -X DELETE \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     "http://localhost:8080/api/v1/clusters/my-cluster?namespace=default"
```

**响应示例**:
```json
{
  "status": "success",
  "message": "Cluster deletion initiated"
}
```

**响应状态码**:
- `200`: 删除操作已启动
- `401`: 认证失败
- `404`: 集群不存在
- `500`: 服务器内部错误

---

### 备份管理

#### GET /api/v1/backups

获取所有备份列表。

**查询参数**:
- `cluster`: 集群名称（可选）
- `namespace`: 命名空间（可选，默认为 default）

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`

**请求示例**:
```bash
curl -X GET \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     "http://localhost:8080/api/v1/backups?cluster=my-cluster&namespace=default"
```

**响应示例**:
```json
{
  "backups": [
    {
      "metadata": {
        "name": "backup-20240115-103000",
        "namespace": "default",
        "creationTimestamp": "2024-01-15T10:30:00Z",
        "labels": {
          "cluster": "my-cluster"
        }
      },
      "spec": {
        "cluster": "my-cluster",
        "type": "full",
        "storageProvider": "s3"
      },
      "status": {
        "phase": "Completed",
        "completionTime": "2024-01-15T10:45:00Z",
        "size": "1.2GB"
      }
    }
  ]
}
```

**响应状态码**:
- `200`: 成功获取备份列表
- `401`: 认证失败
- `500`: 服务器内部错误

---

### 参数管理

#### GET /api/v1/parameters/:name

获取指定集群的参数配置。

**路径参数**:
- `name`: 集群名称

**查询参数**:
- `namespace`: 命名空间（可选，默认为 default）

**请求头**:
- `X-Kubeconfig-B64: <base64_encoded_kubeconfig>`

**请求示例**:
```bash
curl -X GET \
     -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64)" \
     "http://localhost:8080/api/v1/parameters/my-cluster?namespace=default"
```

**响应示例**:
```json
{
  "parameters": {
    "cn": {
      "max_connections": "1000",
      "innodb_buffer_pool_size": "1G",
      "query_cache_size": "256M"
    },
    "dn": {
      "max_connections": "500",
      "innodb_buffer_pool_size": "2G",
      "innodb_log_file_size": "512M"
    }
  }
}
```

**响应状态码**:
- `200`: 成功获取参数配置
- `401`: 认证失败
- `404`: 集群不存在
- `500`: 服务器内部错误

---

## 错误处理

### 错误响应格式

```json
{
  "error": {
    "code": "CLUSTER_NOT_FOUND",
    "message": "The specified cluster was not found",
    "details": {
      "cluster": "my-cluster",
      "namespace": "default"
    }
  }
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| `INVALID_KUBECONFIG` | 401 | Kubeconfig 无效或格式错误 |
| `CLUSTER_NOT_FOUND` | 404 | 指定的集群不存在 |
| `CLUSTER_ALREADY_EXISTS` | 409 | 集群已存在 |
| `INVALID_REQUEST` | 400 | 请求参数无效 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |
| `KUBERNETES_API_ERROR` | 500 | Kubernetes API 调用失败 |

## 使用示例

### 完整工作流示例

```bash
#!/bin/bash

# 1. 验证连接
echo "Testing connection..."
curl -X POST \
     -H "Content-Type: text/plain" \
     --data-binary @~/.kube/config \
     http://localhost:8080/api/v1/connect

# 2. 获取集群列表
echo "Getting cluster list..."
KUBECONFIG_B64=$(cat ~/.kube/config | base64)
curl -X GET \
     -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
     http://localhost:8080/api/v1/clusters

# 3. 创建新集群
echo "Creating new cluster..."
curl -X POST \
     -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "test-cluster",
       "namespace": "default",
       "spec": {
         "topology": {
           "nodes": {
             "cn": {"replicas": 2},
             "dn": {"replicas": 3},
             "gms": {"replicas": 1}
           }
         }
       }
     }' \
     http://localhost:8080/api/v1/clusters

# 4. 获取集群详情
echo "Getting cluster details..."
curl -X GET \
     -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
     http://localhost:8080/api/v1/clusters/test-cluster

# 5. 更新集群
echo "Updating cluster..."
curl -X PUT \
     -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
     -H "Content-Type: application/json" \
     -d '{
       "spec": {
         "topology": {
           "nodes": {
             "cn": {"replicas": 3}
           }
         }
       }
     }' \
     http://localhost:8080/api/v1/clusters/test-cluster

# 6. 获取备份列表
echo "Getting backup list..."
curl -X GET \
     -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
     "http://localhost:8080/api/v1/backups?cluster=test-cluster"

# 7. 删除集群
echo "Deleting cluster..."
curl -X DELETE \
     -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
     http://localhost:8080/api/v1/clusters/test-cluster
```

### JavaScript/TypeScript 示例

```typescript
class PolarDBXClient {
  private baseURL: string;
  private kubeconfigB64: string;

  constructor(baseURL: string, kubeconfig: string) {
    this.baseURL = baseURL;
    this.kubeconfigB64 = btoa(kubeconfig);
  }

  private async request(method: string, path: string, data?: any) {
    const headers: Record<string, string> = {
      'X-Kubeconfig-B64': this.kubeconfigB64,
    };

    if (data) {
      headers['Content-Type'] = 'application/json';
    }

    const response = await fetch(`${this.baseURL}${path}`, {
      method,
      headers,
      body: data ? JSON.stringify(data) : undefined,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  async getClusters() {
    return this.request('GET', '/api/v1/clusters');
  }

  async getCluster(name: string, namespace = 'default') {
    return this.request('GET', `/api/v1/clusters/${name}?namespace=${namespace}`);
  }

  async createCluster(clusterConfig: any) {
    return this.request('POST', '/api/v1/clusters', clusterConfig);
  }

  async updateCluster(name: string, updateConfig: any, namespace = 'default') {
    return this.request('PUT', `/api/v1/clusters/${name}?namespace=${namespace}`, updateConfig);
  }

  async deleteCluster(name: string, namespace = 'default') {
    return this.request('DELETE', `/api/v1/clusters/${name}?namespace=${namespace}`);
  }

  async getBackups(cluster?: string, namespace = 'default') {
    const params = new URLSearchParams();
    if (cluster) params.append('cluster', cluster);
    params.append('namespace', namespace);
    
    return this.request('GET', `/api/v1/backups?${params}`);
  }
}

// 使用示例
const client = new PolarDBXClient('http://localhost:8080', kubeconfigContent);

try {
  const clusters = await client.getClusters();
  console.log('Clusters:', clusters);
} catch (error) {
  console.error('Error:', error);
}
```

## 版本历史

### v1.0.0
- 初始版本
- 支持基本的集群管理功能
- 支持备份列表查询
- 支持参数配置查询

---

更多信息请参考 [README.md](./README.md) 和 [开发环境配置指南](./DEVELOPMENT.md)。