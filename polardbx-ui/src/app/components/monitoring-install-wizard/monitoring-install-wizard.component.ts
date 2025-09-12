import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup } from '@angular/forms';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzStepsModule } from 'ng-zorro-antd/steps';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { NzCollapseModule } from 'ng-zorro-antd/collapse';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-monitoring-install-wizard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NzCardModule,
    NzFormModule,
    NzInputModule,
    NzSelectModule,
    NzButtonModule,
    NzIconModule,
    NzSwitchModule,
    NzSpinModule,
    NzGridModule,
    NzStepsModule,
    NzAlertModule,
    NzCollapseModule,
    NzCheckboxModule
  ],
  template: `
    <div class="wizard">
      <div class="page-header">
        <div class="header-content">
          <h1 class="page-title">
            <i nz-icon nzType="tool" class="page-icon"></i>
            监控安装向导
          </h1>
          <p class="page-description">引导式安装配置 Prometheus、Grafana、Alertmanager 监控堆栈</p>
        </div>
      </div>

      <div class="page-content">
        <nz-steps [nzCurrent]="currentStep" class="wizard-steps">
          <nz-step nzTitle="选择模式" nzDescription="选择安装模式和基本配置"></nz-step>
          <nz-step nzTitle="配置参数" nzDescription="设置命名空间和发布名称"></nz-step>
          <nz-step nzTitle="执行安装" nzDescription="提交安装请求并验证"></nz-step>
        </nz-steps>

        <nz-card class="config-card" nzTitle="安装配置" [nzExtra]="configExtra">
          <ng-template #configExtra>
            <i nz-icon nzType="setting" class="section-icon"></i>
          </ng-template>
          
          <nz-alert 
            nzType="info" 
            nzMessage="安装说明" 
            nzDescription="支持三种安装模式：托管模式由平台自动安装所有组件；协助模式共享部分资源；自带模式仅接入现有服务。"
            nzShowIcon
            [nzCloseable]="true"
            class="info-alert">
          </nz-alert>

          <form [formGroup]="form" class="config-form">
            <nz-collapse [nzBordered]="true" [nzAccordion]="false">
              <nz-collapse-panel nzHeader="基本配置" [nzActive]="true">
                <nz-row [nzGutter]="16">
                  <nz-col [nzSpan]="12">
                    <nz-form-item>
                      <nz-form-label [nzSpan]="6" nzRequired>安装模式</nz-form-label>
                      <nz-form-control [nzSpan]="18">
                        <nz-select formControlName="mode" nzPlaceHolder="选择安装模式">
                          <nz-option nzValue="managed" nzLabel="托管模式 - 由平台自动安装"></nz-option>
                          <nz-option nzValue="assisted" nzLabel="协助模式 - 部分资源共享"></nz-option>
                          <nz-option nzValue="byo" nzLabel="自带模式 - 仅接入现有服务"></nz-option>
                        </nz-select>
                      </nz-form-control>
                    </nz-form-item>
                  </nz-col>
                  <nz-col [nzSpan]="12">
                    <nz-form-item>
                      <nz-form-label [nzSpan]="6">命名空间</nz-form-label>
                      <nz-form-control [nzSpan]="18">
                        <input nz-input formControlName="namespace" placeholder="polardbx-operator-system" />
                      </nz-form-control>
                    </nz-form-item>
                  </nz-col>
                </nz-row>
                
                <nz-row [nzGutter]="16">
                  <nz-col [nzSpan]="12">
                    <nz-form-item>
                      <nz-form-label [nzSpan]="6">Release 名称</nz-form-label>
                      <nz-form-control [nzSpan]="18">
                        <input nz-input formControlName="releaseName" placeholder="kps" />
                      </nz-form-control>
                    </nz-form-item>
                  </nz-col>
                  <nz-col [nzSpan]="12">
                    <nz-form-item>
                      <nz-form-control [nzOffset]="6" [nzSpan]="18">
                        <label nz-checkbox formControlName="dryRun">仅试运行（不实际安装）</label>
                      </nz-form-control>
                    </nz-form-item>
                  </nz-col>
                </nz-row>
              </nz-collapse-panel>

              <nz-collapse-panel nzHeader="模式说明">
                <div class="mode-explanations">
                  <div class="mode-item">
                    <h4><i nz-icon nzType="cloud"></i> 托管模式</h4>
                    <p>平台自动安装和管理 Prometheus、Grafana、Alertmanager 等组件，适合新部署环境。</p>
                  </div>
                  <div class="mode-item">
                    <h4><i nz-icon nzType="share-alt"></i> 协助模式</h4>
                    <p>与现有监控基础设施集成，共享部分资源，适合已有部分监控组件的环境。</p>
                  </div>
                  <div class="mode-item">
                    <h4><i nz-icon nzType="link"></i> 自带模式</h4>
                    <p>仅配置连接到已有的监控服务，不安装任何组件，适合成熟的监控环境。</p>
                  </div>
                </div>
              </nz-collapse-panel>
            </nz-collapse>
          </form>

          <div class="action-buttons">
            <button nz-button nzType="default" (click)="prefill()">
              <i nz-icon nzType="reload"></i>
              重置为默认
            </button>
            <button nz-button nzType="primary" (click)="submit()" [nzLoading]="submitting">
              <i nz-icon nzType="play-circle"></i>
              开始安装
            </button>
          </div>
        </nz-card>

        <nz-card class="tips-card" nzTitle="安装提示" [nzExtra]="tipsExtra">
          <ng-template #tipsExtra>
            <i nz-icon nzType="info-circle" class="section-icon"></i>
          </ng-template>
          
          <nz-alert nzType="success" nzMessage="推荐配置" nzShowIcon class="tip-alert">
            <div class="tip-content">
              <p><strong>命名空间：</strong>建议使用 <code>polardbx-operator-system</code> 以便统一管理</p>
              <p><strong>资源需求：</strong>确保集群有足够的 CPU 和内存资源用于监控组件</p>
              <p><strong>网络策略：</strong>确认防火墙和网络策略允许监控端口访问</p>
            </div>
          </nz-alert>
          
          <nz-alert nzType="warning" nzMessage="注意事项" nzShowIcon class="tip-alert">
            <div class="tip-content">
              <p><strong>安装后验证：</strong>安装完成后请前往 "监控健康检查" 页面验证组件状态</p>
              <p><strong>配置持久化：</strong>托管模式会创建 PVC，确保存储类已配置</p>
              <p><strong>权限要求：</strong>确认当前用户有足够权限在目标命名空间创建资源</p>
            </div>
          </nz-alert>
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
    
    .config-card, .tips-card {
      background: #fff;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.06);
      border: 1px solid #e0e0e0;
    }
    
    .section-icon {
      font-size: 16px;
      color: #1890ff;
    }
    
    .info-alert {
      margin-bottom: 16px;
    }
    
    .config-form {
      margin-bottom: 24px;
    }
    
    .mode-explanations {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    
    .mode-item {
      padding: 12px;
      border: 1px solid #e0e0e0;
      border-radius: 6px;
      background: #fafafa;
    }
    
    .mode-item h4 {
      margin: 0 0 8px 0;
      color: rgba(0, 0, 0, 0.85);
      font-size: 14px;
      font-weight: 500;
      display: flex;
      align-items: center;
      gap: 6px;
    }
    
    .mode-item p {
      margin: 0;
      color: rgba(0, 0, 0, 0.6);
      font-size: 13px;
      line-height: 1.4;
    }
    
    .action-buttons {
      display: flex;
      justify-content: flex-end;
      gap: 12px;
      padding-top: 16px;
      border-top: 1px solid #e0e0e0;
    }
    
    .tip-alert {
      margin-bottom: 12px;
    }
    
    .tip-alert:last-child {
      margin-bottom: 0;
    }
    
    .tip-content p {
      margin: 4px 0;
      font-size: 13px;
    }
    
    .tip-content code {
      background: #f5f5f5;
      padding: 2px 4px;
      border-radius: 3px;
      font-family: 'Monaco', 'Menlo', monospace;
    }
    
    /* 响应式设计 */
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
      
      .action-buttons {
        flex-direction: column;
      }
      
      .mode-explanations {
        gap: 12px;
      }
    }
  `]
})
export class MonitoringInstallWizardComponent implements OnInit {
  form: FormGroup;
  submitting = false;
  currentStep = 0;

  constructor(private fb: FormBuilder, private api: ApiService, private message: NzMessageService) {
    this.form = this.fb.group({
      mode: ['managed'],
      namespace: [''],
      releaseName: ['kps'],
      dryRun: [false]
    });
  }

  ngOnInit(): void {
    this.api.getSystemContext().subscribe(ctx => {
      const ns = ctx?.defaultNamespace || 'polardbx-operator-system';
      this.form.patchValue({ namespace: ns });
    });
  }

  prefill() {
    this.form.patchValue({ mode: 'managed', releaseName: 'kps' });
  }

  submit() {
    const v = this.form.value as { mode: 'managed'|'assisted'|'byo'; namespace?: string; releaseName?: string; dryRun?: boolean };
    this.submitting = true;
    this.currentStep = 2;
    
    this.api.monitoringBootstrap(v).subscribe({
      next: () => {
        this.message.success('监控安装请求已提交，请查看监控健康检查页面确认状态');
        this.currentStep = 3;
      },
      error: (error) => {
        console.error('监控安装失败:', error);
        this.message.error('监控安装请求失败，请检查配置并重试');
        this.currentStep = 1;
      },
      complete: () => { this.submitting = false; }
    });
  }
}