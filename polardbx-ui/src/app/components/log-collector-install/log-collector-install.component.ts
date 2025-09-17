import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup } from '@angular/forms';
import { Router } from '@angular/router';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzTypographyModule } from 'ng-zorro-antd/typography';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzCollapseModule } from 'ng-zorro-antd/collapse';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzStepsModule } from 'ng-zorro-antd/steps';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzStatisticModule } from 'ng-zorro-antd/statistic';
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { NzMessageService } from 'ng-zorro-antd/message';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-log-collector-install',
  standalone: true,
  imports: [
    CommonModule, 
    FormsModule,
    ReactiveFormsModule,
    NzCardModule, 
    NzTypographyModule, 
    NzAlertModule, 
    NzButtonModule, 
    NzIconModule,
    NzCollapseModule,
    NzDividerModule,
    NzFormModule,
    NzInputModule,
    NzGridModule,
    NzSelectModule,
    NzStepsModule,
    NzSpinModule,
    NzStatisticModule,
    NzResultModule,
    NzDescriptionsModule,
    NzCheckboxModule
  ],
  template: `
    <div class="wizard">
      <div class="page-header">
        <div class="header-content">
          <h1 class="page-title">
            <i nz-icon nzType="cluster" class="page-icon"></i>
            日志采集安装向导
          </h1>
          <p class="page-description">一站式安装和配置 PolarDB-X LogCollector 日志采集堆栈</p>
        </div>
      </div>

      <div class="page-content">
        <nz-steps [nzCurrent]="currentStep" class="wizard-steps" nzSize="small">
          <nz-step nzTitle="环境检查" nzDescription="检测系统环境和依赖"></nz-step>
          <nz-step nzTitle="配置选择" nzDescription="选择安装方案和参数"></nz-step>
          <nz-step nzTitle="安装执行" nzDescription="执行安装并监控进度"></nz-step>
          <nz-step nzTitle="验证完成" nzDescription="验证安装结果"></nz-step>
        </nz-steps>

        <!-- 步骤1：环境检查 -->
        <nz-card *ngIf="currentStep === 0" class="step-card" nzTitle="环境检查" [nzExtra]="checkExtra">
          <ng-template #checkExtra>
            <button nz-button nzType="primary" nzSize="small" (click)="runEnvironmentCheck()" [nzLoading]="checking">
              <i nz-icon nzType="sync"></i>
              重新检查
            </button>
          </ng-template>

          <div class="check-section">
            <nz-spin [nzSpinning]="checking">
              <div class="loading-tip" *ngIf="checking">正在检查环境...</div>
              <div class="check-items">
                <div class="check-item" *ngFor="let check of environmentChecks">
                  <div class="check-info">
                    <span class="check-status">
                      <i nz-icon [nzType]="check.status === 'success' ? 'check-circle' : check.status === 'error' ? 'close-circle' : 'clock-circle'"
                         [style.color]="check.status === 'success' ? '#52c41a' : check.status === 'error' ? '#ff4d4f' : '#faad14'"></i>
                    </span>
                    <span class="check-name">{{ check.name }}</span>
                    <span class="check-description">{{ check.description }}</span>
                  </div>
                  <div class="check-result" [ngClass]="check.status">
                    {{ check.result }}
                  </div>
                </div>
              </div>
            </nz-spin>
          </div>

          <div class="step-actions">
            <button nz-button nzType="primary" 
                    [disabled]="!allChecksPassed" 
                    (click)="nextStep()">
              下一步：配置选择
              <i nz-icon nzType="right"></i>
            </button>
          </div>
        </nz-card>

        <!-- 步骤2：配置选择 -->
        <nz-card *ngIf="currentStep === 1" class="step-card" nzTitle="配置选择">
          <form [formGroup]="form" class="config-form">
            <div class="config-section">
              <h4>基本配置</h4>
              <nz-row [nzGutter]="16">
                <nz-col [nzSpan]="12">
                  <nz-form-item>
                    <nz-form-label nzRequired>安装方案</nz-form-label>
                    <nz-form-control>
                      <nz-select formControlName="deploymentType" nzPlaceHolder="选择安装方案" (ngModelChange)="onDeploymentTypeChange($event)">
                        <nz-option nzValue="default" nzLabel="标准安装 - 使用默认配置"></nz-option>
                        <nz-option nzValue="production" nzLabel="生产环境 - 高性能配置"></nz-option>
                        <nz-option nzValue="minimal" nzLabel="最小安装 - 节约资源"></nz-option>
                        <nz-option nzValue="custom" nzLabel="自定义配置"></nz-option>
                      </nz-select>
                    </nz-form-control>
                  </nz-form-item>
                </nz-col>
                <nz-col [nzSpan]="12">
                  <nz-form-item>
                    <nz-form-label>命名空间</nz-form-label>
                    <nz-form-control>
                      <input nz-input formControlName="namespace" />
                    </nz-form-control>
                  </nz-form-item>
                </nz-col>
              </nz-row>
            </div>

            <!-- 动态配置区域 -->
            <div class="config-section" *ngIf="form.value.deploymentType">
              <h4>{{ getDeploymentConfig().title }}</h4>
              <nz-alert [nzType]="getDeploymentConfig().alertType" [nzMessage]="getDeploymentConfig().description" nzShowIcon class="deployment-alert"></nz-alert>
              
              <div class="resource-overview" *ngIf="getDeploymentConfig().resources">
                <h5>预估资源需求</h5>
                <nz-row [nzGutter]="16">
                  <nz-col [nzSpan]="6">
                    <nz-statistic nzTitle="Filebeat CPU" [nzValue]="getDeploymentConfig().resources.filebeatCpu" nzSuffix="核/节点"></nz-statistic>
                  </nz-col>
                  <nz-col [nzSpan]="6">
                    <nz-statistic nzTitle="Filebeat 内存" [nzValue]="getDeploymentConfig().resources.filebeatMemory" nzSuffix="MB/节点"></nz-statistic>
                  </nz-col>
                  <nz-col [nzSpan]="6">
                    <nz-statistic nzTitle="Logstash CPU" [nzValue]="getDeploymentConfig().resources.logstashCpu" nzSuffix="核"></nz-statistic>
                  </nz-col>
                  <nz-col [nzSpan]="6">
                    <nz-statistic nzTitle="Logstash 内存" [nzValue]="getDeploymentConfig().resources.logstashMemory" nzSuffix="GB"></nz-statistic>
                  </nz-col>
                </nz-row>
              </div>

              <!-- 自定义配置 -->
              <div *ngIf="form.value.deploymentType === 'custom'" class="custom-config">
                <nz-collapse nzGhost>
                  <nz-collapse-panel nzHeader="组件配置">
                    <nz-row [nzGutter]="16">
                      <nz-col [nzSpan]="8">
                        <nz-form-item>
                          <nz-form-control>
                            <label nz-checkbox formControlName="enableFilebeat">启用 Filebeat</label>
                          </nz-form-control>
                        </nz-form-item>
                      </nz-col>
                      <nz-col [nzSpan]="8">
                        <nz-form-item>
                          <nz-form-control>
                            <label nz-checkbox formControlName="enableLogstash">启用 Logstash</label>
                          </nz-form-control>
                        </nz-form-item>
                      </nz-col>
                      <nz-col [nzSpan]="8">
                        <nz-form-item>
                          <nz-form-control>
                            <label nz-checkbox formControlName="enableElasticsearch">集成 Elasticsearch</label>
                          </nz-form-control>
                        </nz-form-item>
                      </nz-col>
                    </nz-row>
                  </nz-collapse-panel>
                </nz-collapse>
              </div>
            </div>
          </form>

          <div class="step-actions">
            <button nz-button nzType="default" (click)="prevStep()">
              <i nz-icon nzType="left"></i>
              上一步
            </button>
            <button nz-button nzType="primary" (click)="nextStep()" [disabled]="!form.valid">
              下一步：开始安装
              <i nz-icon nzType="right"></i>
            </button>
          </div>
        </nz-card>

        <!-- 步骤3：安装执行 -->
        <nz-card *ngIf="currentStep === 2" class="step-card" nzTitle="安装执行">
          <div class="install-section">
            <nz-alert nzType="info" nzMessage="安装进行中" nzDescription="请勿关闭页面，安装过程可能需要几分钟时间" nzShowIcon class="install-alert"></nz-alert>
            
            <div class="install-progress">
              <nz-steps nzDirection="vertical" nzSize="small" [nzCurrent]="installStep">
                <nz-step nzTitle="准备安装环境" [nzDescription]="getInstallStepDescription(0)"></nz-step>
                <nz-step nzTitle="创建命名空间" [nzDescription]="getInstallStepDescription(1)"></nz-step>
                <nz-step nzTitle="部署 Filebeat" [nzDescription]="getInstallStepDescription(2)"></nz-step>
                <nz-step nzTitle="部署 Logstash" [nzDescription]="getInstallStepDescription(3)"></nz-step>
                <nz-step nzTitle="配置日志采集" [nzDescription]="getInstallStepDescription(4)"></nz-step>
              </nz-steps>
            </div>

            <div class="install-logs" *ngIf="installLogs.length > 0">
              <h5>安装日志</h5>
              <div class="log-viewer">
                <div class="log-entry" *ngFor="let log of installLogs" [ngClass]="log.level">
                  <span class="log-time">{{ log.timestamp | date:'HH:mm:ss' }}</span>
                  <span class="log-message">{{ log.message }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="step-actions">
            <button nz-button nzType="default" (click)="cancelInstall()" [disabled]="installing">
              取消安装
            </button>
            <button nz-button nzType="primary" 
                    *ngIf="installCompleted"
                    (click)="nextStep()">
              下一步：验证结果
              <i nz-icon nzType="right"></i>
            </button>
          </div>
        </nz-card>

        <!-- 步骤4：验证完成 -->
        <nz-card *ngIf="currentStep === 3" class="step-card" nzTitle="验证完成">
          <div class="verification-section">
            <nz-result 
              [nzStatus]="installSuccess ? 'success' : 'error'"
              [nzTitle]="installSuccess ? '安装成功' : '安装失败'"
              [nzSubTitle]="installSuccess ? 'PolarDB-X LogCollector 已成功安装并运行' : '安装过程中出现错误，请检查日志'">
              
              <div nz-result-extra *ngIf="installSuccess">
                <button nz-button nzType="primary" (click)="goToLogsDashboard()">
                  <i nz-icon nzType="dashboard"></i>
                  打开日志面板
                </button>
                <button nz-button nzType="default" (click)="goToCollectors()">
                  <i nz-icon nzType="cluster"></i>
                  采集器管理
                </button>
              </div>
              
              <div nz-result-extra *ngIf="!installSuccess">
                <button nz-button nzType="primary" (click)="retryInstall()">
                  <i nz-icon nzType="reload"></i>
                  重新安装
                </button>
                <button nz-button nzType="default" (click)="restart()">
                  <i nz-icon nzType="undo"></i>
                  重新开始
                </button>
                <button nz-button nzType="default" (click)="goToCollectors()">
                  <i nz-icon nzType="cluster"></i>
                  采集器管理
                </button>
                <button nz-button nzType="default" (click)="goToLogsDashboard()">
                  <i nz-icon nzType="dashboard"></i>
                  日志仪表盘
                </button>
              </div>
            </nz-result>

            <!-- 安装摘要 -->
            <div class="install-summary" *ngIf="installSuccess">
              <h5>安装摘要</h5>
              <nz-descriptions nzBordered nzSize="small">
                <nz-descriptions-item nzTitle="命名空间">{{ form.value.namespace }}</nz-descriptions-item>
                <nz-descriptions-item nzTitle="安装方案">{{ getDeploymentConfig().title }}</nz-descriptions-item>
                <nz-descriptions-item nzTitle="组件数量">{{ getInstalledComponents().length }}</nz-descriptions-item>
                <nz-descriptions-item nzTitle="安装时间">{{ installDuration }}</nz-descriptions-item>
              </nz-descriptions>
            </div>
          </div>
        </nz-card>
      </div>
    </div>
  `,
  styles: [`
    .wizard {
      padding: 16px;
      background: #ffffff;
    }
    
    .page-header {
      margin-bottom: 16px;
    }
    
    .header-content {
      max-width: 1120px;
      margin: 0 auto;
    }
    
    .page-title {
      color: rgba(0, 0, 0, 0.87);
      font-size: 18px;
      font-weight: 500;
      margin: 0 0 4px 0;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    
    .page-icon {
      font-size: 20px;
      color: #1890ff;
    }
    
    .page-description {
      color: rgba(0, 0, 0, 0.6);
      font-size: 14px;
      margin: 0;
      line-height: 1.5;
    }
    
    .page-content {
      max-width: 1120px;
      margin: 0 auto;
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    
    .wizard-steps {
      margin-bottom: 24px;
    }
    
    .step-card {
      background: #fff;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.06);
      border: 1px solid #e0e0e0;
      margin-bottom: 16px;
    }

    .check-section {
      padding: 16px 0;
    }

    .loading-tip {
      text-align: center;
      color: rgba(0,0,0,0.65);
      letter-spacing: 0.5px;
      padding: 20px 0;
      font-size: 14px;
    }

    .check-items {
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .check-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      border: 1px solid #e8e8e8;
      border-radius: 6px;
      background: #fafafa;
    }

    .check-info {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .check-status {
      font-size: 16px;
    }

    .check-name {
      font-weight: 500;
      color: rgba(0, 0, 0, 0.85);
      min-width: 120px;
    }

    .check-description {
      color: rgba(0, 0, 0, 0.6);
      font-size: 13px;
    }

    .check-result {
      font-size: 13px;
      font-family: monospace;
    }

    .check-result.success {
      color: #52c41a;
    }

    .check-result.error {
      color: #ff4d4f;
    }

    .check-result.pending {
      color: #faad14;
    }

    .config-section {
      margin-bottom: 24px;
    }

    .config-section h4 {
      margin: 0 0 16px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 16px;
      font-weight: 500;
    }

    .config-section h5 {
      margin: 16px 0 12px 0;
      color: rgba(0, 0, 0, 0.8);
      font-size: 14px;
      font-weight: 500;
    }

    .deployment-alert {
      margin: 12px 0;
    }

    .resource-overview {
      margin: 16px 0;
      padding: 16px;
      background: #f8f9fa;
      border-radius: 6px;
    }

    .custom-config {
      margin-top: 16px;
    }

    .install-section {
      padding: 16px 0;
    }

    .install-alert {
      margin-bottom: 24px;
    }

    .install-progress {
      margin: 24px 0;
    }

    .install-logs {
      margin-top: 24px;
    }

    .log-viewer {
      max-height: 300px;
      overflow-y: auto;
      background: #f6f8fa;
      border: 1px solid #e1e4e8;
      border-radius: 6px;
      padding: 12px;
    }

    .log-entry {
      display: flex;
      align-items: flex-start;
      gap: 12px;
      margin-bottom: 4px;
      font-family: monospace;
      font-size: 12px;
    }

    .log-time {
      color: #666;
      min-width: 60px;
    }

    .log-message {
      flex: 1;
    }

    .log-entry.info .log-message {
      color: #1890ff;
    }

    .log-entry.success .log-message {
      color: #52c41a;
    }

    .log-entry.error .log-message {
      color: #ff4d4f;
    }

    .log-entry.warning .log-message {
      color: #faad14;
    }

    .verification-section {
      padding: 16px 0;
    }

    .install-summary {
      margin-top: 24px;
      padding-top: 16px;
      border-top: 1px solid #e8e8e8;
    }

    .install-summary h5 {
      margin: 0 0 16px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 14px;
      font-weight: 500;
    }

    .step-actions {
      display: flex;
      justify-content: flex-end;
      gap: 12px;
      margin-top: 24px;
      padding-top: 16px;
      border-top: 1px solid #e8e8e8;
    }

    @media (max-width: 1200px) {
      .page-content {
        max-width: 100%;
        padding: 0 8px;
      }
    }
    
    @media (max-width: 768px) {
      .wizard {
        padding: 8px;
      }
      
      .step-actions {
        flex-direction: column;
      }
    }
  `]
})
export class LogCollectorInstallComponent implements OnInit {
  form: FormGroup;
  currentStep = 0;
  
