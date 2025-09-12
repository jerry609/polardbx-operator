import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { RestoreJob, RestoreJobWithStatus } from '../../models/restore.model';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Subscription, interval } from 'rxjs';
import { switchMap } from 'rxjs/operators';
import { FormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatPaginator, MatPaginatorModule } from '@angular/material/paginator';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';

@Component({
  selector: 'app-restore-job-management',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatTabsModule,
    MatProgressBarModule,
    MatChipsModule,
    MatDividerModule,
    MatTooltipModule,
    MatFormFieldModule,
    MatInputModule,
    FormsModule,
    MatPaginatorModule,
    MatSortModule,
    MatSlideToggleModule
  ],
  template: `
    <div class="restore-page">
      <mat-card class="page-header-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon>manage_search</mat-icon>
            恢复任务
          </mat-card-title>
          <mat-card-subtitle>管理恢复作业与进度</mat-card-subtitle>
        </mat-card-header>
      </mat-card>

      <!-- 左侧任务列表面板 -->
      <div class="list-panel">
        <mat-card class="list-card">
          <mat-card-header>
            <mat-card-title>任务列表（{{ dataSource?.filteredData?.length || 0 }}）</mat-card-title>
          </mat-card-header>

          <mat-progress-bar *ngIf="loadingService.isLoading(loadingKeys.RESTORE_JOB_LIST)"
                           mode="indeterminate" class="loading-bar"></mat-progress-bar>

          <mat-card-content class="list-content">
            <div class="filter-toolbar">
              <mat-form-field appearance="outline" class="search-field">
                <mat-label>搜索集群</mat-label>
                <input matInput placeholder="输入集群名称" [(ngModel)]="searchTerm" (ngModelChange)="applyFilters()" />
                <button mat-icon-button matSuffix *ngIf="searchTerm" (click)="clearSearch()" aria-label="清空搜索">
                  <mat-icon>close</mat-icon>
                </button>
              </mat-form-field>

              <div class="filters-line">
                <div class="filter-group">
                  <span class="filter-label">状态:</span>
                  <mat-chip-listbox class="chip-group" [multiple]="false">
                    <mat-chip-option [selected]="statusFilter==='all'" (click)="setStatusFilter('all')">全部 ({{statusCounts.all}})</mat-chip-option>
                    <mat-chip-option [selected]="statusFilter==='ongoing'" (click)="setStatusFilter('ongoing')">进行中 ({{statusCounts.ongoing}})</mat-chip-option>
                    <mat-chip-option [selected]="statusFilter==='completed'" (click)="setStatusFilter('completed')">已完成 ({{statusCounts.completed}})</mat-chip-option>
                    <mat-chip-option [selected]="statusFilter==='failed'" (click)="setStatusFilter('failed')">失败 ({{statusCounts.failed}})</mat-chip-option>
                  </mat-chip-listbox>
                </div>

                <div class="filter-group">
                  <span class="filter-label">类型:</span>
                  <mat-chip-listbox class="chip-group" [multiple]="false">
                    <mat-chip-option [selected]="typeFilter==='all'" (click)="setTypeFilter('all')">全部</mat-chip-option>
                    <mat-chip-option [selected]="typeFilter==='backup'" (click)="setTypeFilter('backup')">备份恢复</mat-chip-option>
                    <mat-chip-option [selected]="typeFilter==='pitr'" (click)="setTypeFilter('pitr')">PITR</mat-chip-option>
                  </mat-chip-listbox>
                </div>

                <div class="toolbar-spacer"></div>
                <div class="meta">
                  <mat-slide-toggle [checked]="listAutoRefresh" (change)="toggleAutoRefresh($event.checked)">自动刷新</mat-slide-toggle>
                  <span class="last-updated">上次更新：{{ lastUpdated | date:'HH:mm:ss' }}</span>
                  <button mat-stroked-button (click)="refreshAll()">
                    <mat-icon>refresh</mat-icon>
                    刷新
                  </button>
                </div>
              </div>
            </div>

            <div class="table-container">
              <table mat-table [dataSource]="dataSource" class="restore-table mat-elevation-z1" matSort>
                <ng-container matColumnDef="clusterName">
                  <th mat-header-cell *matHeaderCellDef class="cluster-col" mat-sort-header>集群名称</th>
                  <td mat-cell *matCellDef="let j" class="cluster-cell">
                    <div class="cluster-info">
                      <mat-icon class="cluster-icon">storage</mat-icon>
                      <span class="cluster-name">{{ j.clusterName }}</span>
                    </div>
                  </td>
                </ng-container>
                <ng-container matColumnDef="phase">
                  <th mat-header-cell *matHeaderCellDef class="status-col" mat-sort-header>状态</th>
                  <td mat-cell *matCellDef="let j" class="status-cell">
                    <mat-chip [ngClass]="phaseClass(j.phase)" class="status-chip">
                      {{ j.phase || '-' }}
                    </mat-chip>
                  </td>
                </ng-container>
                <ng-container matColumnDef="restoreType">
                  <th mat-header-cell *matHeaderCellDef class="type-col" mat-sort-header>类型</th>
                  <td mat-cell *matCellDef="let j" class="type-cell">
                    <span class="restore-type">{{ getRestoreType(j) === 'pitr' ? 'PITR' : '备份恢复' }}</span>
                  </td>
                </ng-container>
                <ng-container matColumnDef="created">
                  <th mat-header-cell *matHeaderCellDef class="time-col" mat-sort-header>创建时间</th>
                  <td mat-cell *matCellDef="let j" class="time-cell">
                    <div class="time-info">
                      <span class="time-date">{{ j.creationTimestamp | date:'MM-dd' }}</span>
                      <span class="time-time">{{ j.creationTimestamp | date:'HH:mm' }}</span>
                    </div>
                  </td>
                </ng-container>
                <ng-container matColumnDef="actions">
                  <th mat-header-cell *matHeaderCellDef class="actions-col">操作</th>
                  <td mat-cell *matCellDef="let j" class="actions-cell" (click)="$event.stopPropagation()">
                    <button mat-icon-button (click)="selectJob(j)" class="action-btn" matTooltip="查看详情" aria-label="查看详情">
                      <mat-icon>visibility</mat-icon>
                    </button>
                    <button mat-icon-button color="warn" (click)="cancelJob(j)" class="action-btn" matTooltip="取消任务" aria-label="取消任务" [disabled]="!j.canCancel">
                      <mat-icon>cancel</mat-icon>
                    </button>
                  </td>
                </ng-container>
                <tr mat-header-row *matHeaderRowDef="displayedColumns" class="table-header"></tr>
                <tr mat-row *matRowDef="let row; columns: displayedColumns;" (click)="selectJob(row)" class="table-row" [class.selected]="selectedJob?.clusterName === row.clusterName"></tr>
              </table>
              <div *ngIf="dataSource.data.length === 0" class="empty-list">
                <div class="empty-inner">
                  <mat-icon class="empty-icon">inbox</mat-icon>
                  <p class="empty-text">暂无恢复任务</p>
                </div>
              </div>
              <mat-paginator [length]="dataSource?.filteredData?.length || 0"
                             [pageSize]="pageSize" [pageSizeOptions]="[10, 25, 50]"
                             showFirstLastButtons>
              </mat-paginator>
            </div>
          </mat-card-content>
        </mat-card>
      </div>

      <!-- 右侧详情面板 -->
      <div class="detail-panel" *ngIf="selectedJob; else emptyState">
        <mat-card class="detail-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>article</mat-icon>
              任务详情 · {{ selectedJob?.clusterName }}
            </mat-card-title>
            <mat-card-actions align="end">
              <a *ngIf="grafanaLinkForSelected() as gLink; else noGrafana"
                 [href]="gLink" target="_blank" rel="noopener" mat-stroked-button>
                <mat-icon>open_in_new</mat-icon>
                在 Grafana 打开
              </a>
              <ng-template #noGrafana></ng-template>
              <button mat-stroked-button (click)="reloadSelected()">
                <mat-icon>refresh</mat-icon>
                刷新
              </button>
              <button mat-stroked-button (click)="toggleRaw()">
                <mat-icon>code</mat-icon>
                原始JSON
              </button>
              <button mat-stroked-button (click)="copySelectedJson()">
                <mat-icon>content_copy</mat-icon>
                复制
              </button>
              <button mat-stroked-button color="warn" [disabled]="!selectedJob?.canCancel" (click)="cancelSelected()">
                <mat-icon>cancel</mat-icon>
                取消
              </button>
              <button mat-raised-button color="primary" *ngIf="isFailed(selectedJob)" (click)="diagnoseSelected()">
                <mat-icon>medical_information</mat-icon>
                诊断
              </button>
            </mat-card-actions>
          </mat-card-header>
          
          <mat-card-content class="detail-content">
            <div class="status-section">
              <div class="status-header">
                <mat-chip [ngClass]="phaseClass(selectedJob?.phase)" class="status-chip-large">
                  {{ selectedJob?.phase || '-' }}
                </mat-chip>
                <span class="progress-text">{{ getProgress(selectedJob) }}%</span>
              </div>
              <mat-progress-bar mode="determinate" [value]="getProgress(selectedJob)" class="progress-bar" [ngClass]="getProgressClass(selectedJob)"></mat-progress-bar>
            </div>

            <div class="info-section">
              <h4 class="section-title"><mat-icon>info</mat-icon> 基本信息</h4>
              <div class="info-grid">
                <div class="info-item"><span class="info-label">源集群</span><span class="info-value">{{ selectedJob?.sourceCluster || '-' }}</span></div>
                <div class="info-item"><span class="info-label">命名空间</span><span class="info-value">{{ selectedJob?.namespace || '-' }}</span></div>
                <div class="info-item"><span class="info-label">恢复类型</span><span class="info-value">{{ getRestoreType(selectedJob!) === 'pitr' ? 'PITR恢复' : '备份恢复' }}</span></div>
                <div class="info-item"><span class="info-label">当前阶段</span><span class="info-value">{{ selectedJob?.stage || '-' }}</span></div>
                <div class="info-item"><span class="info-label">开始时间</span><span class="info-value">{{ selectedJob?.creationTimestamp | date:'yyyy-MM-dd HH:mm:ss' }}</span></div>
                <div class="info-item"><span class="info-label">可取消</span><span class="info-value">{{ selectedJob?.canCancel ? '是' : '否' }}</span></div>
              </div>
            </div>

            <div class="conditions-section" *ngIf="selectedJob?.conditions?.length">
              <h4 class="section-title"><mat-icon>event_note</mat-icon> 状态条件</h4>
              <div class="conditions-list">
                <div class="condition-item" *ngFor="let c of selectedJob?.conditions">
                  <div class="condition-header">
                    <span class="condition-type">{{ c.type }}</span>
                    <span class="condition-status" [ngClass]="getConditionStatusClass(c.status)">{{ c.status }}</span>
                    <span class="condition-time">{{ c.lastTransitionTime | date:'MM-dd HH:mm:ss' }}</span>
                  </div>
                  <div class="condition-details" *ngIf="c.reason || c.message">
                    <span class="condition-reason" *ngIf="c.reason">{{ c.reason }}</span>
                    <span class="condition-message" *ngIf="c.message">{{ c.message }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div *ngIf="!selectedJob?.conditions?.length" class="no-conditions">
              <mat-icon>info_outline</mat-icon>
              <span>暂无状态条件</span>
            </div>

            <div class="raw-section" *ngIf="showRaw">
              <pre class="raw-json">{{ selectedJob | json }}</pre>
            </div>
          </mat-card-content>
        </mat-card>
      </div>

      <!-- 空状态 -->
      <ng-template #emptyState>
        <div class="empty-state">
          <mat-icon class="empty-state-icon">touch_app</mat-icon>
          <h3>选择任务查看详情</h3>
          <p>点击左侧任务列表中的任意一行来查看详细信息</p>
        </div>
      </ng-template>
    </div>
  `,
  styles: [`
    .restore-page { display: grid; grid-template-columns: 1fr; gap: 16px; padding: 20px; background: #ffffff; box-sizing: border-box; }
    .page-header-card { margin-bottom: 4px; grid-column: 1 / -1; }
    .list-panel { width: 100%; }
    .detail-panel { width: 100%; }

    @media (min-width: 1200px) {
      .restore-page { grid-template-columns: 52% 1fr; }
      .list-panel { min-width: 620px; }
    }

    .list-card, .detail-card { box-shadow: 0 4px 12px rgba(0,0,0,0.06); border-radius: 12px; }
    .loading-bar { height: 3px; }
    .list-content { padding: 0 8px 8px 8px; }

    .filter-toolbar { display: grid; gap: 12px; padding: 12px; background: #fff; border-bottom: 1px solid #eee; border-radius: 8px; }
    .filters-line { display: flex; gap: 16px; align-items: center; flex-wrap: wrap; }
    .filter-group { display: flex; align-items: center; gap: 8px; }
    .filter-label { color: #666; font-size: 13px; }
    .chip-group .mat-mdc-chip { cursor: pointer; }
    .toolbar-spacer { flex: 1; }
    .meta { display: flex; align-items: center; gap: 8px; color: #999; }
    .last-updated { font-size: 12px; }
    .search-field { width: 100%; max-width: 360px; }

    .table-container { max-height: calc(100vh - 360px); overflow-y: auto; background: #fff; border: 1px solid #e5e7eb; border-radius: 10px; }
    .restore-table { width: 100%; background: white; }
    .table-header { background: #fafafa; font-weight: 600; color: #333; }
    .table-container .mat-mdc-header-row { position: sticky; top: 0; z-index: 2; background: #fafafa; border-bottom: 1px solid #e5e7eb; }
    .table-row { cursor: pointer; transition: background .2s ease; border-bottom: 1px solid #f0f0f0; }
    .table-row:hover { background: #f8f9ff; }
    .table-row.selected { background: #e3f2fd; border-left: 4px solid #2196f3; }

    .cluster-info { display: flex; align-items: center; gap: 8px; }
    .cluster-icon { color: #666; font-size: 20px; }
    .status-chip { font-weight: 500; border-radius: 16px; padding: 4px 12px; font-size: 12px; }

    .table-container mat-paginator { border-top: 1px solid #e5e7eb; padding: 4px 8px; background: #fff; }

    .detail-content { padding: 16px 24px 24px; }
    .status-section { margin-bottom: 16px; padding: 16px; background: #f8f9fa; border-radius: 12px; border: 1px solid #e9ecef; }
    .status-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
    .status-chip-large { font-weight: 600; font-size: 14px; padding: 8px 16px; border-radius: 20px; }
    .progress-text { font-weight: 600; color: #666; font-size: 16px; }
    .progress-bar { height: 8px; border-radius: 4px; }

    .info-section { margin-bottom: 16px; }
    .section-title { display: flex; align-items: center; gap: 8px; font-size: 16px; font-weight: 600; color: #333; margin: 8px 0 12px; padding-bottom: 8px; border-bottom: 2px solid #e3f2fd; }
    .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; }
    .info-item { display: flex; flex-direction: column; gap: 4px; padding: 12px; background: #fafafa; border-radius: 8px; border-left: 4px solid #2196f3; }
    .info-label { font-size: 12px; color: #666; font-weight: 500; letter-spacing: .5px; }
    .info-value { font-size: 14px; color: #333; font-weight: 500; }

    .conditions-section { margin-bottom: 16px; }
    .conditions-list { display: flex; flex-direction: column; gap: 12px; }
    .condition-item { padding: 16px; background: #fafafa; border-radius: 8px; border-left: 4px solid #ff9800; }
    .condition-header { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
    .condition-type { font-weight: 600; color: #333; flex: 1; }
    .condition-status { padding: 4px 8px; border-radius: 12px; font-size: 12px; font-weight: 500; }
    .condition-status.true { background: #e8f5e8; color: #2e7d32; }
    .condition-status.false { background: #ffebee; color: #c62828; }
    .condition-time { font-size: 12px; color: #666; }

    .no-conditions { display: flex; align-items: center; gap: 8px; padding: 20px; color: #999; justify-content: center; background: #fafafa; }

    .raw-section { margin-top: 16px; }
    .raw-json { max-height: 280px; overflow: auto; background: #0f172a; color: #e2e8f0; padding: 12px; border-radius: 8px; font-size: 12px; }

    .phase-warn { background: #ffebee !important; color: #c62828 !important; }
    .phase-primary { background: #e3f2fd !important; color: #1976d2 !important; }
    .phase-accent { background: #f5f5f5 !important; color: #666 !important; }

    /* Empty list styling */
    .empty-list { display: flex; justify-content: center; align-items: center; height: 240px; }
    .empty-inner { display: flex; flex-direction: column; align-items: center; gap: 8px; color: #9aa1a9; }
    .empty-icon { font-size: 44px; width: 44px; height: 44px; color: #c0c6cc; }
    .empty-text { margin: 0; font-size: 14px; color: #97a0aa; }
  `]
})
export class RestoreJobManagementComponent implements OnInit, AfterViewInit, OnDestroy {
  loadingKeys = LoadingKeys;
  displayedColumns = ['clusterName', 'phase', 'restoreType', 'created', 'actions'];
  dataSource = new MatTableDataSource<RestoreJob>([]);
  selectedJob?: RestoreJobWithStatus;
  private pollingSub?: Subscription;
  private listPollingSub?: Subscription;

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  allJobs: RestoreJob[] = [];
  searchTerm = '';
  statusFilter: 'all' | 'ongoing' | 'completed' | 'failed' = 'all';
  typeFilter: 'all' | 'backup' | 'pitr' = 'all';
  statusCounts = { all: 0, ongoing: 0, completed: 0, failed: 0 };
  lastUpdated: Date = new Date();
  showRaw = false;
  listAutoRefresh = true;
  pageSize = 10;

