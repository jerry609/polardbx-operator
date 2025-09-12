import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators, FormsModule } from '@angular/forms';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzInputNumberModule } from 'ng-zorro-antd/input-number';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { NzIconModule } from 'ng-zorro-antd/icon';

import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-log-collector-ilm',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    FormsModule,
    NzCardModule,
    NzFormModule,
    NzInputModule,
    NzInputNumberModule,
    NzSelectModule,
    NzTableModule,
    NzButtonModule,
    NzAlertModule,
    NzIconModule
  ],
  template: `
    <div class="log-ilm">
      <nz-card nzTitle="日志索引生命周期 (ILM) 配置">
        <form [formGroup]="form" nz-form nzLayout="vertical">
          <div nz-row nzGutter="16">
            <div nz-col nzSpan="8">
              <label>保留期(天)</label>
              <nz-input-number formControlName="retentionDays" [nzMin]="1"></nz-input-number>
            </div>
            <div nz-col nzSpan="8">
              <label>冷热分层策略(JSON)</label>
              <input nz-input formControlName="tieringPolicy" placeholder='{"hotDays":7,"coldSink":"s3"}' />
            </div>
            <div nz-col nzSpan="8">
              <label>多副本存储</label>
              <nz-select formControlName="replicaSinks"
                         nzMode="tags"
                         [nzDropdownMatchSelectWidth]="false"
                         [nzMaxTagCount]="5"
                         nzAllowClear
                         nzPlaceHolder="输入或选择存储标识"
                         [nzTokenSeparators]="[',',';','\n']">
                <nz-option *ngFor="let s of availableSinks" [nzValue]="s" [nzLabel]="s"></nz-option>
              </nz-select>
            </div>
          </div>

          <div style="margin-top:12px; display:flex; gap:8px;">
            <button nz-button nzType="default" (click)="load()">重置</button>
            <button nz-button nzType="primary" [disabled]="form.invalid" (click)="save()">保存</button>
          </div>
        </form>
        <nz-alert *ngIf="msg" [nzType]="msgType" [nzMessage]="msg" nzShowIcon style="margin-top:12px;"></nz-alert>
      </nz-card>

      <nz-card nzTitle="连通性测试" style="margin-top:16px;">
        <div style="display:flex; gap:8px; align-items:center;">
          <nz-select [(ngModel)]="testSink" nzPlaceHolder="选择或输入存储" nzAllowClear nzMode="default">
            <nz-option *ngFor="let s of availableSinks" [nzValue]="s" [nzLabel]="s"></nz-option>
          </nz-select>
          <button nz-button nzType="default" (click)="testConnectivity()"><i nz-icon nzType="link"></i> 测试连通性</button>
          <button nz-button nzType="default" (click)="runSampleQuery()"><i nz-icon nzType="search"></i> 样本查询</button>
        </div>
        <nz-alert *ngIf="testMsg" [nzType]="testOk? 'success':'warning'" [nzMessage]="testMsg" nzShowIcon style="margin-top:12px;"></nz-alert>
      </nz-card>
    </div>
  `,
  styles: [`
    .log-ilm { padding: 20px; }
    label { display:block; margin-bottom:4px; color: rgba(0,0,0,.65); }
  `]
})
export class LogCollectorIlmComponent implements OnInit {
  private api = inject(ApiService);
  private fb = inject(FormBuilder);

  form = this.fb.group({
    retentionDays: [7, [Validators.required, Validators.min(1)]],
    tieringPolicy: [''],
    replicaSinks: [[] as string[]]
  });

  availableSinks: string[] = [ 's3-prod', 'oss-archive', 'hpfs-tier1' ];
  testSink: string | null = null;
  msg = '';
  msgType: 'success'|'warning'|'info'|'error' = 'success';
  testMsg = '';
  testOk = false;

  ngOnInit(): void { this.load(); }

  load(): void {
    // 占位：从系统设置读取默认值
    this.api.getSystemSettings().subscribe({
      next: (cfg) => {
        const patch: any = {};
        if (cfg) {
          if (cfg['logs.retentionDays']) patch.retentionDays = Number(cfg['logs.retentionDays']);
          if (cfg['logs.tieringPolicy']) patch.tieringPolicy = cfg['logs.tieringPolicy'];
          if (cfg['logs.replicaSinks']) try { patch.replicaSinks = JSON.parse(cfg['logs.replicaSinks']); } catch {}
        }
        this.form.patchValue(patch);
      },
      error: () => {}
    });
  }

  save(): void {
    if (this.form.invalid) return;
    const v = this.form.value as any;
    const body: Record<string, any> = {
      'logs.retentionDays': v.retentionDays,
      'logs.tieringPolicy': v.tieringPolicy,
      'logs.replicaSinks': JSON.stringify(v.replicaSinks || [])
    };
    this.api.updateSystemSettings(body).subscribe({
      next: () => { this.msgType = 'success'; this.msg = '保存成功'; },
      error: () => { this.msgType = 'error'; this.msg = '保存失败'; }
    });
  }

  testConnectivity(): void {
    const sink = this.testSink || this.form.value.replicaSinks?.[0];
    if (!sink) { this.testOk = false; this.testMsg = '请选择存储'; return; }
    // 复用现有校验接口作为占位
    this.api.validateSink(sink, 's3').subscribe({
      next: (r) => { this.testOk = r?.status === 'ok'; this.testMsg = r?.message || (this.testOk ? '连通性正常' : '连通性异常'); },
      error: () => { this.testOk = false; this.testMsg = '测试失败'; }
    });
  }

  runSampleQuery(): void {
    const sink = this.testSink || this.form.value.replicaSinks?.[0];
    if (!sink) { this.testOk = false; this.testMsg = '请选择存储'; return; }
    // 占位：调用 binlog metrics 作为样本
    this.api.getBinlogMetrics('default').subscribe({
      next: () => { this.testOk = true; this.testMsg = '样本查询成功（占位）'; },
      error: () => { this.testOk = false; this.testMsg = '样本查询失败'; }
    });
  }
}