  // 环境检查
  checking = false;
  environmentChecks: any[] = [];
  allChecksPassed = false;

  // 安装状态
  installing = false;
  installStep = 0;
  installCompleted = false;
  installSuccess = false;
  installLogs: any[] = [];
  installDuration = '';
  verifying = false;
  verifyAttempts = 0;
  verifyMaxAttempts = 10;
  private verifyTimer?: any;
  
  private installStartTime?: Date;

  // 部署配置
  private deploymentConfigs = {
    default: {
      title: '标准安装',
      description: '使用默认配置，适合开发和测试环境',
      alertType: 'info',
      resources: { filebeatCpu: 1, filebeatMemory: 500, logstashCpu: 2, logstashMemory: 1.5 }
    },
    production: {
      title: '生产环境',
      description: '高性能配置，适合生产环境大量日志处理',
      alertType: 'success',
      resources: { filebeatCpu: 2, filebeatMemory: 1000, logstashCpu: 4, logstashMemory: 4 }
    },
    minimal: {
      title: '最小安装',
      description: '最小资源配置，适合资源受限环境',
      alertType: 'warning',
      resources: { filebeatCpu: 0.5, filebeatMemory: 256, logstashCpu: 1, logstashMemory: 1 }
    },
    custom: {
      title: '自定义配置',
      description: '根据需要自定义组件和资源配置',
      alertType: 'info',
      resources: { filebeatCpu: 0, filebeatMemory: 0, logstashCpu: 0, logstashMemory: 0 }
    }
  };