  constructor(
    private apiService: ApiService,
    public loadingService: LoadingService,
    private snackBar: MatSnackBar
  ) {
    this.dataSource.sortingDataAccessor = (item: any, property: string) => {
      switch (property) {
        case 'clusterName': return (item.clusterName || '').toLowerCase();
        case 'phase': return this.mapPhaseOrder(item.phase);
        case 'restoreType': return this.getRestoreType(item) === 'pitr' ? 1 : 0;
        case 'created': return new Date(item.creationTimestamp || 0).getTime();
        default: return item[property];
      }
    };
  }

  ngOnInit(): void {
    this.loadJobs();
    this.startListPolling();
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
  }

  loadJobs(): void {
    this.apiService.listRestoreJobs().subscribe({
      next: (jobs) => {
        const mapped = (jobs || []).map(j => ({
          ...j,
          phase: this.mapPhaseToCN(j.phase)
        } as RestoreJob));
        this.allJobs = mapped;
        this.computeCounts();
        this.applyFilters();
        this.lastUpdated = new Date();
        if (this.selectedJob && this.isTerminal(this.selectedJob)) {
          this.stopPolling();
        }
      },
      error: () => { this.allJobs = []; this.dataSource.data = []; this.computeCounts(); }
    });
  }

  refreshAll(): void { this.loadJobs(); }
  clearSearch(): void { this.searchTerm = ''; this.applyFilters(); }
  setStatusFilter(k: 'all'|'ongoing'|'completed'|'failed'): void { this.statusFilter = k; this.applyFilters(); }
  setTypeFilter(k: 'all'|'backup'|'pitr'): void { this.typeFilter = k; this.applyFilters(); }

