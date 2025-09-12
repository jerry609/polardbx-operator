import { Component, inject, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSortModule } from '@angular/material/sort';
import { MatTabsModule } from '@angular/material/tabs';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatExpansionModule } from '@angular/material/expansion';
import { EmptyStateComponent } from '../empty-state/empty-state.component';
import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { XStoreBackup, XStoreBackupWithStatus, CreateXStoreBackupRequest } from '../../models/xstore-backup.model';
import { XStore } from '../../models/xstore.model';
import { MatTableDataSource } from '@angular/material/table';
import { Observable } from 'rxjs';
import { NamespaceService } from '../../services/namespace.service';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

export interface XStoreBackupDialogData {
  mode: 'create' | 'edit' | 'view';
  backup?: XStoreBackup;
  namespace?: string;
}

@Component({
  selector: 'app-xstore-backup-management',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatCardModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatTabsModule,
    MatChipsModule,
    MatTooltipModule,
    MatButtonToggleModule,
    MatProgressBarModule,
    MatCheckboxModule,
    MatExpansionModule,
    EmptyStateComponent
  ],
  template: `
    <div class="xstore-backup-management">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>storage</mat-icon>
            存储备份管理
          </mat-card-title>
          <mat-card-subtitle>管理XStore存储级备份</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <mat-tab-group [(selectedIndex)]="selectedTab">
            <mat-tab label="存储备份列表">
              <div class="tab-content">
                <div>
                  <button mat-raised-button color="primary" (click)="refreshBackups()">
                    <mat-icon>refresh</mat-icon>
                    刷新
                  </button>
                  <button mat-raised-button color="accent" style="margin-left: 8px" (click)="createNew()">
                    <mat-icon>add</mat-icon>
                    新建备份
                  </button>
                  <mat-button-toggle-group style="margin-left: 12px" [value]="viewMode" (change)="onChangeViewMode($event.value)">
                    <mat-button-toggle value="summary">汇总</mat-button-toggle>
                    <mat-button-toggle value="detail">明细</mat-button-toggle>
                  </mat-button-toggle-group>
                </div>
                <mat-progress-bar *ngIf="loadingService.isLoading(loadingKeys.XSTORE_BACKUP_LIST)" mode="indeterminate"></mat-progress-bar>
                <table mat-table [dataSource]="dataSource" class="full-width-table">
                  <ng-container matColumnDef="name">
                    <th mat-header-cell *matHeaderCellDef>名称</th>
                    <td mat-cell *matCellDef="let b">{{ b.metadata.name }}</td>
                  </ng-container>
                  <ng-container matColumnDef="xStoreName">
                    <th mat-header-cell *matHeaderCellDef>XStore/类别</th>
                    <td mat-cell *matCellDef="let b">
                      {{ getXStoreName(b) }}
                      <span class="category-pill" [matTooltip]="getCategoryTooltip(b)" matTooltipPosition="above" *ngIf="getCategoryDisplay(b) as cat"> · {{ cat }}</span>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="backupType">
                    <th mat-header-cell *matHeaderCellDef>类型</th>
                    <td mat-cell *matCellDef="let b">{{ (b.spec.backupType || 'full') | titlecase }}</td>
                  </ng-container>
                  <ng-container matColumnDef="phase">
                    <th mat-header-cell *matHeaderCellDef>状态</th>
                    <td mat-cell *matCellDef="let b">
                      <mat-chip [color]="getStatusColor(b.status?.phase)">{{ b.displayStatus || (b.status?.phase || '未知') }}</mat-chip>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="size">
                    <th mat-header-cell *matHeaderCellDef>大小</th>
                    <td mat-cell *matCellDef="let b">{{ b.displaySize }}</td>
                  </ng-container>
                  <ng-container matColumnDef="creationTime">
                    <th mat-header-cell *matHeaderCellDef>开始时间</th>
                    <td mat-cell *matCellDef="let b">{{ b.status?.startTime || '-' }}</td>
                  </ng-container>
                  <ng-container matColumnDef="status">
                    <th mat-header-cell *matHeaderCellDef>进度</th>
                    <td mat-cell *matCellDef="let b">
                      <mat-progress-bar [value]="getProgressPercentage(b)" mode="determinate"></mat-progress-bar>
                    </td>
                  </ng-container>
                  <ng-container matColumnDef="actions">
                    <th mat-header-cell *matHeaderCellDef>操作</th>
                    <td mat-cell *matCellDef="let b">
                      <button mat-icon-button (click)="viewBackup(b)"><mat-icon>visibility</mat-icon></button>
                      <button mat-icon-button (click)="editBackup(b)"><mat-icon>edit</mat-icon></button>
                      <button mat-icon-button color="warn" (click)="deleteBackup(b)"><mat-icon>delete</mat-icon></button>
                      <button mat-icon-button color="warn" matTooltip="强制删除（移除finalizers）" (click)="forceDeleteBackup(b)">
                        <mat-icon>delete_forever</mat-icon>
                      </button>
                    </td>
                  </ng-container>
                  <tr mat-header-row *matHeaderRowDef="displayedColumns"></tr>
                  <tr mat-row *matRowDef="let row; columns: displayedColumns"></tr>
                  <tr class="mat-row" *matNoDataRow>
                    <td class="mat-cell" colspan="9999">
                      <app-empty-state icon="storage" title="暂无存储备份" hint="点击“新建备份”创建首个存储备份"></app-empty-state>
                    </td>
                  </tr>
                </table>
              </div>
            </mat-tab>
            <mat-tab label="创建备份配置">
              <div class="tab-content">
                <form [formGroup]="backupForm">
                  <mat-card>
                    <mat-card-title>基础配置</mat-card-title>
                    <mat-card-content>
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half">
                          <mat-label>名称</mat-label>
                          <input matInput formControlName="name">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half">
                          <mat-label>命名空间</mat-label>
                          <input matInput formControlName="namespace">
                        </mat-form-field>
                      </div>
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half">
                          <mat-label>XStore</mat-label>
                          <mat-select formControlName="xStoreName">
                            <mat-option *ngFor="let x of availableXStores" [value]="x.metadata.name">{{ x.metadata.name }}</mat-option>
                          </mat-select>
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="half">
                          <mat-label>备份类型</mat-label>
                          <mat-select formControlName="backupType">
                            <mat-option *ngFor="let t of backupTypes" [value]="t.value">{{ t.label }}</mat-option>
                          </mat-select>
                        </mat-form-field>
                      </div>
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="full">
                          <mat-label>定时表达式（可选）</mat-label>
                          <input matInput formControlName="schedule" placeholder="例如：0 2 * * *">
                        </mat-form-field>
                      </div>
                      <div class="form-row">
                        <mat-checkbox formControlName="compression">启用压缩</mat-checkbox>
                        <mat-checkbox formControlName="enableEncryption" style="margin-left:16px">启用加密</mat-checkbox>
                      </div>
                    </mat-card-content>
                  </mat-card>
                </form>

                <form [formGroup]="storageForm" style="margin-top:16px">
                  <mat-card>
                    <mat-card-title>存储配置</mat-card-title>
                    <mat-card-content>
                      <div class="form-row">
                        <mat-form-field appearance="outline" class="half">
                          <mat-label>存储类型</mat-label>
                          <mat-select formControlName="storageType">
                            <mat-option *ngFor="let s of storageProviders" [value]="s.value">{{ s.label }}</mat-option>
                          </mat-select>
                        </mat-form-field>
                      </div>
                      <div *ngIf="selectedStorageType === 'oss'" formGroupName="ossConfig">
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>AccessKeyId</mat-label>
                            <input matInput formControlName="accessKeyId">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>AccessKeySecret</mat-label>
                            <input matInput formControlName="accessKeySecret">
                          </mat-form-field>
                        </div>
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Bucket</mat-label>
                            <input matInput formControlName="bucket">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Endpoint</mat-label>
                            <input matInput formControlName="endpoint">
                          </mat-form-field>
                        </div>
                      </div>
                      <div *ngIf="selectedStorageType === 's3'" formGroupName="s3Config">
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>AccessKeyId</mat-label>
                            <input matInput formControlName="accessKeyId">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>SecretAccessKey</mat-label>
                            <input matInput formControlName="secretAccessKey">
                          </mat-form-field>
                        </div>
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Bucket</mat-label>
                            <input matInput formControlName="bucket">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Region</mat-label>
                            <input matInput formControlName="region">
                          </mat-form-field>
                        </div>
                      </div>
                      <div *ngIf="selectedStorageType === 'sftp'" formGroupName="sftpConfig">
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Host</mat-label>
                            <input matInput formControlName="host">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Port</mat-label>
                            <input matInput type="number" formControlName="port">
                          </mat-form-field>
                        </div>
                        <div class="form-row">
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Username</mat-label>
                            <input matInput formControlName="username">
                          </mat-form-field>
                          <mat-form-field appearance="outline" class="half">
                            <mat-label>Remote Path</mat-label>
                            <input matInput formControlName="remotePath">
                          </mat-form-field>
                        </div>
                      </div>
                    </mat-card-content>
                  </mat-card>
                </form>

                <form [formGroup]="retentionForm" style="margin-top:16px">
                  <mat-card>
                    <mat-card-title>保留策略</mat-card-title>
                    <mat-card-content>
                      <mat-checkbox formControlName="enableRetention">启用保留策略</mat-checkbox>
                      <div class="form-row" *ngIf="retentionForm.get('enableRetention')?.value">
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>保留数量</mat-label>
                          <input matInput type="number" formControlName="retain">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>保留天数</mat-label>
                          <input matInput type="number" formControlName="retainDays">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>保留小时</mat-label>
                          <input matInput type="number" formControlName="retainHours">
                        </mat-form-field>
                      </div>
                    </mat-card-content>
                  </mat-card>
                </form>

                <form [formGroup]="resourceForm" style="margin-top:16px">
                  <mat-card>
                    <mat-card-title>资源配置</mat-card-title>
                    <mat-card-content>
                      <mat-checkbox formControlName="enableResourceLimits">启用资源限制</mat-checkbox>
                      <div class="form-row" *ngIf="resourceForm.get('enableResourceLimits')?.value">
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>Requests CPU</mat-label>
                          <input matInput formControlName="requestsCpu">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>Requests Memory</mat-label>
                          <input matInput formControlName="requestsMemory">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>Requests Storage</mat-label>
                          <input matInput formControlName="requestsStorage">
                        </mat-form-field>
                      </div>
                      <div class="form-row" *ngIf="resourceForm.get('enableResourceLimits')?.value">
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>Limits CPU</mat-label>
                          <input matInput formControlName="limitsCpu">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>Limits Memory</mat-label>
                          <input matInput formControlName="limitsMemory">
                        </mat-form-field>
                        <mat-form-field appearance="outline" class="third">
                          <mat-label>Limits Storage</mat-label>
                          <input matInput formControlName="limitsStorage">
                        </mat-form-field>
                      </div>
                    </mat-card-content>
                  </mat-card>
                </form>

                <div style="margin-top:16px; display:flex; gap:8px; justify-content:flex-end">
                  <button mat-button type="button" (click)="resetForms()">重置</button>
                  <button mat-raised-button color="primary" [disabled]="!isFormValid() || isProcessing" (click)="saveBackup()">
                    <mat-icon>save</mat-icon>
                    保存
                  </button>
                </div>
              </div>
            </mat-tab>
          </mat-tab-group>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .xstore-backup-management {
      padding: 20px;
    }
    
    .tab-content {
      padding: 20px;
    }
    
    mat-card-title {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .category-pill {
      color: rgba(0,0,0,0.54);
      margin-left: 4px;
      font-size: 12px;
    }
  `]
})
export class XStoreBackupManagementComponent implements OnInit, OnDestroy {
  private fb = inject(FormBuilder);
  private apiService = inject(ApiService);
  public loadingService = inject(LoadingService);
  private ns = inject(NamespaceService);
  loadingKeys = LoadingKeys;
  dialogRef = inject<MatDialogRef<XStoreBackupManagementComponent>>(MatDialogRef, { optional: true });
  data = inject<XStoreBackupDialogData>(MAT_DIALOG_DATA, { optional: true }) || { mode: 'create', namespace: 'default' } as XStoreBackupDialogData;

