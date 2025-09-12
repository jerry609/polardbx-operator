#!/usr/bin/env bash
set -euo pipefail

KNS="${1:-default}"
EXCLUDE_PATTERN="${2:-mybackupsucesstest}"

echo "Namespace: $KNS"
echo "Exclude pattern: $EXCLUDE_PATTERN"

BACKUPS=$(kubectl -n "$KNS" get polardbxbackups -o jsonpath='{.items[*].metadata.name}')
for B in $BACKUPS; do
  if [[ "$B" == *"$EXCLUDE_PATTERN"* ]]; then
    echo "Skip (excluded): $B"
    continue
  fi
  echo "Force deleting backup: $B"

  # 1) 子备份：移除 finalizers 并删除
  XSBS=$(kubectl -n "$KNS" get xstorebackups.polardbx.aliyun.com -l polardbx/top-backup="$B" -o name || true)
  if [[ -n "${XSBS:-}" ]]; then
    while IFS= read -r R; do
      [[ -z "$R" ]] && continue
      kubectl -n "$KNS" patch "$R" --type=merge -p '{"metadata":{"finalizers":[]}}' || true
    done <<< "$XSBS"

    kubectl -n "$KNS" delete xstorebackups.polardbx.aliyun.com -l polardbx/top-backup="$B" --wait=false --ignore-not-found
  fi

  # 2) 关联 Job（按每个子备份的标签删除）
  if [[ -n "${XSBS:-}" ]]; then
    while IFS= read -r R; do
      XSB_NAME="${R##*/}"
      [[ -z "$XSB_NAME" ]] && continue
      kubectl -n "$KNS" delete jobs -l xstore/backup="$XSB_NAME" --wait=false --ignore-not-found || true
    done <<< "$XSBS"
  fi

  # 3) 顶层备份：移除 finalizers 并删除
  kubectl -n "$KNS" patch polardbxbackups "$B" --type=merge -p '{"metadata":{"finalizers":[]}}' || true
  kubectl -n "$KNS" delete polardbxbackups "$B" --wait=false --ignore-not-found || true

  echo "Requested deletion: $B"
done

echo "Done."