  toggleAutoRefresh(checked: boolean): void {
    this.listAutoRefresh = checked;
    if (checked) this.startListPolling(); else this.stopListPolling();
  }

  startListPolling(): void {
    this.stopListPolling();
    if (!this.listAutoRefresh) return;
    this.listPollingSub = interval(10000).subscribe(() => this.loadJobs());
  }

  stopListPolling(): void {
    if (this.listPollingSub) { this.listPollingSub.unsubscribe(); this.listPollingSub = undefined; }
  }

  applyFilters(): void {
    let result = [...this.allJobs];
    const term = this.searchTerm.trim().toLowerCase();
    if (term) {
      result = result.filter(j => (j.clusterName || '').toLowerCase().includes(term));
    }
    if (this.typeFilter !== 'all') {
      result = result.filter(j => this.typeFilter === 'pitr' ? this.getRestoreType(j) === 'pitr' : this.getRestoreType(j) !== 'pitr');
    }
    if (this.statusFilter !== 'all') {
      result = result.filter(j => {
        const p = (j.phase || '');
        if (this.statusFilter === 'ongoing') return p === '已提交' || p === '创建中';
        if (this.statusFilter === 'completed') return p === '已完成';
        if (this.statusFilter === 'failed') return p === '失败';
        return true;
      });
    }
    this.dataSource.data = result;
    if (this.paginator) this.paginator.firstPage();
  }

