#!/usr/bin/env bash
NS="${NS:-polardbx-monitor}"
kubectl get ns "$NS" >/dev/null 2>&1 || NS="monitoring"
kubectl get ns "$NS" >/dev/null 2>&1 || { echo "未找到命名空间 polardbx-monitor/monitoring"; exit 1; }

echo "命名空间: $NS"
echo "Prometheus 相关 Pod："
kubectl -n "$NS" get pods | awk 'NR==1 || /prom|prometheus/i'

# 选择优先使用 Service，其次 Pod 做端口转发
TARGET="$(kubectl -n "$NS" get svc -o name | grep -i prom | head -n1)"
[ -z "$TARGET" ] && TARGET="$(kubectl -n "$NS" get pod -o name | grep -i prom | head -n1)"
[ -z "$TARGET" ] && { echo "未找到 Prometheus Service/Pod"; exit 1; }

echo "端口转发目标: $TARGET"
kubectl -n "$NS" port-forward "$TARGET" 9090:9090 >/tmp/pf_prom.log 2>&1 & PF_PID=$!
sleep 2

http_code() { curl -s -o /dev/null -w "%{http_code}" "$1"; }

R=$(http_code http://127.0.0.1:9090/-/ready)
H=$(http_code http://127.0.0.1:9090/-/healthy)
TJSON=$(curl -s http://127.0.0.1:9090/api/v1/targets)
UP=$(echo "$TJSON" | grep -o '"health":"up"' | wc -l | tr -d ' ')
TOT=$(echo "$TJSON" | grep -o '"health":"' | wc -l | tr -d ' ')
QJSON=$(curl -s "http://127.0.0.1:9090/api/v1/query?query=up")
QS=$(echo "$QJSON" | grep -o '"status":"success"' | wc -l | tr -d ' ')
BJSON=$(curl -s http://127.0.0.1:9090/api/v1/status/buildinfo)
VER=$(echo "$BJSON" | sed -nE 's/.*"version":"([^"]+)".*/\1/p')

echo "就绪(ready): $R  健康(healthy): $H"
echo "采集目标 Targets: up=$UP / total=$TOT"
echo "即时查询 up: $( [ "$QS" = "1" ] && echo success || echo fail )"
echo "版本: ${VER:-unknown}"

OK=1
[ "$R" = "200" ] || { echo "检查失败: /-/ready 非 200"; OK=0; }
[ "$H" = "200" ] || { echo "检查失败: /-/healthy 非 200"; OK=0; }
if [ "$TOT" -gt 0 ]; then
  [ "$UP" = "$TOT" ] || { echo "检查失败: 目标并非全部 up ($UP/$TOT)"; OK=0; }
else
  echo "提示: 当前无活跃采集目标（请确认 ServiceMonitor/PolarDBXMonitor 是否已创建）"
fi
[ "$QS" = "1" ] || { echo "检查失败: PromQL 查询接口异常"; OK=0; }

kill "$PF_PID" >/dev/null 2>&1 || true
wait "$PF_PID" 2>/dev/null || true

[ "$OK" = "1" ] && echo "Prometheus: OK ✅" || { echo "Prometheus: 异常 ❌"; exit 1; }