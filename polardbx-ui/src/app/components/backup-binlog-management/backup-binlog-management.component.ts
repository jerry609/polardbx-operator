import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { ApiService } from '../../services/api.service';
import { LoadingKeys, LoadingService } from '../../services/loading.service';
import { PolarDBXBackupBinlog, CreateBackupBinlogRequest } from '../../models/backup-binlog.model';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';
import { EmptyStateComponent } from '../empty-state/empty-state.component';
import { MatTooltipModule } from '@angular/material/tooltip';
import { NamespaceService } from '../../services/namespace.service';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

@Component({
  selector: 'app-backup-binlog-management',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatTabsModule,
    MatProgressBarModule,
    MatSnackBarModule,
    MatChipsModule,
    MatTooltipModule,
    EmptyStateComponent
  ],
  template: `
    <div class="backup-binlog-management">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>description</mat-icon>
            日志备份管理
          </mat-card-title>
          <mat-card-subtitle>管理数据库二进制日志备份和时间点恢复</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <mat-tab-group [(selectedIndex)]="selectedTab">
            <mat-tab label="日志备份列表">
              <div class="tab-content">
                <div class="actions-toolbar">
                  <button mat-raised-button color="primary" (click)="load()">
                    <mat-icon>refresh</mat-icon>
                    刷新
                  </button>
                  <button mat-raised-button color="accent" (click)="createNew()">
                    <mat-icon>add</mat-icon>
                    创建备份配置
                  </button>
                  <a *ngIf="grafanaLink" [href]="grafanaLink" target="_blank" rel="noopener" mat-button>
                    <mat-icon>open_in_new</mat-icon>
                    在 Grafana 打开
                  </a>
                </div>
                <mat-progress-bar *ngIf="loadingService.isLoading(loadingKeys.BACKUP_BINLOG_LIST)" mode="indeterminate"></mat-progress-bar>
                <table mat-table [dataSource]="binlogs" class="full-width-table" *ngIf="binlogs?.length; else emptyState">
                  <ng-container matColumnDef="name">
                    <th mat-header-cell *matHeaderCellDef>名称</th>
                    <td mat-cell *matCellDef="let b">{{ b.metadata.name }}</td>
                  </ng-container>
                  <ng-container matColumnDef="pxcName">
                    <th mat-header-cell *matHeaderCellDef>集群</th>
                    <td mat-cell *matCellDef="let b">{{ b.spec.pxcName }}</td>
                  </ng-container>
                  <ng-container matColumnDef="provider">
                    <th mat-header-cell *matHeaderCellDef>存储</th>
                    <td mat-cell *matCellDef="let b">{{ b.spec.storageProvider?.storageName || '-' }}</td>
                  </ng-container>
                  <ng-container matColumnDef="phase">
                    <th mat-header-cell *matHeaderCellDef>状态</th>
                    <td mat-cell *matCellDef="let b">
                      <mat-chip [color]="getStatusColor(b.status?.phase)">{{ b.status?.phase || '未知' }}</mat-chip>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="latestBackup">
                    <th mat-header-cell *matHeaderCellDef>最新时间</th>
                    <td mat-cell *matCellDef="let b">
                      <ng-container *ngIf="metricsMap[b.metadata.name] as m; else dashLB">
                        {{ formatLatestTime(m?.latestBackupTime) }}
                      </ng-container>
                      <ng-template #dashLB>-</ng-template>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="rpo">
                    <th mat-header-cell *matHeaderCellDef>RPO</th>
                    <td mat-cell *matCellDef="let b">
                      <ng-container *ngIf="metricsMap[b.metadata.name] as m; else dash">
                        <mat-chip [color]="m.rpoStatus === 'ok' ? 'primary' : 'warn'">{{ formatLag(m) }}</mat-chip>
                      </ng-container>
                      <ng-template #dash>-</ng-template>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="throughput">
                    <th mat-header-cell *matHeaderCellDef>吞吐</th>
                    <td mat-cell *matCellDef="let b">
                      <ng-container *ngIf="metricsMap[b.metadata.name] as m; else dash2">
                        <mat-chip [color]="m.throughputStatus === 'ok' ? 'primary' : 'warn'">{{ throughputValue(m) | number:'1.2-2' }} MB/s</mat-chip>
                      </ng-container>
                      <ng-template #dash2>-</ng-template>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="recentFiles">
                    <th mat-header-cell *matHeaderCellDef>近期文件</th>
                    <td mat-cell *matCellDef="let b">
                      <ng-container *ngIf="metricsMap[b.metadata.name] as m; else dashRF">
                        <span [matTooltip]="(m.recentFiles || []).join('\n')">
                          {{ previewRecentFiles(m) }}
                        </span>
                      </ng-container>
                      <ng-template #dashRF>-</ng-template>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="actions">
                    <th mat-header-cell *matHeaderCellDef>操作</th>
                    <td mat-cell *matCellDef="let b">
                      <button mat-icon-button (click)="openEditDialog(b)" aria-label="编辑"><mat-icon>edit</mat-icon></button>
                      <button mat-icon-button color="warn" (click)="deleteBinlog(b)" aria-label="删除"><mat-icon>delete</mat-icon></button>
                    </td>
                  </ng-container>
                  <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
                  <tr mat-row *matRowDef="let row; columns: displayedColumns;"></tr>
                </table>
                <ng-template #emptyState>
                  <app-empty-state icon="description" title="暂无日志备份配置" hint="点击“创建备份配置”来设置您的第一个日志备份"></app-empty-state>
                </ng-template>
              </div>
            </mat-tab>
            <mat-tab label="创建配置">
              <div class="tab-content">
                <div class="form-container">
                  <h3>创建日志备份配置</h3>
                  <p>日志备份配置创建功能正在开发中...</p>
                  <button mat-button (click)="selectedTab = 0">返回列表</button>
                </div>
              </div>
            </mat-tab>
          </mat-tab-group>
        </mat-card-content>
      </mat-card>
    </div>
  `,
    styles: [`
    .backup-binlog-management {
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
    
    .full-width-table {
      width: 100%;
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
  `]
})
export class BackupBinlogManagementComponent implements OnInit, OnDestroy {
  displayedColumns = ['name', 'pxcName', 'provider', 'phase', 'latestBackup', 'rpo', 'throughput', 'recentFiles', 'actions'];
  binlogs: PolarDBXBackupBinlog[] = [];
  metricsMap: Record<string, any> = {};
  loadingKeys = LoadingKeys;
  selectedTab = 0;
  grafanaURL = '';
  grafanaLink = '';

