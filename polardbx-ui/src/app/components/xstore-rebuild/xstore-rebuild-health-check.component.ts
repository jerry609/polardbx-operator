import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-xstore-rebuild-health-check',
  standalone: true,
  imports: [CommonModule, FormsModule, MatCardModule, MatButtonModule, MatIconModule, MatSnackBarModule, MatChipsModule, MatProgressBarModule],
  template: `
    <div class="page">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>fact_check</mat-icon>
            备库重搭 · 健康检查
          </mat-card-title>
          <mat-card-subtitle>对目标集群/节点进行健康检查，评估重搭可行性</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <div class="form-row">
            <label>命名空间</label>
            <input [(ngModel)]="namespace" placeholder="default"/>
          </div>
          <div class="form-row">
            <label>目标 XStore</label>
            <input [(ngModel)]="xstore" placeholder="cluster-dn-0"/>
          </div>
          <div class="actions">
            <button mat-raised-button color="primary" (click)="run()" [disabled]="loading"><mat-icon>play_arrow</mat-icon> 执行检查</button>
          </div>

          <mat-progress-bar *ngIf="loading" mode="indeterminate" style="margin:12px 0"></mat-progress-bar>

          <div *ngIf="!loading && checked">
            <div class="result-row">
              <span class="k">XStore 对象</span>
              <span class="v">
                <mat-chip [color]="xstoreOk ? 'primary' : 'warn'" selected>{{ xstoreOk ? '存在' : '不存在' }}</mat-chip>
              </span>
            </div>
            <div class="result-row" *ngIf="xstoreOk">
              <span class="k">副本情况</span>
              <span class="v">{{ readyPods }}/{{ totalPods }} Ready</span>
            </div>
            <div class="pods" *ngIf="pods.length">
              <div class="pods-title"><mat-icon>apps</mat-icon> Pods</div>
              <div class="pod" *ngFor="let p of pods">
                <span class="name">{{ p.metadata?.name }}</span>
                <span class="ip">{{ p.status?.podIP || '-' }}</span>
                <mat-chip [color]="(p.status?.phase||'').toLowerCase()==='running' ? 'primary' : ((p.status?.phase||'').toLowerCase()==='pending' ? '' : 'warn')" selected>
                  {{ p.status?.phase || '-' }}
                </mat-chip>
                <span class="role">{{ p.metadata?.labels?.['xstore/role'] || p.metadata?.labels?.['polardbx/role'] || '-' }}</span>
              </div>
            </div>
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .page { padding: 20px; }
    .form-row { display: grid; grid-template-columns: 120px 1fr; gap: 8px; align-items: center; margin: 12px 0; }
    .form-row input { height: 34px; padding: 6px 10px; border: 1px solid #e0e0e0; border-radius: 6px; }
    .actions { display: flex; gap: 8px; align-items: center; }
    .result-row { display: grid; grid-template-columns: 120px 1fr; align-items: center; margin: 8px 0; }
    .result-row .k { color: #666; }
    .pods { margin-top: 12px; border-top: 1px solid #eee; padding-top: 12px; }
    .pods-title { font-weight: 600; color: #333; margin-bottom: 8px; display:flex; align-items:center; gap:6px; }
    .pod { display: grid; grid-template-columns: 1fr 140px 120px 120px; gap: 8px; align-items: center; padding: 6px 0; border-bottom: 1px dashed #f0f0f0; }
    .pod:last-child { border-bottom: 0; }
    .name { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace; }
    .ip { color: #666; }
    .role { color: #888; }
  `]
})
export class XStoreRebuildHealthCheckComponent {
  namespace = 'default';
  xstore = '';
  loading = false;
  checked = false;
  xstoreOk = false;
  readyPods = 0;
  totalPods = 0;
  pods: any[] = [];

  constructor(private snack: MatSnackBar, private api: ApiService) {}

  run(): void {
    if (!this.xstore) { this.snack.open('请填写目标 XStore', '关闭', { duration: 2500 }); return; }
    this.loading = true;
    this.checked = false;
    this.xstoreOk = false;
    this.readyPods = 0;
    this.totalPods = 0;
    this.pods = [];
    // 并行获取 XStore 与 Pods
    this.api.getXStore(this.namespace, this.xstore).subscribe({
      next: (xs) => {
        this.xstoreOk = !!xs;
        const status: any = (xs as any)?.status || {};
        this.readyPods = Number(status.readyPods || 0);
        this.totalPods = Number(status.totalPods || 0);
      },
      error: () => { this.xstoreOk = false; }
    });
    this.api.getXStorePods(this.namespace, this.xstore).subscribe({
      next: (pods) => { this.pods = pods || []; this.finish(); },
      error: () => { this.pods = []; this.finish(); }
    });
  }

  private finish(): void {
    this.loading = false;
    this.checked = true;
  }
}