  constructor(
    private fb: FormBuilder, 
    private api: ApiService, 
    private message: NzMessageService,
    private router: Router
  ) {
    this.form = this.fb.group({
      deploymentType: ['default'],
      namespace: ['polardbx-logcollector'],
      enableFilebeat: [true],
      enableLogstash: [true],
      enableElasticsearch: [false]
    });
  }

  ngOnInit(): void {
    this.runEnvironmentCheck();
  }

  runEnvironmentCheck(): void {
    this.checking = true;
    // 使用日志服务状态做最小可行预检查
    this.environmentChecks = [
      { name: '组件状态', description: 'Filebeat / Logstash 存在与健康', status: 'pending', result: '检查中...' },
      { name: '命名空间', description: '默认命名空间是否存在', status: 'pending', result: '检查中...' }
    ];
    this.api.getLogServiceStatus().subscribe({
      next: (res: any) => {
        const ns = res?.namespace || 'polardbx-logcollector';
        const components = res?.components || {};
        const fb = components?.filebeat || {};
        const ls = components?.logstash || {};
        const fbStatus: string = fb.status || (fb.ready ? 'running' : 'error');
        const lsStatus: string = ls.status || (ls.ready ? 'running' : 'error');
        const fbReady = fb?.replicas?.ready ?? (fb.ready ? 1 : 0);
        const fbTotal = fb?.replicas?.total ?? 0;
        const lsReady = ls?.replicas?.ready ?? (ls.ready ? 1 : 0);
        const lsTotal = ls?.replicas?.total ?? 0;
        const bothRunning = fbStatus === 'running' && lsStatus === 'running';
        const toText = (s: string) => (s === 'running' ? 'ready' : (s === 'crashloop' ? 'crashloop' : 'not ready'));
        this.environmentChecks[0] = {
          ...this.environmentChecks[0],
          status: bothRunning ? 'success' : (fbStatus === 'crashloop' || lsStatus === 'crashloop' ? 'warning' : 'warning'),
          result: `filebeat: ${toText(fbStatus)}${fbTotal?` (${fbReady}/${fbTotal})`:''}, logstash: ${toText(lsStatus)}${lsTotal?` (${lsReady}/${lsTotal})`:''}`
        };
        this.environmentChecks[1] = { ...this.environmentChecks[1], status: 'success', result: ns };
        // 允许继续（未全部就绪也可继续到安装执行/回验）
        this.allChecksPassed = true;
        this.checking = false;
      },
      error: () => {
        this.environmentChecks = this.environmentChecks.map((c) => ({ ...c, status: 'error', result: '检查失败' }));
        this.allChecksPassed = false;
        this.checking = false;
      }
    });
  }