  computeCounts(): void {
    const jobs = this.allJobs;
    const ongoing = jobs.filter(j => (j.phase || '') === '已提交' || (j.phase || '') === '创建中').length;
    const completed = jobs.filter(j => (j.phase || '') === '已完成').length;
    const failed = jobs.filter(j => (j.phase || '') === '失败').length;
    this.statusCounts = { all: jobs.length, ongoing, completed, failed };
  }

  selectJob(job: RestoreJob): void {
    this.selectedJob = job as RestoreJobWithStatus;
    this.startPolling();
    this.reloadSelected();
  }

  cancelJob(job: RestoreJob): void {
    if (!confirm(`确定取消恢复任务 (cluster=${job.clusterName}) 吗？`)) return;
    this.apiService.cancelRestoreJob(job.namespace || 'default', job.clusterName).subscribe({
      next: () => { this.snackBar.open('取消请求已提交', '关闭', { duration: 2000 }); this.loadJobs(); },
      error: () => { this.snackBar.open('取消失败', '关闭', { duration: 2500 }); }
    });
  }

  phaseClass(phase?: string): string {
    const p = (phase || '').toLowerCase();
    if (p === 'failed') return 'phase-warn';
    if (p === 'completed' || p === 'running' || p === '已完成' || p === '运行中' || p === '创建中' || p === '恢复中') return 'phase-primary';
    return 'phase-accent';
  }

