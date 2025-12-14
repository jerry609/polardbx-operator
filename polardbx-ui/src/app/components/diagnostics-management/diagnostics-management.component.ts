import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzProgressModule } from 'ng-zorro-antd/progress';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzPageHeaderModule } from 'ng-zorro-antd/page-header';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzPopconfirmModule } from 'ng-zorro-antd/popconfirm';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { NzModalModule } from 'ng-zorro-antd/modal';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { NzEmptyModule } from 'ng-zorro-antd/empty';
import { ApiService } from '../../services/api.service';
import { interval, Subscription } from 'rxjs';
import { ActivatedRoute } from '@angular/router';

interface DiagnosisReport {
  id: string;
  namespace: string;
  cluster?: string;
  startedAt?: string;
  completedAt?: string;
  status?: string; // running/succeeded/failed/pending
  progress?: number;
  message?: string;
  outputPath?: string;
}

@Component({
  selector: 'app-diagnostics-management',
  standalone: true,
  imports: [
    CommonModule, FormsModule, NzCardModule, NzButtonModule, NzIconModule, 
    NzTableModule, NzProgressModule, NzTagModule, NzFormModule, NzInputModule, 
    NzSpinModule, NzPageHeaderModule, NzGridModule, NzPopconfirmModule,
    NzToolTipModule, NzModalModule, NzDescriptionsModule, NzEmptyModule
  ],
  template: `
    <div class="diagnostics-page">
      <nz-page-header [nzGhost]="false" nzTitle="诊断与排障" nzSubtitle="触发诊断、查看历史报告并下载">
        <nz-page-header-extra>
          <i nz-icon nzType="bug" class="page-icon"></i>
        </nz-page-header-extra>
      </nz-page-header>

      <div class="content-grid">
        <nz-card class="action-card" nzTitle="启动诊断">
          <ng-template #title>
            <i nz-icon nzType="play-circle"></i>
            <span>启动诊断</span>
          </ng-template>
          
          <div class="info-box">
            <i nz-icon nzType="info-circle" nzTheme="outline"></i>
            <span>诊断功能会创建一个临时 Pod 来收集集群信息，包括：Pod 日志、配置信息、事件、资源状态等。</span>
          </div>
          
          <form nz-form nzLayout="vertical" class="diagnostic-form">
            <nz-form-item>
              <nz-form-label nzRequired>命名空间</nz-form-label>
              <nz-form-control>
                <input nz-input [(ngModel)]="namespace" name="namespace" placeholder="default" />
              </nz-form-control>
            </nz-form-item>
            
            <nz-form-item>
              <nz-form-label nzRequired>集群名</nz-form-label>
              <nz-form-control>
                <input nz-input [(ngModel)]="cluster" name="cluster" placeholder="输入集群名称" />
              </nz-form-control>
            </nz-form-item>
            
            <nz-form-item>
              <nz-form-control>
                <div class="form-actions">
                  <button nz-button nzType="primary" (click)="start()" [nzLoading]="starting" [disabled]="!cluster || !namespace">
                    <i nz-icon nzType="play-circle"></i>
                    {{ starting ? '启动中...' : '启动诊断' }}
                  </button>
                  <button nz-button nzType="default" (click)="refresh()" [nzLoading]="loading">
                    <i nz-icon nzType="reload"></i>
                    刷新列表
                  </button>
                </div>
              </nz-form-control>
            </nz-form-item>
          </form>
          
          <nz-progress *ngIf="starting" [nzPercent]="0" nzStatus="active" nzSize="small"></nz-progress>
        </nz-card>

        <nz-card class="list-card">
          <ng-template #title>
            <i nz-icon nzType="history"></i>
            <span>历史报告（{{ reports.length }}）</span>
          </ng-template>
          
          <nz-empty *ngIf="reports.length === 0 && !loading" nzNotFoundContent="暂无诊断报告"></nz-empty>
          
          <nz-table 
            *ngIf="reports.length > 0" 
            [nzData]="reports" 
            [nzShowPagination]="reports.length > 10"
            [nzPageSize]="10"
            [nzLoading]="loading"
            class="reports-table">
            <thead>
              <tr>
                <th>报告ID</th>
                <th>命名空间</th>
                <th>集群</th>
                <th>开始时间</th>
                <th>状态</th>
                <th>进度</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let r of reports">
                <td>
                  <a nz-tooltip [nzTooltipTitle]="r.id" (click)="showDetail(r)">
                    {{ r.id | slice:0:20 }}{{ r.id.length > 20 ? '...' : '' }}
                  </a>
                </td>
                <td>{{ r.namespace }}</td>
                <td>{{ r.cluster || '-' }}</td>
                <td>{{ r.startedAt ? (r.startedAt | date:'yyyy-MM-dd HH:mm:ss') : '-' }}</td>
                <td>
                  <nz-tag [nzColor]="statusColor(r.status)">
                    <i nz-icon [nzType]="statusIcon(r.status)"></i>
                    {{ statusText(r.status) }}
                  </nz-tag>
                </td>
                <td>
                  <nz-progress 
                    *ngIf="r.status === 'running' || r.status === 'pending'" 
                    [nzPercent]="r.progress || 0" 
                    nzSize="small"
                    [nzStatus]="r.status === 'running' ? 'active' : 'normal'">
                  </nz-progress>
                  <span *ngIf="r.status === 'succeeded'">100%</span>
                  <span *ngIf="r.status === 'failed'">-</span>
                </td>
                <td>
                  <div class="action-buttons">
                    <button 
                      nz-button nzType="link" nzSize="small" 
                      (click)="download(r)"
                      [disabled]="r.status !== 'succeeded'"
                      nz-tooltip nzTooltipTitle="下载报告">
                      <i nz-icon nzType="download"></i>
                    </button>
                    <button 
                      nz-button nzType="link" nzSize="small" 
                      (click)="showDetail(r)"
                      nz-tooltip nzTooltipTitle="查看详情">
                      <i nz-icon nzType="eye"></i>
                    </button>
                    <button 
                      nz-button nzType="link" nzSize="small" nzDanger
                      nz-popconfirm
                      nzPopconfirmTitle="确定要删除此诊断报告吗？"
                      (nzOnConfirm)="delete(r)"
                      [disabled]="r.status === 'running'"
                      nz-tooltip nzTooltipTitle="删除">
                      <i nz-icon nzType="delete"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </nz-table>
        </nz-card>
      </div>
    </div>

    <!-- 详情弹窗 -->
    <nz-modal
      [(nzVisible)]="detailVisible"
      [nzTitle]="'诊断报告详情'"
      [nzWidth]="700"
      [nzFooter]="null"
      (nzOnCancel)="detailVisible = false">
      <ng-container *nzModalContent>
        <nz-descriptions *ngIf="selectedReport" nzBordered [nzColumn]="2">
          <nz-descriptions-item nzTitle="报告ID" [nzSpan]="2">{{ selectedReport.id }}</nz-descriptions-item>
          <nz-descriptions-item nzTitle="命名空间">{{ selectedReport.namespace }}</nz-descriptions-item>
          <nz-descriptions-item nzTitle="集群">{{ selectedReport.cluster || '-' }}</nz-descriptions-item>
          <nz-descriptions-item nzTitle="状态">
            <nz-tag [nzColor]="statusColor(selectedReport.status)">{{ statusText(selectedReport.status) }}</nz-tag>
          </nz-descriptions-item>
          <nz-descriptions-item nzTitle="进度">{{ selectedReport.progress || 0 }}%</nz-descriptions-item>
          <nz-descriptions-item nzTitle="开始时间">{{ selectedReport.startedAt | date:'yyyy-MM-dd HH:mm:ss' }}</nz-descriptions-item>
          <nz-descriptions-item nzTitle="完成时间">{{ selectedReport.completedAt ? (selectedReport.completedAt | date:'yyyy-MM-dd HH:mm:ss') : '-' }}</nz-descriptions-item>
          <nz-descriptions-item nzTitle="消息" [nzSpan]="2">{{ selectedReport.message || '-' }}</nz-descriptions-item>
          <nz-descriptions-item nzTitle="输出路径" [nzSpan]="2" *ngIf="selectedReport.outputPath">
            <code>{{ selectedReport.outputPath }}</code>
          </nz-descriptions-item>
        </nz-descriptions>
        
        <div class="modal-actions" *ngIf="selectedReport?.status === 'succeeded'">
          <button nz-button nzType="primary" (click)="download(selectedReport!)">
            <i nz-icon nzType="download"></i> 下载报告
          </button>
        </div>
      </ng-container>
    </nz-modal>
  `,
  styles: [`
    .diagnostics-page { 
      padding: 16px 24px; 
      background: #f5f5f5; 
      min-height: 100vh; 
    }
    
    .page-icon { 
      font-size: 16px; 
      color: #1890ff; 
    }
    
    .content-grid { 
      display: grid; 
      grid-template-columns: 400px 1fr; 
      gap: 24px; 
      margin-top: 16px; 
    }
    
    .action-card, .list-card {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
      border-radius: 8px;
    }
    
    .info-box {
      background: #e6f7ff;
      border: 1px solid #91d5ff;
      border-radius: 4px;
      padding: 12px;
      margin-bottom: 16px;
      display: flex;
      align-items: flex-start;
      gap: 8px;
      font-size: 13px;
      color: #1890ff;
    }
    
    .info-box i {
      margin-top: 2px;
    }
    
    .diagnostic-form {
      margin-top: 8px;
    }
    
    .form-actions { 
      display: flex; 
      gap: 12px; 
      align-items: center; 
      margin-top: 8px;
    }
    
    .reports-table { 
      width: 100%; 
    }
    
    .reports-table th {
      background: #fafafa;
      font-weight: 600;
      color: #262626;
    }
    
    .reports-table td {
      border-bottom: 1px solid #f0f0f0;
    }
    
    .action-buttons {
      display: flex;
      gap: 4px;
    }
    
    .modal-actions {
      margin-top: 16px;
      text-align: right;
    }
    
    @media (max-width: 1200px) { 
      .content-grid { 
        grid-template-columns: 1fr; 
      } 
    }
  `]
})
export class DiagnosticsManagementComponent implements OnInit, OnDestroy {
  namespace = 'default';
  cluster = '';
  reports: DiagnosisReport[] = [];
  starting = false;
  loading = false;
  detailVisible = false;
  selectedReport: DiagnosisReport | null = null;
  