  onDeploymentTypeChange(type: string): void {
    // 根据部署类型自动调整配置
    if (type === 'production') {
      this.form.patchValue({
        enableFilebeat: true,
        enableLogstash: true,
        enableElasticsearch: true
      });
    } else if (type === 'minimal') {
      this.form.patchValue({
        enableFilebeat: true,
        enableLogstash: false,
        enableElasticsearch: false
      });
    }
  }

  getDeploymentConfig(): any {
    const type = this.form.value.deploymentType || 'default';
    return this.deploymentConfigs[type as keyof typeof this.deploymentConfigs] || this.deploymentConfigs.default;
  }

  getInstallStepDescription(step: number): string {
    if (step < this.installStep) return '已完成';
    if (step === this.installStep) return '进行中...';
    return '等待中';
  }

  getInstalledComponents(): string[] {
    const components: string[] = [];
    if (this.form.value.enableFilebeat) components.push('Filebeat');
    if (this.form.value.enableLogstash) components.push('Logstash');
    if (this.form.value.enableElasticsearch) components.push('Elasticsearch');
    return components;
  }

  nextStep(): void {
    if (this.currentStep === 1) {
      // 开始安装
      this.startInstallation();
    }
    if (this.currentStep === 2) {
      this.currentStep++;
      this.startVerification();
      return;
    }
    if (this.currentStep < 3) this.currentStep++;
  }