  private mapPhaseToCN(phase?: string): string {
    const p = (phase || '').toLowerCase();
    switch (p) {
      case 'pending':
        return '已提交';
      case 'restoring':
      case 'creating':
        return '创建中';
      case 'running':
      case 'completed':
        return '已完成';
      case 'failed':
        return '失败';
      default:
        return phase || '-';
    }
  }

  private mapPhaseOrder(phase?: string): number {
    const p = (phase || '').toLowerCase();
    if (p === 'failed' || p === '失败') return 3;
    if (p === 'completed' || p === '已完成') return 2;
    if (p === 'creating' || p === 'restoring' || p === '创建中' || p === '恢复中') return 1;
    return 0; // 已提交/其他
  }

  getProgress(job?: RestoreJob): number {
    const p = (job?.phase || '').toLowerCase();
    switch (p) {
      case 'pending':
      case '已提交':
        return 10;
      case 'creating':
      case 'restoring':
      case '创建中':
      case '恢复中':
        return 60;
      case 'running':
      case 'completed':
      case '已完成':
        return 100;
      case 'failed':
      case '失败':
        return 0;
    }
    return 30;
  }

  reloadSelected(): void {
    if (!this.selectedJob) return;
    this.apiService.getRestoreJob(this.selectedJob.namespace || 'default', this.selectedJob.clusterName).subscribe({
      next: (full) => {
        this.selectedJob = full;
        const idx = this.allJobs.findIndex(j => j.clusterName === full.clusterName && j.namespace === full.namespace);
        if (idx >= 0) {
          const clone = [...this.allJobs];
          clone[idx] = full as any;
          this.allJobs = clone;
          this.computeCounts();
          this.applyFilters();
        }
        if (this.isTerminal(full)) this.stopPolling();
      }
    });
  }

