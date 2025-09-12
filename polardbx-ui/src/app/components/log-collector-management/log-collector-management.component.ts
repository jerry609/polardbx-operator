import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzCollapseModule } from 'ng-zorro-antd/collapse';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';
import { NzProgressModule } from 'ng-zorro-antd/progress';
import { NzPageHeaderModule } from 'ng-zorro-antd/page-header';
import { Subject } from 'rxjs';
import { takeUntil, finalize } from 'rxjs/operators';

import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { 
  PolarDBXLogCollector, 
  PolarDBXLogCollectorList,
  CreateLogCollectorRequest,
  UpdateLogCollectorRequest,
  COMPONENT_PRESETS,
  NAMING_PATTERNS,
  ComponentPreset,
  ComponentNamingPattern,
  validateComponentName,
  getCollectorDescription,
  getReadinessPercentage,
  getCollectorStatusColor,
  getCollectorStatusText,
  generateComponentName
} from '../../models/log-collector.model';
// Removed ConfirmationDialogComponent import - using native confirm() instead

@Component({
  selector: 'app-log-collector-management',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NzTabsModule,
    NzTableModule,
    NzCardModule,
    NzButtonModule,
    NzIconModule,
    NzInputModule,
    NzSelectModule,
    NzFormModule,
    NzTagModule,
    NzSpinModule,
    NzToolTipModule,
    NzSwitchModule,
    NzDividerModule,
    NzCollapseModule,
    NzDropDownModule,
    NzProgressModule,
    NzPageHeaderModule
  ],
  template: `
    <div class="log-collector-management">
      <nz-page-header [nzGhost]="false" nzTitle="日志采集器管理" nzSubtitle="管理 FileBeat 和 LogStash 组件的日志采集和处理"></nz-page-header>

      <nz-tabset class="main-tabs" [(nzSelectedIndex)]="selectedTab" (nzSelectedIndexChange)="onTabChange($event)">
        <!-- 日志采集器列表选项卡 -->
        <nz-tab nzTitle="日志采集器">
            <div class="tab-content">
              <div class="actions-toolbar">
                <button nz-button nzType="default" (click)="refreshCollectors()" [disabled]="isLoading('LOG_COLLECTOR_LIST')">
                  <i nz-icon nzType="reload"></i>
                  <span>刷新</span>
                </button>
                <button nz-button nzType="primary" (click)="selectedTab = 1">
                  <i nz-icon nzType="plus"></i>
                  <span>创建采集器</span>
                </button>
              </div>

              <nz-card class="table-card">
                <div class="table-container" *ngIf="!isLoading('LOG_COLLECTOR_LIST'); else loadingTemplate">
                  <nz-table [nzData]="logCollectors" [nzShowPagination]="false" class="collectors-table">
                    <thead>
                      <tr>
                        <th>采集器名称</th>
                        <th>命名空间</th>
                        <th>组件</th>
                        <th>状态</th>
                        <th>配置</th>
                        <th>创建时间</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr *ngFor="let collector of logCollectors">
                        <td>
                          <div class="collector-name">
                            <i nz-icon nzType="book"></i>
                            <span>{{ collector.metadata.name }}</span>
                          </div>
                        </td>
                        <td>{{ collector.metadata.namespace }}</td>
                        <td>
                          <div class="components-info">
                            <div *ngIf="collector.spec.fileBeatName" class="component-item">
                              <i nz-icon nzType="file-text" class="filebeat-icon"></i>
                              <span>{{ collector.spec.fileBeatName }}</span>
                            </div>
                            <div *ngIf="collector.spec.logStashName" class="component-item">
                              <i nz-icon nzType="deployment-unit" class="logstash-icon"></i>
                              <span>{{ collector.spec.logStashName }}</span>
                            </div>
                            <div *ngIf="!collector.spec.fileBeatName && !collector.spec.logStashName" class="no-components">
                              <i nz-icon nzType="warning"></i>
                              <span>无组件</span>
                            </div>
                          </div>
                        </td>
                        <td>
                          <div class="status-info">
                            <div class="status-summary">
                              <nz-tag [nzColor]="getStatusColor(collector)">{{ getStatusText(collector) }}</nz-tag>
                            </div>
                            <div class="readiness-bar" *ngIf="hasStatus(collector)">
                              <nz-progress [nzPercent]="getReadiness(collector)" [nzStatus]="getReadiness(collector) >= 100 ? 'success' : 'active'"></nz-progress>
                              <span class="readiness-text">{{ getReadiness(collector) }}% 就绪</span>
                            </div>
                          </div>
                        </td>
                        <td>
                          <div class="config-info">
                            <div *ngIf="collector.status?.configStatus?.fileBeatConfigId" class="config-item">
                              <i nz-icon nzType="setting"></i>
                              <span>FB: {{ (collector.status?.configStatus?.fileBeatConfigId || '').substring(0, 8) }}...</span>
                            </div>
                            <div *ngIf="collector.status?.configStatus?.logStashConfigId" class="config-item">
                              <i nz-icon nzType="sliders"></i>
                              <span>LS: {{ (collector.status?.configStatus?.logStashConfigId || '').substring(0, 8) }}...</span>
                            </div>
                          </div>
                        </td>
                        <td>{{ formatDate(collector.metadata.creationTimestamp) }}</td>
                        <td>
                          <button nz-button nzType="default" nz-dropdown [nzDropdownMenu]="collectorMenu" [disabled]="isLoading('LOG_COLLECTOR_UPDATE')">
                            <span>操作</span>
                            <i nz-icon nzType="down"></i>
                          </button>
                          <nz-dropdown-menu #collectorMenu="nzDropdownMenu">
                            <ul nz-menu>
                              <li nz-menu-item (click)="viewCollectorDetails(collector)"><i nz-icon nzType="eye"></i> 查看详情</li>
                              <li nz-menu-item (click)="editCollector(collector)"><i nz-icon nzType="edit"></i> 编辑采集器</li>
                              <li nz-menu-item (click)="viewConfiguration(collector)" [nzDisabled]="!hasConfiguration(collector)"><i nz-icon nzType="code"></i> 查看配置</li>
                              <li nz-menu-item (click)="deleteCollector(collector)" class="delete-action"><i nz-icon nzType="delete"></i> 删除采集器</li>
                            </ul>
                          </nz-dropdown-menu>
                        </td>
                      </tr>
                    </tbody>
                  </nz-table>

                  <div class="no-data" *ngIf="logCollectors.length === 0">
                    <i nz-icon nzType="book"></i>
                    <p>未找到日志采集器。创建您的第一个采集器开始使用。</p>
                  </div>
                </div>
              </nz-card>
            </div>
        </nz-tab>

        <!-- 创建/编辑采集器选项卡 -->
        <nz-tab nzTitle="创建采集器">
            <div class="tab-content">
              <nz-card class="form-card" [nzTitle]="editingCollector ? '编辑日志采集器' : '创建新日志采集器'">
                <form [formGroup]="collectorForm" class="collector-form">
                    <!-- 基本配置 -->
                    <nz-collapse class="form-section" [nzAccordion]="false">
                      <nz-collapse-panel nzHeader="基本配置" [nzActive]="true">

                      <div class="form-row">
                        <nz-form-item class="full-width">
                          <nz-form-label [nzSpan]="6" nzFor="name" nzRequired>采集器名称</nz-form-label>
                          <nz-form-control [nzSpan]="18" nzHasFeedback [nzErrorTip]="nameErrorTpl">
                            <input nz-input id="name" formControlName="name" placeholder="输入采集器名称" />
                            <ng-template #nameErrorTpl let-control>
                              <ng-container *ngIf="control.hasError('required')">采集器名称是必填项</ng-container>
                              <ng-container *ngIf="control.hasError('pattern')">名称必须是合法的 Kubernetes 资源名称</ng-container>
                            </ng-template>
                          </nz-form-control>
                        </nz-form-item>
                      </div>

                      <div class="form-row">
                        <nz-form-item class="half-width">
                          <nz-form-label [nzSpan]="6" nzFor="namespace">命名空间</nz-form-label>
                          <nz-form-control [nzSpan]="18">
                            <input nz-input formControlName="namespace" placeholder="default" />
                          </nz-form-control>
                        </nz-form-item>
                        <nz-form-item class="half-width">
                          <nz-form-label [nzSpan]="6" nzFor="preset">组件预设</nz-form-label>
                          <nz-form-control [nzSpan]="18">
                            <nz-select formControlName="preset">
                              <nz-option [nzValue]="null" nzLabel="自定义配置"></nz-option>
                              <nz-option *ngFor="let preset of componentPresets" [nzValue]="preset.name" [nzLabel]="preset.label + ' - ' + preset.description"></nz-option>
                            </nz-select>
                          </nz-form-control>
                        </nz-form-item>
                      </div>
                      </nz-collapse-panel>
                    </nz-collapse>

                    <!-- FileBeat 配置 -->
                    <nz-collapse class="form-section">
                      <nz-collapse-panel nzHeader="FileBeat 配置" [nzActive]="false">

                      <div class="component-section">
                        <div class="form-row">
                          <mat-slide-toggle formControlName="enableFileBeat" class="component-toggle">
                            启用 FileBeat 组件
                          </mat-slide-toggle>
                        </div>

                        <div *ngIf="collectorForm.get('enableFileBeat')?.value" class="component-config">
                          <div class="form-row">
                            <nz-form-item class="full-width">
                              <nz-form-label [nzSpan]="6" nzFor="fileBeatName">FileBeat 名称</nz-form-label>
                              <nz-form-control [nzSpan]="18" nzHasFeedback [nzExtra]="'用于日志采集的 FileBeat 组件名称'">
                                <input nz-input formControlName="fileBeatName" placeholder="filebeat-main" />
                              </nz-form-control>
                            </nz-form-item>
                          </div>

                          <div class="info-section">
                            <i nz-icon nzType="info-circle"></i>
                            <div class="info-content">
                              <h4>FileBeat 组件信息</h4>
                              <ul>
                                <li><strong>用途</strong>: 从各种来源采集和转发日志文件</li>
                                <li><strong>功能</strong>: 监控日志文件并将日志事件发送到处理管道</li>
                                <li><strong>配置</strong>: 根据集群设置自动配置</li>
                                <li><strong>性能</strong>: 轻量级日志托运程序，资源占用极小</li>
                              </ul>
                            </div>
                          </div>
                        </div>
                      </div>
                      </nz-collapse-panel>
                    </nz-collapse>

                    <!-- LogStash 配置 -->
                    <nz-collapse class="form-section">
                      <nz-collapse-panel nzHeader="LogStash 配置" [nzActive]="false">

                      <div class="component-section">
                        <div class="form-row">
                          <mat-slide-toggle formControlName="enableLogStash" class="component-toggle">
                            启用 LogStash 组件
                          </mat-slide-toggle>
                        </div>

                        <div *ngIf="collectorForm.get('enableLogStash')?.value" class="component-config">
                          <div class="form-row">
                            <nz-form-item class="full-width">
                              <nz-form-label [nzSpan]="6" nzFor="logStashName">LogStash 名称</nz-form-label>
                              <nz-form-control [nzSpan]="18" nzHasFeedback [nzExtra]="'用于日志处理的 LogStash 组件名称'">
                                <input nz-input formControlName="logStashName" placeholder="logstash-main" />
                              </nz-form-control>
                            </nz-form-item>
                          </div>

                          <div class="info-section">
                            <i nz-icon nzType="info-circle"></i>
                            <div class="info-content">
                              <h4>LogStash 组件信息</h4>
                              <ul>
                                <li><strong>用途</strong>: 处理、转换和丰富日志数据</li>
                                <li><strong>功能</strong>: 对日志事件应用过滤器、解析和路由</li>
                                <li><strong>管道</strong>: 可配置的输入、过滤器和输出插件</li>
                                <li><strong>可扩展性</strong>: 支持高吞吐量场景的水平扩展</li>
                              </ul>
                            </div>
                          </div>
                        </div>
                      </div>
                      </nz-collapse-panel>
                    </nz-collapse>

                    <!-- 命名模式 -->
                    <nz-collapse class="form-section">
                      <nz-collapse-panel nzHeader="命名模式" [nzActive]="false">

                      <div class="naming-patterns">
                        <div class="form-row">
                          <nz-form-item class="full-width">
                            <nz-form-label [nzSpan]="6" nzFor="pattern">命名模式</nz-form-label>
                            <nz-form-control [nzSpan]="18">
                              <nz-select formControlName="pattern">
                                <nz-option [nzValue]="null" nzLabel="自定义名称"></nz-option>
                                <nz-option *ngFor="let pattern of namingPatterns" [nzValue]="pattern" [nzLabel]="pattern.label + ' - ' + pattern.description"></nz-option>
                              </nz-select>
                            </nz-form-control>
                          </nz-form-item>
                        </div>

                        <div class="pattern-variables" *ngIf="selectedNamingPattern">
                          <div class="form-row">
                            <nz-form-item class="half-width">
                              <nz-form-label [nzSpan]="6" nzFor="patternEnv">环境</nz-form-label>
                              <nz-form-control [nzSpan]="18">
                                <input nz-input formControlName="patternEnv" placeholder="prod" />
                              </nz-form-control>
                            </nz-form-item>
                            <nz-form-item class="half-width">
                              <nz-form-label [nzSpan]="6" nzFor="patternCluster">集群 ID</nz-form-label>
                              <nz-form-control [nzSpan]="18">
                                <input nz-input formControlName="patternCluster" placeholder="main" />
                              </nz-form-control>
                            </nz-form-item>
                          </div>
                          <div class="form-row">
                            <nz-form-item class="full-width">
                              <nz-form-label [nzSpan]="6" nzFor="patternFunction">功能/用途</nz-form-label>
                              <nz-form-control [nzSpan]="18">
                                <input nz-input formControlName="patternFunction" placeholder="database-logs" />
                              </nz-form-control>
                            </nz-form-item>
                          </div>
                          <button nz-button nzType="default" nzGhost type="button" (click)="generateNames()" class="generate-btn">
                            <i nz-icon nzType="highlight"></i>
                            <span>生成组件名称</span>
                          </button>
                        </div>
                      </div>
                      </nz-collapse-panel>
                    </nz-collapse>
                </form>
                <div class="actions" style="display:flex; gap:12px; justify-content:flex-end; padding:12px 0;">
                  <button nz-button nzType="default" (click)="resetForm()" [disabled]="isLoading('LOG_COLLECTOR_CREATE')">重置</button>
                  <button nz-button nzType="primary" (click)="submitCollector()" [disabled]="collectorForm.invalid || isLoading('LOG_COLLECTOR_CREATE')">
                    <i nz-icon [nzType]="editingCollector ? 'save' : 'plus'"></i>
                    <span>{{ editingCollector ? '更新采集器' : '创建采集器' }}</span>
                  </button>
                </div>
              </nz-card>
            </div>
        </nz-tab>
      </nz-tabset>
    </div>

    <!-- 加载模板 -->
    <ng-template #loadingTemplate>
      <div class="loading-container">
        <nz-spin nzSimple></nz-spin>
        <p>正在加载日志采集器...</p>
      </div>
    </ng-template>
  `,
  styleUrl: './log-collector-management.component.scss'
})
export class LogCollectorManagementComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  
  logCollectors: PolarDBXLogCollector[] = [];
  displayedColumns: string[] = ['name', 'namespace', 'components', 'status', 'configuration', 'createdTime', 'actions'];
  selectedTab = 0;
  editingCollector: PolarDBXLogCollector | null = null;
  selectedNamingPattern: ComponentNamingPattern | null = null;
  
  componentPresets = COMPONENT_PRESETS;
  namingPatterns = NAMING_PATTERNS;
  
  collectorForm: FormGroup;

  constructor(
    private apiService: ApiService,
    private loadingService: LoadingService,
    private fb: FormBuilder,
    private message: NzMessageService
  ) {
    this.collectorForm = this.createCollectorForm();
  }

  ngOnInit(): void {
    this.loadLogCollectors();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private createCollectorForm(): FormGroup {
    return this.fb.group({
      name: ['', [Validators.required, Validators.pattern(/^[a-z0-9-]+$/)]],
      namespace: ['default'],
      preset: [null],
      enableFileBeat: [false],
      fileBeatName: [''],
      enableLogStash: [false],
      logStashName: [''],
      pattern: [null],
      patternEnv: [''],
      patternCluster: [''],
      patternFunction: ['']
    });
  }

  isLoading(key: keyof typeof LoadingKeys): boolean {
    return this.loadingService.isLoading(LoadingKeys[key]);
  }

  onTabChange(index: number): void {
    this.selectedTab = index;
    if (index === 0) {
      this.editingCollector = null;
      this.resetForm();
    }
  }

  loadLogCollectors(): void {
    this.apiService.getLogCollectors()
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => {})
      )
      .subscribe({
        next: (collectors: PolarDBXLogCollector[]) => {
          this.logCollectors = collectors || [];
        },
        error: (error) => {
          console.error('加载日志采集器失败:', error);
          this.message.error('加载日志采集器失败');
        }
      });
  }

  refreshCollectors(): void {
    this.loadLogCollectors();
  }

  applyComponentPreset(presetName: string | null): void {
    if (!presetName) return;
    
    const preset = this.componentPresets.find(p => p.name === presetName);
    if (!preset) return;

    this.collectorForm.patchValue({
      enableFileBeat: !!preset.fileBeatName,
      fileBeatName: preset.fileBeatName || '',
      enableLogStash: !!preset.logStashName,
      logStashName: preset.logStashName || ''
    });
  }

  applyNamingPattern(pattern: ComponentNamingPattern | null): void {
    this.selectedNamingPattern = pattern;
    if (pattern) {
      this.collectorForm.patchValue({
        patternEnv: 'prod',
        patternCluster: 'main',
        patternFunction: 'database-logs'
      });
    }
  }

  generateNames(): void {
    if (!this.selectedNamingPattern) return;

    const variables = {
      env: this.collectorForm.get('patternEnv')?.value || 'prod',
      cluster: this.collectorForm.get('patternCluster')?.value || 'main',
      function: this.collectorForm.get('patternFunction')?.value || 'logs'
    };

    const fileBeatName = generateComponentName(this.selectedNamingPattern.fileBeatPattern, variables);
    const logStashName = generateComponentName(this.selectedNamingPattern.logStashPattern, variables);

    this.collectorForm.patchValue({
      fileBeatName: fileBeatName,
      logStashName: logStashName,
      enableFileBeat: true,
      enableLogStash: true
    });
  }

  submitCollector(): void {
    if (this.collectorForm.invalid) return;

    const formValue = this.collectorForm.value;
    const collectorRequest: CreateLogCollectorRequest = {
      name: formValue.name,
      namespace: formValue.namespace || 'default',
      fileBeatName: formValue.enableFileBeat ? formValue.fileBeatName : undefined,
      logStashName: formValue.enableLogStash ? formValue.logStashName : undefined
    };

    const operation = this.editingCollector
      ? this.apiService.updateLogCollector(this.editingCollector.metadata.namespace!, {
          ...this.editingCollector,
          spec: {
            fileBeatName: collectorRequest.fileBeatName,
            logStashName: collectorRequest.logStashName
          }
        })
      : this.apiService.createLogCollector(collectorRequest.namespace!, collectorRequest);

    operation.pipe(
      takeUntil(this.destroy$),
      finalize(() => {})
    ).subscribe({
      next: (collector) => {
            const messageText = this.editingCollector ? '日志采集器更新成功' : '日志采集器创建成功';
    this.message.success(messageText);
        this.resetForm();
        this.selectedTab = 0;
        this.loadLogCollectors();
      },
      error: (error) => {
        console.error('保存日志采集器失败:', error);
        this.message.error('保存日志采集器失败');
      }
    });
  }

  editCollector(collector: PolarDBXLogCollector): void {
    this.editingCollector = collector;
    this.collectorForm.patchValue({
      name: collector.metadata.name,
      namespace: collector.metadata.namespace,
      enableFileBeat: !!collector.spec.fileBeatName,
      fileBeatName: collector.spec.fileBeatName || '',
      enableLogStash: !!collector.spec.logStashName,
      logStashName: collector.spec.logStashName || ''
    });
    this.selectedTab = 1;
  }

  deleteCollector(collector: PolarDBXLogCollector): void {
    if (confirm(`删除日志采集器\n\n确定要删除日志采集器 "${collector.metadata.name}" 吗？`)) {
      this.apiService.deleteLogCollector(collector.metadata.namespace!, collector.metadata.name)
        .pipe(
          takeUntil(this.destroy$),
          finalize(() => {})
        )
        .subscribe({
          next: () => {
            this.message.success('日志采集器删除成功');
            this.loadLogCollectors();
          },
          error: (error) => {
            console.error('删除日志采集器失败:', error);
            this.message.error('删除日志采集器失败');
          }
        });
    }
  }

  viewCollectorDetails(collector: PolarDBXLogCollector): void {
    // TODO: 实现采集器详情对话框
    console.log('查看采集器详情:', collector);
  }

  viewConfiguration(collector: PolarDBXLogCollector): void {
    // TODO: 实现配置查看器对话框
    console.log('查看配置:', collector);
  }

  resetForm(): void {
    this.editingCollector = null;
    this.selectedNamingPattern = null;
    this.collectorForm.reset({
      name: '',
      namespace: 'default',
      enableFileBeat: false,
      fileBeatName: '',
      enableLogStash: false,
      logStashName: '',
      patternEnv: '',
      patternCluster: '',
      patternFunction: ''
    });
  }

  formatDate(dateString?: string): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  }

  getCollectorDescription(collector: PolarDBXLogCollector): string {
    return getCollectorDescription(collector);
  }

  getReadiness(collector: PolarDBXLogCollector): number {
    return getReadinessPercentage(collector.status?.configStatus);
  }

  getStatusColor(collector: PolarDBXLogCollector): string {
    return getCollectorStatusColor(collector);
  }

  getStatusText(collector: PolarDBXLogCollector): string {
    return getCollectorStatusText(collector);
  }

  hasStatus(collector: PolarDBXLogCollector): boolean {
    return !!collector.status?.configStatus;
  }

  hasConfiguration(collector: PolarDBXLogCollector): boolean {
    return !!(collector.status?.configStatus?.fileBeatConfigId || collector.status?.configStatus?.logStashConfigId);
  }
}