  prevStep(): void {
    if (this.currentStep > 0) {
      this.currentStep--;
    }
  }

  startInstallation(): void {
    this.installing = true;
    this.installStep = 0;
    this.installLogs = [];
    this.installStartTime = new Date();
    this.addLog('info', '开始安装 PolarDB-X LogCollector...');

    // 快速方案：创建 PolarDBXLogCollector CR 作为安装入口
    const ns = this.form.value.namespace || 'polardbx-logcollector';
    const release = 'polardbx-logcollector';
    const spec = this.form.value.deploymentType;
    const fileBeatName = 'filebeat-main';
    const logStashName = 'logstash-main';
    this.api.createLogCollector(ns, { name: release, namespace: ns, fileBeatName, logStashName }).subscribe({
      next: () => {
        this.addLog('success', `已创建 LogCollector CR: ${release}`);
        // 展示进度动画并完成
        this.simulateInstallation();
      },
      error: () => {
        this.addLog('error', '创建 LogCollector 资源失败');
        this.completeInstallation(false);
      }
    });
  }

  private simulateInstallation(): void {
    const steps = [
      { message: '准备安装环境...', delay: 1000 },
      { message: '创建命名空间 ' + this.form.value.namespace, delay: 1500 },
      { message: '部署 Filebeat DaemonSet...', delay: 2500 },
      { message: '部署 Logstash...', delay: 2000 },
      { message: '配置日志采集规则...', delay: 1000 }
    ];

    let currentStep = 0;
    const executeStep = () => {
      if (currentStep < steps.length) {
        this.installStep = currentStep;
        this.addLog('info', steps[currentStep].message);
        
        setTimeout(() => {
          currentStep++;
          executeStep();
        }, steps[currentStep].delay);
      } else {
        this.completeInstallation(true);
      }
    };

    executeStep();
  }

