import { AfterViewInit, Component, ElementRef, OnDestroy, OnInit, ViewChild, inject } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, ActivatedRoute, RouterModule } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatTabsModule } from '@angular/material/tabs';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatInputModule } from '@angular/material/input';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatFormFieldModule } from '@angular/material/form-field';
import { Chart, registerables, ChartConfiguration, ChartType } from 'chart.js';
import * as dagre from 'dagre';

import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { PolarDBXCluster } from '../../models/cluster.model';
import { Pod } from '../../models/pod.model';
import { PolarDBXBackup, BackupInfo, CreateBackupRequest } from '../../models/backup.model';

Chart.register(...registerables);

interface NodeInfo {
  name: string;
  type: string;
  status: string;
  cpu: string;
  memory: string;
  ip: string;
  resources?: string; // 显示的资源信息
  readyContainers?: number; // 就绪容器数量
  totalContainers?: number; // 总容器数量
  totalRestarts?: number; // 总重启次数
  hasUnhealthyContainers?: boolean; // 是否有不健康的容器
  containerStatuses?: ContainerStatus[]; // 容器状态详情
}

interface ContainerStatus {
  name: string;
  ready: boolean;
  restartCount: number;
  state?: string; // Running, Waiting, Terminated
  reason?: string; // 状态原因
  message?: string; // 详细信息
}