  startPolling(): void {
    this.stopPolling();
    if (!this.selectedJob) return;
    this.pollingSub = interval(5000).pipe(
      switchMap(() => this.apiService.getRestoreJob(this.selectedJob!.namespace || 'default', this.selectedJob!.clusterName))
    ).subscribe({ next: (full) => {
      this.selectedJob = full;
      if (this.isTerminal(full)) this.stopPolling();
    }});
  }

  stopPolling(): void {
    if (this.pollingSub) { this.pollingSub.unsubscribe(); this.pollingSub = undefined; }
  }

  cancelSelected(): void {
    if (!this.selectedJob) return;
    this.cancelJob(this.selectedJob);
  }

  isTerminal(job: RestoreJob): boolean {
    const p = (job.phase || '').toLowerCase();
    return p === 'failed' || p === 'completed' || p === '已完成' || p === '失败';
  }

  getProgressClass(job?: RestoreJob): string {
    const p = (job?.phase || '').toLowerCase();
    if (p === 'failed' || p === '失败') return 'phase-warn';
    if (p === 'completed' || p === '已完成' || p === 'running') return 'phase-primary';
    return 'phase-accent';
  }

  getConditionStatusClass(status?: string): string {
    return status?.toLowerCase() || 'unknown';
  }

  getRestoreType(job: RestoreJob | RestoreJobWithStatus): 'pitr' | 'backup' {
    return job?.restoreSpec?.time ? 'pitr' : 'backup';
  }