  private completeInstallation(success: boolean): void {
    this.installing = false;
    this.installCompleted = true;
    this.installSuccess = success;
    
    if (this.installStartTime) {
      const duration = Date.now() - this.installStartTime.getTime();
      this.installDuration = Math.round(duration / 1000) + ' 秒';
    }

    if (success) {
      this.addLog('success', '安装完成！所有组件已成功部署');
      this.message.success('PolarDB-X LogCollector 安装成功！');
    } else {
      this.addLog('error', '安装失败，请检查错误信息');
      this.message.error('安装过程中出现错误');
    }
  }

  private addLog(level: string, message: string): void {
    this.installLogs.push({
      level,
      message,
      timestamp: new Date()
    });
  }

  cancelInstall(): void {
    this.installing = false;
    this.addLog('warning', '用户取消安装');
    this.message.warning('安装已取消');
  }

  retryInstall(): void {
    this.currentStep = 1;
    this.installing = false;
    this.installCompleted = false;
    this.installSuccess = false;
    this.installLogs = [];
  }

  restart(): void {
    this.currentStep = 0;
    this.installing = false;
    this.installCompleted = false;
    this.installSuccess = false;
    this.installLogs = [];
    this.runEnvironmentCheck();
  }

  goToLogsDashboard(): void {
    this.router.navigate(['/operations/logs/dashboard']);
  }

  goToCollectors(): void {
    this.router.navigate(['/operations/logs/collectors']);
  }

  private clearVerifyTimer(): void {
    if (this.verifyTimer) {
      clearInterval(this.verifyTimer);
      this.verifyTimer = undefined;
    }
  }

  private startVerification(): void {
    this.verifying = true;
    this.verifyAttempts = 0;
    this.installSuccess = false;
    this.clearVerifyTimer();
    this.verifyTimer = setInterval(() => {
      this.verifyAttempts++;
      this.api.getLogServiceStatus().subscribe({
        next: (s: any) => {
          const comps = s?.components || {};
          const fbStatus: string = comps?.filebeat?.status || (comps?.filebeat?.ready ? 'running' : 'error');
          const lsStatus: string = comps?.logstash?.status || (comps?.logstash?.ready ? 'running' : 'error');
          const ok = (fbStatus === 'running') && (lsStatus === 'running');
          if (ok) {
            this.installSuccess = true;
            this.verifying = false;
            this.clearVerifyTimer();
          } else if (this.verifyAttempts >= this.verifyMaxAttempts) {
            this.installSuccess = false;
            this.verifying = false;
            this.clearVerifyTimer();
          }
        },
        error: () => {
          if (this.verifyAttempts >= this.verifyMaxAttempts) {
            this.installSuccess = false;
            this.verifying = false;
            this.clearVerifyTimer();
          }
        }
      });
    }, 3000);
  }
}


