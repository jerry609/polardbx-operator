import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzProgressModule } from 'ng-zorro-antd/progress';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { NzStepsModule } from 'ng-zorro-antd/steps';
import { NzEmptyModule } from 'ng-zorro-antd/empty';
import { NzMessageService } from 'ng-zorro-antd/message';
import { Subject, timer } from 'rxjs';
import { takeUntil, switchMap } from 'rxjs/operators';

import { ApiService } from '../../services/api.service';
import { XStoreFollower } from '../../models/xstore-follower.model';

@Component({
  selector: 'app-rebuild-task-detail',
  standalone: true,
  imports: [
    CommonModule,
    NzCardModule,
    NzTabsModule,
    NzButtonModule,
    NzIconModule,
    NzTagModule,
    NzSpinModule,
    NzProgressModule,
    NzDescriptionsModule,
    NzStepsModule,
    NzEmptyModule
  ],
  template: `
    <div class="rebuild-task-detail-container">
      <div *ngIf="!isLoading && task; else loadingTemplate">
        <!-- 头部信息 -->
        <nz-card class="header-card">
          <div class="task-header">
            <div class="header-left">
              <button nz-button nzType="text" (click)="goBack()" class="back-button">
                <i nz-icon nzType="arrow-left"></i>
                返回任务列表
              </button>
              <h2 class="task-title">
                <i nz-icon [nzType]="getTaskIcon()" class="task-icon"></i>
                {{ task.metadata.name }}
              </h2>
            </div>
            <div class="header-right">
              <nz-tag [nzColor]="getStatusColor()" class="status-tag">
                {{ getDisplayStatus() }}
              </nz-tag>
              <nz-tag [nzColor]="getRoleColor()" class="role-tag">
                {{ getRoleDisplayName() }}
              </nz-tag>
            </div>
          </div>
          
          <!-- 进度条 -->
          <div class="progress-section" *ngIf="!isEndPhase()">
            <nz-progress 
              [nzPercent]="getProgressPercent()" 
              [nzStatus]="getProgressStatus()"
              [nzStrokeWidth]="8">
            </nz-progress>
          </div>
        </nz-card>

        <!-- 详情标签页 -->
        <nz-card class="detail-card">
          <nz-tabset nzType="card">
            <!-- 概览 -->
            <nz-tab nzTitle="概览">
              <div class="overview-content">
                <nz-descriptions nzTitle="基本信息" nzBordered [nzColumn]="2">
                  <nz-descriptions-item nzTitle="任务名">{{ task.metadata.name }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="命名空间">{{ task.metadata.namespace }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="角色类型">{{ getRoleDisplayName() }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="XStore">{{ task.spec.xStoreName }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="构建方式">
                    <nz-tag [nzColor]="task.spec.local ? 'blue' : 'orange'">
                      {{ task.spec.local ? '本机构建' : '跨机构建' }}
                    </nz-tag>
                  </nz-descriptions-item>
                  <nz-descriptions-item nzTitle="创建时间">{{ formatTime(task.metadata.creationTimestamp) }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="源 Pod" *ngIf="task.spec.fromPodName">{{ task.spec.fromPodName }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="目标 Pod" *ngIf="getTargetPodName()">{{ getTargetPodName() }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="目标节点" *ngIf="getTargetNodeName()">{{ getTargetNodeName() }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="当前状态">
                    <nz-tag [nzColor]="getStatusColor()">{{ getDisplayStatus() }}</nz-tag>
                  </nz-descriptions-item>
                  <nz-descriptions-item nzTitle="状态消息" *ngIf="task.status?.message">{{ task.status?.message }}</nz-descriptions-item>
                  <nz-descriptions-item nzTitle="当前任务" *ngIf="task.status?.currentJobName">{{ task.status?.currentJobName }}</nz-descriptions-item>
                </nz-descriptions>

                <!-- 操作按钮 -->
                <div class="action-section">
                  <button 
                    nz-button 
                    nzType="default" 
                    *ngIf="!isEndPhase()"
                    (click)="stopTask()">
                    <i nz-icon nzType="stop"></i>
                    停止任务
                  </button>
                  <button 
                    nz-button 
                    nzType="primary" 
                    *ngIf="canRetry()"
                    (click)="retryTask()">
                    <i nz-icon nzType="redo"></i>
                    重试任务
                  </button>
                  <button 
                    nz-button 
                    nzDanger
                    *ngIf="isEndPhase()"
                    (click)="deleteTask()">
                    <i nz-icon nzType="delete"></i>
                    删除任务
                  </button>
                </div>
              </div>
            </nz-tab>

            <!-- 步骤 -->
            <nz-tab nzTitle="执行步骤">
              <div class="steps-content">
                <nz-steps [nzCurrent]="getCurrentStepIndex()" nzDirection="vertical">
                  <nz-step 
                    *ngFor="let step of getSteps()" 
                    [nzTitle]="step.title"
                    [nzDescription]="step.description"
                    [nzStatus]="step.status">
                  </nz-step>
                </nz-steps>
              </div>
            </nz-tab>

            <!-- 日志 -->
            <nz-tab nzTitle="任务日志">
              <div class="logs-content">
                <nz-empty nzNotFoundContent="日志功能开发中，请通过 kubectl 查看相关 Pod 日志">
                  <div nz-empty-footer>
                    <p>当前任务相关信息：</p>
                    <ul>
                      <li *ngIf="task.status?.currentJobName">当前任务: {{ task.status?.currentJobName }}</li>
                      <li *ngIf="task.status?.currentJobTask">当前子任务: {{ task.status?.currentJobTask }}</li>
                      <li *ngIf="task.status?.rebuildPodName">重建 Pod: {{ task.status?.rebuildPodName }}</li>
                    </ul>
                  </div>
                </nz-empty>
              </div>
            </nz-tab>

            <!-- 事件 -->
            <nz-tab nzTitle="事件">
              <div class="events-content">
                <nz-empty nzNotFoundContent="事件功能开发中">
                  <div nz-empty-footer>
                    <p>可通过以下命令查看相关事件：</p>
                    <code>kubectl get events -n {{ task.metadata.namespace }} --field-selector involvedObject.name={{ task.metadata.name }}</code>
                  </div>
                </nz-empty>
              </div>
            </nz-tab>
          </nz-tabset>
        </nz-card>
      </div>

      <!-- 加载模板 -->
      <ng-template #loadingTemplate>
        <div class="loading-container">
          <nz-spin nzSize="large" nzTip="加载任务详情中..."></nz-spin>
        </div>
      </ng-template>
    </div>
  `,
  styleUrls: ['./rebuild-task-detail.component.scss']
})
export class RebuildTaskDetailComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  
  task: XStoreFollower | null = null;
  isLoading = false;
  namespace = '';
  taskName = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService,
    private message: NzMessageService
  ) {}

  ngOnInit(): void {
    this.route.params.pipe(takeUntil(this.destroy$)).subscribe(params => {
      this.namespace = params['namespace'];
      this.taskName = params['name'];
      this.startPolling();
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private startPolling(): void {
    // 立即加载一次，然后每3秒轮询（如果不是终态）
    timer(0, 3000)
      .pipe(
        switchMap(() => this.loadTask()),
        takeUntil(this.destroy$)
      )
      .subscribe(task => {
        if (task && this.isEndPhase()) {
          // 如果是终态，停止轮询
          this.destroy$.next();
        }
      });
  }

  private async loadTask(): Promise<XStoreFollower | null> {
    if (this.isLoading) return null;
    
    this.isLoading = true;
    try {
      const task = await this.apiService.getXStoreFollower(this.namespace, this.taskName).toPromise();
      this.task = task || null;
      return task || null;
    } catch (error) {
      console.error('Failed to load task detail:', error);
      this.message.error('加载任务详情失败');
      return null;
    } finally {
      this.isLoading = false;
    }
  }

  getTaskIcon(): string {
    if (!this.task) return 'setting';
    switch (this.task.spec.role) {
      case 'learner': return 'experiment';
      case 'logger': return 'file-text';
      case 'follower': return 'share-alt';
      default: return 'setting';
    }
  }

  getRoleColor(): string {
    if (!this.task) return 'default';
    switch (this.task.spec.role) {
      case 'learner': return 'purple';
      case 'logger': return 'cyan';
      case 'follower': return 'blue';
      default: return 'default';
    }
  }

  getRoleDisplayName(): string {
    if (!this.task) return '';
    switch (this.task.spec.role) {
      case 'learner': return 'Learner';
      case 'logger': return 'Logger';
      case 'follower': return 'Follower';
      default: return this.task.spec.role || '';
    }
  }

  getDisplayStatus(): string {
    if (!this.task) return '';
    const phase = this.task.status?.phase || '';
    const statusMap: { [key: string]: string } = {
      '': '初始化中',
      'FollowerPhaseNew': '已创建',
      'FollowerPhaseCheck': '检查中',
      'FollowerPhaseBackupPrepare': '备份准备',
      'FollowerPhaseBackupStart': '开始备份',
      'FollowerPhaseBackup': '备份中',
      'FollowerPhaseLoggerCreate': '创建日志器',
      'FollowerPhaseLoggerRebuild': '重建日志',
      'FollowerCreateRemotePod': '创建远程Pod',
      'FollowerPhaseMonitorBackup': '监控备份',
      'FollowerPhaseBeforeRestore': '准备恢复',
      'FollowerPhaseRestore': '恢复中',
      'FollowerPhaseAfterRestore': '完成恢复',
      'FollowerPhaseWaitSwitch': '等待切换',
      'FollowerPhaseSuccess': '成功',
      'FollowerPhaseFailed': '失败',
      'FollowerPhaseDeleting': '删除中'
    };
    return statusMap[phase] || '初始化中';
  }

  getStatusColor(): string {
    if (!this.task) return 'default';
    const phase = this.task.status?.phase || '';
    switch (phase) {
      case 'FollowerPhaseSuccess': return 'green';
      case 'FollowerPhaseFailed': return 'red';
      case 'FollowerPhaseDeleting':
      case 'FollowerPhaseWaitSwitch': return 'orange';
      default: return 'blue';
    }
  }

  getProgressPercent(): number {
    if (!this.task) return 0;
    const phase = this.task.status?.phase || '';
    const progressMap: { [key: string]: number } = {
      '': 0,
      'FollowerPhaseNew': 5,
      'FollowerPhaseCheck': 10,
      'FollowerPhaseBackupPrepare': 20,
      'FollowerPhaseBackupStart': 25,
      'FollowerPhaseBackup': 40,
      'FollowerPhaseLoggerCreate': 35,
      'FollowerPhaseLoggerRebuild': 50,
      'FollowerCreateRemotePod': 45,
      'FollowerPhaseMonitorBackup': 60,
      'FollowerPhaseBeforeRestore': 70,
      'FollowerPhaseRestore': 80,
      'FollowerPhaseAfterRestore': 90,
      'FollowerPhaseWaitSwitch': 95,
      'FollowerPhaseSuccess': 100,
      'FollowerPhaseFailed': 0,
      'FollowerPhaseDeleting': 0
    };
    return progressMap[phase] || 0;
  }

  getProgressStatus(): 'success' | 'exception' | 'active' | 'normal' {
    if (!this.task) return 'active';
    const phase = this.task.status?.phase || '';
    if (phase === 'FollowerPhaseSuccess') return 'success';
    if (phase === 'FollowerPhaseFailed') return 'exception';
    return 'active';
  }

  isEndPhase(): boolean {
    if (!this.task) return false;
    const phase = this.task.status?.phase || '';
    return ['FollowerPhaseSuccess', 'FollowerPhaseFailed', 'FollowerPhaseDeleting'].includes(phase);
  }

  canRetry(): boolean {
    return this.task?.status?.phase === 'FollowerPhaseFailed';
  }

  getTargetPodName(): string {
    if (!this.task) return '';
    return this.task.status?.targetPodName || this.task.spec.targetPodName || '';
  }

  getTargetNodeName(): string {
    if (!this.task) return '';
    return this.task.status?.rebuildNodeName || this.task.spec.nodeName || '';
  }

  formatTime(timestamp?: string): string {
    if (!timestamp) return '-';
    return new Date(timestamp).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  }

  getSteps(): Array<{title: string, description: string, status: string}> {
    if (!this.task) return [];
    
    const currentPhase = this.task.status?.phase || '';
    const steps = [
      { phase: 'FollowerPhaseNew', title: '任务创建', description: '创建重搭任务' },
      { phase: 'FollowerPhaseCheck', title: '环境检查', description: '检查重搭环境和条件' },
      { phase: 'FollowerPhaseBackupPrepare', title: '备份准备', description: '准备备份操作' },
      { phase: 'FollowerPhaseBackupStart', title: '开始备份', description: '启动数据备份' },
      { phase: 'FollowerPhaseBackup', title: '执行备份', description: '执行数据备份操作' },
      { phase: 'FollowerPhaseLoggerCreate', title: '创建日志器', description: '创建日志收集器' },
      { phase: 'FollowerPhaseLoggerRebuild', title: '重建日志', description: '重建日志节点' },
      { phase: 'FollowerCreateRemotePod', title: '创建远程Pod', description: '创建远程执行Pod' },
      { phase: 'FollowerPhaseMonitorBackup', title: '监控备份', description: '监控备份进度' },
      { phase: 'FollowerPhaseBeforeRestore', title: '准备恢复', description: '准备数据恢复' },
      { phase: 'FollowerPhaseRestore', title: '执行恢复', description: '执行数据恢复操作' },
      { phase: 'FollowerPhaseAfterRestore', title: '完成恢复', description: '完成数据恢复' },
      { phase: 'FollowerPhaseWaitSwitch', title: '等待切换', description: '等待角色切换' },
      { phase: 'FollowerPhaseSuccess', title: '任务完成', description: '重搭任务成功完成' }
    ];

    return steps.map(step => {
      let status = 'wait';
      if (step.phase === currentPhase) {
        status = currentPhase === 'FollowerPhaseFailed' ? 'error' : 'process';
      } else if (this.isPhaseReached(step.phase, currentPhase)) {
        status = 'finish';
      }
      
      return {
        title: step.title,
        description: step.description,
        status
      };
    });
  }

  getCurrentStepIndex(): number {
    if (!this.task) return 0;
    const steps = this.getSteps();
    const currentPhase = this.task.status?.phase || '';
    return steps.findIndex(step => step.status === 'process');
  }

  private isPhaseReached(targetPhase: string, currentPhase: string): boolean {
    const phaseOrder = [
      'FollowerPhaseNew', 'FollowerPhaseCheck', 'FollowerPhaseBackupPrepare',
      'FollowerPhaseBackupStart', 'FollowerPhaseBackup', 'FollowerPhaseLoggerCreate',
      'FollowerPhaseLoggerRebuild', 'FollowerCreateRemotePod', 'FollowerPhaseMonitorBackup',
      'FollowerPhaseBeforeRestore', 'FollowerPhaseRestore', 'FollowerPhaseAfterRestore',
      'FollowerPhaseWaitSwitch', 'FollowerPhaseSuccess'
    ];
    
    const targetIndex = phaseOrder.indexOf(targetPhase);
    const currentIndex = phaseOrder.indexOf(currentPhase);
    
    return targetIndex < currentIndex;
  }

  goBack(): void {
    this.router.navigate(['/storage/xstore-rebuild/rebuild/tasks']);
  }

  stopTask(): void {
    // TODO: 实现停止任务
    this.message.info('停止任务功能开发中');
  }

  retryTask(): void {
    // TODO: 实现重试任务
    this.message.info('重试任务功能开发中');
  }

  deleteTask(): void {
    // TODO: 实现删除任务
    this.message.info('删除任务功能开发中');
  }
}
