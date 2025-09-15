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
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { NzProgressModule } from 'ng-zorro-antd/progress';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzInputNumberModule } from 'ng-zorro-antd/input-number';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { XStoreBackup, XStoreBackupWithStatus, CreateXStoreBackupRequest } from '../../models/xstore-backup.model';
import { XStore } from '../../models/xstore.model';
import { MatTableDataSource } from '@angular/material/table';
import { Observable } from 'rxjs';
import { debounceTime, distinctUntilChanged, switchMap } from 'rxjs/operators';
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
    EmptyStateComponent,
    NzCardModule,
    NzButtonModule,
    NzIconModule,
    NzTagModule,
    NzProgressModule,
    NzTableModule,
    NzToolTipModule,
    NzGridModule,
    NzFormModule,
    NzInputModule,
    NzSelectModule,
    NzCheckboxModule,
    NzDividerModule,
    NzInputNumberModule,
    NzAlertModule
  ],
  template: `
    <div class="xstore-backup-management">
      <nz-card nzTitle="存储备份管理" [nzExtra]="headerExtra" class="header-card">
        <ng-template #headerExtra>
          <span class="subtitle">管理 XStore 存储级备份</span>
        </ng-template>
        <div class="mat-like-tabs">
          <mat-tab-group [(selectedIndex)]="selectedTab">
            <mat-tab label="存储备份列表">
              <div class="tab-content">
                <div class="list-actions">
                  <button nz-button nzType="default" (click)="refreshBackups()">
                    <i nz-icon nzType="reload"></i>
                    刷新
                  </button>
                  <button nz-button nzType="primary" style="margin-left:8px" (click)="createNew()">
                    <i nz-icon nzType="plus"></i>
                    新建备份
                  </button>
                  <div class="view-toggle">
                    <button nz-button [nzType]="viewMode==='summary' ? 'primary' : 'default'" (click)="onChangeViewMode('summary')">汇总</button>
                    <button nz-button [nzType]="viewMode==='detail' ? 'primary' : 'default'" (click)="onChangeViewMode('detail')">明细</button>
                </div>
                </div>
                <nz-table #xTable [nzData]="backups" nzSize="middle" [nzFrontPagination]="true" [nzPageSize]="10">
                  <thead>
                    <tr>
                      <th>名称</th>
                      <th>XStore/类别</th>
                      <th>类型</th>
                      <th>状态</th>
                      <th>大小</th>
                      <th>开始时间</th>
                      <th style="width:160px">进度</th>
                      <th style="width:140px">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr *ngFor="let b of xTable.data">
                      <td>{{ b.metadata.name }}</td>
                      <td>
                      {{ getXStoreName(b) }}
                        <nz-tag nzColor="geekblue" *ngIf="getCategoryDisplay(b) as cat" nz-tooltip [nzTooltipTitle]="getCategoryTooltip(b)">{{ cat }}</nz-tag>
                    </td>
                      <td>{{ (b.spec.backupType || 'full') | titlecase }}</td>
                      <td>
                        <nz-tag [nzColor]="getNzStatusColor(b.status?.phase)">{{ b.displayStatus || (b.status?.phase || '未知') }}</nz-tag>
                    </td>
                      <td>{{ b.displaySize }}</td>
                      <td>{{ b.status?.startTime || '-' }}</td>
                      <td>
                        <nz-progress [nzPercent]="getProgressPercentage(b)" nzSize="small"></nz-progress>
                    </td>
                      <td>
                        <button nz-button nzType="link" (click)="viewBackup(b)"><i nz-icon nzType="eye"></i></button>
                        <button nz-button nzType="link" (click)="editBackup(b)"><i nz-icon nzType="edit"></i></button>
                        <button nz-button nzType="link" nzDanger (click)="deleteBackup(b)"><i nz-icon nzType="delete"></i></button>
                        <button nz-button nzType="link" nzDanger nz-tooltip nzTooltipTitle="强制删除（移除finalizers）" (click)="forceDeleteBackup(b)"><i nz-icon nzType="delete"></i></button>
                    </td>
                    </tr>
                    <tr *ngIf="!xTable.data?.length">
                      <td colspan="8">
                      <app-empty-state icon="hdd" title="暂无存储备份" hint="点击“新建备份”创建首个存储备份"></app-empty-state>
                    </td>
                  </tr>
                  </tbody>
                </nz-table>
              </div>
            </mat-tab>
            <mat-tab label="创建备份配置">
              <div class="tab-content">
                <form nz-form [formGroup]="backupForm">
                  <nz-card nzTitle="基础配置" [nzExtra]="basicHelpTemplate">
                      <div class="form-row">
                      <div class="half">
                        <nz-form-item>
                          <nz-form-label [nzRequired]="true">备份名称</nz-form-label>
                          <nz-form-control nzHasFeedback [nzValidateStatus]="backupForm.get('name')?.invalid && backupForm.get('name')?.touched ? 'error' : ''">
                            <input nz-input formControlName="name" placeholder="输入备份配置名称" />
                            <div *ngIf="backupForm.get('name')?.invalid && backupForm.get('name')?.touched" style="color: #ff4d4f; font-size: 12px; margin-top: 4px;">
                              请输入有效的备份名称
                            </div>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                      <div class="half">
                        <nz-form-item>
                          <nz-form-label [nzRequired]="true">命名空间</nz-form-label>
                          <nz-form-control nzHasFeedback>
                            <input nz-input formControlName="namespace" placeholder="例如：default" />
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                      </div>
                      <div class="form-row">
                      <div class="half">
                        <nz-form-item>
                          <nz-form-label [nzRequired]="true">目标 XStore</nz-form-label>
                          <nz-form-control>
                            <nz-select formControlName="xStoreName" nzPlaceHolder="选择要备份的 XStore 实例" nzShowSearch>
                              <nz-option *ngFor="let x of availableXStores" [nzValue]="x.metadata.name" [nzLabel]="x.metadata.name">
                                <i nz-icon nzType="database" style="margin-right: 8px;"></i>
                                {{ x.metadata.name }}
                                <nz-tag nzColor="blue" style="margin-left: 8px;">{{ x.status?.phase || '未知' }}</nz-tag>
                              </nz-option>
                            </nz-select>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                      <div class="half">
                        <nz-form-item>
                          <nz-form-label [nzRequired]="true">备份类型</nz-form-label>
                          <nz-form-control>
                            <nz-select formControlName="backupType">
                              <nz-option *ngFor="let t of backupTypes" [nzValue]="t.value" [nzLabel]="t.label">
                                <i nz-icon [nzType]="getBackupTypeIcon(t.value)" style="margin-right: 8px;"></i>
                                {{ t.label }}
                              </nz-option>
                            </nz-select>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                      </div>
                      <div class="form-row">
                      <div class="full">
                        <nz-form-item>
                          <nz-form-label>
                            <span>定时表达式（可选）</span>
                            <i nz-icon nzType="question-circle" nz-tooltip nzTooltipTitle="Cron 表达式格式：分 时 日 月 周，例如每天凌晨2点：0 2 * * *" style="margin-left: 4px; color: #999;"></i>
                          </nz-form-label>
                          <nz-form-control>
                            <nz-input-group [nzPrefix]="cronPrefixTemplate" [nzSuffix]="cronSuffixTemplate">
                              <input nz-input formControlName="schedule" placeholder="例如：0 2 * * * （每天凌晨2点执行）" />
                            </nz-input-group>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                    </div>

                    <nz-divider nzText="备份选项" nzOrientation="left"></nz-divider>
                    <div class="form-row" style="align-items: center;">
                      <div style="flex: 1;">
                        <label nz-checkbox formControlName="compression">
                          <span style="margin-left: 8px;">启用压缩</span>
                        </label>
                        <div style="color: #666; font-size: 12px; margin-top: 4px; margin-left: 24px;">
                          压缩备份文件以节省存储空间，但会增加 CPU 使用量
                        </div>
                      </div>
                      <div style="flex: 1;">
                        <label nz-checkbox formControlName="enableEncryption">
                          <span style="margin-left: 8px;">启用加密</span>
                        </label>
                        <div style="color: #666; font-size: 12px; margin-top: 4px; margin-left: 24px;">
                          对备份文件进行加密保护，提高数据安全性
                        </div>
                      </div>
                    </div>

                    <!-- 模板定义 -->
                    <ng-template #basicHelpTemplate>
                      <i nz-icon nzType="question-circle" nz-tooltip nzTooltipTitle="配置备份任务的基本参数"></i>
                    </ng-template>

                    <ng-template #cronPrefixTemplate>
                      <i nz-icon nzType="schedule"></i>
                    </ng-template>

                    <ng-template #cronSuffixTemplate>
                      <i nz-icon nzType="info-circle" nz-tooltip nzTooltipTitle="留空表示立即执行一次性备份"></i>
                    </ng-template>
                  </nz-card>
                </form>

                <form nz-form [formGroup]="storageForm" style="margin-top:16px">
                  <nz-card nzTitle="存储配置" [nzExtra]="storageHelpTemplate">
                      <div class="form-row">
                      <div class="half">
                        <nz-form-item>
                          <nz-form-label [nzRequired]="true">存储类型</nz-form-label>
                          <nz-form-control>
                            <nz-select formControlName="storageType" nzPlaceHolder="选择存储类型">
                              <nz-option *ngFor="let s of storageProviders" [nzValue]="s.value" [nzLabel]="s.label">
                                <i nz-icon [nzType]="getStorageIcon(s.value)" style="margin-right: 8px;"></i>
                                {{ s.label }}
                              </nz-option>
                            </nz-select>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                      <div class="half">
                        <nz-form-item>
                          <nz-form-label [nzRequired]="true">Sink 名称</nz-form-label>
                          <nz-form-control>
                            <nz-select formControlName="sinkName" nzPlaceHolder="选择 HPFS 中已配置的 sink 名称" nzShowSearch>
                              <nz-option *ngFor="let s of hpfsSinks.filter(x => x.type?.toLowerCase() === (selectedStorageType||'').toLowerCase())"
                                         [nzValue]="s.name" [nzLabel]="s.name">
                                <i nz-icon [nzType]="getStorageIcon(s.type)"></i>
                                {{ s.name }}
                                <span style="color:#999;margin-left:8px">{{ s.endpoint || s.host || '' }}</span>
                              </nz-option>
                            </nz-select>
                            <div style="margin-top:6px; font-size:12px;">
                              <ng-container [ngSwitch]="sinkStatus">
                                <span *ngSwitchCase="'valid'" style="color:#52c41a">已校验：sink 存在</span>
                                <span *ngSwitchCase="'invalid'" style="color:#ff4d4f">未找到该 sink，请检查 HPFS 配置</span>
                                <span *ngSwitchCase="'checking'" style="color:#1890ff">正在校验...</span>
                                <span *ngSwitchDefault style="color:#999">从 HPFS ConfigMap 中选择 sink</span>
                              </ng-container>
                            </div>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                    </div>

                      <div *ngIf="selectedStorageType === 'oss'" formGroupName="ossConfig">
                      <nz-divider nzText="阿里云 OSS 配置" nzOrientation="left"></nz-divider>
                      <div class="help-text">该信息应配置在 <code>polardbx-operator-system/polardbx-hpfs-config</code> 的 <code>config.yaml</code> 中（前端不保存凭据）。例如：</div>
                      <pre class="config-summary" [innerText]="exampleOssYaml"></pre>
                        <div class="form-row">
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Access Key ID</nz-form-label>
                            <nz-form-control nzHasFeedback>
                              <input nz-input formControlName="accessKeyId" placeholder="请输入 Access Key ID" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Access Key Secret</nz-form-label>
                            <nz-form-control nzHasFeedback>
                              <nz-input-group [nzSuffix]="secretSuffixTemplate">
                                <input nz-input [type]="showSecret ? 'text' : 'password'" formControlName="accessKeySecret" placeholder="请输入 Access Key Secret" />
                              </nz-input-group>
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        </div>
                        <div class="form-row">
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Bucket</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="bucket" placeholder="存储桶名称" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Endpoint</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="endpoint" placeholder="例如：oss-cn-beijing.aliyuncs.com" />
                            </nz-form-control>
                          </nz-form-item>
                      </div>
                      </div>
                    </div>

                      <div *ngIf="selectedStorageType === 's3'" formGroupName="s3Config">
                      <nz-divider nzText="Amazon S3 配置" nzOrientation="left"></nz-divider>
                      <div class="help-text">在 HPFS ConfigMap 中配置 S3 sink，例如：</div>
                      <pre class="config-summary" [innerText]="exampleS3Yaml"></pre>
                        <div class="form-row">
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Access Key ID</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="accessKeyId" placeholder="请输入 Access Key ID" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Secret Access Key</nz-form-label>
                            <nz-form-control>
                              <nz-input-group [nzSuffix]="secretSuffixTemplate">
                                <input nz-input [type]="showSecret ? 'text' : 'password'" formControlName="secretAccessKey" placeholder="请输入 Secret Access Key" />
                              </nz-input-group>
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        </div>
                        <div class="form-row">
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Bucket</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="bucket" placeholder="存储桶名称" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">Region</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="region" placeholder="例如：us-east-1" />
                            </nz-form-control>
                          </nz-form-item>
                      </div>
                      </div>
                    </div>

                      <div *ngIf="selectedStorageType === 'sftp'" formGroupName="sftpConfig">
                      <nz-divider nzText="SFTP 配置" nzOrientation="left"></nz-divider>
                      <div class="help-text">在 HPFS ConfigMap 中配置 SFTP sink，例如：</div>
                      <pre class="config-summary" [innerText]="exampleSftpYaml"></pre>
                        <div class="form-row">
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">主机地址</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="host" placeholder="SFTP 服务器地址" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">端口</nz-form-label>
                            <nz-form-control>
                              <nz-input-number formControlName="port" [nzMin]="1" [nzMax]="65535" nzPlaceHolder="22" style="width: 100%;"></nz-input-number>
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        </div>
                        <div class="form-row">
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">用户名</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="username" placeholder="SFTP 用户名" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="half">
                          <nz-form-item>
                            <nz-form-label [nzRequired]="true">远程路径</nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="remotePath" placeholder="/backup/path" />
                            </nz-form-control>
                          </nz-form-item>
                      </div>
                      </div>
                    </div>

                    <!-- 模板定义 -->
                    <ng-template #storageHelpTemplate>
                      <i nz-icon nzType="question-circle" nz-tooltip nzTooltipTitle="选择合适的存储后端来保存备份文件"></i>
                    </ng-template>

                    <ng-template #secretSuffixTemplate>
                      <i nz-icon [nzType]="showSecret ? 'eye-invisible' : 'eye'" (click)="showSecret = !showSecret" style="cursor: pointer;"></i>
                    </ng-template>
                  </nz-card>
                </form>

                <form nz-form [formGroup]="retentionForm" style="margin-top:16px">
                  <nz-card nzTitle="保留策略" [nzExtra]="retentionHelpTemplate">
                    <div style="margin-bottom: 16px;">
                      <label nz-checkbox formControlName="enableRetention">
                        <span style="margin-left: 8px;">启用保留策略</span>
                      </label>
                      <div style="color: #666; font-size: 12px; margin-top: 4px;">
                        配置备份文件的自动清理规则，避免存储空间过度占用
                      </div>
                    </div>
                    
                    <div *ngIf="retentionForm.get('enableRetention')?.value">
                      <nz-alert 
                        nzType="info" 
                        nzMessage="保留策略说明" 
                        nzDescription="系统会按照以下规则保留备份文件，达到任一条件即触发清理。建议根据业务需求合理配置。"
                        nzShowIcon
                        style="margin-bottom: 16px;">
                      </nz-alert>
                      
                      <div class="form-row">
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="保留最近的 N 个备份文件">
                              <i nz-icon nzType="info-circle" style="margin-right: 4px;"></i>
                              保留数量
                            </nz-form-label>
                            <nz-form-control>
                              <nz-input-number 
                                formControlName="retain" 
                                [nzMin]="1" 
                                [nzMax]="999" 
                                nzPlaceHolder="10"
                                style="width: 100%;">
                              </nz-input-number>
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="保留指定天数内的备份文件">
                              <i nz-icon nzType="calendar" style="margin-right: 4px;"></i>
                              保留天数
                            </nz-form-label>
                            <nz-form-control>
                              <nz-input-number 
                                formControlName="retainDays" 
                                [nzMin]="1" 
                                [nzMax]="365" 
                                nzPlaceHolder="30"
                                style="width: 100%;">
                              </nz-input-number>
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="保留指定小时内的备份文件">
                              <i nz-icon nzType="clock-circle" style="margin-right: 4px;"></i>
                              保留小时
                            </nz-form-label>
                            <nz-form-control>
                              <nz-input-number 
                                formControlName="retainHours" 
                                [nzMin]="1" 
                                [nzMax]="8760" 
                                nzPlaceHolder="72"
                                style="width: 100%;">
                              </nz-input-number>
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                      </div>
                    </div>

                    <ng-template #retentionHelpTemplate>
                      <i nz-icon nzType="question-circle" nz-tooltip nzTooltipTitle="自动清理过期的备份文件，节省存储空间"></i>
                    </ng-template>
                  </nz-card>
                </form>

                <form nz-form [formGroup]="resourceForm" style="margin-top:16px">
                  <nz-card nzTitle="资源配置" [nzExtra]="resourceHelpTemplate">
                    <div style="margin-bottom: 16px;">
                      <label nz-checkbox formControlName="enableResourceLimits">
                        <span style="margin-left: 8px;">启用资源限制</span>
                      </label>
                      <div style="color: #666; font-size: 12px; margin-top: 4px;">
                        为备份任务分配合适的计算资源，确保备份性能和集群稳定性
                      </div>
                      </div>
                    
                    <div *ngIf="resourceForm.get('enableResourceLimits')?.value">
                      <nz-divider nzText="资源请求 (Requests)" nzOrientation="left"></nz-divider>
                      <div class="form-row">
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="容器启动时保证分配的 CPU 核数">
                              <i nz-icon nzType="dashboard" style="margin-right: 4px;"></i>
                              CPU
                            </nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="requestsCpu" placeholder="例如：100m" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="容器启动时保证分配的内存大小">
                              <i nz-icon nzType="database" style="margin-right: 4px;"></i>
                              Memory
                            </nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="requestsMemory" placeholder="例如：256Mi" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="备份过程中需要的临时存储空间">
                              <i nz-icon nzType="hdd" style="margin-right: 4px;"></i>
                              Storage
                            </nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="requestsStorage" placeholder="例如：1Gi" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                      </div>

                      <nz-divider nzText="资源上限 (Limits)" nzOrientation="left"></nz-divider>
                      <div class="form-row">
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="容器可使用的 CPU 上限">
                              <i nz-icon nzType="dashboard" style="margin-right: 4px;"></i>
                              CPU
                            </nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="limitsCpu" placeholder="例如：500m" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="容器可使用的内存上限">
                              <i nz-icon nzType="database" style="margin-right: 4px;"></i>
                              Memory
                            </nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="limitsMemory" placeholder="例如：512Mi" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                        <div class="third">
                          <nz-form-item>
                            <nz-form-label nzTooltipTitle="备份过程中可使用的存储上限">
                              <i nz-icon nzType="hdd" style="margin-right: 4px;"></i>
                              Storage
                            </nz-form-label>
                            <nz-form-control>
                              <input nz-input formControlName="limitsStorage" placeholder="例如：2Gi" />
                            </nz-form-control>
                          </nz-form-item>
                        </div>
                      </div>
                    </div>

                    <ng-template #resourceHelpTemplate>
                      <i nz-icon nzType="question-circle" nz-tooltip nzTooltipTitle="配置备份任务的计算资源分配"></i>
                    </ng-template>
                  </nz-card>
                </form>

                <nz-card nzTitle="操作确认" style="margin-top:16px">
                  <div style="margin-bottom: 16px;">
                    <nz-alert 
                      nzType="info" 
                      nzMessage="配置检查" 
                      [nzDescription]="getConfigSummary()"
                      nzShowIcon>
                    </nz-alert>
                  </div>
                  
                  <div style="display:flex; gap:12px; justify-content:flex-end; align-items: center;">
                    <button nz-button type="button" (click)="resetForms()" [nzLoading]="isProcessing">
                      <i nz-icon nzType="reload"></i>
                      重置配置
                    </button>
                    <button nz-button nzType="default" (click)="previewConfig()" [disabled]="!isFormValid()">
                      <i nz-icon nzType="eye"></i>
                      预览配置
                    </button>
                    <button nz-button nzType="primary" [disabled]="!isFormValid() || isProcessing" (click)="saveBackup()" [nzLoading]="isProcessing">
                      <i nz-icon nzType="save"></i>
                      {{ isProcessing ? '创建中...' : '创建备份配置' }}
                  </button>
                </div>
                </nz-card>
              </div>
            </mat-tab>
          </mat-tab-group>
        </div>
      </nz-card>
    </div>
  `,
  styles: [`
    .xstore-backup-management {
      padding: 20px;
    }
    .header-card .subtitle { color: rgba(0,0,0,0.45); }
    .mat-like-tabs { margin-top: 8px; }
    
    .tab-content {
      padding: 20px;
    }
    .list-actions { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
    .view-toggle { margin-left: auto; display: flex; gap: 8px; }
    
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

    /* 表单样式优化 */
    .form-row {
      display: flex;
      gap: 16px;
      margin-bottom: 16px;
      align-items: flex-start;
    }
    .form-row .half {
      flex: 1;
    }
    .form-row .third {
      flex: 1;
    }
    .form-row .full {
      flex: 1;
    }

    /* 卡片间距优化 */
    nz-card {
      margin-bottom: 0;
    }
    
    /* 表单验证反馈优化 */
    nz-form-explain {
      font-size: 12px;
      margin-top: 4px;
    }

    /* 帮助提示样式 */
    .help-text {
      color: #666;
      font-size: 12px;
      margin-top: 4px;
    }

    /* 配置汇总样式 */
    .config-summary {
      background: #f5f5f5;
      padding: 12px;
      border-radius: 6px;
      margin-bottom: 16px;
      font-family: monospace;
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
  hpfsSinks: Array<{ name: string; type: string; endpoint?: string; bucket?: string; host?: string; port?: number; rootPath?: string; bucketLookupType?: string }> = [];
  sinkStatus: 'idle' | 'valid' | 'invalid' | 'checking' = 'idle';
  backups: XStoreBackupWithStatus[] = [];
  dataSource = new MatTableDataSource<XStoreBackupWithStatus>();
  // 汇总/明细视图切换，默认汇总
  viewMode: 'summary' | 'detail' = 'summary';
  // 明细计数与名称用于汇总视图显示类别细分（DN/GMS）
  private detailCountsByParent: Record<string, { dn: number; gms: number; dnNames: string[]; gmsNames: string[] }> = {};
  
  // UI State
  isProcessing = false;
  selectedTab = 0;
  showSecret = false;
  private destroy$ = new Subject<void>();
  // Examples for HPFS ConfigMap snippets
  exampleOssYaml: string = `sinks:\n  - name: default\n    type: oss\n    endpoint: oss-cn-beijing.aliyuncs.com\n    accessKey: <OSS_AK>\n    accessSecret: <OSS_SK>\n    bucket: my-bucket\n`;
  exampleS3Yaml: string = `sinks:\n  - name: default\n    type: s3\n    endpoint: play.min.io\n    useSSL: true\n    bucketLookupType: dns\n    accessKey: <S3_AK>\n    secretKey: <S3_SK>\n    bucket: my-bucket\n`;
  exampleSftpYaml: string = `sinks:\n  - name: default\n    type: sftp\n    host: sftp.example.com\n    port: 22\n    user: backup\n    password: <SFTP_PASSWORD>\n    rootPath: /data/backup\n`;
  
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
      sinkName: ['', Validators.required],
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
    // 后端契约：仅需 storageProvider.storageName + storageProvider.sink
    const sinkCtl = this.storageForm.get('sinkName');
    sinkCtl?.setValidators([Validators.required]);
    sinkCtl?.updateValueAndValidity();
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
      
      // Load HPFS sinks for selection
      const sinksResp = await this.apiService.getHpfsSinks().toPromise();
      this.hpfsSinks = (sinksResp?.sinks || []).filter((s: any) => !!s?.name && !!s?.type);

      // Wire sink validation on change
      const sinkCtl = this.storageForm.get('sinkName');
      const typeCtl = this.storageForm.get('storageType');
      if (sinkCtl && typeCtl) {
        sinkCtl.valueChanges.pipe(debounceTime(200), distinctUntilChanged(), takeUntil(this.destroy$)).subscribe(async (val: string) => {
          this.sinkStatus = 'checking';
          try {
            const name = (val || '').trim();
            const type = (typeCtl.value || '').toString();
            if (!name || !type) { this.sinkStatus = 'idle'; return; }
            const res = await this.apiService.validateSink(name, type).toPromise();
            this.sinkStatus = (res?.status === 'ok') ? 'valid' : 'invalid';
          } catch {
            this.sinkStatus = 'invalid';
          }
        });
        typeCtl.valueChanges.pipe(takeUntil(this.destroy$)).subscribe(() => {
          // Reset sink selection when type changes
          sinkCtl.setValue('');
          this.sinkStatus = 'idle';
        });
      }

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
    const sink = (storageValue.sinkName || localStorage.getItem('xstoreBackupSinkName') || localStorage.getItem('backupSinkName') || 'default').trim();

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

  // Map material chip color semantics to ng-zorro tag colors
  getNzStatusColor(phase?: string): string {
    if (!phase) return 'default';
    const ok = ['Completed', 'Finished'];
    const running = ['Running', 'Pending', 'Backuping', 'Collecting', 'Calculating', 'Binloging', 'MetadataBackuping', 'Waiting', 'Deleting'];
    const fail = ['Failed'];
    if (ok.includes(phase)) return 'green';
    if (fail.includes(phase)) return 'red';
    if (running.includes(phase)) return 'blue';
    return 'default';
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

  getStorageIcon(type: string): string {
    switch (type) {
      case 'oss': return 'cloud';
      case 's3': return 'amazon';
      case 'sftp': return 'share-alt';
      default: return 'database';
    }
  }

  getBackupTypeIcon(type: string): string {
    switch (type) {
      case 'full': return 'database';
      case 'incremental': return 'diff';
      default: return 'file-sync';
    }
  }

  getConfigSummary(): string {
    const name = this.backupForm.get('name')?.value || '未设置';
    const xstore = this.backupForm.get('xStoreName')?.value || '未选择';
    const type = this.backupForm.get('backupType')?.value || '未选择';
    const storage = this.storageForm.get('storageType')?.value || '未选择';
    const schedule = this.backupForm.get('schedule')?.value ? '定时备份' : '立即备份';
    
    return `备份名称：${name} | 目标：${xstore} | 类型：${type} | 存储：${storage} | 模式：${schedule}`;
  }

  previewConfig(): void {
    // 可以在这里显示配置预览对话框
    console.log('预览配置:', {
      backup: this.backupForm.value,
      storage: this.storageForm.value,
      retention: this.retentionForm.value,
      resource: this.resourceForm.value
    });
  }
}