  private statusPoll?: Subscription;
  private message = inject(NzMessageService);
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);

  ngOnInit(): void {
    // Parse query parameters to auto-fill and trigger when navigating from external pages
    const qp = this.route.snapshot.queryParamMap;
    const ns = qp.get('namespace');
    const cl = qp.get('cluster');
    const auto = qp.get('autoStart');
    if (ns) this.namespace = ns;
    if (cl) this.cluster = cl;
    this.refresh();
    if (cl && (auto === '1' || auto === 'true')) {
      setTimeout(() => this.start(), 0);
    }
  }

  ngOnDestroy(): void {
    this.stopPolling();
  }

  statusColor(s?: string): string { 
    switch (s) {
      case 'running': return 'processing';
      case 'succeeded': case 'completed': return 'success';
      case 'failed': return 'error';
      case 'pending': return 'warning';
      default: return 'default';
    }
  }

  statusIcon(s?: string): string {
    switch (s) {
      case 'running': return 'loading';
      case 'succeeded': case 'completed': return 'check-circle';
      case 'failed': return 'close-circle';
      case 'pending': return 'clock-circle';
      default: return 'question-circle';
    }
  }

  statusText(s?: string): string {
    switch (s) {
      case 'running': return '运行中';
      case 'succeeded': case 'completed': return '已完成';
      case 'failed': return '失败';
      case 'pending': return '等待中';
      default: return '未知';
    }
  }

  refresh(): void {
    this.loading = true;
    this.api.listDiagnosisReports().subscribe({
      next: (items) => { 
        this.reports = (items || []) as DiagnosisReport[]; 
        this.loading = false;
      },
      error: () => { 
        this.reports = []; 
        this.loading = false;
      }
    });
  }

  start(): void {
    if (!this.cluster || !this.namespace) return;
    this.starting = true;
    this.api.startDiagnosis(this.namespace, this.cluster).subscribe({
      next: (res: any) => {
        this.message.success('诊断已触发');
        const id = res?.id;
        this.refresh();
        if (id) this.pollStatus(id);
        this.starting = false;
      },
      error: (err) => { 
        this.message.error(err?.error?.message || '诊断触发失败'); 
        this.starting = false; 
      }
    });
  }

  pollStatus(id: string): void {
    this.stopPolling();
    this.statusPoll = interval(5000).subscribe(() => {
      this.api.getDiagnosisStatus(this.namespace, id).subscribe({
        next: (s: any) => {
          const status = s?.status || s?.phase;
          const progress = s?.progress || 0;
          // Update local list item status
          const idx = this.reports.findIndex(r => r.id === id);
          if (idx >= 0) {
            const copy = [...this.reports];
            copy[idx] = { ...copy[idx], status, progress };
            this.reports = copy;
          }
          if (status === 'succeeded' || status === 'failed' || status === 'completed') {
            this.stopPolling();
            this.message.info(`诊断任务${status === 'succeeded' || status === 'completed' ? '完成' : '失败'}`);
          }
        }
      });
    });
  }

  stopPolling(): void { 
    if (this.statusPoll) { 
      this.statusPoll.unsubscribe(); 
      this.statusPoll = undefined; 
    } 
  }

  download(r: DiagnosisReport): void {
    if (!r?.id) return;
    this.api.downloadDiagnosisReport(r.namespace || 'default', r.id).subscribe({
      next: (resp: any) => {
        const url = resp?.url;
        const command = resp?.command;
        if (url && url !== '/api/v1/diagnostics/' + r.namespace + '/' + r.id + '/file') {
          window.open(url, '_blank');
        } else if (command) {
          // Display kubectl cp command
          this.message.info('请使用命令下载: ' + command, { nzDuration: 10000 });
        } else {
          this.message.warning('下载链接不可用');
        }
      },
      error: () => this.message.error('下载失败')
    });
  }

  delete(r: DiagnosisReport): void {
    if (!r?.id) return;
    this.api.deleteDiagnosisReport(r.namespace || 'default', r.id).subscribe({
      next: () => {
        this.message.success('诊断报告已删除');
        this.refresh();
      },
      error: (err) => this.message.error(err?.error?.message || '删除失败')
    });
  }

  showDetail(r: DiagnosisReport): void {
    this.selectedReport = r;
    this.detailVisible = true;
  }
}