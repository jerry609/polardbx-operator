import { Component, inject, OnInit, OnDestroy, TemplateRef, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { RouterModule } from '@angular/router';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { PolarDBXBackup } from '../../models/backup.model';
import { EmptyStateComponent } from '../empty-state/empty-state.component';

@Component({
  selector: 'app-backup-management',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatTabsModule,
    MatProgressBarModule,
    MatChipsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatCheckboxModule,
    RouterModule,
    MatSnackBarModule,
    MatDialogModule,
    EmptyStateComponent
  ],
  template: `
    <div class="backup-management">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>save</mat-icon>
            手动备份管理
          </mat-card-title>
          <mat-card-subtitle>管理集群手动备份</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <mat-tab-group [(selectedIndex)]="selectedTab">
            <mat-tab label="手动备份列表">
              <div class="tab-content">
                <div class="actions-toolbar">
                  <button mat-raised-button color="primary" (click)="refreshBackups()">
                    <mat-icon>refresh</mat-icon>
                    刷新
                  </button>
                  <button mat-raised-button color="accent" (click)="createNew()">
                    <mat-icon>add</mat-icon>
                    新建备份
                  </button>
                  <a *ngIf="grafanaURL" [href]="grafanaURL" target="_blank" rel="noopener" mat-button>
                    <mat-icon>open_in_new</mat-icon>
                    在 Grafana 打开
                  </a>
                </div>
                <mat-progress-bar *ngIf="loadingService.isLoading(loadingKeys.BACKUPS_LIST)" mode="indeterminate"></mat-progress-bar>
                
                <app-empty-state *ngIf="backups.length === 0 && !loadingService.isLoading(loadingKeys.BACKUPS_LIST)"
                                  icon="backup"
                                  title="暂无手动备份记录"
                                  hint="点击“新建备份”创建您的第一个手动备份"></app-empty-state>
                
                <mat-table *ngIf="backups.length > 0" [dataSource]="backups" class="backup-table">
                  <ng-container matColumnDef="name">
                    <mat-header-cell *matHeaderCellDef>备份名称</mat-header-cell>
                    <mat-cell *matCellDef="let backup">
                      {{ backup.metadata.name }}
                      <div *ngIf="isRunningEx(backup)" class="inline-progress">
                        <mat-progress-bar mode="indeterminate"></mat-progress-bar>
                      </div>
                    </mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="cluster">
                    <mat-header-cell *matHeaderCellDef>集群</mat-header-cell>
                    <mat-cell *matCellDef="let backup">{{ backup.spec?.cluster?.name || '-' }}</mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="status">
                    <mat-header-cell *matHeaderCellDef>状态</mat-header-cell>
                    <mat-cell *matCellDef="let backup">
                      <mat-chip [color]="getBackupStatusColorEx(backup)">
                        {{ mapPhaseTextEx(backup) }}
                      </mat-chip>
                      <div class="status-hint" *ngIf="isRunningEx(backup)">
                        进度：{{ backupProgress[backup.metadata.name]?.progressPercent ?? '—' }}%
                        <span *ngIf="backupProgress[backup.metadata.name]?.estimated" style="opacity:0.7">（估算）</span>
                        ｜ <a (click)="openBackupDetail(backup)" style="cursor:pointer;">查看详情</a>
                      </div>
                    </mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="createdTime">
                    <mat-header-cell *matHeaderCellDef>创建时间</mat-header-cell>
                    <mat-cell *matCellDef="let backup">{{ backup.metadata.creationTimestamp | date:'medium' }}</mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="actions">
                    <mat-header-cell *matHeaderCellDef>操作</mat-header-cell>
                    <mat-cell *matCellDef="let backup">
                      <button mat-icon-button color="warn" (click)="deleteBackup(backup)" title="删除备份">
                        <mat-icon>delete</mat-icon>
                      </button>
                      <button mat-icon-button color="warn" (click)="forceDeleteBackup(backup)" title="强制删除（移除finalizers）" style="margin-left: 4px;">
                        <mat-icon>delete_forever</mat-icon>
                      </button>
                    </mat-cell>
                  </ng-container>

                  <mat-header-row *matHeaderRowDef="['name', 'cluster', 'status', 'createdTime', 'actions']"></mat-header-row>
                  <mat-row *matRowDef="let row; columns: ['name', 'cluster', 'status', 'createdTime', 'actions']"></mat-row>
                </mat-table>
              </div>
            </mat-tab>
            <mat-tab label="创建备份">
              <div class="tab-content">
                <div class="form-container">
                  <h3>创建手动备份</h3>
                  
                  <form [formGroup]="backupForm" (ngSubmit)="createBackup()">
                    <div class="form-row">
                      <mat-form-field appearance="outline" class="full-width">
                        <mat-label>备份名称</mat-label>
                        <input matInput formControlName="name" placeholder="输入备份名称（可留空自动命名）">
                        <mat-hint>留空将自动命名；仅包含小写字母、数字和连字符；最长 63</mat-hint>
                        
                        <mat-error *ngIf="backupForm.get('name')?.hasError('pattern')">
                          名称只能包含小写字母、数字和连字符
                        </mat-error>
                        <mat-error *ngIf="backupForm.get('name')?.hasError('maxlength')">
                          名称长度不能超过 63 个字符
                        </mat-error>
                      </mat-form-field>
                    </div>

                    <div class="form-row">
                      <mat-form-field appearance="outline" class="half-width">
                        <mat-label>命名空间</mat-label>
                        <mat-select formControlName="namespace">
                          <mat-option value="default">default</mat-option>
                          <mat-option *ngFor="let ns of namespaces" [value]="ns">{{ ns }}</mat-option>
                        </mat-select>
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="half-width">
                        <mat-label>目标集群</mat-label>
                        <mat-select formControlName="cluster" (selectionChange)="onClusterChange($event.value)">
                          <mat-option *ngFor="let cluster of clusters" [value]="cluster.metadata.name">
                            {{ cluster.metadata.name }}
                          </mat-option>
                        </mat-select>
                        <mat-error *ngIf="backupForm.get('cluster')?.hasError('required')">
                          请选择目标集群
                        </mat-error>
                      </mat-form-field>
                    </div>

                    <div class="form-row">
                      <mat-form-field appearance="outline" class="half-width">
                        <mat-label>存储提供商</mat-label>
                        <mat-select formControlName="storageProvider" (selectionChange)="onProviderChange($event.value)">
                          <mat-option value="s3">Amazon S3</mat-option>
                          <mat-option value="oss">Alibaba Cloud OSS</mat-option>
                          <mat-option value="sftp">SFTP</mat-option>
                        </mat-select>
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="half-width">
                        <mat-label>备份类型</mat-label>
                        <mat-select formControlName="backupType">
                          <mat-option value="Snapshot">快照备份</mat-option>
                          <mat-option value="PhysicalBackup">物理备份</mat-option>
                        </mat-select>
                      </mat-form-field>
                    </div>

                    <div class="form-row">
                      <ng-container *ngIf="getSinksForProvider(backupForm.get('storageProvider')?.value).length > 0; else sinkInputTpl">
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>存储配置 (sink 名称)</mat-label>
                          <mat-select formControlName="storageSink" (selectionChange)="onSinkChange()">
                            <mat-option *ngFor="let s of getSinksForProvider(backupForm.get('storageProvider')?.value)" [value]="s.name">
                              {{ s.name }} ({{ s.type }})
                            </mat-option>
                          </mat-select>
                          <mat-hint>
                            这是 HPFS/filestream 的“sink 名称”（例如：default、lyfz-polardbx-backup）。
                            <ng-container *ngIf="secretTemplateUrl">
                              ｜ Secret 模板：
                              <a [href]="secretTemplateUrl" target="_blank" download>下载</a>
                            </ng-container>
                            <span *ngIf="sinkStatusText" [style.color]="sinkInvalid ? '#d32f2f' : '#2e7d32'" style="margin-left:8px;">{{ sinkStatusText }}</span>
                          </mat-hint>
                          <mat-error *ngIf="backupForm.get('storageSink')?.hasError('required')">
                            请选择 sink 名称（非 URL）
                          </mat-error>
                        </mat-form-field>
                      </ng-container>
                      <ng-template #sinkInputTpl>
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>存储配置 (sink 名称)</mat-label>
                          <input matInput formControlName="storageSink" [placeholder]="sinkPlaceholder" (blur)="onSinkChange()">
                          <mat-hint>
                            这是 HPFS/filestream 的“sink 名称”（例如：default、lyfz-polardbx-backup）。
                            <ng-container *ngIf="secretTemplateUrl">
                              ｜ Secret 模板：
                              <a [href]="secretTemplateUrl" target="_blank" download>下载</a>
                            </ng-container>
                            <span *ngIf="sinkStatusText" [style.color]="sinkInvalid ? '#d32f2f' : '#2e7d32'" style="margin-left:8px;">{{ sinkStatusText }}</span>
                          </mat-hint>
                          <mat-error *ngIf="backupForm.get('storageSink')?.hasError('required')">
                            请填写 sink 名称（非 URL）
                          </mat-error>
                        </mat-form-field>
                      </ng-template>
                    </div>

                    <div class="form-row">
                      <mat-checkbox formControlName="compression">启用压缩</mat-checkbox>
                      <mat-checkbox formControlName="encryption" style="margin-left: 16px;">启用加密</mat-checkbox>
                      <mat-checkbox formControlName="preferFollower" style="margin-left: 16px;">仅在 follower 节点执行</mat-checkbox>
                      <mat-checkbox formControlName="skipPrecheck" style="margin-left: 16px;">跳过预校验</mat-checkbox>
                    </div>

                    <div class="form-actions">
                      <button mat-button type="button" (click)="resetBackupForm()">重置</button>
                      <button mat-button type="button" (click)="selectedTab = 0">取消</button>
                      <button mat-raised-button color="primary" type="submit" 
                              [disabled]="backupForm.invalid || isCreatingBackup || sinkInvalid">
                        <mat-icon>save</mat-icon>
                        {{ isCreatingBackup ? '创建中...' : '创建备份' }}
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            </mat-tab>
          </mat-tab-group>
        </mat-card-content>
      </mat-card>
    </div>

    <ng-template #backupDetailsDialog>
      <h2 mat-dialog-title>备份详情</h2>
      <mat-dialog-content>
        <div><strong>名称：</strong>{{ selectedBackup?.metadata?.name }}</div>
        <div><strong>命名空间：</strong>{{ selectedBackup?.metadata?.namespace || 'default' }}</div>
        <div><strong>集群：</strong>{{ selectedBackup?.spec?.cluster?.name || '-' }}</div>
        <div style="margin:8px 0">
          <mat-chip [color]="getBackupStatusColorEx(selectedBackup!)">{{ mapPhaseTextEx(selectedBackup!) }}</mat-chip>
        </div>
        <div *ngIf="selectedBackup">
          <div><strong>进度：</strong>{{ backupProgress[selectedBackup.metadata.name]?.progressPercent ?? '—' }}%</div>
          <div><strong>大小：</strong>{{ (backupProgress[selectedBackup.metadata.name]?.sizeBytes || 0) / 1048576 | number:'1.0-0' }} MB</div>
        </div>
        <div style="margin-top:8px" *ngIf="grafanaLink">
          <a [href]="grafanaLink" target="_blank" rel="noopener">
            <mat-icon style="vertical-align:middle; margin-right:4px">open_in_new</mat-icon>在 Grafana 打开
          </a>
        </div>
      </mat-dialog-content>
      <mat-dialog-actions align="end">
        <button mat-button mat-dialog-close>关闭</button>
      </mat-dialog-actions>
    </ng-template>
  `,
  styles: [`
    .backup-management {
      padding: 20px;
    }
    
    .tab-content {
      padding: 20px 0;
    }
    
    .actions-toolbar {
      margin-bottom: 16px;
      display: flex;
      gap: 8px;
    }
    
    mat-card-title {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    
    .empty-state {
      text-align: center;
      padding: 40px;
      color: #666;
    }
    
    .empty-state mat-icon {
      font-size: 48px;
      width: 48px;
      height: 48px;
      margin-bottom: 16px;
      opacity: 0.5;
    }
    
    .empty-state .hint {
      font-size: 0.9em;
      opacity: 0.7;
    }
    
    .form-container {
      padding: 20px;
    }
    
    .form-row {
      display: flex;
      gap: 16px;
      margin-bottom: 16px;
    }
    
    .full-width {
      width: 100%;
    }
    
    .half-width {
      flex: 1;
    }
    
    .form-actions {
      display: flex;
      gap: 8px;
      justify-content: flex-end;
      margin-top: 24px;
      padding-top: 16px;
      border-top: 1px solid #e0e0e0;
    }
  `]
})
export class BackupManagementComponent implements OnInit, OnDestroy {
    selectedTab = 0;
    loadingKeys = LoadingKeys;
    backups: PolarDBXBackup[] = [];
    // 估算进度缓存
    backupProgress: Record<string, { progressPercent: number; estimated: boolean; sizeBytes: number | null; phase: string }> = {};
    // 备份详情与进度订阅（简单 SSE 骨架）
    private eventSources: Record<string, EventSource> = {};
    selectedBackup: PolarDBXBackup | null = null;
    grafanaLink = '';
    @ViewChild('backupDetailsDialog') backupDetailsDialog!: TemplateRef<any>;
    private dialog = inject(MatDialog);
    // 详情抽屉/对话框
    openBackupDetail(backup: PolarDBXBackup): void {
        this.selectedBackup = backup;
        const ns = backup.metadata.namespace || 'default';
        const cluster = backup.spec?.cluster?.name || '';
        const base = localStorage.getItem('grafanaURL') || '';
        this.grafanaLink = base ? `${base}/d/polardbx-monitor?orgId=1&var-namespace=${encodeURIComponent(ns)}${cluster ? `&var-cluster=${encodeURIComponent(cluster)}` : ''}` : '';
        this.subscribeBackup(ns, backup.metadata.name);
        this.dialog.open(this.backupDetailsDialog);
    }