  toggleRaw(): void { this.showRaw = !this.showRaw; }

  copySelectedJson(): void {
    if (!this.selectedJob) return;
    const text = JSON.stringify(this.selectedJob, null, 2);
    if (navigator?.clipboard?.writeText) {
      navigator.clipboard.writeText(text).then(() => {
        this.snackBar.open('已复制 JSON', '关闭', { duration: 2000 });
      }).catch(() => this.snackBar.open('复制失败', '关闭', { duration: 2000 }));
    } else {
      const ta = document.createElement('textarea');
      ta.value = text;
      document.body.appendChild(ta);
      ta.select();
      try { document.execCommand('copy'); this.snackBar.open('已复制 JSON', '关闭', { duration: 2000 }); } catch {}
      document.body.removeChild(ta);
    }
  }

  isFailed(job?: RestoreJob): boolean {
    const p = (job?.phase || '').toLowerCase();
    return p === 'failed' || p === '失败';
  }

  diagnoseSelected(): void {
    if (!this.selectedJob) return;
    const ns = this.selectedJob.namespace || 'default';
    const cluster = this.selectedJob.clusterName;
    // 跳转到诊断页并自动触发，便于查看历史与下载
    const url = `/operations/diagnostics?namespace=${encodeURIComponent(ns)}&cluster=${encodeURIComponent(cluster)}&autoStart=1`;
    window.location.href = url;
  }

  grafanaLinkForSelected(): string | null {
    const base = localStorage.getItem('grafanaURL') || '';
    if (!base || !this.selectedJob) return null;
    const ns = this.selectedJob.namespace || 'default';
    const cluster = this.selectedJob.clusterName || '';
    const clusterParam = cluster ? `&var-cluster=${encodeURIComponent(cluster)}` : '';
    return `${base}/d/polardbx-monitor?orgId=1&var-namespace=${encodeURIComponent(ns)}${clusterParam}`;
  }

  ngOnDestroy(): void {
    this.stopPolling();
    this.stopListPolling();
  }
}