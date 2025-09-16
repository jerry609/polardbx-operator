import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzTagModule } from 'ng-zorro-antd/tag';
import { Subject } from 'rxjs';
import { takeUntil, finalize } from 'rxjs/operators';

import { ApiService } from '../../services/api.service';
import { LoadingService } from '../../services/loading.service';
import { CreateXStoreFollowerRequest } from '../../models/xstore-follower.model';
import { XStore } from '../../models/xstore.model';

interface RebuildFormData {
  namespace: string;
  role: 'learner' | 'logger' | 'follower';
  xStoreName: string;
  targetPodName?: string;
  fromPodName?: string;
  nodeName?: string;
  local: boolean;
  description?: string;
}

@Component({
  selector: 'app-rebuild-form',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NzCardModule,
    NzButtonModule,
    NzIconModule,
    NzInputModule,
    NzSelectModule,
    NzFormModule,
    NzSwitchModule,
    NzSpinModule,
    NzToolTipModule,
    NzAlertModule,
    NzDividerModule,
    NzTagModule
  ],
  template: `
    <div class="rebuild-form-container">
      <nz-card class="form-card" nzTitle="创建重搭任务">
        <div class="form-content">
          <form nz-form [formGroup]="rebuildForm" (ngSubmit)="onSubmit()">
            
            <!-- 基础信息 -->
            <div class="form-section">
              <h4 class="section-title">
                <i nz-icon nzType="setting"></i>
                <span>基础配置</span>
              </h4>
              
              <nz-form-item>
                <nz-form-label [nzSpan]="6" nzFor="namespace" nzRequired>命名空间</nz-form-label>
                <nz-form-control [nzSpan]="18" nzErrorTip="请选择命名空间">
                  <nz-select 
                    id="namespace" 
                    formControlName="namespace" 
                    nzPlaceHolder="选择命名空间"
                    (ngModelChange)="onNamespaceChange($event)">
                    <nz-option *ngFor="let ns of namespaces" [nzValue]="ns" [nzLabel]="ns"></nz-option>
                  </nz-select>
                </nz-form-control>
              </nz-form-item>

              <nz-form-item>
                <nz-form-label [nzSpan]="6" nzFor="role" nzRequired>重搭角色</nz-form-label>
                <nz-form-control [nzSpan]="18" nzErrorTip="请选择重搭角色">
                  <nz-select 
                    id="role" 
                    formControlName="role" 
                    nzPlaceHolder="选择角色类型"
                    (ngModelChange)="onRoleChange($event)">
                    <nz-option nzValue="learner" nzLabel="Learner（学习者节点）">
                      <div class="role-option">
                        <i nz-icon nzType="experiment"></i>
                        <span>Learner（学习者节点）</span>
                      </div>
                    </nz-option>
                    <nz-option nzValue="logger" nzLabel="Logger（日志节点）">
                      <div class="role-option">
                        <i nz-icon nzType="file-text"></i>
                        <span>Logger（日志节点）</span>
                      </div>
                    </nz-option>
                    <nz-option nzValue="follower" nzLabel="Follower（从节点）">
                      <div class="role-option">
                        <i nz-icon nzType="share-alt"></i>
                        <span>Follower（从节点）</span>
                      </div>
                    </nz-option>
                  </nz-select>
                </nz-form-control>
              </nz-form-item>

              <nz-form-item>
                <nz-form-label [nzSpan]="6" nzFor="xStoreName" nzRequired>目标 XStore</nz-form-label>
                <nz-form-control [nzSpan]="18" nzErrorTip="请选择目标 XStore">
                  <nz-select 
                    id="xStoreName" 
                    formControlName="xStoreName" 
                    nzPlaceHolder="选择 XStore"
                    (ngModelChange)="onXStoreChange($event)"
                    [nzLoading]="isLoadingXStores">
                    <nz-option *ngFor="let xstore of availableXStores" [nzValue]="xstore.metadata.name" [nzLabel]="xstore.metadata.name">
                      <div class="xstore-option">
                        <span class="xstore-name">{{ xstore.metadata.name }}</span>
                        <nz-tag [nzColor]="getXStoreStatusColor(xstore)" class="status-tag">
                          {{ getXStoreStatusText(xstore) }}
                        </nz-tag>
                      </div>
                    </nz-option>
                  </nz-select>
                </nz-form-control>
              </nz-form-item>
            </div>

            <nz-divider></nz-divider>

            <!-- 高级配置 -->
            <div class="form-section">
              <h4 class="section-title">
                <i nz-icon nzType="tool"></i>
                <span>高级配置</span>
              </h4>

              <!-- 本机构建选项 -->
              <nz-form-item>
                <nz-form-label [nzSpan]="6" nzFor="local">本机构建</nz-form-label>
                <nz-form-control [nzSpan]="18">
                  <nz-switch 
                    id="local" 
                    formControlName="local"
                    [nzLoading]="false">
                  </nz-switch>
                  <span class="switch-hint">
                    {{ rebuildForm.get('local')?.value ? '在当前节点进行重搭' : '跨节点或新建节点重搭' }}
                  </span>
                </nz-form-control>
              </nz-form-item>

              <!-- 目标 Pod 选择 -->
              <nz-form-item *ngIf="shouldShowTargetPod()">
                <nz-form-label [nzSpan]="6" nzFor="targetPodName" [nzRequired]="isTargetPodRequired()">
                  目标 Pod
                  <span nz-tooltip="选择要重搭的目标 Pod，建议选择非 Leader 且处于 Running 状态的 Pod">
                    <i nz-icon nzType="question-circle"></i>
                  </span>
                </nz-form-label>
                <nz-form-control [nzSpan]="18" [nzErrorTip]="getTargetPodErrorTip()">
                  <nz-select 
                    id="targetPodName" 
                    formControlName="targetPodName" 
                    nzPlaceHolder="选择目标 Pod"
                    [nzLoading]="isLoadingPods"
                    nzAllowClear>
                    <nz-option *ngFor="let pod of filteredTargetPods" [nzValue]="pod.metadata.name" [nzLabel]="pod.metadata.name">
                      <div class="pod-option">
                        <span class="pod-name">{{ pod.metadata.name }}</span>
                        <div class="pod-info">
                          <nz-tag [nzColor]="getPodStatusColor(pod)" class="status-tag">
                            {{ pod.status?.phase || '未知' }}
                          </nz-tag>
                          <nz-tag *ngIf="getPodRole(pod)" [nzColor]="getRoleColor(getPodRole(pod))" class="role-tag">
                            {{ getPodRole(pod) }}
                          </nz-tag>
                        </div>
                      </div>
                    </nz-option>
                  </nz-select>
                </nz-form-control>
              </nz-form-item>

              <!-- 源 Pod 选择（Logger 重搭时显示） -->
              <nz-form-item *ngIf="shouldShowFromPod()">
                <nz-form-label [nzSpan]="6" nzFor="fromPodName">
                  源 Pod
                  <span nz-tooltip="选择数据源 Pod，留空将使用系统默认策略（可能产生较重的数据传输）">
                    <i nz-icon nzType="question-circle"></i>
                  </span>
                </nz-form-label>
                <nz-form-control [nzSpan]="18">
                  <nz-select 
                    id="fromPodName" 
                    formControlName="fromPodName" 
                    nzPlaceHolder="选择源 Pod（可选）"
                    [nzLoading]="isLoadingPods"
                    nzAllowClear>
                    <nz-option *ngFor="let pod of filteredSourcePods" [nzValue]="pod.metadata.name" [nzLabel]="pod.metadata.name">
                      <div class="pod-option">
                        <span class="pod-name">{{ pod.metadata.name }}</span>
                        <div class="pod-info">
                          <nz-tag [nzColor]="getPodStatusColor(pod)" class="status-tag">
                            {{ pod.status?.phase || '未知' }}
                          </nz-tag>
                          <nz-tag *ngIf="getPodRole(pod)" [nzColor]="getRoleColor(getPodRole(pod))" class="role-tag">
                            {{ getPodRole(pod) }}
                          </nz-tag>
                        </div>
                      </div>
                    </nz-option>
                  </nz-select>
                </nz-form-control>
              </nz-form-item>

              <!-- 节点选择（跨机构建时显示） -->
              <nz-form-item *ngIf="shouldShowNodeName()">
                <nz-form-label [nzSpan]="6" nzFor="nodeName">目标节点</nz-form-label>
                <nz-form-control [nzSpan]="18">
                  <nz-input 
                    id="nodeName" 
                    formControlName="nodeName" 
                    nzPlaceHolder="输入目标节点名称（可选）">
                  </nz-input>
                </nz-form-control>
              </nz-form-item>

              <!-- 备注 -->
              <nz-form-item>
                <nz-form-label [nzSpan]="6" nzFor="description">备注</nz-form-label>
                <nz-form-control [nzSpan]="18">
                  <nz-input 
                    id="description" 
                    formControlName="description" 
                    nzPlaceHolder="可选的任务描述">
                  </nz-input>
                </nz-form-control>
              </nz-form-item>
            </div>

            <!-- 提示信息 -->
            <div class="form-section" *ngIf="getRoleHints().length > 0">
              <nz-alert 
                *ngFor="let hint of getRoleHints()" 
                [nzType]="hint.type" 
                [nzMessage]="hint.message" 
                [nzShowIcon]="true"
                class="role-hint">
              </nz-alert>
            </div>

            <!-- 操作按钮 -->
            <nz-form-item class="submit-buttons">
              <nz-form-control [nzOffset]="6" [nzSpan]="18">
                <button 
                  nz-button 
                  nzType="primary" 
                  [nzLoading]="isSubmitting"
                  [disabled]="!rebuildForm.valid"
                  type="submit">
                  <i nz-icon nzType="play-circle"></i>
                  <span>创建重搭任务</span>
                </button>
                <button 
                  nz-button 
                  nzType="default" 
                  (click)="onCancel()"
                  [disabled]="isSubmitting"
                  style="margin-left: 8px;">
                  <i nz-icon nzType="rollback"></i>
                  <span>取消</span>
                </button>
                <button 
                  nz-button 
                  nzType="default" 
                  (click)="goToTaskList()"
                  style="margin-left: 8px;">
                  <i nz-icon nzType="bars"></i>
                  <span>查看任务列表</span>
                </button>
              </nz-form-control>
            </nz-form-item>

          </form>
        </div>
      </nz-card>
    </div>
  `,
  styleUrls: ['./rebuild-form.component.scss']
})
export class RebuildFormComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  
  rebuildForm!: FormGroup;
  namespaces: string[] = [];
  availableXStores: XStore[] = [];
  targetPods: any[] = [];
  filteredTargetPods: any[] = [];
  filteredSourcePods: any[] = [];
  
  isLoadingXStores = false;
  isLoadingPods = false;
  isSubmitting = false;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private route: ActivatedRoute,
    private apiService: ApiService,
    private loadingService: LoadingService,
    private message: NzMessageService
  ) {
    this.initForm();
  }

  ngOnInit(): void {
    this.loadNamespaces();
    this.handleQueryParams();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private initForm(): void {
    this.rebuildForm = this.fb.group({
      namespace: ['default', [Validators.required]],
      role: ['learner', [Validators.required]],
      xStoreName: ['', [Validators.required]],
      targetPodName: [''],
      fromPodName: [''],
      nodeName: [''],
      local: [true],
      description: ['']
    });

    // Set validators based on role
    this.rebuildForm.get('role')?.valueChanges.subscribe(role => {
      this.updateFormValidators(role);
    });
  }

  private handleQueryParams(): void {
    this.route.queryParams.pipe(takeUntil(this.destroy$)).subscribe(params => {
      if (params['role']) {
        this.rebuildForm.patchValue({ role: params['role'] });
      }
      if (params['namespace']) {
        this.rebuildForm.patchValue({ namespace: params['namespace'] });
        this.onNamespaceChange(params['namespace']);
      }
      if (params['xstore']) {
        this.rebuildForm.patchValue({ xStoreName: params['xstore'] });
        this.onXStoreChange(params['xstore']);
      }
      if (params['targetPod']) {
        this.rebuildForm.patchValue({ targetPodName: params['targetPod'] });
      }
      if (params['fromPod']) {
        this.rebuildForm.patchValue({ fromPodName: params['fromPod'] });
      }
    });
  }

  private async loadNamespaces(): Promise<void> {
    try {
      // TODO: 实现获取命名空间的 API
      const namespaces = ['default', 'polardbx-operator-system'];
      this.namespaces = namespaces || ['default'];
    } catch (error) {
      console.error('Failed to load namespaces:', error);
      this.namespaces = ['default'];
    }
  }

  async onNamespaceChange(namespace: string): Promise<void> {
    if (!namespace) return;
    
    this.isLoadingXStores = true;
    try {
      const xstores = await this.apiService.getXStores(namespace).toPromise();
      this.availableXStores = xstores || [];
    } catch (error) {
      console.error('Failed to load XStores:', error);
      this.availableXStores = [];
    } finally {
      this.isLoadingXStores = false;
    }
  }

  async onXStoreChange(xstoreName: string): Promise<void> {
    if (!xstoreName) return;
    
    const namespace = this.rebuildForm.get('namespace')?.value;
    if (!namespace) return;

    this.isLoadingPods = true;
    try {
      const pods = await this.apiService.getXStorePods(namespace, xstoreName).toPromise();
      this.targetPods = pods || [];
      this.filterPods();
    } catch (error) {
      console.error('Failed to load XStore pods:', error);
      this.targetPods = [];
      this.filteredTargetPods = [];
      this.filteredSourcePods = [];
    } finally {
      this.isLoadingPods = false;
    }
  }

  onRoleChange(role: string): void {
    this.updateFormValidators(role);
    this.filterPods();
  }

  private updateFormValidators(role: string): void {
    const targetPodControl = this.rebuildForm.get('targetPodName');
    
    if (role === 'learner') {
      // Learner requires targetPodName
      targetPodControl?.setValidators([Validators.required]);
      this.rebuildForm.patchValue({ local: true });
    } else {
      // Logger and follower don't require targetPodName
      targetPodControl?.clearValidators();
    }
    
    targetPodControl?.updateValueAndValidity();
  }

  private filterPods(): void {
    const role = this.rebuildForm.get('role')?.value;
    
    // Filter target pods (prefer running and non-leader)
    this.filteredTargetPods = this.targetPods.filter(pod => {
      const phase = pod.status?.phase;
      const podRole = this.getPodRole(pod);
      return phase === 'Running' && podRole !== 'leader';
    });
    
    // If no suitable target pods, show all running pods
    if (this.filteredTargetPods.length === 0) {
      this.filteredTargetPods = this.targetPods.filter(pod => pod.status?.phase === 'Running');
    }
    
    // Filter source pods (prefer running pods for logger rebuild)
    this.filteredSourcePods = this.targetPods.filter(pod => pod.status?.phase === 'Running');
    
    // Auto-select single candidate for learner
    if (role === 'learner' && this.filteredTargetPods.length === 1 && !this.rebuildForm.get('targetPodName')?.value) {
      this.rebuildForm.patchValue({ targetPodName: this.filteredTargetPods[0].metadata.name });
    }
  }

  shouldShowTargetPod(): boolean {
    const role = this.rebuildForm.get('role')?.value;
    return role === 'learner' || role === 'logger';
  }

  shouldShowFromPod(): boolean {
    return this.rebuildForm.get('role')?.value === 'logger';
  }

  shouldShowNodeName(): boolean {
    return !this.rebuildForm.get('local')?.value;
  }

  isTargetPodRequired(): boolean {
    return this.rebuildForm.get('role')?.value === 'learner';
  }

  getTargetPodErrorTip(): string {
    const role = this.rebuildForm.get('role')?.value;
    return role === 'learner' ? '请选择目标 Pod' : '';
  }

  getRoleHints(): Array<{type: 'success' | 'info' | 'warning' | 'error', message: string}> {
    const role = this.rebuildForm.get('role')?.value;
    const local = this.rebuildForm.get('local')?.value;
    const hints = [];

    switch (role) {
      case 'learner':
        hints.push({
          type: 'info' as const,
          message: 'Learner 重搭通常在本机进行，需要选择具体的目标 Pod。建议选择非 Leader 且处于 Running 状态的 Pod。'
        });
        break;
      case 'logger':
        hints.push({
          type: 'warning' as const,
          message: 'Logger 重搭如果不指定源 Pod，可能会触发较重的数据传输。建议选择就近的健康源 Pod。'
        });
        break;
      case 'follower':
        hints.push({
          type: 'info' as const,
          message: 'Follower 重搭可以跨机进行，系统将自动选择合适的配置。'
        });
        break;
    }

    if (!local) {
      hints.push({
        type: 'info' as const,
        message: '跨机重搭将在其他节点创建新的 Pod，可以指定目标节点名称。'
      });
    }

    return hints;
  }

  getPodRole(pod: any): string {
    return pod.metadata?.labels?.['xstore/role'] || '';
  }

  getPodStatusColor(pod: any): string {
    const phase = pod.status?.phase;
    switch (phase) {
      case 'Running': return 'green';
      case 'Pending': return 'blue';
      case 'Failed': return 'red';
      default: return 'default';
    }
  }

  getRoleColor(role: string): string {
    switch (role) {
      case 'leader': return 'gold';
      case 'follower': return 'blue';
      case 'logger': return 'cyan';
      default: return 'default';
    }
  }

  getXStoreStatusColor(xstore: XStore): string {
    const phase = (xstore.status as any)?.phase;
    switch (phase) {
      case 'Running': return 'green';
      case 'Pending': return 'blue';
      case 'Failed': return 'red';
      default: return 'default';
    }
  }

  getXStoreStatusText(xstore: XStore): string {
    return (xstore.status as any)?.phase || '未知';
  }

  async onSubmit(): Promise<void> {
    if (!this.rebuildForm.valid) {
      Object.values(this.rebuildForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
      return;
    }

    this.isSubmitting = true;
    const formValue = this.rebuildForm.value as RebuildFormData;
    
    try {
      const request: CreateXStoreFollowerRequest = {
        xStoreName: formValue.xStoreName,
        role: formValue.role,
        local: formValue.local,
        targetPodName: formValue.targetPodName,
        fromPodName: formValue.fromPodName,
        nodeName: formValue.nodeName
      };

      const result = await this.apiService.createXStoreFollower(formValue.namespace, request).toPromise();
      
      this.message.success('重搭任务创建成功');
      
      // Navigate to task detail
      if (result?.metadata?.name) {
        this.router.navigate(['/storage/xstore-rebuild/rebuild/tasks', formValue.namespace, result.metadata.name]);
      } else {
        this.router.navigate(['/storage/xstore-rebuild/rebuild/tasks']);
      }
      
    } catch (error) {
      console.error('Failed to create rebuild task:', error);
      this.message.error('创建重搭任务失败: ' + (error as any)?.message || '未知错误');
    } finally {
      this.isSubmitting = false;
    }
  }

  onCancel(): void {
    this.goToTaskList();
  }

  goToTaskList(): void {
    this.router.navigate(['/storage/xstore-rebuild/rebuild/tasks']);
  }
}