  private apiService = inject(ApiService);
  public loadingService = inject(LoadingService);
  private snackBar = inject(MatSnackBar);
  private ns = inject(NamespaceService);
  private destroy$ = new Subject<void>();

  ngOnInit(): void {
    this.grafanaURL = localStorage.getItem('grafanaURL') || '';
    this.grafanaLink = this.grafanaURL ? `${this.grafanaURL}/d/polardbx-monitor?orgId=1&var-namespace=default` : '';
    this.load();
    this.ns.activeNamespace$
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => this.load());
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  load(): void {
    const ns = 'default';
    this.apiService.getBackupBinlogs(ns).subscribe({
      next: (list: PolarDBXBackupBinlog[]) => {
        this.binlogs = list;
        // 取指标（估算吞吐）, 按名称映射
        this.apiService.getBinlogMetrics(ns, true, 600).subscribe({
          next: (metrics: any[]) => {
            const map: Record<string, any> = {};
            metrics.forEach((m: any) => { map[m.name] = m; });
            this.metricsMap = map;
          },
          error: () => { this.metricsMap = {}; }
        });
      },
      error: () => {
        this.binlogs = [];
        this.metricsMap = {};
      }
    });
  }

  createNew(): void {
    this.selectedTab = 1;
  }

  openCreateDialog(): void {
    this.selectedTab = 1;
  }

  openEditDialog(binlog: PolarDBXBackupBinlog): void {
    this.snackBar.open('编辑功能正在开发中...', '关闭', { duration: 3000 });
  }

  deleteBinlog(b: PolarDBXBackupBinlog): void {
    if (confirm(`确定删除 Binlog 配置 "${b.metadata.name}" 吗？`)) {
      this.apiService.deleteBackupBinlog(b.metadata.namespace || 'default', b.metadata.name).subscribe({
        next: () => {
          this.snackBar.open('删除成功!', '关闭', { duration: 3000 });
          this.load();
        },
        error: (err) => this.snackBar.open(`删除失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 })
      });
    }
  }

  getStatusColor(phase?: string): string {
    switch (phase?.toLowerCase()) {
      case 'running':
        return 'primary';
      case 'succeeded':
        return 'accent';
      case 'failed':
        return 'warn';
      default:
        return '';
    }
  }

  formatLag(m?: any): string {
    const v = m?.lagSeconds;
    return (typeof v === 'number' && isFinite(v)) ? `${v}s` : '-';
  }

  throughputValue(m?: any): number {
    const v = m?.throughputMBps;
    return (typeof v === 'number' && isFinite(v)) ? v : 0;
  }

  formatLatestTime(v?: any): string {
    if (!v || v === 'pending_implementation') return '-';
    try { return new Date(v).toLocaleString(); } catch {
      return '-';
    }
  }

  previewRecentFiles(m?: any): string {
    const files: string[] = Array.isArray(m?.recentFiles) ? m.recentFiles : [];
    if (!files.length) return '-';
    if (files.length <= 3) return files.join(', ');
    return `${files.slice(0, 3).join(', ')} 等 ${files.length} 个`;
  }
}