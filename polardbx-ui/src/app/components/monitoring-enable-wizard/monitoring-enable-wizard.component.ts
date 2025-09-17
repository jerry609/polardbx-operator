import { Component, OnInit, ChangeDetectionStrategy, ChangeDetectorRef, ViewChild, TemplateRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzMessageModule, NzMessageService } from 'ng-zorro-antd/message';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';

import { WizardShellComponent, WizardStep, WizardAction } from '../wizard-shell/wizard-shell.component';
import { YamlPreviewComponent } from '../yaml-preview/yaml-preview.component';
import { ApiService } from '../../services/api.service';

interface PreflightCheck {
  name: string;
  description: string;
  status: 'pending' | 'success' | 'warning' | 'error';
  result: string;
  command?: string;
}

@Component({
  selector: 'app-monitoring-enable-wizard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NzFormModule,
    NzInputModule,
    NzSelectModule,
    NzButtonModule,
    NzIconModule,
    NzAlertModule,
    NzSpinModule,
    NzGridModule,
    NzDescriptionsModule,
    NzResultModule,
    NzMessageModule,
    NzModalModule,
    WizardShellComponent,
    YamlPreviewComponent
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <app-wizard-shell
      title="监控一键开启向导"
      subtitle="快速启用 PolarDB-X 集群监控（企业版 PolarDBXMonitor / 标准版 ServiceMonitor）"
      titleIcon="dashboard"
      [namespace]="form.value.namespace"
      [objectName]="getObjectName()"
      objectLabel="目标"
      docLink="https://docs.polardbx.com/monitoring"
      [steps]="wizardSteps"
      [currentStepIndex]="currentStep"
      [actions]="getStepActions()"
      [loading]="stepLoading">

      <!-- 步骤1：选择目标 -->
      <ng-template #step1Template>
        <div class="step-content">
          <nz-alert 
            nzType="info"
            nzMessage="选择监控目标"
            nzDescription="选择要启用监控的集群类型和具体目标。企业版使用 PolarDBXMonitor，标准版使用 ServiceMonitor。"
            nzShowIcon
            class="step-alert">
          </nz-alert>

          <form [formGroup]="form" class="config-form">
            <nz-row [nzGutter]="16">
              <nz-col [nzSpan]="12">
                <nz-form-item>
                  <nz-form-label [nzSpan]="6" nzRequired>监控类型</nz-form-label>
                  <nz-form-control [nzSpan]="18">
                    <nz-select 
                      formControlName="monitoringType" 
                      nzPlaceholder="选择监控类型"
                      (ngModelChange)="onMonitoringTypeChange($event)">
                      <nz-option nzValue="enterprise" nzLabel="企业版 (PolarDBXMonitor)"></nz-option>
                      <nz-option nzValue="standard" nzLabel="标准版 (ServiceMonitor)"></nz-option>
                    </nz-select>
                  </nz-form-control>
                </nz-form-item>
              </nz-col>
              <nz-col [nzSpan]="12">
                <nz-form-item>
                  <nz-form-label [nzSpan]="6" nzRequired>命名空间</nz-form-label>
                  <nz-form-control [nzSpan]="18">
                    <nz-select 
                      formControlName="namespace" 
                      nzPlaceholder="选择命名空间"
                      nzShowSearch
                      nzAllowClear>
                      <nz-option 
                        *ngFor="let ns of namespaces" 
                        [nzValue]="ns" 
                        [nzLabel]="ns">
                      </nz-option>
                    </nz-select>
                  </nz-form-control>
                </nz-form-item>
              </nz-col>
            </nz-row>

            <nz-row [nzGutter]="16" *ngIf="form.value.monitoringType">
              <nz-col [nzSpan]="12">
                <nz-form-item>
                  <nz-form-label [nzSpan]="6" nzRequired>{{ getTargetLabel() }}</nz-form-label>
                  <nz-form-control [nzSpan]="18">
                    <nz-select 
                      formControlName="targetName" 
                      nzPlaceholder="选择目标"
                      nzShowSearch
                      nzAllowClear
                      [nzLoading]="loadingTargets">
                      <nz-option 
                        *ngFor="let target of targets" 
                        [nzValue]="target" 
                        [nzLabel]="target">
                      </nz-option>
                    </nz-select>
                  </nz-form-control>
                </nz-form-item>
              </nz-col>
              <nz-col [nzSpan]="12">
                <nz-form-item>
                  <nz-form-label [nzSpan]="6">Monitor 名称</nz-form-label>
                  <nz-form-control [nzSpan]="18">
                    <input 
                      nz-input 
                      formControlName="monitorName"
                      placeholder="自动生成（可自定义）">
                  </nz-form-control>
                </nz-form-item>
              </nz-col>
            </nz-row>

            <div class="type-description" *ngIf="form.value.monitoringType">
              <h4>{{ getTypeDescription().title }}</h4>
              <p>{{ getTypeDescription().description }}</p>
            </div>
          </form>
        </div>
      </ng-template>

      <!-- 步骤2：前置检测 -->
      <ng-template #step2Template>
        <div class="step-content">
          <nz-alert 
            nzType="info"
            nzMessage="环境检测"
            nzDescription="检查 CRD、权限和监控组件状态，确保监控配置可以正常应用。"
            nzShowIcon
            class="step-alert">
          </nz-alert>

          <div class="preflight-section">
            <nz-spin [nzSpinning]="runningPreflight" nzTip="正在检查环境...">
              <div class="check-items">
                <div class="check-item" *ngFor="let check of preflightChecks">
                  <div class="check-info">
                    <span class="check-status">
                      <i nz-icon 
                        [nzType]="getCheckIcon(check.status)" 
                        [style.color]="getCheckColor(check.status)">
                      </i>
                    </span>
                    <span class="check-name">{{ check.name }}</span>
                    <span class="check-description">{{ check.description }}</span>
                  </div>
                  <div class="check-result" [ngClass]="check.status">
                    {{ check.result }}
                  </div>
                  <div class="check-command" *ngIf="check.command">
                    <button 
                      nz-button 
                      nzType="dashed" 
                      nzSize="small"
                      (click)="copyCommand(check.command!)">
                      <i nz-icon nzType="copy"></i>
                      复制命令
                    </button>
                  </div>
                </div>
              </div>
            </nz-spin>
          </div>
        </div>
      </ng-template>

      <!-- 步骤3：采集参数 -->
      <ng-template #step3Template>
        <div class="step-content">
          <nz-alert 
            nzType="info"
            nzMessage="监控参数配置"
            nzDescription="配置监控采集间隔、超时时间等参数。使用默认值即可满足大多数场景。"
            nzShowIcon
            class="step-alert">
          </nz-alert>

          <form [formGroup]="form" class="config-form">
            <div class="config-section">
              <h4>{{ form.value.monitoringType === 'enterprise' ? 'PolarDBXMonitor 参数' : 'ServiceMonitor 参数' }}</h4>
              
              <nz-row [nzGutter]="16">
                <nz-col [nzSpan]="12">
                  <nz-form-item>
                    <nz-form-label [nzSpan]="6">采集间隔</nz-form-label>
                    <nz-form-control [nzSpan]="18">
                      <input 
                        nz-input 
                        formControlName="scrapeInterval"
                        placeholder="例如: 30s">
                    </nz-form-control>
                  </nz-form-item>
                </nz-col>
                <nz-col [nzSpan]="12">
                  <nz-form-item>
                    <nz-form-label [nzSpan]="6">超时时间</nz-form-label>
                    <nz-form-control [nzSpan]="18">
                      <input 
                        nz-input 
                        formControlName="scrapeTimeout"
                        placeholder="例如: 10s">
                    </nz-form-control>
                  </nz-form-item>
                </nz-col>
              </nz-row>

              <div *ngIf="form.value.monitoringType === 'standard'">
                <nz-row [nzGutter]="16">
                  <nz-col [nzSpan]="24">
                    <nz-form-item>
                      <nz-form-label [nzSpan]="3">标签选择器</nz-form-label>
                      <nz-form-control [nzSpan]="21">
                        <input 
                          nz-input 
                          formControlName="selectorLabels"
                          placeholder="自动填充 XStore 标签（可自定义）"
                          readonly>
                      </nz-form-control>
                    </nz-form-item>
                  </nz-col>
                </nz-row>
              </div>
            </div>

            <div class="config-preview">
              <h4>配置预览</h4>
              <nz-descriptions nzBordered nzSize="small">
                <nz-descriptions-item nzTitle="监控类型">
                  {{ form.value.monitoringType === 'enterprise' ? 'PolarDBXMonitor' : 'ServiceMonitor' }}
                </nz-descriptions-item>
                <nz-descriptions-item nzTitle="目标">{{ getObjectName() }}</nz-descriptions-item>
                <nz-descriptions-item nzTitle="命名空间">{{ form.value.namespace }}</nz-descriptions-item>
                <nz-descriptions-item nzTitle="采集间隔">{{ form.value.scrapeInterval }}</nz-descriptions-item>
                <nz-descriptions-item nzTitle="超时时间">{{ form.value.scrapeTimeout }}</nz-descriptions-item>
              </nz-descriptions>
            </div>
          </form>
        </div>
      </ng-template>

      <!-- 步骤4：YAML 预览 -->
      <ng-template #step4Template>
        <div class="step-content">
          <app-yaml-preview
            [yamlContent]="generatedYaml"
            [filename]="getYamlFilename()"
            [loading]="generatingYaml"
            [readonly]="true">
          </app-yaml-preview>
        </div>
      </ng-template>

      <!-- 步骤5：应用与验证 -->
      <ng-template #step5Template>
        <div class="step-content">
          <nz-result 
            [nzStatus]="applyResult?.success ? 'success' : (applyResult ? 'error' : 'info')"
            [nzTitle]="getResultTitle()"
            [nzSubTitle]="getResultSubtitle()">
            
            <div nz-result-content *ngIf="!applyResult">
              <div class="apply-options">
                <h4>应用方式</h4>
                <nz-alert 
                  nzType="info"
                  nzMessage="选择应用方式"
                  nzDescription="您可以复制命令手动执行，或者让系统自动应用配置。"
                  nzShowIcon
                  class="apply-alert">
                </nz-alert>

                <div class="kubectl-command">
                  <h5>kubectl 命令</h5>
                  <div class="command-block">
                    <pre>{{ getKubectlCommand() }}</pre>
                    <button 
                      nz-button 
                      nzType="dashed" 
                      nzSize="small"
                      (click)="copyKubectlCommand()">
                      <i nz-icon nzType="copy"></i>
                      复制命令
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div nz-result-extra *ngIf="applyResult?.success">
              <button nz-button nzType="primary" (click)="goToMonitoring()">
                <i nz-icon nzType="dashboard"></i>
                查看监控
              </button>
              <button nz-button nzType="default" (click)="goToPrometheus()">
                <i nz-icon nzType="line-chart"></i>
                Prometheus
              </button>
              <button nz-button nzType="default" (click)="goToGrafana()">
                <i nz-icon nzType="bar-chart"></i>
                Grafana
              </button>
            </div>

            <div nz-result-extra *ngIf="applyResult && !applyResult.success">
              <button nz-button nzType="primary" (click)="retryApply()">
                <i nz-icon nzType="reload"></i>
                重试
              </button>
              <button nz-button nzType="default" (click)="goToPrevStep()">
                <i nz-icon nzType="left"></i>
                上一步
              </button>
            </div>
          </nz-result>

          <div class="verification-tips" *ngIf="applyResult?.success">
            <h4>验证指引</h4>
            <nz-alert 
              nzType="success"
              nzMessage="常见验证步骤"
              nzDescription="监控配置已应用，您可以通过以下方式验证是否生效："
              nzShowIcon>
            </nz-alert>
            <ul class="tips-list">
              <li>检查 ServiceMonitor/PolarDBXMonitor 资源状态</li>
              <li>访问 Prometheus Targets 页面确认目标发现</li>
              <li>在 Grafana 中查看相关面板数据</li>
              <li>检查 Alertmanager 告警规则加载情况</li>
            </ul>
          </div>
        </div>
      </ng-template>
    </app-wizard-shell>
  `,
  styles: [`
    .step-content {
      padding: 0;
    }

    .step-alert {
      margin-bottom: 24px;
    }

    .config-form {
      margin-bottom: 24px;
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

    .type-description {
      margin-top: 20px;
      padding: 16px;
      background: #f8f9fa;
      border-radius: 6px;
      border: 1px solid #e8e8e8;
    }

    .type-description h4 {
      margin: 0 0 8px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 14px;
      font-weight: 500;
    }

    .type-description p {
      margin: 0;
      color: rgba(0, 0, 0, 0.65);
      font-size: 13px;
      line-height: 1.5;
    }

    .preflight-section {
      padding: 16px 0;
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
      flex: 1;
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
      flex: 1;
    }

    .check-result {
      font-size: 13px;
      font-family: monospace;
      margin-right: 12px;
    }

    .check-result.success {
      color: #52c41a;
    }

    .check-result.error {
      color: #ff4d4f;
    }

    .check-result.warning {
      color: #faad14;
    }

    .check-result.pending {
      color: #1890ff;
    }

    .check-command {
      flex-shrink: 0;
    }

    .config-preview {
      margin-top: 24px;
      padding-top: 16px;
      border-top: 1px solid #e8e8e8;
    }

    .config-preview h4 {
      margin: 0 0 16px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 14px;
      font-weight: 500;
    }

    .apply-options {
      text-align: left;
      max-width: 600px;
      margin: 0 auto;
    }

    .apply-options h4 {
      margin: 0 0 16px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 16px;
      font-weight: 500;
    }

    .apply-alert {
      margin-bottom: 24px;
    }

    .kubectl-command h5 {
      margin: 16px 0 12px 0;
      color: rgba(0, 0, 0, 0.8);
      font-size: 14px;
      font-weight: 500;
    }

    .command-block {
      position: relative;
      background: #f6f8fa;
      border: 1px solid #e1e4e8;
      border-radius: 6px;
      padding: 12px;
    }

    .command-block pre {
      margin: 0;
      font-family: 'SFMono-Regular', 'Monaco', 'Menlo', 'Courier New', monospace;
      font-size: 13px;
      line-height: 1.4;
      color: #24292e;
      word-wrap: break-word;
      white-space: pre-wrap;
    }

    .command-block button {
      position: absolute;
      top: 8px;
      right: 8px;
    }

    .verification-tips {
      margin-top: 32px;
      padding-top: 24px;
      border-top: 1px solid #e8e8e8;
      text-align: left;
      max-width: 600px;
      margin-left: auto;
      margin-right: auto;
    }

    .verification-tips h4 {
      margin: 0 0 16px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 16px;
      font-weight: 500;
    }

    .tips-list {
      margin: 16px 0 0 0;
      padding-left: 20px;
    }

    .tips-list li {
      margin-bottom: 8px;
      color: rgba(0, 0, 0, 0.65);
      font-size: 14px;
      line-height: 1.5;
    }
  `]
})
export class MonitoringEnableWizardComponent implements OnInit {
  @ViewChild('step1Template', { read: TemplateRef }) step1Template!: TemplateRef<any>;
  @ViewChild('step2Template', { read: TemplateRef }) step2Template!: TemplateRef<any>;
  @ViewChild('step3Template', { read: TemplateRef }) step3Template!: TemplateRef<any>;
  @ViewChild('step4Template', { read: TemplateRef }) step4Template!: TemplateRef<any>;
  @ViewChild('step5Template', { read: TemplateRef }) step5Template!: TemplateRef<any>;

  form: FormGroup;
  currentStep = 0;
  stepLoading = false;

  // 数据源
  namespaces: string[] = [];
  targets: string[] = [];
  loadingTargets = false;

  // 前置检测
  runningPreflight = false;
  preflightChecks: PreflightCheck[] = [];

  // YAML 生成
  generatingYaml = false;
  generatedYaml = '';

  // 应用结果
  applyResult: { success: boolean; message: string } | null = null;

  wizardSteps: WizardStep[] = [];

  constructor(
    private fb: FormBuilder,
    private api: ApiService,
    private message: NzMessageService,
    private modal: NzModalService,
    private router: Router,
    private cdr: ChangeDetectorRef
  ) {
    this.form = this.fb.group({
      monitoringType: ['enterprise', Validators.required],
      namespace: ['polardbx-monitor', Validators.required],
      targetName: ['', Validators.required],
      monitorName: [''],
      scrapeInterval: ['30s'],
      scrapeTimeout: ['10s'],
      selectorLabels: ['']
    });
  }

  ngOnInit(): void {
    this.initializeWizardSteps();
    this.loadNamespaces();
    this.setupFormWatchers();
  }

  private initializeWizardSteps(): void {
    // 将在 ngAfterViewInit 中设置模板
    this.wizardSteps = [
      { id: 'target', title: '选择目标', description: '监控类型与目标' },
      { id: 'preflight', title: '前置检测', description: '环境检查' },
      { id: 'config', title: '采集参数', description: '监控配置' },
      { id: 'yaml', title: 'YAML 预览', description: '配置预览' },
      { id: 'apply', title: '应用验证', description: '应用与验证' }
    ];
  }

  ngAfterViewInit(): void {
    // 设置步骤模板
    this.wizardSteps[0].template = this.step1Template;
    this.wizardSteps[1].template = this.step2Template;
    this.wizardSteps[2].template = this.step3Template;
    this.wizardSteps[3].template = this.step4Template;
    this.wizardSteps[4].template = this.step5Template;
    this.cdr.detectChanges();
  }

  private setupFormWatchers(): void {
    // 监听监控类型变化
    this.form.get('monitoringType')?.valueChanges.subscribe(() => {
      this.loadTargets();
    });

    // 监听命名空间变化
    this.form.get('namespace')?.valueChanges.subscribe(() => {
      this.loadTargets();
    });

    // 监听目标名称变化，自动生成监控名称
    this.form.get('targetName')?.valueChanges.subscribe((targetName) => {
      if (targetName && !this.form.get('monitorName')?.value) {
        const monitorName = `${targetName}-monitor`;
        this.form.patchValue({ monitorName });
      }
      this.updateSelectorLabels();
    });
  }

  private loadNamespaces(): void {
    this.api.getNamespaces().subscribe({
      next: (namespaces) => {
        this.namespaces = namespaces || [];
        this.cdr.markForCheck();
      },
      error: (error) => {
        console.error('加载命名空间失败:', error);
        this.message.error('加载命名空间失败');
      }
    });
  }

  private loadTargets(): void {
    const monitoringType = this.form.value.monitoringType;
    const namespace = this.form.value.namespace;
    
    if (!monitoringType || !namespace) {
      this.targets = [];
      return;
    }

    this.loadingTargets = true;
    
    if (monitoringType === 'enterprise') {
      // 加载 PolarDB-X 集群
      this.api.getClusters(namespace).subscribe({
        next: (clusters) => {
          this.targets = clusters.map((c: any) => c.metadata?.name).filter(Boolean) || [];
          this.loadingTargets = false;
          this.cdr.markForCheck();
        },
        error: (error) => {
          console.error('加载集群失败:', error);
          this.targets = [];
          this.loadingTargets = false;
          this.cdr.markForCheck();
        }
      });
    } else {
      // 加载 XStore
      this.api.getXStores(namespace).subscribe({
        next: (xstores) => {
          this.targets = xstores.map((x: any) => x.metadata?.name).filter(Boolean) || [];
          this.loadingTargets = false;
          this.cdr.markForCheck();
        },
        error: (error) => {
          console.error('加载 XStore 失败:', error);
          this.targets = [];
          this.loadingTargets = false;
          this.cdr.markForCheck();
        }
      });
    }
  }

  private updateSelectorLabels(): void {
    if (this.form.value.monitoringType === 'standard' && this.form.value.targetName) {
      const labels = `xstore/name=${this.form.value.targetName}`;
      this.form.patchValue({ selectorLabels: labels });
    }
  }

  onMonitoringTypeChange(type: string): void {
    // 重置相关字段
    this.form.patchValue({
      targetName: '',
      monitorName: '',
      selectorLabels: ''
    });
    this.loadTargets();
  }

  getTargetLabel(): string {
    return this.form.value.monitoringType === 'enterprise' ? '集群名称' : 'XStore 名称';
  }

  getObjectName(): string {
    const targetName = this.form.value.targetName;
    const monitoringType = this.form.value.monitoringType;
    if (!targetName) return '';
    return `${targetName} (${monitoringType === 'enterprise' ? 'PolarDBX' : 'XStore'})`;
  }

  getTypeDescription(): { title: string; description: string } {
    const type = this.form.value.monitoringType;
    if (type === 'enterprise') {
      return {
        title: '企业版监控',
        description: '使用 PolarDBXMonitor CRD 为 PolarDB-X 集群启用监控。适用于完整的集群级别监控。'
      };
    } else {
      return {
        title: '标准版监控',
        description: '使用 ServiceMonitor CRD 为 XStore 启用监控。适用于特定 XStore 实例的监控。'
      };
    }
  }

  getStepActions(): WizardAction[] {
    const actions: WizardAction[] = [];
    
    // 上一步按钮
    if (this.currentStep > 0) {
      actions.push({
        text: '上一步',
        icon: 'left',
        handler: () => this.prevStep()
      });
    }

    // 根据当前步骤添加特定按钮
    switch (this.currentStep) {
      case 0: // 选择目标
        actions.push({
          text: '下一步：环境检测',
          type: 'primary',
          icon: 'right',
          disabled: !this.form.valid,
          handler: () => this.nextStep()
        });
        break;
      
      case 1: // 前置检测
        actions.push({
          text: '重新检测',
          icon: 'sync',
          loading: this.runningPreflight,
          handler: () => this.runPreflightChecks()
        });
        actions.push({
          text: '下一步：参数配置',
          type: 'primary',
          icon: 'right',
          disabled: this.hasPreflightErrors(),
          handler: () => this.nextStep()
        });
        break;
      
      case 2: // 采集参数
        actions.push({
          text: '下一步：YAML 预览',
          type: 'primary',
          icon: 'right',
          handler: () => this.nextStep()
        });
        break;
      
      case 3: // YAML 预览
        actions.push({
          text: '重新生成',
          icon: 'sync',
          loading: this.generatingYaml,
          handler: () => this.generateYaml()
        });
        actions.push({
          text: '下一步：应用配置',
          type: 'primary',
          icon: 'right',
          disabled: !this.generatedYaml,
          handler: () => this.nextStep()
        });
        break;
      
      case 4: // 应用验证
        if (!this.applyResult) {
          actions.push({
            text: '自动应用',
            type: 'primary',
            icon: 'check',
            handler: () => this.applyConfiguration()
          });
        }
        actions.push({
          text: '完成',
          type: 'default',
          icon: 'check-circle',
          handler: () => this.finish()
        });
        break;
    }

    return actions;
  }

  nextStep(): void {
    if (this.currentStep < this.wizardSteps.length - 1) {
      this.currentStep++;
      
      // 进入特定步骤时的自动操作
      switch (this.currentStep) {
        case 1: // 进入前置检测
          this.runPreflightChecks();
          break;
        case 3: // 进入 YAML 预览
          this.generateYaml();
          break;
      }
      
      this.cdr.markForCheck();
    }
  }

  prevStep(): void {
    if (this.currentStep > 0) {
      this.currentStep--;
      this.cdr.markForCheck();
    }
  }

  goToPrevStep(): void {
    this.prevStep();
  }

  runPreflightChecks(): void {
    this.runningPreflight = true;
    
    // 初始化检查项
    this.preflightChecks = [
      {
        name: 'CRD 检查',
        description: this.form.value.monitoringType === 'enterprise' ? 
          '检查 PolarDBXMonitor CRD' : '检查 ServiceMonitor CRD',
        status: 'pending',
        result: '检查中...'
      },
      {
        name: 'RBAC 权限',
        description: '检查 K8s API 访问权限',
        status: 'pending',
        result: '检查中...'
      },
      {
        name: 'Prometheus 状态',
        description: '检查 Prometheus 运行状态',
        status: 'pending',
        result: '检查中...'
      }
    ];

    // 模拟检查过程
    setTimeout(() => {
      this.preflightChecks[0] = {
        ...this.preflightChecks[0],
        status: 'success',
        result: 'CRD 已安装'
      };
      
      this.preflightChecks[1] = {
        ...this.preflightChecks[1],
        status: 'warning',
        result: '权限检查需要手动确认',
        command: 'kubectl auth can-i create servicemonitors --as=system:serviceaccount:default:prometheus'
      };
      
      this.preflightChecks[2] = {
        ...this.preflightChecks[2],
        status: 'success',
        result: 'Prometheus 运行正常'
      };
      
      this.runningPreflight = false;
      this.cdr.markForCheck();
    }, 2000);
  }

  hasPreflightErrors(): boolean {
    return this.preflightChecks.some(check => check.status === 'error');
  }

  getCheckIcon(status: string): string {
    switch (status) {
      case 'success': return 'check-circle';
      case 'error': return 'close-circle';
      case 'warning': return 'exclamation-circle';
      default: return 'loading';
    }
  }

  getCheckColor(status: string): string {
    switch (status) {
      case 'success': return '#52c41a';
      case 'error': return '#ff4d4f';
      case 'warning': return '#faad14';
      default: return '#1890ff';
    }
  }

  copyCommand(command: string): void {
    navigator.clipboard.writeText(command).then(() => {
      this.message.success('命令已复制到剪贴板');
    }).catch(() => {
      this.message.error('复制失败');
    });
  }

  generateYaml(): void {
    this.generatingYaml = true;
    
    // 根据表单数据生成 YAML
    const config = this.form.value;
    let yaml = '';
    
    if (config.monitoringType === 'enterprise') {
      yaml = `apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXMonitor
metadata:
  name: ${config.monitorName}
  namespace: ${config.namespace}
spec:
  clusterName: ${config.targetName}
  monitorInterval: ${config.scrapeInterval}
  scrapeTimeout: ${config.scrapeTimeout}
  enabled: true`;
    } else {
      yaml = `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ${config.monitorName}
  namespace: ${config.namespace}
spec:
  selector:
    matchLabels:
      ${config.selectorLabels}
  endpoints:
  - port: metrics
    interval: ${config.scrapeInterval}
    scrapeTimeout: ${config.scrapeTimeout}
    path: /metrics`;
    }
    
    // 模拟生成过程
    setTimeout(() => {
      this.generatedYaml = yaml;
      this.generatingYaml = false;
      this.cdr.markForCheck();
    }, 1000);
  }

  getYamlFilename(): string {
    const config = this.form.value;
    const type = config.monitoringType === 'enterprise' ? 'polardbxmonitor' : 'servicemonitor';
    return `${config.monitorName}-${type}.yaml`;
  }

  getKubectlCommand(): string {
    const filename = this.getYamlFilename();
    return `# 保存 YAML 内容到文件\nkubectl apply -f ${filename}\n\n# 或者直接应用\ncat <<EOF | kubectl apply -f -\n${this.generatedYaml}\nEOF`;
  }

  copyKubectlCommand(): void {
    const command = this.getKubectlCommand();
    navigator.clipboard.writeText(command).then(() => {
      this.message.success('命令已复制到剪贴板');
    }).catch(() => {
      this.message.error('复制失败');
    });
  }

  applyConfiguration(): void {
    this.modal.confirm({
      nzTitle: '确认应用配置？',
      nzContent: `将自动应用 ${this.form.value.monitoringType === 'enterprise' ? 'PolarDBXMonitor' : 'ServiceMonitor'} 配置`,
      nzOkText: '确认应用',
      nzCancelText: '取消',
      nzOkType: 'primary',
      nzOnOk: () => this.doApplyConfiguration()
    });
  }

  private doApplyConfiguration(): void {
    this.stepLoading = true;
    
    // 模拟应用过程
    setTimeout(() => {
      this.applyResult = {
        success: true,
        message: '监控配置已成功应用'
      };
      this.stepLoading = false;
      this.message.success('监控配置应用成功！');
      this.cdr.markForCheck();
    }, 2000);
  }

  retryApply(): void {
    this.applyResult = null;
    this.applyConfiguration();
  }

  getResultTitle(): string {
    if (!this.applyResult) {
      return '准备应用配置';
    }
    return this.applyResult.success ? '配置应用成功' : '配置应用失败';
  }

  getResultSubtitle(): string {
    if (!this.applyResult) {
      return '选择应用方式以启用监控配置';
    }
    return this.applyResult.message;
  }

  goToMonitoring(): void {
    this.router.navigate(['/operations/monitoring/overview']);
  }

  goToPrometheus(): void {
    window.open('http://prometheus.example.com', '_blank');
  }

  goToGrafana(): void {
    window.open('http://grafana.example.com', '_blank');
  }

  finish(): void {
    this.router.navigate(['/operations/monitoring']);
  }
}