@Component({
  selector: 'app-cluster-detail',
  templateUrl: './cluster-detail.component.html',
  styleUrls: ['./cluster-detail.component.scss'],
  standalone: true,
  imports: [
    CommonModule,
    DatePipe,
    FormsModule,
    ReactiveFormsModule,
    RouterModule,
    MatButtonModule,
    MatCardModule,
    MatIconModule,
    MatSelectModule,
    MatSnackBarModule,
    MatTabsModule,
    MatTableModule,
    MatInputModule,
    MatChipsModule,
    MatTooltipModule,
    MatFormFieldModule,
    MatDialogModule
  ]
})
export class ClusterDetailComponent implements OnInit, AfterViewInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private fb = inject(FormBuilder);
  private snackBar = inject(MatSnackBar);
  private dialog = inject(MatDialog);
  private apiService = inject(ApiService);
  private loadingService = inject(LoadingService);

  @ViewChild('cpuChart') cpuChartRef!: ElementRef<HTMLCanvasElement>;
  @ViewChild('memChart') memChartRef!: ElementRef<HTMLCanvasElement>;
  @ViewChild('topologySvg') topologySvgRef!: ElementRef<SVGSVGElement>;

  // 集群基本信息
  clusterName = '';
  clusterNamespace = '';
  clusterStatus = 'Running';
  clusterVersion = '5.4.18';

  // 架构信息
  cnReplicas = 2;
  dnReplicas = 2;
  gmsReplicas = 1;
  cdcReplicas = 1;

  // 选项卡索引
  selectedTabIndex = 0;

  // 节点数据
  nodeDataSource = new MatTableDataSource<NodeInfo>([]);
  nodeColumns: string[] = ['name', 'type', 'status', 'ready', 'restarts', 'resources', 'ip'];
  // 增加操作列
  constructorColumns() {
    if (!this.nodeColumns.includes('actions')) this.nodeColumns.push('actions');
  }

  // 备份数据
  backupDataSource = new MatTableDataSource<BackupInfo>([]);
  backupColumns: string[] = ['id', 'completedTime', 'type', 'status', 'actions'];

  // 表单
  configForm: FormGroup;
  upgradeForm: FormGroup;

  // 日志相关
  selectedLogNode = 'all';
  logContent = '请选择一个节点和容器，然后点击“查看日志”来加载日志。';
  isLoadingLogs = false;

  // 图表实例
  private cpuChart?: Chart;
  private memChart?: Chart;

  cluster: PolarDBXCluster | null = null;
  cnCount = 0;
  dnCount = 0;
  gmsCount = 0;
  cdcCount = 0;
  allPods: Pod[] = [];
  selectedPod = '';
  selectedContainer = '';
  selectedPodContainers: { name: string }[] = [];
  rowContainer: Record<string, string> = {};
  alerts = { critical: 0, warning: 0, info: 0, total: 0, source: 'none' };
  // 明确声明键，避免索引签名导致 TS4111
  roleSummary: {
    cn: { total: number; ready: number; pods: Pod[] };
    dn: { total: number; ready: number; pods: Pod[] };
    gms: { total: number; ready: number; pods: Pod[] };
    cdc: { total: number; ready: number; pods: Pod[] };
  } = {
    cn: { total: 0, ready: 0, pods: [] },
    dn: { total: 0, ready: 0, pods: [] },
    gms: { total: 0, ready: 0, pods: [] },
    cdc: { total: 0, ready: 0, pods: [] }
  };

  @ViewChild('resizeHandle', { static: false }) resizeHandle!: ElementRef;
  @ViewChild('leftPanel', { static: false }) leftPanel!: ElementRef;

  constructor() {
    // 初始化表单
    this.configForm = this.fb.group({
      cnReplicas: [this.cnReplicas, [Validators.required, Validators.min(1)]],
      cnCpu: ['1', [Validators.required]],
      cnMemory: ['2', [Validators.required]],
      dnReplicas: [this.dnReplicas, [Validators.required, Validators.min(1)]],
      dnCpu: ['1', [Validators.required]],
      dnMemory: ['2', [Validators.required]]
    });

    this.upgradeForm = this.fb.group({
      targetVersion: ['5.4.18', Validators.required]
    });
  }

  ngOnInit(): void {
    // 强制清除所有可能残留的加载状态
    this.loadingService.clearAll();
    this.constructorColumns();
    
    // 延迟再次清除，确保状态被正确重置
    setTimeout(() => {
      this.loadingService.clearAll();
    }, 100);
    
    // 从路由参数获取集群信息
    this.route.params.subscribe(params => {
      this.clusterName = params['name'] || 'demo-cluster';
    });
    
    // 从查询参数获取命名空间
    this.route.queryParams.subscribe(queryParams => {
      this.clusterNamespace = queryParams['namespace'] || 'default';
      this.loadClusterData();
    });
  }

  ngAfterViewInit(): void {
    // 延迟初始化图表，确保DOM已渲染
    setTimeout(() => {
      this.initCharts();
    }, 100);
  }

  private loadClusterData(): void {
    // 设置超时清除loading状态，防止API卡住
    setTimeout(() => {
      this.loadingService.clearAll();
    }, 10000); // 10秒后强制清除
    
    this.apiService.getCluster(this.clusterNamespace, this.clusterName).subscribe({
      next: (cluster) => {
        this.cluster = cluster;
      },
      error: (err) => {
        this.snackBar.open(`Error loading cluster details: ${err.message}`, 'Close', { duration: 5000 });
      }
    });
    
    this.apiService.getPodsForCluster(this.clusterNamespace, this.clusterName).subscribe({
      next: (pods) => {
        this.cnCount = 0;
        this.dnCount = 0;
        this.gmsCount = 0;
        this.cdcCount = 0;
        const nodes: NodeInfo[] = pods.map(pod => {
          const nodeInfo = this.transformPodToNodeInfo(pod);
          switch (nodeInfo.type) {
            case 'CN':
              this.cnCount++;
              break;
            case 'DN':
              this.dnCount++;
              break;
            case 'GMS':
              this.gmsCount++;
              break;
            case 'CDC':
              this.cdcCount++;
              break;
          }
          return nodeInfo;
        });
        this.nodeDataSource.data = nodes;
        this.allPods = pods;
        // 计算拓扑实时摘要（按 Pod 名含 -cn-/-dn-/-gms-/-cdc- 分组）
        this.roleSummary = { cn: { total: 0, ready: 0, pods: [] }, dn: { total: 0, ready: 0, pods: [] }, gms: { total: 0, ready: 0, pods: [] }, cdc: { total: 0, ready: 0, pods: [] } };
        for (const p of pods) {
          const name = p.metadata?.name || '';
          let role: 'cn'|'dn'|'gms'|'cdc'|'' = '';
          if (name.includes('-cn-')) role = 'cn';
          else if (name.includes('-dn-')) role = 'dn';
          else if (name.includes('-gms-')) role = 'gms';
          else if (name.includes('-cdc-')) role = 'cdc';
          if (role) {
            this.roleSummary[role].pods.push(p);
            this.roleSummary[role].total++;
            const phase = (p.status?.phase || '').toLowerCase();
            if (phase === 'running') this.roleSummary[role].ready++;
          }
        }
        // 重新渲染拓扑，确保 DN → CDC 连线与计数更新
        this.renderTopology();
        if (this.allPods.length > 0) {
          this.selectedPod = this.allPods[0].metadata.name;
          this.onPodSelectionChange();
        }
      },
      error: (err) => {
        this.snackBar.open(`Error loading pods: ${err.message}`, 'Close', { duration: 5000 });
      }
    });

    this.loadBackupData();
    this.loadLogData();

    // 加载告警汇总（可通过 localStorage.alertmanager 设置地址）
    const am = localStorage.getItem('alertmanager') || '';
    this.apiService.getClusterAlertsSummary(this.clusterNamespace, this.clusterName, am || undefined).subscribe({
      next: (s) => this.alerts = s,
      error: () => this.alerts = { critical: 0, warning: 0, info: 0, total: 0, source: 'none' }
    });
  }

  getUptime(): string {
    const start = (this.cluster as any)?.status?.startTime || this.cluster?.metadata?.creationTimestamp;
    if (!start) return '未知';
    const startTime = new Date(start).getTime();
    const now = Date.now();
    let diff = Math.max(0, now - startTime);
    const days = Math.floor(diff / (24 * 3600 * 1000)); diff -= days * 24 * 3600 * 1000;
    const hours = Math.floor(diff / (3600 * 1000)); diff -= hours * 3600 * 1000;
    const mins = Math.floor(diff / (60 * 1000));
    if (days > 0) return `${days}天${hours}小时`;
    if (hours > 0) return `${hours}小时${mins}分钟`;
    return `${mins}分钟`;
  }

  goToBackupsTab(): void {
    // 切换到“备份与恢复”选项卡（索引根据模板位置可能为 3）
    this.selectedTabIndex = 3;
  }

  getAlertCount(): number {
    return this.alerts.total || 0;
  }

  getAvailabilityPercent(): number {
    // 预估可用性：就绪 Pod / 总 Pod
    const total = this.allPods?.length || 0;
    if (total === 0) return 0;
    const ready = this.allPods.filter(p => (p.status?.phase || '').toLowerCase() === 'running').length;
    return Math.round((ready / total) * 100);
  }

  getAvailabilityColor(): 'primary' | 'accent' | 'warn' {
    const v = this.getAvailabilityPercent();
    if (v > 99) return 'primary';
    if (v >= 95) return 'accent';
    return 'warn';
  }

  copyExecCommand(node: NodeInfo): void {
    const pod = node.name;
    const container = (this.selectedPodContainers?.[0]?.name) || '';
    const ns = this.clusterNamespace || 'default';
    const command = `kubectl exec -it ${pod} -n ${ns} ${container ? '-c ' + container + ' ' : ''}-- /bin/bash`;
    navigator.clipboard.writeText(command).then(() => {
      this.snackBar.open('已复制 kubectl exec 命令', '关闭', { duration: 2000 });
    }).catch(() => {
      this.snackBar.open('复制失败，请手动复制', '关闭', { duration: 2000 });
    });
  }

  openExecDialog(node: NodeInfo): void {
    import('../../components/exec-command-dialog/exec-command-dialog.component').then(m => {
      const containers = this.allPods.find(p => p.metadata.name === node.name)?.spec?.containers?.map(c => c.name) || [];
      const ref = this.dialog.open(m.ExecCommandDialogComponent, {
        width: '700px',
        data: {
          namespace: this.clusterNamespace || 'default',
          pod: node.name,
          containers,
          defaultContainer: containers[0] || ''
        }
      });
      ref.afterClosed().subscribe((result) => {
        if (result && result.runOnce) {
          this.snackBar.open('命令已执行', '关闭', { duration: 2000 });
        }
      });
    });
  }

  openWebShell(node: NodeInfo): void {
    import('../../components/webshell-dialog/webshell-dialog.component').then(m => {
      const containers = this.getContainersOf(node);
      const prefer = this.getPreferredContainers(containers)[0] || '';
      const chosen = this.rowContainer[node.name] || prefer;
      this.dialog.open(m.WebShellDialogComponent, {
        width: '900px',
        height: '600px',
        data: {
          namespace: this.clusterNamespace || 'default',
          pod: node.name,
          container: chosen,
          containers: containers
        }
      });
    });
  }

  openPodDetail(node: NodeInfo): void {
    const podName = node.name;
    const ns = this.clusterNamespace || 'default';
    this.router.navigate(['/operations','nodes', ns, podName]);
  }

  getContainersOf(node: NodeInfo): string[] {
    return this.allPods.find(p => p.metadata.name === node.name)?.spec?.containers?.map(c => c.name) || [];
  }
  getPreferredContainers(containers: string[]): string[] {
    const items = (containers || []).slice();
    const lower = (s: string) => (s||'').toLowerCase();
    const negatives = ['prober','probe','exporter','sidecar','pause','proxy','metrics','reloader','kube-rbac-proxy','configmap-reload'];
    const positives = ['engine','mysql','xstore','server','main','app','dn','cn','gms','cdc'];
    const preferred = items.filter(c => positives.some(p => lower(c) === p || lower(c).includes(p)));
    const others = items.filter(c => !preferred.includes(c) && !negatives.some(n => lower(c).includes(n)));
    const sidecars = items.filter(c => negatives.some(n => lower(c).includes(n)));
    const dedup = (arr: string[]) => Array.from(new Set(arr));
    return dedup([...preferred, ...others, ...sidecars]);
  }
  setRowContainer(podName: string, container: string): void {
    this.rowContainer[podName] = container;
  }
  openWebShellWith(node: NodeInfo, container: string): void {
    this.rowContainer[node.name] = container;
    this.openWebShell(node);
  }

  private getBackupPreferences(): { storageName: string; sink: string; retentionTime: string } {
    // 允许通过本地存储覆盖，便于在不同环境（如 lyfz 或默认 minio）之间切换
    const storageName = (localStorage.getItem('backupStorageName') || 's3').trim();
    const sink = (localStorage.getItem('backupSink') || 'default').trim();
    // 默认 240 小时（10 天），符合 Go 的 time.Duration 格式
    const retentionTime = (localStorage.getItem('backupRetentionTime') || '240h').trim();
    return { storageName, sink, retentionTime };
  }

  private transformPodToNodeInfo(pod: Pod): NodeInfo {
    // 使用智能角色检测器
    const roleInfo = this.detectPodRole(pod);
    
    // Placeholder for resource requests
    const cpu = pod.spec.containers?.[0]?.resources?.requests?.cpu || 'N/A';
    const memory = pod.spec.containers?.[0]?.resources?.requests?.memory || 'N/A';

    // 解析容器状态
    const containerStatuses: ContainerStatus[] = [];
    let readyContainers = 0;
    let totalRestarts = 0;
    let hasUnhealthyContainers = false;

    if (pod.status?.containerStatuses) {
      for (const cs of pod.status.containerStatuses) {
        const containerStatus: ContainerStatus = {
          name: cs.name,
          ready: cs.ready || false,
          restartCount: cs.restartCount || 0,
          state: this.getContainerState(cs),
          reason: this.getContainerReason(cs),
          message: this.getContainerMessage(cs)
        };
        
        containerStatuses.push(containerStatus);
        
        if (cs.ready) readyContainers++;
        totalRestarts += cs.restartCount || 0;
        
        // 检查是否不健康（未就绪或重启次数过多）
        if (!cs.ready || (cs.restartCount && cs.restartCount > 3)) {
          hasUnhealthyContainers = true;
        }
      }
    }

    const totalContainers = pod.spec?.containers?.length || 0;

    return {
      name: pod.metadata.name,
      type: roleInfo.role,
      status: pod.status?.phase || 'Unknown',
      cpu: cpu,
      memory: memory,
      ip: pod.status?.podIP || 'Unknown',
      resources: `${cpu}/${memory}`,
      readyContainers,
      totalContainers,
      totalRestarts,
      hasUnhealthyContainers,
      containerStatuses
    };
  }

  // 获取容器状态
  private getContainerState(containerStatus: any): string {
    if (containerStatus.state?.running) return 'Running';
    if (containerStatus.state?.waiting) return 'Waiting';
    if (containerStatus.state?.terminated) return 'Terminated';
    return 'Unknown';
  }

  // 获取容器状态原因
  private getContainerReason(containerStatus: any): string {
    if (containerStatus.state?.waiting?.reason) return containerStatus.state.waiting.reason;
    if (containerStatus.state?.terminated?.reason) return containerStatus.state.terminated.reason;
    return '';
  }

  // 获取容器状态消息
  private getContainerMessage(containerStatus: any): string {
    if (containerStatus.state?.waiting?.message) return containerStatus.state.waiting.message;
    if (containerStatus.state?.terminated?.message) return containerStatus.state.terminated.message;
    return '';
  }

  // 临时角色检测方法（简化版）
  private detectPodRole(pod: Pod): { role: string; category: string } {
    if (!pod || !pod.metadata) {
      return { role: 'Unknown', category: 'unknown' };
    }

    const name = pod.metadata.name || '';
    const labels = pod.metadata.labels || {};

    // 从标签获取角色
    if (labels['polardbx/role']) {
      return { role: labels['polardbx/role'].toUpperCase(), category: 'compute' };
    }

    // 从名称模式推断
    if (name.includes('-cn-')) return { role: 'CN', category: 'compute' };
    if (name.includes('-dn-')) return { role: 'DN', category: 'storage' };
    if (name.includes('-gms-')) return { role: 'GMS', category: 'service' };
    if (name.includes('-cdc-')) return { role: 'CDC', category: 'service' };
    if (name.includes('-columnar-')) return { role: 'Columnar', category: 'storage' };
    if (name.toLowerCase().includes('minio')) return { role: 'MinIO', category: 'storage' };
    if (name.toLowerCase().includes('sftp')) return { role: 'SFTP', category: 'service' };
    if (name.toLowerCase().includes('hpfs')) return { role: 'HPFS', category: 'service' };

    return { role: 'Unknown', category: 'unknown' };
  }

  private loadBackupData(): void {
    // 如果集群信息还没加载，使用路由参数
    const namespace = this.cluster?.metadata.namespace || this.clusterNamespace;
    const clusterName = this.cluster?.metadata.name || this.clusterName;

    if (!namespace || !clusterName) {
      console.warn('无法加载备份数据：缺少集群信息');
      return;
    }

    this.apiService.getBackups(namespace, clusterName)
      .subscribe({
        next: (backups: PolarDBXBackup[]) => {
          console.log('获取到备份列表:', backups);
          const backupInfos: BackupInfo[] = backups.map(backup => ({
            id: backup.metadata.name,
            name: backup.metadata.name,
            namespace: backup.metadata.namespace,
            completedTime: backup.status?.completionTime || backup.metadata.creationTimestamp,
            type: backup.spec.backupType || 'Snapshot',
            status: backup.status?.phase || 'Unknown',
            phase: backup.status?.phase,
            message: backup.status?.message,
            backupObject: backup
          }));
          this.backupDataSource.data = backupInfos;
          console.log('备份数据已设置到表格:', backupInfos);
        },
        error: (error) => {
          console.error('获取备份列表失败:', error);
          this.snackBar.open('获取备份列表失败', '关闭', { duration: 3000 });
        }
      });
  }

  private loadLogData(): void {
    // 模拟日志数据
    this.logContent = `[2024-12-01 15:30:01] INFO: PolarDB-X cluster ${this.clusterName} is running normally
[2024-12-01 15:29:58] INFO: All nodes are healthy
[2024-12-01 15:29:55] INFO: Connection pool status: 50/100 active connections
[2024-12-01 15:29:52] INFO: Memory usage: 65% of allocated memory
[2024-12-01 15:29:49] INFO: CPU usage: 45% average across all nodes
[2024-12-01 15:29:46] INFO: Storage usage: 2.5GB / 100GB (2.5%)
[2024-12-01 15:29:43] INFO: Backup job completed successfully
[2024-12-01 15:29:40] INFO: Replication lag: 0.2ms average`;
  }

  private initCharts(): void {
    this.initCpuChart();
    this.initMemChart();
    this.renderTopology();
  }

  private initCpuChart(): void {
    if (!this.cpuChartRef?.nativeElement) return;

    const ctx = this.cpuChartRef.nativeElement.getContext('2d');
    if (!ctx) return;

    // 生成模拟数据
    const labels = [];
    const data = [];
    const now = new Date();
    
    for (let i = 29; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60000); // 每分钟一个数据点
      labels.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }));
      data.push(Math.random() * 30 + 20); // 20-50% 的随机CPU使用率
    }

    const config: ChartConfiguration = {
      type: 'line' as ChartType,
      data: {
        labels: labels,
        datasets: [{
          label: 'CPU 使用率 (%)',
          data: data,
          borderColor: '#1976d2',
          backgroundColor: 'rgba(25, 118, 210, 0.1)',
          fill: true,
          tension: 0.4
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: {
            beginAtZero: true,
            max: 100,
            ticks: {
              callback: function(value: any) {
                return value + '%';
              }
            }
          }
        },
        plugins: {
          legend: {
            display: false
          }
        }
      }
    };

    this.cpuChart = new Chart(ctx, config);
  }

  private initMemChart(): void {
    if (!this.memChartRef?.nativeElement) return;

    const ctx = this.memChartRef.nativeElement.getContext('2d');
    if (!ctx) return;

    // 生成模拟数据
    const labels = [];
    const data = [];
    const now = new Date();
    
    for (let i = 29; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60000);
      labels.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }));
      data.push(Math.random() * 5 + 10); // 10-15 GiB 的随机内存使用
    }

    const config: ChartConfiguration = {
      type: 'line' as ChartType,
      data: {
        labels: labels,
        datasets: [{
          label: '内存使用 (GiB)',
          data: data,
          borderColor: '#4caf50',
          backgroundColor: 'rgba(76, 175, 80, 0.1)',
          fill: true,
          tension: 0.4
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: {
            beginAtZero: true,
            ticks: {
              callback: function(value: any) {
                return value + ' GiB';
              }
            }
          }
        },
        plugins: {
          legend: {
            display: false
          }
        }
      }
    };

    this.memChart = new Chart(ctx, config);
  }

  private renderTopology(): void {
    if (!this.topologySvgRef) return;
    const svg = this.topologySvgRef.nativeElement;
    // 清空
    while (svg.firstChild) svg.removeChild(svg.firstChild);
    // 定义箭头样式
    const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs');
    const marker = document.createElementNS('http://www.w3.org/2000/svg', 'marker');
    marker.setAttribute('id', 'arrow');
    marker.setAttribute('markerWidth', '10');
    marker.setAttribute('markerHeight', '7');
    marker.setAttribute('refX', '10');
    marker.setAttribute('refY', '3.5');
    marker.setAttribute('orient', 'auto');
    const arrowPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    arrowPath.setAttribute('d', 'M0,0 L10,3.5 L0,7 Z');
    arrowPath.setAttribute('fill', '#9aa6b2');
    marker.appendChild(arrowPath);
    defs.appendChild(marker);
    svg.appendChild(defs);
    const g = new dagre.graphlib.Graph();
    g.setGraph({ rankdir: 'LR', nodesep: 20, ranksep: 40 });
    g.setDefaultEdgeLabel(() => ({}));

    // 定义节点 (只显示有实际Pod的节点)
    const roles: Array<{ id: 'cn'|'dn'|'gms'|'cdc'; label: string; count: number }> = [
      { id: 'cn', label: `CN (${this.roleSummary['cn'].ready}/${this.roleSummary['cn'].total})`, count: this.roleSummary['cn'].total },
      { id: 'dn', label: `DN (${this.roleSummary['dn'].ready}/${this.roleSummary['dn'].total})`, count: this.roleSummary['dn'].total },
      { id: 'gms', label: `GMS (${this.roleSummary['gms'].ready}/${this.roleSummary['gms'].total})`, count: this.roleSummary['gms'].total }
    ];
    
    // 只有当CDC有实际Pod时才显示
    if (this.roleSummary['cdc'].total > 0) {
      roles.push({ id: 'cdc', label: `CDC (${this.roleSummary['cdc'].ready}/${this.roleSummary['cdc'].total})`, count: this.roleSummary['cdc'].total });
    }
    
    for (const r of roles) {
      const width = 90 + r.label.length * 6;
      const height = 36;
      g.setNode(r.id, { label: r.label, width, height });
    }
    // 简单连线：CN->DN->GMS 以及 DN->CDC（如有）
    g.setEdge('cn', 'dn');
    g.setEdge('dn', 'gms');
    if (this.roleSummary['cdc'].total > 0) {
      g.setEdge('dn', 'cdc');
    }

    dagre.layout(g);

    // 画节点（带健康度颜色 + 悬浮提示 Pod 列表）
    for (const v of g.nodes()) {
      const n = g.node(v) as any;
      const rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
      rect.setAttribute('x', String(n.x - n.width / 2));
      rect.setAttribute('y', String(n.y - n.height / 2));
      rect.setAttribute('width', String(n.width));
      rect.setAttribute('height', String(n.height));
      rect.setAttribute('rx', '8');
      const health = this.getRoleHealth(v as 'cn'|'dn'|'gms'|'cdc');
      const fill = health >= 100 ? '#eef7ff' : (health >= 95 ? '#fff8e1' : '#ffeaea');
      const stroke = health >= 100 ? '#2f5d8a' : (health >= 95 ? '#e6a23c' : '#f56c6c');
      rect.setAttribute('fill', fill);
      rect.setAttribute('stroke', stroke);
      rect.setAttribute('stroke-width', '1');
      svg.appendChild(rect);

      const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      text.setAttribute('x', String(n.x));
      text.setAttribute('y', String(n.y + 4));
      text.setAttribute('text-anchor', 'middle');
      text.setAttribute('font-size', '12');
      text.textContent = n.label;
      svg.appendChild(text);

      // 内置 tooltip（原生 title）列出 Pod
      const title = document.createElementNS('http://www.w3.org/2000/svg', 'title');
      const pods = this.roleSummary[v as 'cn'|'dn'|'gms'|'cdc'].pods
        .map(p => `${p.metadata?.name} [${p.status?.phase}]`).join('\n');
      title.textContent = pods || '无 Pod';
      rect.appendChild(title);
      text.appendChild(title.cloneNode(true));
    }
    // 画边
    for (const e of g.edges()) {
      const edge = g.edge(e) as any;
      const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
      const points = edge.points as Array<{ x: number; y: number }>;
      const d = points.map((p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`)).join(' ');
      path.setAttribute('d', d);
      path.setAttribute('fill', 'none');
      // 边颜色：取下游角色健康度
      const target = (e.w as 'cn'|'dn'|'gms');
      const health = this.getRoleHealth(target);
      const stroke = health >= 100 ? '#2f5d8a' : (health >= 95 ? '#e6a23c' : '#f56c6c');
      path.setAttribute('stroke', stroke);
      path.setAttribute('stroke-width', '1.5');
      path.setAttribute('marker-end', 'url(#arrow)');
      // 边 tooltip（占位，未来可填充流量/延迟）
      const t = document.createElementNS('http://www.w3.org/2000/svg', 'title');
      t.textContent = `链路 ${e.v} → ${e.w}`;
      path.appendChild(t);
      svg.appendChild(path);
    }
  }

  private getRoleHealth(role: 'cn'|'dn'|'gms'|'cdc'): number {
    const s = this.roleSummary[role];
    if (!s || s.total === 0) return 0;
    return Math.round((s.ready / s.total) * 100);
  }

  // 事件处理方法
  deleteCluster(): void {
    if (confirm(`确定要删除集群 ${this.clusterName} 吗？此操作不可撤销。`)) {
      // 预检查
      this.apiService.getPrechangeChecklist(this.clusterNamespace, this.clusterName).subscribe({
        next: () => {
          // 通过后再删除
          this.apiService.deleteCluster(this.clusterNamespace, this.clusterName).subscribe({
            next: () => {
              this.snackBar.open('集群删除成功', '关闭', { duration: 3000 });
              this.router.navigate(['/clusters']);
            },
            error: (err) => {
              this.snackBar.open(`Error deleting cluster: ${err.message}`, 'Close', { duration: 5000 });
            }
          });
        },
        error: (err) => {
          const checks = err?.error?.checks || {};
          const msg = `前置检查未通过：\n` +
            `- 最近全备：${checks.hasRecentBackup ? '是' : '否'}\n` +
            `- 存储连通：${checks.storage || '未知'}\n` +
            `- RPO滞后(秒)：${typeof checks.rpoLagSeconds==='number'?checks.rpoLagSeconds:'未知'}`;
          this.snackBar.open(msg, '关闭', { duration: 6000 });
        }
      });
    }
  }

  applyConfig(): void {
    if (!this.configForm.valid || !this.cluster) {
      this.snackBar.open('请检查配置信息', '关闭', { duration: 3000 });
      return;
    }

    const config = this.configForm.value;
    const scalingRequest: any = {};

    // 检查哪些副本数发生了变化
    if (config.cnReplicas !== this.cnReplicas) {
      scalingRequest.cnReplicas = config.cnReplicas;
    }
    if (config.dnReplicas !== this.dnReplicas) {
      scalingRequest.dnReplicas = config.dnReplicas;
    }

    // 如果没有变化，不需要发送请求
    if (Object.keys(scalingRequest).length === 0) {
      this.snackBar.open('配置没有变化', '关闭', { duration: 3000 });
      return;
    }

    // 确认对话框
    const changeList = [];
    if (scalingRequest.cnReplicas !== undefined) {
      changeList.push(`CN 节点: ${this.cnReplicas} → ${scalingRequest.cnReplicas}`);
    }
    if (scalingRequest.dnReplicas !== undefined) {
      changeList.push(`DN 节点: ${this.dnReplicas} → ${scalingRequest.dnReplicas}`);
    }

    if (confirm(`确定要应用以下配置变更吗？\n${changeList.join('\n')}`)) {
      // 预检查
      this.apiService.getPrechangeChecklist(this.clusterNamespace, this.clusterName)
        .subscribe({
          next: () => {
            this.apiService.scaleCluster(this.clusterNamespace, this.clusterName, scalingRequest)
        .subscribe({
          next: (response) => {
            this.snackBar.open('集群扩缩容任务已启动', '关闭', { 
              duration: 5000,
              panelClass: ['success-snackbar']
            });
            // 更新本地数据
            if (scalingRequest.cnReplicas !== undefined) {
              this.cnReplicas = scalingRequest.cnReplicas;
            }
            if (scalingRequest.dnReplicas !== undefined) {
              this.dnReplicas = scalingRequest.dnReplicas;
            }
            // 刷新集群数据
            setTimeout(() => {
              this.loadClusterData();
            }, 2000);
          },
          error: (error) => {
            console.error('扩缩容失败:', error);
            this.snackBar.open(`扩缩容失败: ${error.message || '未知错误'}`, '关闭', { 
              duration: 5000,
              panelClass: ['error-snackbar']
            });
          }
        });
          },
          error: (err) => {
            const checks = err?.error?.checks || {};
            const msg = `前置检查未通过：\n` +
              `- 最近全备：${checks.hasRecentBackup ? '是' : '否'}\n` +
              `- 存储连通：${checks.storage || '未知'}\n` +
              `- RPO滞后(秒)：${typeof checks.rpoLagSeconds==='number'?checks.rpoLagSeconds:'未知'}`;
            this.snackBar.open(msg, '关闭', { duration: 6000 });
          }
        });
    }
  }

  resetConfig(): void {
    // 重置表单到当前集群状态
    this.configForm.patchValue({
      cnReplicas: this.cnReplicas,
      cnCpu: '1',
      cnMemory: '2',
      dnReplicas: this.dnReplicas,
      dnCpu: '1',
      dnMemory: '2'
    });
    this.snackBar.open('配置已重置', '关闭', { duration: 2000 });
  }

  resetUpgrade(): void {
    // 重置升级表单
    this.upgradeForm.patchValue({
      targetVersion: this.clusterVersion
    });
    this.snackBar.open('升级配置已重置', '关闭', { duration: 2000 });
  }

  // 检查是否有配置变更
  hasConfigChanges(): boolean {
    const config = this.configForm.value;
    return config.cnReplicas !== this.cnReplicas || 
           config.dnReplicas !== this.dnReplicas;
  }

  // 预估成本变化
  previewCost(): void {
    const config = this.configForm.value;
    const currentTotal = this.cnReplicas + this.dnReplicas + 1; // +1 for GMS
    const newTotal = config.cnReplicas + config.dnReplicas + 1;
    const costChange = ((newTotal - currentTotal) / currentTotal * 100).toFixed(1);
    
    const message = newTotal > currentTotal 
      ? `扩容后成本预计增加 ${costChange}%`
      : newTotal < currentTotal 
        ? `缩容后成本预计减少 ${Math.abs(parseFloat(costChange))}%`
        : '配置无变化，成本不变';
    
    this.snackBar.open(message, '关闭', { duration: 5000 });
  }

  // 节点管理功能
  restartNode(node: NodeInfo): void {
    if (confirm(`确定要重启节点 ${node.name} 吗？这会短暂中断该节点的服务。`)) {
      this.apiService.deletePod(this.clusterNamespace, node.name).subscribe({
        next: () => {
          this.snackBar.open(`节点 ${node.name} 重启中...`, '关闭', { duration: 3000 });
          setTimeout(() => this.loadClusterData(), 2000);
        },
        error: (error) => {
          this.snackBar.open(`重启节点失败: ${error.message}`, '关闭', { duration: 5000 });
        }
      });
    }
  }

  deleteAndRecreateNode(node: NodeInfo): void {
    const warningMessage = node.type === 'GMS' 
      ? `⚠️ 警告：GMS节点是集群的管理核心，删除重建可能导致集群短时间不可用。\n\n确定要删除重建 ${node.name} 吗？`
      : `确定要删除重建节点 ${node.name} 吗？该节点将被重新创建。`;
      
    if (confirm(warningMessage)) {
      this.snackBar.open(`正在删除重建节点 ${node.name}...`, '关闭', { duration: 3000 });
      
      this.apiService.deletePod(this.clusterNamespace, node.name).subscribe({
        next: () => {
          this.snackBar.open(`节点 ${node.name} 已删除，正在重建...`, '关闭', { duration: 5000 });
          // 延迟刷新以显示重建过程
          setTimeout(() => this.loadClusterData(), 3000);
          setTimeout(() => this.loadClusterData(), 10000);
        },
        error: (error) => {
          this.snackBar.open(`删除节点失败: ${error.message}`, '关闭', { duration: 5000 });
        }
      });
    }
  }

  isolateNode(node: NodeInfo): void {
    if (node.type === 'GMS') {
      this.snackBar.open('GMS节点不能被隔离', '关闭', { duration: 3000 });
      return;
    }
    
    if (confirm(`确定要隔离节点 ${node.name} 吗？隔离后该节点将不再处理新的请求。`)) {
      // 这里应该调用相应的API来设置节点为隔离状态
      // 具体实现取决于PolarDB-X的隔离机制
      this.snackBar.open(`节点隔离功能正在开发中...`, '关闭', { duration: 3000 });
    }
  }

  // 打开告警详情对话框
  openAlertsDialog(): void {
    if (this.getAlertCount() === 0) {
      this.snackBar.open('当前无告警信息', '关闭', { duration: 2000 });
      return;
    }

    import('../../components/alerts-detail-dialog/alerts-detail-dialog.component').then(m => {
      const ref = this.dialog.open(m.AlertsDetailDialogComponent, {
        width: '80vw',
        maxWidth: '1200px',
        height: '70vh',
        data: {
          cluster: this.cluster,
          alerts: this.alerts,
          namespace: this.clusterNamespace,
          clusterName: this.clusterName
        }
      });

      ref.afterClosed().subscribe(result => {
        if (result?.refreshAlerts) {
          // 重新加载告警数据
          const am = localStorage.getItem('alertmanager') || '';
          this.apiService.getClusterAlertsSummary(this.clusterNamespace, this.clusterName, am || undefined).subscribe({
            next: (s) => this.alerts = s,
            error: () => this.alerts = { critical: 0, warning: 0, info: 0, total: 0, source: 'none' }
          });
        }
      });
    }).catch(error => {
      console.error('Failed to load alerts dialog:', error);
      this.snackBar.open('告警详情功能正在开发中...', '关闭', { duration: 3000 });
    });
  }

  startUpgrade(): void {
    if (!this.upgradeForm.valid || !this.cluster) {
      this.snackBar.open('请检查升级配置', '关闭', { duration: 3000 });
      return;
    }

    const targetVersion = this.upgradeForm.value.targetVersion;
    const currentVersion = this.clusterVersion;

    // 版本比较
    if (targetVersion === currentVersion) {
      this.snackBar.open('目标版本与当前版本相同', '关闭', { duration: 3000 });
      return;
    }

    const upgradeRequest = {
      targetVersion: targetVersion,
      strategy: 'rolling', // 默认使用滚动升级
      maxUnavailable: 1
    };

    const confirmMessage = `确定要将集群升级吗？\n\n` +
                          `当前版本: ${currentVersion}\n` +
                          `目标版本: ${targetVersion}\n` +
                          `升级策略: 滚动升级\n\n` +
                          `⚠️ 升级过程中可能会有短暂的服务中断`;

    if (confirm(confirmMessage)) {
      this.apiService.getPrechangeChecklist(this.clusterNamespace, this.clusterName)
        .subscribe({
          next: () => {
            this.apiService.upgradeCluster(this.clusterNamespace, this.clusterName, upgradeRequest)
        .subscribe({
          next: (response) => {
            this.snackBar.open('集群升级任务已启动，请关注升级进度', '关闭', { 
              duration: 5000,
              panelClass: ['success-snackbar']
            });
            
            // 更新本地显示版本
            this.clusterVersion = targetVersion;
            
            // 刷新集群状态
            setTimeout(() => {
              this.loadClusterData();
            }, 2000);
          },
          error: (error) => {
            console.error('升级失败:', error);
            this.snackBar.open(`升级失败: ${error.message || '未知错误'}`, '关闭', { 
              duration: 5000,
              panelClass: ['error-snackbar']
            });
          }
        });
          },
          error: (err) => {
            const checks = err?.error?.checks || {};
            const msg = `前置检查未通过：\n` +
              `- 最近全备：${checks.hasRecentBackup ? '是' : '否'}\n` +
              `- 存储连通：${checks.storage || '未知'}\n` +
              `- RPO滞后(秒)：${typeof checks.rpoLagSeconds==='number'?checks.rpoLagSeconds:'未知'}`;
            this.snackBar.open(msg, '关闭', { duration: 6000 });
          }
        });
    }
  }

  createBackup(): void {
    if (!this.cluster) return;

    // 生成符合Kubernetes规范的备份名称 (最大63字符，小写字母数字和连字符)
    const now = new Date();
    const shortTimestamp = now.getFullYear().toString().slice(-2) + 
                          (now.getMonth() + 1).toString().padStart(2, '0') +
                          now.getDate().toString().padStart(2, '0') +
                          now.getHours().toString().padStart(2, '0') +
                          now.getMinutes().toString().padStart(2, '0');
    
    // 确保名称不超过63字符且符合DNS标签规范
    const clusterName = this.cluster.metadata.name;
    const baseName = `${clusterName}-bak-${shortTimestamp}`;
    
    // 如果名称过长，截断集群名称部分
    const backupName = baseName.length > 63 ? 
                       `${clusterName.slice(0, 63 - 16)}-bak-${shortTimestamp}` : 
                       baseName;

    // 偏好设置：存储类型/ Sink / 保留时长
    const prefs = this.getBackupPreferences();

    // 构造完整的PolarDBXBackup对象（符合Kubernetes资源格式）
    const backupObject: any = {
      apiVersion: "polardbx.aliyun.com/v1",
      kind: "PolarDBXBackup",
      metadata: {
        name: backupName,
        namespace: this.cluster.metadata.namespace
      },
      spec: {
        cluster: {
          name: this.cluster.metadata.name
        },
        retentionTime: prefs.retentionTime,
        storageProvider: {
          storageName: prefs.storageName,
          sink: prefs.sink
        },
        preferredBackupRole: "follower"
      }
    };

    this.apiService.createBackup(this.cluster.metadata.namespace, this.cluster.metadata.name, backupObject)
      .subscribe({
        next: (backup: PolarDBXBackup) => {
          console.log('备份创建成功:', backup);
          this.snackBar.open(`备份 ${backup.metadata.name} 创建成功`, '关闭', { duration: 3000 });
          this.loadBackupData(); // 重新加载备份数据
        },
        error: (error) => {
          console.error('创建备份失败:', error);
          this.snackBar.open('创建备份失败', '关闭', { duration: 3000 });
        }
      });
  }

  restoreBackup(backup: BackupInfo): void {
    if (confirm(`确定要从备份 ${backup.id} 恢复集群吗？此操作将覆盖当前数据。`)) {
      // 模拟恢复备份
      this.snackBar.open('恢复任务已启动', '关闭', { duration: 3000 });
    }
  }

  deleteBackup(backup: BackupInfo): void {
    if (confirm(`确定要删除备份 ${backup.id} 吗？`)) {
      this.apiService.deleteBackup(backup.namespace, backup.name)
        .subscribe({
          next: () => {
            console.log('备份删除成功:', backup.name);
            this.snackBar.open(`备份 ${backup.name} 删除成功`, '关闭', { duration: 3000 });
            this.loadBackupData(); // 重新加载备份数据
          },
          error: (error) => {
            console.error('删除备份失败:', error);
            this.snackBar.open('删除备份失败', '关闭', { duration: 3000 });
          }
        });
    }
  }

  forceDeleteBackup(backup: BackupInfo): void {
    if (confirm(`强制删除将直接移除 finalizers 并清理对象。确定对备份 ${backup.name} 执行吗？`)) {
      this.apiService.forceDeleteBackup(backup.namespace, backup.name)
        .subscribe({
          next: () => {
            this.snackBar.open(`备份 ${backup.name} 强制删除成功`, '关闭', { duration: 3000 });
            this.loadBackupData();
          },
          error: (error) => {
            console.error('强制删除备份失败:', error);
            this.snackBar.open('强制删除备份失败', '关闭', { duration: 3000 });
          }
        });
    }
  }

  onPodSelectionChange(): void {
    const selectedPod = this.allPods.find(p => p.metadata.name === this.selectedPod);
    if (selectedPod && selectedPod.spec.containers) {
      this.selectedPodContainers = selectedPod.spec.containers;
      if (this.selectedPodContainers.length > 0) {
        this.selectedContainer = this.selectedPodContainers[0].name;
      } else {
        this.selectedContainer = '';
      }
    } else {
      this.selectedPodContainers = [];
      this.selectedContainer = '';
    }
  }

  fetchLogs(): void {
    if (!this.selectedPod || !this.selectedContainer) {
      this.snackBar.open('请先选择一个 Pod 和 Container', '关闭', { duration: 3000 });
      return;
    }
    this.isLoadingLogs = true;
    this.logContent = '正在加载日志...';
    this.apiService.getPodLogs(this.clusterNamespace, this.selectedPod, this.selectedContainer).subscribe({
      next: (logs) => {
        this.logContent = logs || '该 Pod 没有可显示的日志。';
        this.isLoadingLogs = false;
      },
      error: (err) => {
        this.logContent = `加载日志失败: ${err.error?.error || err.message}`;
        this.isLoadingLogs = false;
        this.snackBar.open(`加载日志失败: ${err.message}`, '关闭', {
          duration: 5000,
        });
      }
    });
  }

  // 辅助方法
  getNodeStatusColor(status: string): string {
    switch (status.toLowerCase()) {
      case 'running':
        return 'primary';
      case 'pending':
        return 'accent';
      case 'failed':
        return 'warn';
      default:
        return '';
    }
  }

  getBackupStatusColor(status: string): string {
    switch (status?.toLowerCase()) {
      case 'completed':
      case '完成':
        return 'primary';
      case 'running':
      case 'pending':
      case '进行中':
        return 'accent';
      case 'failed':
      case '失败':
        return 'warn';
      default:
        return '';
    }
  }

  ngOnDestroy(): void {
    // 清除所有加载状态以防内存泄漏
    this.loadingService.clearAll();
    
    // 清理图表实例
    if (this.cpuChart) {
      this.cpuChart.destroy();
    }
    if (this.memChart) {
      this.memChart.destroy();
    }
  }

  refreshData() {
    // 重新加载集群数据
    this.loadClusterData();
    this.loadBackupData();
    this.loadLogData();
  }

  // 处理选项卡切换
  onTabChange(index: number): void {
    this.selectedTabIndex = index;
    if (index === 2) { // 备份与恢复选项卡的索引（从0开始）
      // 确保备份数据已加载
      setTimeout(() => {
        this.loadBackupData();
      }, 100);
    }
  }

  // 获取节点类型颜色
  getNodeTypeColor(type: string): string {
    switch (type) {
      case 'CN':
        return 'primary';
      case 'DN':
        return 'accent';
      case 'GMS':
        return 'warn';
      case 'CDC':
        return 'basic';
      default:
        return 'basic';
    }
  }

  // 获取状态颜色
  getStatusColor(status: string): string {
    switch (status) {
      case 'Running':
      case 'Ready':
      case 'Completed':
        return 'primary';
      case 'Pending':
      case 'Creating':
        return 'accent';
      case 'Failed':
      case 'Error':
        return 'warn';
      default:
        return 'basic';
    }
  }

  // 获取Ready状态颜色
  getReadyColor(ready: number = 0, total: number = 0): string {
    if (total === 0) return 'basic';
    const ratio = ready / total;
    if (ratio === 1) return 'primary';  // 全部就绪 - 绿色
    if (ratio >= 0.5) return 'accent';  // 部分就绪 - 蓝色
    return 'warn';                      // 大部分未就绪 - 红色
  }

  // 获取Ready状态提示信息
  getReadyTooltip(ready: number = 0, total: number = 0, containers: ContainerStatus[] = []): string {
    if (total === 0) return '无容器';
    
    const notReadyContainers = containers.filter(c => !c.ready);
    if (notReadyContainers.length === 0) {
      return '所有容器都已就绪';
    }
    
    const notReadyInfo = notReadyContainers.map(c => 
      `${c.name}: ${c.state || '未知'} ${c.reason ? '(' + c.reason + ')' : ''}`
    ).join('\n');
    
    return `未就绪容器:\n${notReadyInfo}`;
  }

  // 获取不健康容器信息
  getUnhealthyContainersInfo(containers: ContainerStatus[] = []): string {
    const unhealthy = containers.filter(c => !c.ready || (c.restartCount > 3));
    if (unhealthy.length === 0) return '';
    
    return '不健康容器:\n' + unhealthy.map(c => 
      `${c.name}: 重启${c.restartCount}次 ${c.reason ? '(' + c.reason + ')' : ''}`
    ).join('\n');
  }

  // 获取重启次数提示信息
  getRestartsTooltip(containers: ContainerStatus[] = []): string {
    if (containers.length === 0) return '无容器信息';
    
    return '各容器重启次数:\n' + containers.map(c => 
      `${c.name}: ${c.restartCount}次`
    ).join('\n');
  }

}
