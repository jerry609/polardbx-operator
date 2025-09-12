import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-monitoring-preflight',
  standalone: true,
  imports: [CommonModule, NzCardModule, NzButtonModule, NzIconModule, NzSpinModule, NzTagModule, NzGridModule, NzDescriptionsModule],
  template: `
    <div class="preflight">
      <div class="page-header">
        <div class="header-content">
          <h1 class="page-title">
            <i nz-icon nzType="safety" class="page-icon"></i>
            监控安装前健康体检
          </h1>
          <p class="page-description">检查时钟同步、可用区分布与 IOPS 性能，确保监控组件安装前环境就绪</p>
        </div>
      </div>

      <div class="page-content">
        <nz-card class="control-panel" nzTitle="体检操作">
          <div class="toolbar">
            <button nz-button nzType="primary" (click)="reload()" [nzLoading]="loading">
              <i nz-icon nzType="reload"></i>
              刷新检查
            </button>
            <span class="timestamp" *ngIf="!loading && data?.timestamp">基准时间：{{ data?.timestamp }}</span>
          </div>
        </nz-card>

        <div *ngIf="loading" class="loading-container">
          <nz-spin nzSize="large" nzTip="正在检查环境..."></nz-spin>
        </div>

        <div class="check-sections" *ngIf="!loading">
          <nz-card class="check-card" nzTitle="可用区分布检查" [nzExtra]="azExtra">
            <ng-template #azExtra>
              <i nz-icon nzType="global" class="section-icon"></i>
            </ng-template>
            <nz-descriptions nzBordered [nzColumn]="2">
              <nz-descriptions-item nzTitle="节点总数">{{ data?.az?.nodeCount || 0 }}</nz-descriptions-item>
              <nz-descriptions-item nzTitle="有区标记节点">{{ data?.az?.nodesWithZone || 0 }}</nz-descriptions-item>
              <nz-descriptions-item nzTitle="可用区数量">{{ data?.az?.count || 0 }}</nz-descriptions-item>
              <nz-descriptions-item nzTitle="多可用区部署">
                <nz-tag [nzColor]="data?.az?.hasMultiple ? 'success' : 'warning'">
                  {{ data?.az?.hasMultiple ? '是' : '否' }}
                </nz-tag>
              </nz-descriptions-item>
              <nz-descriptions-item nzTitle="可用区列表" [nzSpan]="2">
                {{ (data?.az?.zones || []).join(', ') || '未检测到' }}
              </nz-descriptions-item>
            </nz-descriptions>
          </nz-card>

          <nz-card class="check-card" nzTitle="时钟同步检查" [nzExtra]="clockExtra">
            <ng-template #clockExtra>
              <i nz-icon nzType="clock-circle" class="section-icon"></i>
            </ng-template>
            <nz-descriptions nzBordered [nzColumn]="2">
              <nz-descriptions-item nzTitle="控制面时间">{{ data?.clock?.controllerTime || '-' }}</nz-descriptions-item>
              <nz-descriptions-item nzTitle="漂移校验">{{ data?.clock?.skewAssessed ? '已完成' : '未完成' }}</nz-descriptions-item>
              <nz-descriptions-item nzTitle="同步状态" [nzSpan]="2">
                <nz-tag [nzColor]="data?.clock?.ok ? 'success' : 'error'">
                  {{ data?.clock?.ok ? '正常' : '异常' }}
                </nz-tag>
                <span style="margin-left: 8px;">{{ data?.clock?.message || '-' }}</span>
              </nz-descriptions-item>
            </nz-descriptions>
          </nz-card>

          <nz-card class="check-card" nzTitle="IOPS 性能检查" [nzExtra]="iopsExtra">
            <ng-template #iopsExtra>
              <i nz-icon nzType="dashboard" class="section-icon"></i>
            </ng-template>
            <nz-descriptions nzBordered [nzColumn]="2">
              <nz-descriptions-item nzTitle="性能估算">{{ data?.iops?.estimated ? '已完成' : '未完成' }}</nz-descriptions-item>
              <nz-descriptions-item nzTitle="IOPS 状态">
                <nz-tag [nzColor]="data?.iops?.ok ? 'success' : 'warning'">
                  {{ data?.iops?.ok ? '满足要求' : '未知/不足' }}
                </nz-tag>
              </nz-descriptions-item>
              <nz-descriptions-item nzTitle="详细说明" [nzSpan]="2">
                {{ data?.iops?.message || '暂无详细信息' }}
              </nz-descriptions-item>
            </nz-descriptions>
          </nz-card>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .preflight {
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
    }
    
    .control-panel {
      margin-bottom: 16px;
      background: #fff;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.06);
      border: 1px solid #e0e0e0;
    }
    
    .toolbar {
      display: flex;
      align-items: center;
      gap: 16px;
    }
    
    .timestamp {
      color: rgba(0, 0, 0, 0.6);
      font-size: 12px;
    }
    
    .loading-container {
      display: flex;
      justify-content: center;
      align-items: center;
      padding: 80px 0;
    }
    
    .check-sections {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    
    .check-card {
      background: #fff;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.06);
      border: 1px solid #e0e0e0;
    }
    
    .section-icon {
      font-size: 16px;
      color: #1890ff;
    }
    
    /* 响应式设计 */
    @media (max-width: 768px) {
      .preflight {
        padding: 16px;
      }
      
      .toolbar {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
      }
    }
  `]
})
export class MonitoringPreflightComponent implements OnInit {
  loading = false;
  data: any = null;

  constructor(private api: ApiService) {}

  ngOnInit(): void { this.reload(); }

  reload(): void {
    this.loading = true;
    this.api.getMonitoringPreflight().subscribe({
      next: (d) => { this.data = d; },
      error: () => { this.data = null; },
      complete: () => { this.loading = false; }
    });
  }
}

