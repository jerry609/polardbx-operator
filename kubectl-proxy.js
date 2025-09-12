#!/usr/bin/env node

// kubectl proxy 封装器，用于优化终端性能
const { spawn } = require('child_process');
const WebSocket = require('ws');
const express = require('express');
const cors = require('cors');

const app = express();
app.use(cors());
app.use(express.json());

// 启动 kubectl proxy
const kubectlProxy = spawn('kubectl', ['proxy', '--port=8001'], {
  stdio: ['pipe', 'pipe', 'pipe']
});

kubectlProxy.stdout.on('data', (data) => {
  console.log(`kubectl proxy: ${data}`);
});

kubectlProxy.stderr.on('data', (data) => {
  console.error(`kubectl proxy error: ${data}`);
});

// WebSocket代理服务器
const wss = new WebSocket.Server({ port: 8004 });

wss.on('connection', (ws, req) => {
  const url = new URL(req.url, 'http://localhost');
  const namespace = url.searchParams.get('namespace');
  const pod = url.searchParams.get('pod');
  const container = url.searchParams.get('container') || '';
  
  console.log(`Terminal connection: ${namespace}/${pod}/${container}`);
  
  // 构建kubectl exec命令，直接连接，避免额外的代理层
  const kubectlExec = spawn('kubectl', [
    'exec', 
    '-it',
    `${pod}`,
    `-n${namespace}`,
    ...(container ? ['-c', container] : []),
    '--',
    '/bin/bash'
  ], {
    stdio: ['pipe', 'pipe', 'pipe'],
    env: { ...process.env, TERM: 'xterm-256color' }
  });
  
  // 直接管道连接，最小化延迟
  ws.on('message', (data) => {
    if (kubectlExec.stdin.writable) {
      kubectlExec.stdin.write(data);
    }
  });
  
  kubectlExec.stdout.on('data', (data) => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(data);
    }
  });
  
  kubectlExec.stderr.on('data', (data) => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(data);
    }
  });
  
  kubectlExec.on('close', () => {
    ws.close();
  });
  
  ws.on('close', () => {
    kubectlExec.kill();
  });
});

// 健康检查端点
app.get('/health', (req, res) => {
  res.json({ status: 'ok', proxy: 'kubectl-direct' });
});

app.listen(8005, () => {
  console.log('kubectl proxy wrapper running on port 8005');
  console.log('WebSocket terminal server on port 8004');
  console.log('Direct kubectl exec bypass enabled');
});

process.on('SIGINT', () => {
  console.log('Shutting down kubectl proxy wrapper...');
  kubectlProxy.kill();
  process.exit(0);
});