    // 订阅备份进度（SSE + 定时轮询混合）
    public subscribeBackup(namespace: string, name: string): void {
        const key = `${namespace}/${name}`;
        
        // 避免重复订阅
        if (this.eventSources[key]) {
            return;
        }

        // 立即获取一次进度
        this.fetchBackupProgress(namespace, name);

        // SSE 订阅 phase 变化
        try {
            const kubeconfigB64 = localStorage.getItem('kubeconfig-b64') || sessionStorage.getItem('kubeconfig-b64');
            const sseUrl = `http://localhost:8080/api/v1/backups/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/stream${kubeconfigB64 ? `?k=${encodeURIComponent(kubeconfigB64)}` : ''}`;
            
            const eventSource = new EventSource(sseUrl);
            this.eventSources[key] = eventSource;

            eventSource.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    if (data.type === 'phaseChanged') {
                        // Phase 变化时重新获取详细进度
                        this.fetchBackupProgress(namespace, name);
                    }
                } catch (e) {
                    console.warn('SSE 数据解析失败:', e);
                }
            };

            eventSource.onerror = () => {
                console.warn(`SSE 连接错误: ${key}`);
                eventSource.close();
                delete this.eventSources[key];
            };

        } catch (e) {
            console.warn('SSE 不可用，使用轮询模式:', e);
        }

        // 备用轮询机制（SSE失败或不支持时）
        if (!this.eventSources[key]) {
            const interval = setInterval(() => {
                this.fetchBackupProgress(namespace, name);
            }, 3000);
            
            // 存储间隔ID以便清理
            (this.eventSources as any)[`${key}_interval`] = interval;
        }
    }

    // 获取备份详细进度
    private fetchBackupProgress(namespace: string, name: string): void {
        this.apiService.getBackupMetrics(namespace, name).subscribe({
            next: (metrics) => {
                this.backupProgress[name] = {
                    progressPercent: metrics.progressPercent || 0,
                    estimated: metrics.estimated ?? true,
                    sizeBytes: metrics.sizeBytes,
                    phase: metrics.phase || 'unknown'
                };

                // 如果备份已完成，停止订阅
                if (['finished', 'failed', 'succeeded', 'completed'].includes(metrics.phase?.toLowerCase() || '')) {
                    this.unsubscribeBackup(namespace, name);
                }
            },
            error: (error) => {
                console.warn(`获取备份进度失败 (${name}):`, error);
                // 如果是404，说明备份可能已删除
                if (error.status === 404) {
                    this.unsubscribeBackup(namespace, name);
                }
            }
        });
    }

    // 取消订阅
    private unsubscribeBackup(namespace: string, name: string): void {
        const key = `${namespace}/${name}`;
        
        // 关闭 SSE 连接
        if (this.eventSources[key]) {
            this.eventSources[key].close();
            delete this.eventSources[key];
        }

        // 清理轮询间隔
        const intervalKey = `${key}_interval`;
        if ((this.eventSources as any)[intervalKey]) {
            clearInterval((this.eventSources as any)[intervalKey]);
            delete (this.eventSources as any)[intervalKey];
        }
    }
    clusters: any[] = [];
    namespaces: string[] = ['default'];
    isCreatingBackup = false;
    backupForm: FormGroup;
    sinkPlaceholder = '存储端点配置，例如: s3://bucket-name/path';
    sinkExample = 's3://my-bucket/backups/cluster-A/';
    secretTemplateUrl = '';
    hpfsSinks: Array<{ name: string; type: string; endpoint?: string; bucket?: string; bucketLookupType?: string; host?: string; port?: number; rootPath?: string; }> = [];
    sinkInvalid: boolean = false;
    sinkStatusText: string = '';
    grafanaURL = '';
    
    private fb = inject(FormBuilder);
    private snackBar = inject(MatSnackBar);
    public loadingService = inject(LoadingService);
    private apiService = inject(ApiService);

    constructor() {
        this.backupForm = this.initBackupForm();
    }

    ngOnInit(): void {
        this.grafanaURL = localStorage.getItem('grafanaURL') || '';
        this.loadClusters();
        this.loadHpfsSinks();
        this.refreshBackups();
    }

    ngOnDestroy(): void {
        // 清理所有订阅
        Object.keys(this.eventSources).forEach(key => {
            if (key.includes('_interval')) {
                clearInterval((this.eventSources as any)[key]);
            } else {
                this.eventSources[key]?.close();
            }
        });
        this.eventSources = {};
    }

    private loadHpfsSinks(): void {
        // 读取 kubeconfig 并以 Header 方式调用，沿用 ApiService 默认 headers
        this.apiService.getHpfsSinks().subscribe({
            next: (resp) => {
                this.hpfsSinks = resp.sinks || [];
                // 默认策略：若存在 s3 类型且 name=default 的 sink，则默认 provider=s3，sink=default
                const defaultS3 = this.hpfsSinks.find(s => (s.type || '').toLowerCase() === 's3' && s.name === 'default');
                if (defaultS3) {
                    this.backupForm.patchValue({ storageProvider: 's3', storageSink: 'default' });
                    this.sinkInvalid = false;
                    this.sinkStatusText = '已验证：default (s3)';
                }
            },
            error: () => {
                // 静默失败，不阻塞表单
            }
        });
    }

    getSinksForProvider(provider: string): Array<{ name: string; type: string; endpoint?: string; bucket?: string; bucketLookupType?: string; host?: string; port?: number; rootPath?: string; }> {
        const p = (provider || '').toLowerCase();
        return (this.hpfsSinks || []).filter(s => (s.type || '').toLowerCase() === p);
    }

    private async decideRoleByAdvice(namespace: string, clusterName: string): Promise<'leader'|'follower'|undefined> {
        try {
            const advice = await this.apiService.getBackupAdvice(namespace, clusterName).toPromise();
            if (advice?.role === 'leader' || advice?.role === 'follower') {
                return advice.role;
            }
        } catch {}
        return undefined;
    }

    onSinkChange(): void {
        const provider = String(this.backupForm.get('storageProvider')?.value || '').toLowerCase();
        const sinkName = String(this.backupForm.get('storageSink')?.value || '').trim();
        if (!provider || !sinkName) {
            this.sinkInvalid = true;
            this.sinkStatusText = '未填写 sink';
            return;
        }
        // 只做存在性校验（不含连通性）
        this.apiService.validateSink(sinkName, provider).subscribe({
            next: (res) => {
                if (res.status === 'ok') {
                    this.sinkInvalid = false;
                    this.sinkStatusText = `已验证：${res.name} (${res.type})`;
                } else {
                    this.sinkInvalid = true;
                    this.sinkStatusText = '未找到该 sink，检查 HPFS 配置';
                }
            },
            error: () => {
                this.sinkInvalid = true;
                this.sinkStatusText = '校验失败';
            }
        });
    }

    loadClusters(): void {
        this.apiService.getClusters().subscribe({
            next: (clusters) => {
                this.clusters = clusters;
                if (clusters.length > 0) {
                    // 加载第一个集群的备份
                    this.loadBackups(clusters[0].metadata.namespace, clusters[0].metadata.name);
                }
            },
            error: (err) => {
                this.snackBar.open(`加载集群列表失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
            }
        });
    }

    loadBackups(namespace: string, clusterName: string): void {
        this.apiService.getBackups(namespace, clusterName).subscribe({
            next: (backups) => {
                this.backups = backups;
                // 对运行中的备份，自动尝试建立 SSE 订阅（骨架）
                backups.forEach(b => {
                    if (this.isRunning(b.status?.phase)) {
                        this.ensureSubscribed(b.metadata.namespace || 'default', b.metadata.name);
                    }
                    // 拉取估算进度
                    const ns = b.metadata.namespace || 'default';
                    const name = b.metadata.name;
                    this.apiService.getBackupMetrics(ns, name).subscribe({
                        next: (m) => { this.backupProgress[name] = m as any; },
                        error: () => { /* 忽略 */ }
                    });
                });
            },
            error: (err) => {
                this.snackBar.open(`加载备份列表失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
                this.backups = [];
            }
        });
    }

    refreshBackups(): void {
        if (this.clusters.length > 0) {
            const firstCluster = this.clusters[0];
            this.loadBackups(firstCluster.metadata.namespace, firstCluster.metadata.name);
            this.snackBar.open('备份列表已刷新', '关闭', { duration: 2000 });
        } else {
            this.snackBar.open('没有可用的集群', '关闭', { duration: 2000 });
        }
    }

    createNew(): void {
        this.selectedTab = 1;
    }

    ensureSubscribed(namespace: string, name: string) {
        const key = `${namespace}/${name}`;
        if (this.eventSources[key]) {
            return;
        }
        this.subscribeBackup(namespace, name);
    }



    deleteBackup(backup: PolarDBXBackup): void {
        if (confirm(`确定删除备份 "${backup.metadata.name}" 吗？`)) {
            this.apiService.deleteBackup(backup.metadata.namespace || 'default', backup.metadata.name).subscribe({
                next: () => {
                    this.snackBar.open('删除成功!', '关闭', { duration: 3000 });
                    this.refreshBackups();
                },
                error: (err) => {
                    this.snackBar.open(`删除失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
                }
            });
        }
    }

    forceDeleteBackup(backup: PolarDBXBackup): void {
        if (confirm(`强制删除将直接移除 finalizers 并清理对象。确定对备份 "${backup.metadata.name}" 执行吗？`)) {
            this.apiService.forceDeleteBackup(backup.metadata.namespace || 'default', backup.metadata.name).subscribe({
                next: () => {
                    this.snackBar.open('强制删除成功!', '关闭', { duration: 3000 });
                    this.refreshBackups();
                },
                error: (err) => {
                    this.snackBar.open(`强制删除失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
                }
            });
        }
    }

    getBackupStatusColor(status?: string): string {
        // PolarDBXBackup v1 枚举："", FullBackuping, Collecting, Calculating, BinlogBackuping, MetadataBackuping, Finished, Failed, Deleting
        const p = (status || '').toLowerCase();
        if (!p) return 'accent'; // 创建中
        if (['fullbackuping','collecting','calculating','binlogbackuping','metadatabackuping'].includes(p)) return 'primary';
        if (['finished','succeeded','completed'].includes(p)) return 'accent';
        if (['deleting','failed'].includes(p)) return 'warn';
        return '';
    }
    mapPhaseText(phase?: string): string {
        const p = (phase || '').toLowerCase();
        if (!p) return '创建中';
        switch (p) {
            case 'fullbackuping':
                return '全量备份中';
            case 'collecting':
                return '收集中';
            case 'calculating':
                return '计算中';
            case 'binlogbackuping':
                return '日志备份中';
            case 'metadatabackuping':
                return '元数据备份中';
            case 'finished':
            case 'succeeded':
            case 'completed':
                return '已完成';
            case 'failed':
                return '失败';
            case 'deleting':
                return '删除中';
            default:
                return '未知';
        }
    }

    isRunning(phase?: string): boolean {
        const p = (phase || '').toLowerCase();
        // 仅 v1 的非终态视为运行中；空串视为“创建中”，不显示“进行中…”字样
        return ['fullbackuping','collecting','calculating','binlogbackuping','metadatabackuping'].includes(p);
    }

    // 增强版：当 phase 为空时，尝试推断阶段（如果后端返回了 Message 或 StartTime）
    private inferPhaseWhenEmpty(backup: PolarDBXBackup): string {
        const phase = (backup?.status?.phase || '').toString();
        if (phase) return phase;
        // 无相位：按是否已经开始推断“创建中”/“收集中”
        // 这里保守返回空串，由 mapPhaseTextEx 统一呈现“创建中”
        return '';
    }

    mapPhaseTextEx(backup: PolarDBXBackup): string {
        const phase = (this.inferPhaseWhenEmpty(backup) || '').toString();
        return this.mapPhaseText(phase);
    }

    getBackupStatusColorEx(backup: PolarDBXBackup): string {
        const phase = (this.inferPhaseWhenEmpty(backup) || '').toString();
        return this.getBackupStatusColor(phase);
    }

    isRunningEx(backup: PolarDBXBackup): boolean {
        const phase = (this.inferPhaseWhenEmpty(backup) || '').toString();
        return this.isRunning(phase);
    }

    private initBackupForm(): FormGroup {
    const prefs = this.getBackupPreferences();
    const fg = this.fb.group({
        name: ['', [Validators.pattern(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/), Validators.maxLength(63)]],
        namespace: ['default', Validators.required],
        cluster: ['', Validators.required],
        storageProvider: [prefs.storageName || 's3', Validators.required],
        backupType: ['Snapshot', Validators.required],
        storageSink: [prefs.sink || '', []],
        compression: [false],
        encryption: [false],
        preferFollower: [false],
        skipPrecheck: [true]
    });
    // 初始化联动与校验
    this.applyProviderRules(String(fg.get('storageProvider')?.value ?? ''), fg);
    fg.get('storageProvider')?.valueChanges.subscribe(v => this.applyProviderRules(String(v || ''), fg));
    return fg;
}

    private getBackupPreferences(): { storageName: string; sink: string; retentionTime: string } {
    const storageName = (localStorage.getItem('backupStorageName') || 's3').trim();
    const sink = (localStorage.getItem('backupSink') || '').trim();
    const retentionTime = (localStorage.getItem('backupRetentionTime') || '240h').trim();
    return { storageName, sink, retentionTime };
}

    onClusterChange(clusterName: string): void {
    // 可以根据选择的集群更新命名空间
    const selectedCluster = this.clusters.find(c => c.metadata.name === clusterName);
    if (selectedCluster) {
        this.backupForm.patchValue({ namespace: selectedCluster.metadata.namespace });
    }
    // 使用后端建议展示提示（真正设置在提交时执行）
    const ns = selectedCluster?.metadata?.namespace || 'default';
    this.apiService.getBackupAdvice(ns, clusterName).subscribe({
        next: (advice) => {
            if (!this.backupForm.get('preferFollower')?.value && advice?.role === 'leader') {
                this.snackBar.open('后端建议：单副本（无 follower），将默认在 leader 上执行。', '关闭', { duration: 3000 });
            }
        },
        error: () => { /* 忽略错误，仅用于提示 */ }
    });
}

    resetBackupForm(): void {
    this.backupForm.reset();
    this.backupForm.patchValue({
        namespace: 'default',
        storageProvider: 's3',
        backupType: 'Snapshot',
        compression: false,
        encryption: false
    });
}

    async createBackup(): Promise<void> {
    if (this.backupForm.invalid) {
        this.snackBar.open('请检查表单填写是否正确', '关闭', { duration: 3000 });
        return;
    }


    this.isCreatingBackup = true;
    const formValue = this.backupForm.value;
    const prefs = this.getBackupPreferences();

    try {
        const backupObject: any = {
            apiVersion: 'polardbx.aliyun.com/v1',
            kind: 'PolarDBXBackup',
            metadata: {
                name: formValue.name,
                namespace: formValue.namespace
            },
            spec: {
                cluster: { name: formValue.cluster },
                retentionTime: prefs.retentionTime,
                storageProvider: {
                    storageName: formValue.storageProvider || prefs.storageName,
                    sink: formValue.storageSink || prefs.sink
                }
            }
        };
        if (formValue.preferFollower) {
            backupObject.spec.preferredBackupRole = 'follower';
        } else {
            // 基于简单启发式：若当前所选集群推断为单副本（探测 dn-0 的 totalPods<=1），则强制 leader
            try {
                const ns = this.backupForm.get('namespace')?.value || 'default';
                const dnName = `${formValue.cluster}-dn-0`;
                const xs = await this.apiService.getXStore(ns, dnName).toPromise().catch(() => null);
                const totalPods = Number((xs as any)?.status?.totalPods || 0);
                if (totalPods <= 1) {
                    (backupObject.spec as any).preferredBackupRole = 'leader';
                }
            } catch { /* 忽略探测错误，保持不传 */ }
        }
        // 无法可靠获取 follower 副本时，不强制 leader；但当用户未勾选 follower 且集群明显为单副本时，可考虑默认 leader。
        // 简化：保持未勾选不传；后续可基于集群拓扑补强。

        // Advice：按后端建议设置角色（未勾选 follower 时）
        if (!formValue.preferFollower) {
            const adviceRole = await this.decideRoleByAdvice(formValue.namespace, formValue.cluster);
            if (adviceRole) {
                (backupObject.spec as any).preferredBackupRole = adviceRole;
            }
        }

        // 先进行后端 dry-run 预校验（触发 webhook），通过后再创建
        if (!formValue.skipPrecheck) {
            const controller = new AbortController();
            const timer = setTimeout(() => controller.abort(), 15000);
            try {
                await this.apiService.validateBackup(formValue.namespace, backupObject).toPromise();
            } finally {
                clearTimeout(timer);
            }
        }
        await this.apiService.createBackup(formValue.namespace, formValue.cluster, backupObject).toPromise();
        
        this.snackBar.open('备份创建成功', '关闭', { duration: 3000 });
        this.resetBackupForm();
        this.selectedTab = 0;
        this.refreshBackups();
        
    } catch (error: any) {
        console.error('创建备份失败:', error);
        this.snackBar.open(`创建备份失败: ${error.error?.message || error.message}`, '关闭', { duration: 5000 });
    } finally {
        this.isCreatingBackup = false;
    }
}

    onProviderChange(provider: string) {
        this.applyProviderRules(provider, this.backupForm);
    }

    private applyProviderRules(provider: string, formGroup?: FormGroup) {
        const targetForm = formGroup || this.backupForm;
        if (!targetForm) return;
        const control = targetForm.get('storageSink');
        if (!control) return;
        // 清理旧校验
        control.clearValidators();
        const validators = [] as any[];
        // 基本规则：不同 provider 不同示例与格式
        switch ((provider || '').toLowerCase()) {
            case 's3':
                this.sinkPlaceholder = '例如：default 或自定义 sink 名（非 URL）';
                this.sinkExample = 'default（或 lyfz-polardbx-backup 等）';
                this.secretTemplateUrl = '/assets/examples/secret-s3.yaml';
                validators.push(Validators.required);
                break;
            case 'oss':
                this.sinkPlaceholder = '例如：default-oss 或自定义 sink 名（非 URL）';
                this.sinkExample = 'default-oss（或自定义名称）';
                this.secretTemplateUrl = '/assets/examples/secret-oss.yaml';
                validators.push(Validators.required);
                break;
            case 'sftp':
                this.sinkPlaceholder = '例如：default-sftp 或自定义 sink 名（非 URL）';
                this.sinkExample = 'default-sftp（或自定义名称）';
                this.secretTemplateUrl = '/assets/examples/secret-sftp.yaml';
                validators.push(Validators.required);
                break;
            default:
                this.sinkPlaceholder = '请输入 HPFS/filestream 的 sink 名称（例如：default）';
                this.sinkExample = 'default';
                this.secretTemplateUrl = '';
                validators.push(Validators.required);
        }
        control.setValidators(validators);
        control.updateValueAndValidity();
  }
}