  // Forms
  backupForm!: FormGroup;
  storageForm!: FormGroup;
  retentionForm!: FormGroup;
  resourceForm!: FormGroup;

  // Data
  availableXStores: XStore[] = [];
  backups: XStoreBackupWithStatus[] = [];
  dataSource = new MatTableDataSource<XStoreBackupWithStatus>();
  // 汇总/明细视图切换，默认汇总
  viewMode: 'summary' | 'detail' = 'summary';
  // 明细计数与名称用于汇总视图显示类别细分（DN/GMS）
  private detailCountsByParent: Record<string, { dn: number; gms: number; dnNames: string[]; gmsNames: string[] }> = {};
  
  // UI State
  isProcessing = false;
  selectedTab = 0;
  private destroy$ = new Subject<void>();
  
  // Table configuration
  displayedColumns: string[] = [
    'name', 
    'xStoreName', 
    'backupType', 
    'phase', 
    'size', 
    'creationTime',
    'status',
    'actions'
  ];

  // Options
  backupTypes = [
    { value: 'full', label: '全量备份' },
    { value: 'incremental', label: '增量备份' }
  ];

  storageProviders = [
    { value: 'oss', label: '阿里云 OSS' },
    { value: 's3', label: 'Amazon S3' },
    { value: 'sftp', label: 'SFTP' }
  ];

