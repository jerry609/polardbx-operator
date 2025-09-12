import { Component, OnInit, OnDestroy, TemplateRef, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatTabsModule } from '@angular/material/tabs';
import { MatTableModule } from '@angular/material/table';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBarModule, MatSnackBar } from '@angular/material/snack-bar';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatDividerModule } from '@angular/material/divider';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatMenuModule } from '@angular/material/menu';
import { Subject } from 'rxjs';
import { EmptyStateComponent } from '../empty-state/empty-state.component';
import { takeUntil, finalize } from 'rxjs/operators';

import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { 
  PolarDBXBackupSchedule, 
  PolarDBXBackupScheduleList,
  CreateBackupScheduleRequest,
  UpdateBackupScheduleRequest,
  PREDEFINED_CRON_SCHEDULES,
  STORAGE_PROVIDER_OPTIONS,
  CLEAN_POLICY_OPTIONS,
  BACKUP_ROLE_OPTIONS,
  BackupStorage,
  CleanPolicyType
} from '../../models/backup-schedule.model';
// Removed ConfirmationDialogComponent import - using native confirm() instead

@Component({
  selector: 'app-backup-schedule-management',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatTabsModule,
    MatTableModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatSelectModule,
    MatFormFieldModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    MatDialogModule,
    MatTooltipModule,
    MatSlideToggleModule,
    MatDividerModule,
    MatExpansionModule,
    MatMenuModule,
    EmptyStateComponent
  ],
  template: `
    <div class="backup-schedule-management">
      <mat-card class="header-card">
        <mat-card-header>
          <mat-card-title>
            <mat-icon>event</mat-icon>
            定时备份管理
          </mat-card-title>
          <mat-card-subtitle>
            为 PolarDB-X 集群配置自动备份计划
          </mat-card-subtitle>
        </mat-card-header>
      </mat-card>

      <mat-tab-group class="main-tabs" [(selectedIndex)]="selectedTab" (selectedTabChange)="onTabChange($event)">
        <!-- 定时备份列表选项卡 -->
        <mat-tab label="备份计划">
          <ng-template matTabContent>
            <div class="tab-content">
              <div class="actions-toolbar">
                <button mat-raised-button color="primary" (click)="refreshSchedules()" 
                        [disabled]="isLoading('BACKUP_SCHEDULE_LIST')">
                  <mat-icon>refresh</mat-icon>
                  刷新
                </button>
                <button mat-raised-button color="accent" (click)="selectedTab = 1">
                  <mat-icon>add</mat-icon>
                  创建计划
                </button>
                <a *ngIf="grafanaURL" [href]="grafanaURL" target="_blank" rel="noopener" mat-button>
                  <mat-icon>open_in_new</mat-icon>
                  在 Grafana 打开
                </a>
              </div>

              <mat-card class="table-card">
                <div class="table-container" *ngIf="!isLoading('BACKUP_SCHEDULE_LIST'); else loadingTemplate">
                  <table mat-table [dataSource]="backupSchedules" class="schedules-table">
                    <!-- 名称列 -->
                    <ng-container matColumnDef="name">
                      <th mat-header-cell *matHeaderCellDef>名称</th>
                      <td mat-cell *matCellDef="let schedule">
                        <div class="schedule-name">
                          <mat-icon [color]="schedule.spec.suspend ? 'warn' : 'primary'">
                            {{ schedule.spec.suspend ? 'pause_circle_outline' : 'schedule' }}
                          </mat-icon>
                          <span>{{ schedule.metadata.name }}</span>
                        </div>
                      </td>
                    </ng-container>

                    <!-- 命名空间列 -->
                    <ng-container matColumnDef="namespace">
                      <th mat-header-cell *matHeaderCellDef>命名空间</th>
                      <td mat-cell *matCellDef="let schedule">{{ schedule.metadata.namespace }}</td>
                    </ng-container>

                    <!-- 目标集群列 -->
                    <ng-container matColumnDef="cluster">
                      <th mat-header-cell *matHeaderCellDef>目标集群</th>
                      <td mat-cell *matCellDef="let schedule">
                        <mat-chip color="primary">{{ schedule.spec.backupSpec.cluster.name }}</mat-chip>
                      </td>
                    </ng-container>

                    <!-- 计划列 -->
                    <ng-container matColumnDef="schedule">
                      <th mat-header-cell *matHeaderCellDef>计划</th>
                      <td mat-cell *matCellDef="let schedule">
                        <div class="schedule-info">
                          <code>{{ schedule.spec.schedule }}</code>
                          <small>{{ getCronDescription(schedule.spec.schedule) }}</small>
                        </div>
                      </td>
                    </ng-container>

                    <!-- 状态列 -->
                    <ng-container matColumnDef="status">
                      <th mat-header-cell *matHeaderCellDef>状态</th>
                      <td mat-cell *matCellDef="let schedule">
                        <mat-chip [color]="schedule.spec.suspend ? 'warn' : 'primary'">
                          {{ schedule.spec.suspend ? '已暂停' : '活动' }}
                        </mat-chip>
                      </td>
                    </ng-container>

                    <!-- 上次备份列 -->
                    <ng-container matColumnDef="lastBackup">
                      <th mat-header-cell *matHeaderCellDef>上次备份</th>
                      <td mat-cell *matCellDef="let schedule">
                        <div class="backup-time">
                          <div>{{ formatDate(schedule.status?.lastBackupTime) }}</div>
                          <small>下次: {{ formatDate(schedule.status?.nextBackupTime) }}</small>
                        </div>
                      </td>
                    </ng-container>

                    <!-- 存储列 -->
                    <ng-container matColumnDef="storage">
                      <th mat-header-cell *matHeaderCellDef>存储</th>
                      <td mat-cell *matCellDef="let schedule">
                        <div class="storage-info">
                          <mat-icon>{{ getStorageIcon(schedule.spec.backupSpec.storageProvider?.storageName) }}</mat-icon>
                          <span>{{ getStorageLabel(schedule.spec.backupSpec.storageProvider?.storageName) }}</span>
                        </div>
                      </td>
                    </ng-container>

                    <!-- 操作列 -->
                    <ng-container matColumnDef="actions">
                      <th mat-header-cell *matHeaderCellDef>操作</th>
                      <td mat-cell *matCellDef="let schedule">
                        <button mat-icon-button [matMenuTriggerFor]="scheduleMenu" 
                                [disabled]="isLoading('BACKUP_SCHEDULE_UPDATE')">
                          <mat-icon>more_vert</mat-icon>
                        </button>
                        <mat-menu #scheduleMenu="matMenu">
                          <button mat-menu-item (click)="viewScheduleDetails(schedule)">
                            <mat-icon>visibility</mat-icon>
                            查看详情
                          </button>
                          <button mat-menu-item (click)="editSchedule(schedule)">
                            <mat-icon>edit</mat-icon>
                            编辑计划
                          </button>
                          <button mat-menu-item (click)="toggleSchedule(schedule)">
                            <mat-icon>{{ schedule.spec.suspend ? 'play_arrow' : 'pause' }}</mat-icon>
                            {{ schedule.spec.suspend ? '恢复' : '暂停' }}
                          </button>
                          <button mat-menu-item (click)="runNow(schedule)">
                            <mat-icon>play_arrow</mat-icon>
                            立即执行一次
                          </button>
                          <button mat-menu-item (click)="deleteSchedule(schedule)" class="delete-action">
                            <mat-icon>delete</mat-icon>
                            删除计划
                          </button>
                        </mat-menu>
                      </td>
                    </ng-container>

                    <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
                    <tr mat-row *matRowDef="let row; columns: displayedColumns;"></tr>
                  </table>

                  <app-empty-state *ngIf="backupSchedules.length === 0"
                                   icon="schedule"
                                   title="未找到备份计划"
                                   hint="点击“创建计划”开始使用"></app-empty-state>
                </div>
              </mat-card>
            </div>
          </ng-template>
        </mat-tab>

        <!-- 创建/编辑计划选项卡 -->
        <mat-tab label="创建计划">
          <ng-template matTabContent>
            <div class="tab-content">
              <mat-card class="form-card">
                <mat-card-header>
                  <mat-card-title>
                    {{ editingSchedule ? '编辑备份计划' : '创建新备份计划' }}
                  </mat-card-title>
                  <mat-card-subtitle>
                    通过灵活的计划和存储选项配置自动备份
                  </mat-card-subtitle>
                </mat-card-header>

                <mat-card-content>
                  <form [formGroup]="scheduleForm" class="schedule-form">
                    <!-- 基本配置 -->
                    <mat-expansion-panel class="form-section" [expanded]="true">
                      <mat-expansion-panel-header>
                        <mat-panel-title>基本配置</mat-panel-title>
                      </mat-expansion-panel-header>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>计划名称</mat-label>
                          <input matInput formControlName="name" placeholder="输入计划名称">
                          <mat-error *ngIf="scheduleForm.get('name')?.hasError('required')">
                            计划名称是必填项
                          </mat-error>
                          <mat-error *ngIf="scheduleForm.get('name')?.hasError('pattern')">
                            名称必须是合法的 Kubernetes 资源名称
                          </mat-error>
                        </mat-form-field>
                      </div>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>命名空间</mat-label>
                          <input matInput formControlName="namespace" placeholder="default">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>目标集群</mat-label>
                          <mat-select formControlName="clusterName" placeholder="选择集群">
                            <mat-option *ngFor="let name of availableClusters" [value]="name">{{ name }}</mat-option>
                          </mat-select>
                          <mat-error *ngIf="scheduleForm.get('clusterName')?.hasError('required')">
                            集群名称是必填项
                          </mat-error>
                        </mat-form-field>
                      </div>
                    </mat-expansion-panel>

                    <!-- 计划配置 -->
                    <mat-expansion-panel class="form-section">
                      <mat-expansion-panel-header>
                        <mat-panel-title>计划配置</mat-panel-title>
                      </mat-expansion-panel-header>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>预定义计划</mat-label>
                          <mat-select (selectionChange)="applyPredefinedSchedule($event.value)">
                            <mat-option [value]="null">自定义 Cron 表达式</mat-option>
                            <mat-option *ngFor="let option of predefinedSchedules" [value]="option.value">
                              {{ option.label }} - {{ option.description }}
                            </mat-option>
                          </mat-select>
                        </mat-form-field>
                      </div>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>Cron 表达式</mat-label>
                          <input matInput formControlName="schedule" placeholder="0 2 * * *">
                          <mat-hint>格式: 分 时 日 月 周 (例如 "0 2 * * *" 表示每天凌晨2点)</mat-hint>
                          <mat-error *ngIf="scheduleForm.get('schedule')?.hasError('required')">
                            计划是必填项
                          </mat-error>
                        </mat-form-field>
                      </div>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>最大备份数</mat-label>
                          <input matInput type="number" formControlName="maxBackupCount" 
                                 placeholder="保留最近的备份" min="1">
                          <mat-hint>要保留的最大备份数量</mat-hint>
                        </mat-form-field>
                        <div class="half-width">
                          <mat-slide-toggle formControlName="suspend">
                            暂停计划
                          </mat-slide-toggle>
                        </div>
                      </div>
                    </mat-expansion-panel>

                    <!-- 存储配置 -->
                    <mat-expansion-panel class="form-section">
                      <mat-expansion-panel-header>
                        <mat-panel-title>存储配置</mat-panel-title>
                      </mat-expansion-panel-header>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>存储提供商</mat-label>
                          <mat-select formControlName="storageProvider">
                            <mat-option *ngFor="let option of storageProviders" [value]="option.value">
                              <mat-icon>{{ getStorageIcon(option.value) }}</mat-icon>
                              {{ option.label }}
                            </mat-option>
                          </mat-select>
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>存储 Sink</mat-label>
                          <input matInput formControlName="storageSink" 
                                 placeholder="存储端点/路径">
                          <mat-hint>存储端点 URL 或路径配置</mat-hint>
                        </mat-form-field>
                      </div>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>清理策略</mat-label>
                          <mat-select formControlName="cleanPolicy">
                            <mat-option *ngFor="let option of cleanPolicyOptions" [value]="option.value">
                              {{ option.label }}
                            </mat-option>
                          </mat-select>
                          <mat-hint>完成后如何处理备份文件</mat-hint>
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half-width">
                          <mat-label>首选备份角色</mat-label>
                          <mat-select formControlName="preferredBackupRole">
                            <mat-option *ngFor="let option of backupRoleOptions" [value]="option.value">
                              {{ option.label }}
                            </mat-option>
                          </mat-select>
                          <mat-hint>使用哪个数据库角色进行备份</mat-hint>
                        </mat-form-field>
                      </div>

                      <div class="form-row">
                        <mat-form-field appearance="outline" class="full-width">
                          <mat-label>保留时间 (天)</mat-label>
                          <input matInput type="number" formControlName="retentionDays" 
                                 placeholder="30" min="1">
                          <mat-hint>备份文件保留多长时间 (以天为单位)</mat-hint>
                        </mat-form-field>
                      </div>
                    </mat-expansion-panel>
                  </form>
                </mat-card-content>

                <mat-card-actions align="end">
                  <button mat-button (click)="resetForm()" [disabled]="isLoading('BACKUP_SCHEDULE_CREATE')">
                    重置
                  </button>
                  <button mat-raised-button color="primary" 
                          (click)="submitSchedule()" 
                          [disabled]="scheduleForm.invalid || isLoading('BACKUP_SCHEDULE_CREATE')">
                    <mat-icon>{{ editingSchedule ? 'save' : 'add' }}</mat-icon>
                    {{ editingSchedule ? '更新计划' : '创建计划' }}
                  </button>
                </mat-card-actions>
              </mat-card>
            </div>
          </ng-template>
        </mat-tab>
      </mat-tab-group>
    </div>

    <!-- 加载模板 -->
    <ng-template #loadingTemplate>
      <div class="loading-container">
        <mat-progress-spinner mode="indeterminate"></mat-progress-spinner>
        <p>正在加载备份计划...</p>
      </div>
    </ng-template>

    <!-- 详情对话框模板 -->
    <ng-template #detailsDialog let-data>
      <h2 mat-dialog-title>备份计划详情</h2>
      <mat-dialog-content>
        <div class="detail-row"><strong>名称:</strong> {{ data?.metadata?.name }}</div>
        <div class="detail-row"><strong>命名空间:</strong> {{ data?.metadata?.namespace }}</div>
        <div class="detail-row"><strong>目标集群:</strong> {{ data?.spec?.backupSpec?.cluster?.name }}</div>
        <div class="detail-row"><strong>Cron:</strong> <code>{{ data?.spec?.schedule }}</code> ({{ getCronDescription(data?.spec?.schedule) }})</div>
        <div class="detail-row"><strong>状态:</strong>
          <mat-chip [color]="data?.spec?.suspend ? 'warn' : 'primary'">{{ data?.spec?.suspend ? '已暂停' : '活动' }}</mat-chip>
        </div>
        <div class="detail-row"><strong>上次备份:</strong> {{ formatDate(data?.status?.lastBackupTime) }}</div>
        <div class="detail-row"><strong>下次备份:</strong> {{ formatDate(data?.status?.nextBackupTime) }}</div>
        <div class="detail-row"><strong>存储:</strong> {{ getStorageLabel(data?.spec?.backupSpec?.storageProvider?.storageName) }}
          <small style="margin-left:6px">{{ data?.spec?.backupSpec?.storageProvider?.sink }}</small>
        </div>
        <div class="detail-row"><strong>清理策略:</strong> {{ data?.spec?.backupSpec?.cleanPolicy || '保留' }}</div>
        <div class="detail-row"><strong>首选角色:</strong> {{ data?.spec?.backupSpec?.preferredBackupRole || '-' }}</div>
        <div class="detail-row" *ngIf="grafanaLinkFor(data) as gLink">
          <a [href]="gLink" target="_blank" rel="noopener">
            <mat-icon style="vertical-align:middle; margin-right:4px">open_in_new</mat-icon>在 Grafana 打开
          </a>
        </div>
      </mat-dialog-content>
      <mat-dialog-actions align="end">
        <button mat-button mat-dialog-close>关闭</button>
      </mat-dialog-actions>
    </ng-template>
  `,
  styleUrl: './backup-schedule-management.component.scss'
})
export class BackupScheduleManagementComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  
  backupSchedules: PolarDBXBackupSchedule[] = [];
  displayedColumns: string[] = ['name', 'namespace', 'cluster', 'schedule', 'status', 'lastBackup', 'storage', 'actions'];
  selectedTab = 0;
  editingSchedule: PolarDBXBackupSchedule | null = null;
  
  predefinedSchedules = PREDEFINED_CRON_SCHEDULES;
  storageProviders = STORAGE_PROVIDER_OPTIONS;
  cleanPolicyOptions = CLEAN_POLICY_OPTIONS;
  backupRoleOptions = BACKUP_ROLE_OPTIONS;
  
  scheduleForm: FormGroup;
  availableClusters: string[] = [];
  grafanaURL: string = '';

  constructor(
    private apiService: ApiService,
    private loadingService: LoadingService,
    private fb: FormBuilder,
    private snackBar: MatSnackBar,
    private dialog: MatDialog
  ) {
    this.scheduleForm = this.createScheduleForm();
    const defaultStorage = (localStorage.getItem('backupStorageName') || 's3') as BackupStorage;
    const defaultSink = (localStorage.getItem('backupSinkName') || localStorage.getItem('xstoreBackupSinkName') || 'lyfz-polardbx-backup');
    this.scheduleForm.patchValue({ storageProvider: defaultStorage, storageSink: defaultSink });
  }

  ngOnInit(): void {
    this.grafanaURL = (localStorage.getItem('grafanaURL') || '') + '/d/polardbx-monitor?orgId=1&var-namespace=default';
    this.loadBackupSchedules();
    this.loadClusters();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private createScheduleForm(): FormGroup {
    return this.fb.group({
      name: ['', [Validators.required, Validators.pattern(/^[a-z0-9-]+$/), Validators.maxLength(30)]],
      namespace: ['default'],
      clusterName: ['', Validators.required],
      schedule: ['0 2 * * *', Validators.required],
      suspend: [false],
      maxBackupCount: [7, [Validators.min(1)]],
      storageProvider: ['oss' as BackupStorage],
      storageSink: [''],
      cleanPolicy: ['Retain' as CleanPolicyType],
      preferredBackupRole: ['follower'],
      retentionDays: [30, [Validators.min(1)]]
    });
  }

  private loadClusters(): void {
    this.apiService.getClusters().pipe(takeUntil(this.destroy$)).subscribe({
      next: (clusters) => {
        this.availableClusters = (clusters || []).map(c => c.metadata?.name).filter(Boolean);
      },
      error: () => {}
    });
  }

  isLoading(key: keyof typeof LoadingKeys): boolean {
    return this.loadingService.isLoading(LoadingKeys[key]);
  }

  onTabChange(event: any): void {
    this.selectedTab = event.index;
    // 避免隐藏面板内仍保留焦点导致 aria-hidden 警告
    setTimeout(() => {
      const active = document.activeElement as HTMLElement | null;
      if (active && active.blur) active.blur();
    }, 0);
    if (event.index === 0) {
      this.editingSchedule = null;
      this.resetForm();
    }
  }

  loadBackupSchedules(): void {
    this.apiService.getBackupSchedules()
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => {})
      )
      .subscribe({
        next: (schedules: PolarDBXBackupSchedule[]) => {
          this.backupSchedules = schedules || [];
        },
        error: (error) => {
          console.error('加载备份计划失败:', error);
          this.snackBar.open('加载备份计划失败', '关闭', { duration: 3000 });
        }
      });
  }

  refreshSchedules(): void {
    this.loadBackupSchedules();
  }

  applyPredefinedSchedule(schedule: string | null): void {
    if (schedule) {
      this.scheduleForm.patchValue({ schedule });
    }
  }

  submitSchedule(): void {
    if (this.scheduleForm.invalid) return;

    const formValue = this.scheduleForm.value;
    // 补齐缺省 sink / provider
    let storageSink: string = (formValue.storageSink || '').trim();
    if (!storageSink) {
      storageSink = (localStorage.getItem('backupSinkName') || localStorage.getItem('xstoreBackupSinkName') || '').trim();
    }
    const storageProvider: BackupStorage = formValue.storageProvider || (localStorage.getItem('backupStorageName') as BackupStorage) || 's3';

    const scheduleRequest: CreateBackupScheduleRequest = {
      name: formValue.name,
      namespace: formValue.namespace || 'default',
      schedule: formValue.schedule,
      suspend: formValue.suspend,
      maxBackupCount: formValue.maxBackupCount,
      clusterName: formValue.clusterName,
      cleanPolicy: formValue.cleanPolicy,
      preferredBackupRole: formValue.preferredBackupRole,
      retentionTime: formValue.retentionDays ? { duration: formValue.retentionDays * 24 * 60 * 60 * 1000 * 1000 * 1000 } : undefined,
      storageProvider: {
        storageName: storageProvider,
        sink: storageSink
      }
    };

    const operation = this.editingSchedule
      ? this.apiService.updateBackupSchedule(this.editingSchedule.metadata.namespace!, this.editingSchedule)
      : this.apiService.createBackupSchedule(scheduleRequest.namespace!, scheduleRequest);

    operation.pipe(
      takeUntil(this.destroy$),
      finalize(() => {})
    ).subscribe({
      next: (schedule) => {
        const message = this.editingSchedule ? '备份计划更新成功' : '备份计划创建成功';
        this.snackBar.open(message, '关闭', { duration: 3000 });
        this.resetForm();
        this.selectedTab = 0;
        this.loadBackupSchedules();
      },
      error: (error) => {
        console.error('保存备份计划失败:', error);
        this.snackBar.open('保存备份计划失败', '关闭', { duration: 3000 });
      }
    });
  }

  editSchedule(schedule: PolarDBXBackupSchedule): void {
    this.editingSchedule = schedule;
    const retentionDays = schedule.spec.backupSpec.retentionTime ? 
      Math.floor(schedule.spec.backupSpec.retentionTime.duration / (24 * 60 * 60 * 1000000000)) : 30;
    
    this.scheduleForm.patchValue({
      name: schedule.metadata.name,
      namespace: schedule.metadata.namespace,
      clusterName: schedule.spec.backupSpec.cluster.name,
      schedule: schedule.spec.schedule,
      suspend: schedule.spec.suspend || false,
      maxBackupCount: schedule.spec.maxBackupCount || 7,
      storageProvider: schedule.spec.backupSpec.storageProvider?.storageName || 'oss',
      storageSink: schedule.spec.backupSpec.storageProvider?.sink || '',
      cleanPolicy: schedule.spec.backupSpec.cleanPolicy || 'Retain',
      preferredBackupRole: schedule.spec.backupSpec.preferredBackupRole || 'follower',
      retentionDays: retentionDays
    });
    this.selectedTab = 1;
  }

  toggleSchedule(schedule: PolarDBXBackupSchedule): void {
    const updatedSchedule = { ...schedule };
    updatedSchedule.spec.suspend = !schedule.spec.suspend;

    this.apiService.updateBackupSchedule(schedule.metadata.namespace!, updatedSchedule)
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => {})
      )
      .subscribe({
        next: () => {
          const action = updatedSchedule.spec.suspend ? '暂停成功' : '恢复成功';
          this.snackBar.open(`备份计划 ${action}`, '关闭', { duration: 3000 });
          this.loadBackupSchedules();
        },
        error: (error) => {
          console.error('切换备份计划失败:', error);
          this.snackBar.open('切换备份计划失败', '关闭', { duration: 3000 });
        }
      });
  }

  deleteSchedule(schedule: PolarDBXBackupSchedule): void {
    if (confirm(`删除备份计划\n\n您确定要删除备份计划 "${schedule.metadata.name}" 吗？`)) {
      this.apiService.deleteBackupSchedule(schedule.metadata.namespace!, schedule.metadata.name)
        .pipe(
          takeUntil(this.destroy$),
          finalize(() => {})
        )
        .subscribe({
          next: () => {
            this.snackBar.open('备份计划删除成功', '关闭', { duration: 3000 });
            this.loadBackupSchedules();
          },
          error: (error) => {
            console.error('删除备份计划失败:', error);
            this.snackBar.open('删除备份计划失败', '关闭', { duration: 3000 });
          }
        });
    }
  }

  viewScheduleDetails(schedule: PolarDBXBackupSchedule): void {
    this.dialog.open(this.detailsDialogTpl, { data: schedule });
  }

  runNow(schedule: PolarDBXBackupSchedule): void {
    const ns = schedule.metadata.namespace || 'default';
    const cluster = schedule.spec.backupSpec.cluster.name;
    // 使用计划的 backupSpec 生成一次手动备份对象
    const backupObj: any = {
      apiVersion: 'polardbx.aliyun.com/v1',
      kind: 'PolarDBXBackup',
      metadata: {
        name: `${cluster}-manual-${Math.random().toString(36).slice(2,6)}`.toLowerCase(),
        namespace: ns
      },
      spec: {
        cluster: { name: cluster },
        storageProvider: schedule.spec.backupSpec.storageProvider,
        cleanPolicy: schedule.spec.backupSpec.cleanPolicy,
        preferredBackupRole: schedule.spec.backupSpec.preferredBackupRole,
        retentionTime: schedule.spec.backupSpec.retentionTime
      }
    };
    this.apiService.createBackup(ns, cluster, backupObj).subscribe({
      next: () => this.snackBar.open('已触发一次手动备份', '关闭', { duration: 3000 }),
      error: () => this.snackBar.open('触发失败', '关闭', { duration: 3000 })
    });
  }

  resetForm(): void {
    this.editingSchedule = null;
    this.scheduleForm.reset({
      name: '',
      namespace: 'default',
      clusterName: '',
      schedule: '0 2 * * *',
      suspend: false,
      maxBackupCount: 7,
      storageProvider: 'oss',
      storageSink: '',
      cleanPolicy: 'Retain',
      preferredBackupRole: 'follower',
      retentionDays: 30
    });
  }

  // References
  @ViewChild('detailsDialog') detailsDialogTpl!: TemplateRef<any>;

  getCronDescription(cronExpression: string): string {
    const predefined = this.predefinedSchedules.find(s => s.value === cronExpression);
    return predefined ? predefined.description : '自定义计划';
  }

  getStorageIcon(storage?: BackupStorage): string {
    switch (storage) {
      case 'oss': return 'cloud';
      case 's3': return 'cloud_circle';
      case 'sftp': return 'folder_shared';
      default: return 'storage';
    }
  }

  getStorageLabel(storage?: BackupStorage): string {
    const option = this.storageProviders.find(p => p.value === storage);
    return option ? option.label : storage || '未指定';
  }

  formatDate(dateString?: string): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  }

  grafanaLinkFor(schedule: PolarDBXBackupSchedule): string {
    const base = localStorage.getItem('grafanaURL') || '';
    if (!base) return '';
    const ns = schedule?.metadata?.namespace || 'default';
    const cluster = schedule?.spec?.backupSpec?.cluster?.name || '';
    const clusterParam = cluster ? `&var-cluster=${encodeURIComponent(cluster)}` : '';
    return `${base}/d/polardbx-monitor?orgId=1&var-namespace=${encodeURIComponent(ns)}${clusterParam}`;
  }
}