  cpuOptions = ['0.5', '1', '2', '4', '8'];
  memoryOptions = ['1Gi', '2Gi', '4Gi', '8Gi', '16Gi'];
  storageOptions = ['10Gi', '20Gi', '50Gi', '100Gi', '200Gi'];

  constructor() {
    this.initializeForms();
  }

  ngOnInit(): void {
    this.loadInitialData();
    if (this.data.mode === 'edit' && this.data.backup) {
      this.populateFormFromBackup(this.data.backup);
    }
    this.ns.activeNamespace$.pipe(takeUntil(this.destroy$)).subscribe(() => this.loadBackups());
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private initializeForms(): void {
    // Main backup configuration form
    this.backupForm = this.fb.group({
      name: ['', [
        Validators.required, 
        Validators.pattern(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/),
        Validators.maxLength(30)
      ]],
      namespace: [this.data.namespace || 'default', Validators.required],
      xStoreName: ['', Validators.required],
      backupType: ['full', Validators.required],
      schedule: [''],
      compression: [true],
      enableEncryption: [false]
    });

    // Storage configuration form
    this.storageForm = this.fb.group({
      storageType: ['oss', Validators.required],
      ossConfig: this.fb.group({
        accessKeyId: ['', Validators.required],
        accessKeySecret: ['', Validators.required],
        bucket: ['', Validators.required],
        endpoint: ['', Validators.required],
        prefix: ['']
      }),
      s3Config: this.fb.group({
        accessKeyId: ['', Validators.required],
        secretAccessKey: ['', Validators.required],
        bucket: ['', Validators.required],
        region: ['', Validators.required],
        endpoint: [''],
        prefix: ['']
      }),
      sftpConfig: this.fb.group({
        host: ['', Validators.required],
        port: [22, [Validators.required, Validators.min(1), Validators.max(65535)]],
        username: ['', Validators.required],
        password: [''],
        privateKey: [''],
        remotePath: ['', Validators.required]
      })
    });

    // Retention policy form
    this.retentionForm = this.fb.group({
      enableRetention: [false],
      retain: [7, [Validators.min(1)]],
      retainDays: [30, [Validators.min(1)]],
      retainHours: [168, [Validators.min(1)]]
    });

    // Resource configuration form
    this.resourceForm = this.fb.group({
      enableResourceLimits: [false],
      requestsCpu: ['1'],
      requestsMemory: ['2Gi'],
      requestsStorage: ['10Gi'],
      limitsCpu: ['2'],
      limitsMemory: ['4Gi'],
      limitsStorage: ['20Gi'],
      nodeSelector: this.fb.group({
        enabled: [false],
        key: [''],
        value: ['']
      })
    });

    // Watch for form changes
    this.storageForm.get('storageType')?.valueChanges.subscribe(() => {
      this.updateStorageValidators();
    });

    this.retentionForm.get('enableRetention')?.valueChanges.subscribe(enabled => {
      this.updateRetentionValidators(enabled);
    });

    this.resourceForm.get('enableResourceLimits')?.valueChanges.subscribe(enabled => {
      this.updateResourceValidators(enabled);
    });

    this.backupForm.get('enableEncryption')?.valueChanges.subscribe(enabled => {
      this.updateEncryptionValidators(enabled);
    });
  }

  private updateStorageValidators(): void {
    const storageType = this.storageForm.get('storageType')?.value;

    // Clear all validators first
    ['ossConfig','s3Config','sftpConfig'].forEach(groupName => {
      const configGroup = this.storageForm.get(groupName) as FormGroup;
      if (!configGroup) return;
      Object.keys(configGroup.controls).forEach(field => {
        const control = configGroup.get(field);
        control?.clearValidators();
        control?.updateValueAndValidity();
      });
    });

    // Minimal validators that align with backend (sink-only)
    if (storageType === 's3') {
      const g = this.storageForm.get('s3Config') as FormGroup;
      g.get('bucket')?.setValidators([Validators.required]);
      g.get('bucket')?.updateValueAndValidity();
    } else if (storageType === 'oss') {
      const g = this.storageForm.get('ossConfig') as FormGroup;
      g.get('bucket')?.setValidators([Validators.required]);
      g.get('bucket')?.updateValueAndValidity();
    } else if (storageType === 'sftp') {
      const g = this.storageForm.get('sftpConfig') as FormGroup;
      g.get('host')?.setValidators([Validators.required]);
      g.get('username')?.setValidators([Validators.required]);
      g.get('remotePath')?.setValidators([Validators.required]);
      g.get('host')?.updateValueAndValidity();
      g.get('username')?.updateValueAndValidity();
      g.get('remotePath')?.updateValueAndValidity();
    }
  }

  private updateRetentionValidators(enabled: boolean): void {
    const fields = ['retain', 'retainDays', 'retainHours'];
    fields.forEach(field => {
      const control = this.retentionForm.get(field);
      if (enabled) {
        control?.setValidators([Validators.required, Validators.min(1)]);
      } else {
        control?.clearValidators();
      }
      control?.updateValueAndValidity();
    });
  }

  private updateResourceValidators(enabled: boolean): void {
    const resourceFields = [
      'requestsCpu', 'requestsMemory', 'requestsStorage',
      'limitsCpu', 'limitsMemory', 'limitsStorage'
    ];

    resourceFields.forEach(field => {
      const control = this.resourceForm.get(field);
      if (enabled) {
        control?.setValidators([Validators.required]);
      } else {
        control?.clearValidators();
      }
      control?.updateValueAndValidity();
    });
  }

  private updateEncryptionValidators(enabled: boolean): void {
    // Could add encryption-specific validators here
    // For now, encryption is just a boolean flag
  }

  private async loadInitialData(): Promise<void> {
    try {
      // Load available XStores
      this.availableXStores = await this.apiService.getXStores(this.data.namespace).toPromise() || [];
      
      // Load existing backups
      await this.loadBackups();
    } catch (error) {
      console.error('Failed to load initial data:', error);
    }
  }

  private async loadBackups(): Promise<void> {
    try {
      if (this.viewMode === 'summary') {
        const [summary, detail] = await Promise.all([
          this.apiService.listXStoreBackups(this.data.namespace as string, 'summary').toPromise(),
          this.apiService.listXStoreBackups(this.data.namespace as string, 'detail').toPromise()
        ]);
        const safeSummary = summary || [];
        const safeDetail = detail || [];
        // 统计明细类别计数
        this.detailCountsByParent = {};
        for (const b of safeDetail as any[]) {
          const name: string = b?.metadata?.name || '';
          const key = this.getParentKey(name);
          if (!this.detailCountsByParent[key]) this.detailCountsByParent[key] = { dn: 0, gms: 0, dnNames: [], gmsNames: [] };
          if (name.includes('-dn-')) { this.detailCountsByParent[key].dn++; this.detailCountsByParent[key].dnNames.push(name); }
          if (name.endsWith('-gms') || name.includes('-gms-')) { this.detailCountsByParent[key].gms++; this.detailCountsByParent[key].gmsNames.push(name); }
        }
        this.backups = safeSummary.map(backup => this.enrichBackupWithStatus(backup as any));
        this.dataSource.data = this.backups;
      } else {
        const backups = await this.apiService.listXStoreBackups(this.data.namespace as string, 'detail').toPromise() || [];
        this.backups = backups.map(backup => this.enrichBackupWithStatus(backup));
        this.dataSource.data = this.backups;
      }
    } catch (error) {
      console.error('Failed to load XStore backups:', error);
    }
  }

  private enrichBackupWithStatus(backup: XStoreBackup): XStoreBackupWithStatus {
    const status = backup.status;
    const provider: any = (backup as any).spec?.storageProvider || {};
    const storageName: string = (provider.type || provider.storageName || '').toString();

    return {
      ...backup,
      isRunning: status?.phase === 'Running',
      isCompleted: status?.phase === 'Completed',
      isFailed: status?.phase === 'Failed',
      displayStatus: this.getDisplayStatus(status?.phase as any, (status as any)?.stage),
      displaySize: this.formatBytes((status as any)?.backupSize),
      displayDuration: this.calculateDuration(status?.startTime as any, (status as any)?.completionTime),
      canRestore: status?.phase === 'Completed',
      storageType: storageName ? storageName.toUpperCase() : undefined,
      compressionRatio: this.calculateCompressionRatio((status as any)?.backupSize, (status as any)?.compressedSize)
    } as any;
  }

  private getDisplayStatus(phase?: string, stage?: string): string {
    if (!phase) return '未知';

    // 兼容 XStoreBackup 与旧/新阶段名称
    const map: Record<string, string> = {
      Pending: '等待中',
      Running: '备份中',
      Completed: '已完成',
      Failed: '失败',

      // XStore 专有阶段
      Backuping: '备份中',
      Collecting: '收集中',
      Calculating: '计算中',
      Binloging: '日志备份中',
      MetadataBackuping: '元数据备份中',
      Waiting: '等待中',
      Finished: '已完成',
      Deleting: '删除中',
      Dummy: '占位'
    };

    let status = map[phase] || phase;
    if (stage && stage !== phase) status += ` (${stage})`;
    return status;
  }

  private formatBytes(bytes?: number): string {
    if (!bytes) return '-';
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
  }

  private calculateDuration(startTime?: string, endTime?: string): string {
    if (!startTime) return '-';
    
    const start = new Date(startTime);
    const end = endTime ? new Date(endTime) : new Date();
    const diffMs = end.getTime() - start.getTime();
    
    const hours = Math.floor(diffMs / (1000 * 60 * 60));
    const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  }

  private calculateCompressionRatio(originalSize?: number, compressedSize?: number): number {
    if (!originalSize || !compressedSize) return 0;
    return Math.round((1 - compressedSize / originalSize) * 100);
  }

  getXStoreName(b: any): string {
    // Prefer spec.xStoreName; fallback to parsing from legacy names like "<backupName>-<xstoreName>-<role>"
    const direct = b?.spec?.xStoreName || (b?.spec?.xstore?.name);
    if (direct) return direct;
    const name: string = b?.metadata?.name || '';
    // Try to strip suffix like "-dn-0" or "-gms"
    const parts = name.split('-');
    if (parts.length > 4) {
      // Heuristic: remove trailing 3 segments for role/pod index
      return parts.slice(0, parts.length - 3).join('-');
    }
    return '未知';
  }

  private getParentKey(name: string): string {
    const dnIdx = name.lastIndexOf('-dn-');
    if (dnIdx > 0) return name.substring(0, dnIdx);
    if (name.endsWith('-gms')) return name.substring(0, name.length - 4);
    const gmsMid = name.lastIndexOf('-gms-');
    if (gmsMid > 0) return name.substring(0, gmsMid);
    return name;
  }

  getCategoryDisplay(b: XStoreBackupWithStatus): string {
    const name = b?.metadata?.name || '';
    if (this.viewMode === 'detail') {
      if (name.includes('-dn-')) return 'DN';
      if (name.endsWith('-gms') || name.includes('-gms-')) return 'GMS';
      return '未知';
    }
    // 汇总模式：显示混合及计数
    const key = this.getParentKey(name);
    const c = this.detailCountsByParent[key] || { dn: 0, gms: 0, dnNames: [], gmsNames: [] };
    if (c.dn > 0 && c.gms > 0) return `混合 (DN ${c.dn}, GMS ${c.gms})`;
    if (c.dn > 0) return `DN (${c.dn})`;
    if (c.gms > 0) return `GMS (${c.gms})`;
    return '未知';
  }

  getCategoryTooltip(b: XStoreBackupWithStatus): string {
    const name = b?.metadata?.name || '';
    if (this.viewMode === 'detail') {
      if (name.includes('-dn-')) return 'DN';
      if (name.endsWith('-gms') || name.includes('-gms-')) return 'GMS';
      return '';
    }
    const key = this.getParentKey(name);
    const c = this.detailCountsByParent[key];
    if (!c) return '';
    const parts: string[] = [];
    if (c.dnNames?.length) parts.push(`DN: ${c.dnNames.join(', ')}`);
    if (c.gmsNames?.length) parts.push(`GMS: ${c.gmsNames.join(', ')}`);
    return parts.join(' | ');
  }

  private populateFormFromBackup(backup: XStoreBackup): void {
    // Populate main form
    this.backupForm.patchValue({
      name: backup.metadata.name,
      namespace: backup.metadata.namespace,
      xStoreName: backup.spec.xStoreName,
      backupType: backup.spec.backupType,
      schedule: backup.spec.schedule,
      compression: backup.spec.compression,
      enableEncryption: backup.spec.encryption?.enabled
    });

    // Populate storage form
    const storageProvider = backup.spec.storageProvider;
    this.storageForm.patchValue({
      storageType: storageProvider.type
    });

    switch (storageProvider.type) {
      case 'oss':
        this.storageForm.get('ossConfig')?.patchValue(storageProvider.oss || {});
        break;
      case 's3':
        this.storageForm.get('s3Config')?.patchValue(storageProvider.s3 || {});
        break;
      case 'sftp':
        this.storageForm.get('sftpConfig')?.patchValue(storageProvider.sftp || {});
        break;
    }

    // Populate retention form
    const retentionPolicy = backup.spec.retentionPolicy;
    if (retentionPolicy) {
      this.retentionForm.patchValue({
        enableRetention: true,
        retain: retentionPolicy.retain,
        retainDays: retentionPolicy.retainDays,
        retainHours: retentionPolicy.retainHours
      });
    }

    // Populate resource form
    const resources = backup.spec.resources;
    if (resources) {
      this.resourceForm.patchValue({
        enableResourceLimits: true,
        requestsCpu: resources.requests?.cpu,
        requestsMemory: resources.requests?.memory,
        requestsStorage: resources.requests?.storage,
        limitsCpu: resources.limits?.cpu,
        limitsMemory: resources.limits?.memory,
        limitsStorage: resources.limits?.storage
      });
    }

    // Populate node selector
    const nodeSelector = backup.spec.nodeSelector;
    if (nodeSelector && Object.keys(nodeSelector).length > 0) {
      const firstKey = Object.keys(nodeSelector)[0];
      this.resourceForm.patchValue({
        nodeSelector: {
          enabled: true,
          key: firstKey,
          value: nodeSelector[firstKey]
        }
      });
    }
  }

  // Create or update XStore backup
  async saveBackup(): Promise<void> {
    if (!this.isFormValid()) {
      return;
    }

    this.isProcessing = true;

    try {
      const request = this.buildBackupRequest();

      if (this.data.mode === 'create') {
        await this.apiService.createXStoreBackup(request.namespace, request).toPromise();
      } else if (this.data.mode === 'edit' && this.data.backup) {
        const updatedBackup = this.buildUpdatedBackup(request);
        await this.apiService.updateXStoreBackup(request.namespace, updatedBackup).toPromise();
      }

      // Reload the backups list
      if (this.selectedTab === 0) {
        await this.loadBackups();
      }

      this.dialogRef?.close(true);
    } catch (error) {
      console.error('Failed to save XStore backup:', error);
    } finally {
      this.isProcessing = false;
    }
  }

  private buildBackupRequest(): any & { namespace: string } {
    const backupValue = this.backupForm.value;
    const storageValue = this.storageForm.value;
    const retentionValue = this.retentionForm.value;
    const resourceValue = this.resourceForm.value;

    // Use sink NAME configured in HPFS (not URL). Default to 'default'.
    const storageType: 'oss' | 's3' | 'sftp' = storageValue.storageType;
    const sink = (localStorage.getItem('xstoreBackupSinkName') || localStorage.getItem('backupSinkName') || 'default').trim();

    // Convert retention to duration string (hours) if enabled
    let retentionTime: string | undefined;
    if (retentionValue?.enableRetention) {
      const days = Number(retentionValue.retainDays) || 0;
      const hours = Number(retentionValue.retainHours) || 0;
      const totalHours = days * 24 + hours;
      if (totalHours > 0) retentionTime = `${totalHours}h`;
    }

    // Build CR object matching backend
    const cr: any = {
      apiVersion: 'polardbx.aliyun.com/v1',
      kind: 'XStoreBackup',
      metadata: {
        name: backupValue.name,
        namespace: backupValue.namespace
      },
      spec: {
        xstore: { name: backupValue.xStoreName },
        storageProvider: { storageName: storageType, sink: sink },
        preferredBackupRole: 'follower'
      }
    };
    if (retentionTime) cr.spec.retentionTime = retentionTime;

    return { namespace: backupValue.namespace, ...cr };
  }

  private buildUpdatedBackup(request: CreateXStoreBackupRequest): XStoreBackup {
    return {
      ...this.data.backup!,
      spec: {
        xStoreName: request.xStoreName,
        backupType: request.backupType,
        storageProvider: request.storageProvider,
        retentionPolicy: request.retentionPolicy,
        resources: request.resources,
        schedule: request.schedule,
        compression: request.compression,
        encryption: request.encryption,
        nodeSelector: request.nodeSelector
      }
    };
  }

  // Delete XStore backup
  async deleteBackup(backup: XStoreBackupWithStatus): Promise<void> {
    if (!confirm(`确定要删除备份 "${backup.metadata.name}" 吗？此操作不可撤销。`)) {
      return;
    }

    try {
      await this.apiService.deleteXStoreBackup(
        backup.metadata.namespace, 
        backup.metadata.name
      ).toPromise();
      
      await this.loadBackups();
    } catch (error) {
      console.error('Failed to delete XStore backup:', error);
    }
  }

  async forceDeleteBackup(backup: XStoreBackupWithStatus): Promise<void> {
    if (!confirm(`强制删除将直接移除 finalizers 并清理对象。确定对备份 "${backup.metadata.name}" 执行吗？`)) {
      return;
    }
    try {
      await this.apiService.forceDeleteXStoreBackup(
        backup.metadata.namespace,
        backup.metadata.name
      ).toPromise();
      await this.loadBackups();
    } catch (error) {
      console.error('Failed to force delete XStore backup:', error);
    }
  }

  // Other methods (edit, view, refresh, etc.)
  editBackup(backup: XStoreBackupWithStatus): void {
    this.data.mode = 'edit';
    this.data.backup = backup;
    this.populateFormFromBackup(backup);
    this.selectedTab = 1;
  }

  viewBackup(backup: XStoreBackupWithStatus): void {
    console.log('View backup details:', backup);
  }

  async refreshBackups(): Promise<void> {
    await this.loadBackups();
  }

  getStatusColor(phase?: string): string {
    if (!phase) return '';
    const ok = ['Completed', 'Finished'];
    const running = ['Running', 'Pending', 'Backuping', 'Collecting', 'Calculating', 'Binloging', 'MetadataBackuping', 'Waiting', 'Deleting'];
    const fail = ['Failed'];
    if (ok.includes(phase)) return 'primary';
    if (fail.includes(phase)) return 'warn';
    if (running.includes(phase)) return 'accent';
    return '';
  }

  getProgressPercentage(backup: XStoreBackupWithStatus): number {
    return backup.status?.progress?.percentage || 0;
  }

  onChangeViewMode(mode: 'summary' | 'detail'): void {
    if (this.viewMode !== mode) {
      this.viewMode = mode;
      this.loadBackups();
    }
  }

  isFormValid(): boolean {
    return this.backupForm.valid && 
           this.storageForm.valid &&
           (!this.retentionForm.get('enableRetention')?.value || this.retentionForm.valid) &&
           (!this.resourceForm.get('enableResourceLimits')?.value || this.resourceForm.valid);
  }

  cancel(): void {
    this.dialogRef?.close();
  }

  getFieldError(formGroup: FormGroup, fieldName: string): string {
    const field = formGroup.get(fieldName);
    if (field?.errors && field.touched) {
      if (field.errors['required']) return '此字段为必填项';
      if (field.errors['pattern']) return '格式不正确';
      if (field.errors['min']) return `最小值为 ${field.errors['min'].min}`;
      if (field.errors['max']) return `最大值为 ${field.errors['max'].max}`;
    }
    return '';
  }

  createNew(): void {
    this.data.mode = 'create';
    this.resetForms();
    this.selectedTab = 1;
  }

  resetForms(): void {
    this.backupForm.reset({
      namespace: this.data.namespace || 'default',
      backupType: 'full',
      compression: true,
      enableEncryption: false
    });
    
    this.storageForm.reset({
      storageType: 'oss'
    });
    
    this.retentionForm.reset({
      enableRetention: false,
      retain: 7,
      retainDays: 30,
      retainHours: 168
    });
    
    this.resourceForm.reset({
      enableResourceLimits: false,
      requestsCpu: '1',
      requestsMemory: '2Gi',
      requestsStorage: '10Gi',
      limitsCpu: '2',
      limitsMemory: '4Gi',
      limitsStorage: '20Gi',
      nodeSelector: { enabled: false, key: '', value: '' }
    });
  }

  getTitle(): string {
    switch (this.data.mode) {
      case 'create': return '创建 XStore 备份';
      case 'edit': return '编辑 XStore 备份';
      case 'view': return 'XStore 备份详情';
      default: return 'XStore 备份管理';
    }
  }

  get selectedStorageType(): string {
    return this.storageForm.get('storageType')?.value || 'oss';
  }
}