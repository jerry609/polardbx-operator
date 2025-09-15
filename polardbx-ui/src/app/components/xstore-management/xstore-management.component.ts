import { Component, OnInit, inject } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { ApiService } from '../../services/api.service';
import { LoadingService, LoadingKeys } from '../../services/loading.service';
import { XStore } from '../../models/xstore.model';
import { EmptyStateComponent } from '../empty-state/empty-state.component';

@Component({
  selector: 'app-xstore-management',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatTabsModule,
    MatProgressBarModule,
    MatChipsModule,
    MatSnackBarModule,
    MatFormFieldModule,
    MatInputModule,
    ReactiveFormsModule,
    EmptyStateComponent
  ],
  template: `
    <div class="xstore-management">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>dns</mat-icon>
            存储节点管理
          </mat-card-title>
          <mat-card-subtitle>管理XStore存储节点和拓扑</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <mat-tab-group [(selectedIndex)]="selectedTab">
            <mat-tab label="存储节点列表">
              <div class="tab-content">
                <div class="actions-toolbar">
                  <button mat-raised-button color="primary" (click)="refreshXStores()">
                    <mat-icon>refresh</mat-icon>
                    刷新
                  </button>
                  <button mat-raised-button color="accent" (click)="createNew()">
                    <mat-icon>add</mat-icon>
                    新建节点
                  </button>
                </div>
                <mat-progress-bar *ngIf="loadingService.isLoading(loadingKeys.XSTORE_LIST)" mode="indeterminate"></mat-progress-bar>
                
                <app-empty-state *ngIf="xstores.length === 0 && !loadingService.isLoading(loadingKeys.XSTORE_LIST)"
                                  icon="dns"
                                  title="暂无存储节点"
                                  hint="点击“新建节点”创建您的第一个存储节点"></app-empty-state>
                
                <mat-table *ngIf="xstores.length > 0" [dataSource]="xstores" class="xstore-table">
                  <ng-container matColumnDef="name">
                    <mat-header-cell *matHeaderCellDef>节点名称</mat-header-cell>
                    <mat-cell *matCellDef="let xstore">{{ xstore.metadata.name }}</mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="namespace">
                    <mat-header-cell *matHeaderCellDef>命名空间</mat-header-cell>
                    <mat-cell *matCellDef="let xstore">{{ xstore.metadata.namespace }}</mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="status">
                    <mat-header-cell *matHeaderCellDef>状态</mat-header-cell>
                    <mat-cell *matCellDef="let xstore">
                      <mat-chip [color]="getStatusColor(xstore.status?.phase)">
                        {{ xstore.status?.phase || '未知' }}
                      </mat-chip>
                    </mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="replicas">
                    <mat-header-cell *matHeaderCellDef>副本数</mat-header-cell>
                    <mat-cell *matCellDef="let xstore">{{ getReplicaDisplay(xstore) }}</mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="createdTime">
                    <mat-header-cell *matHeaderCellDef>创建时间</mat-header-cell>
                    <mat-cell *matCellDef="let xstore">{{ xstore.metadata.creationTimestamp | date:'medium' }}</mat-cell>
                  </ng-container>

                  <ng-container matColumnDef="actions">
                    <mat-header-cell *matHeaderCellDef>操作</mat-header-cell>
                    <mat-cell *matCellDef="let xstore">
                      <button mat-icon-button color="primary" (click)="viewXStoreDetails(xstore)" title="查看详情">
                        <mat-icon>visibility</mat-icon>
                      </button>
                      <button mat-icon-button color="warn" (click)="deleteXStore(xstore)" title="删除节点">
                        <mat-icon>delete</mat-icon>
                      </button>
                    </mat-cell>
                  </ng-container>

                  <mat-header-row *matHeaderRowDef="['name', 'namespace', 'status', 'replicas', 'createdTime', 'actions']"></mat-header-row>
                  <mat-row *matRowDef="let row; columns: ['name', 'namespace', 'status', 'replicas', 'createdTime', 'actions']"></mat-row>
                </mat-table>
              </div>
            </mat-tab>
            <mat-tab label="创建存储节点">
              <div class="form-container">
                <mat-card class="create-card">
                  <mat-card-header>
                    <mat-card-title>创建存储节点</mat-card-title>
                  </mat-card-header>
                  <mat-card-content>
                    <form [formGroup]="createForm" (ngSubmit)="submit()" class="form-grid">
                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>名称</mat-label>
                        <input matInput formControlName="name" placeholder="xstore-name">
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>命名空间</mat-label>
                        <input matInput formControlName="namespace" placeholder="default">
                      </mat-form-field>

                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>引擎</mat-label>
                        <input matInput formControlName="engine" placeholder="galaxy">
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>节点数</mat-label>
                        <input matInput type="number" min="1" formControlName="nodeCount" placeholder="2">
                      </mat-form-field>

                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>CPU (limits)</mat-label>
                        <input matInput formControlName="cpu" placeholder="2">
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>内存 (limits)</mat-label>
                        <input matInput formControlName="memory" placeholder="4Gi">
                      </mat-form-field>

                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>数据卷大小</mat-label>
                        <input matInput formControlName="diskQuota" placeholder="100Gi">
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>存储类</mat-label>
                        <input matInput formControlName="storageClass" placeholder="(可选)">
                      </mat-form-field>

                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>CN 副本数</mat-label>
                        <input type="number" min="0" matInput formControlName="cnReplicas" placeholder="0">
                      </mat-form-field>
                      <mat-form-field appearance="outline" class="grid-item">
                        <mat-label>服务类型</mat-label>
                        <input matInput formControlName="serviceType" placeholder="NodePort">
                      </mat-form-field>

                      <div class="form-actions grid-full">
                        <button mat-raised-button color="primary" type="submit" [disabled]="createForm.invalid">创建</button>
                        <button mat-button type="button" (click)="selectedTab = 0">返回列表</button>
                      </div>
                    </form>
                  </mat-card-content>
                </mat-card>
              </div>
            </mat-tab>
          </mat-tab-group>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .xstore-management { padding: 20px; }
    .tab-content { padding: 20px; }
    .actions-toolbar { margin-bottom: 16px; display: flex; gap: 8px; }
    .empty-state { text-align: center; padding: 40px; color: #666; }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; color: #ccc; }
    .empty-state .hint { font-size: 0.9em; opacity: 0.7; }
    .form-container { padding: 20px; }
    mat-card-title { display: flex; align-items: center; gap: 8px; }

    .form-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: 16px 24px;
    }
    .grid-item { width: 100%; }
    .grid-full { grid-column: 1 / -1; display: flex; gap: 8px; justify-content: flex-end; margin-top: 8px; }

    @media (min-width: 1280px) {
      .form-grid { gap: 20px 32px; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); }
    }
  `]
})
export class XStoreManagementComponent implements OnInit {
  selectedTab = 0;
  loadingKeys = LoadingKeys;
  xstores: XStore[] = [];
  createForm!: FormGroup;
  
  private snackBar = inject(MatSnackBar);
  public loadingService = inject(LoadingService);
  private apiService = inject(ApiService);
  private fb = inject(FormBuilder);

  ngOnInit(): void {
    this.createForm = this.fb.group({
      name: ['', [Validators.required, Validators.pattern(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/)]],
      namespace: ['default', Validators.required],
      engine: ['galaxy'],
      nodeCount: [2, [Validators.required, Validators.min(1)]],
      cpu: ['2'],
      memory: ['4Gi'],
      diskQuota: ['100Gi'],
      storageClass: [''],
      cnReplicas: [0, [Validators.min(0)]],
      serviceType: ['NodePort']
    });
    this.loadXStores();
  }

  loadXStores(): void {
    // 使用全局选择的命名空间（ApiService 内部会从 localStorage.activeNamespace 读取）
    this.apiService.getXStores().subscribe({
      next: (xstores) => {
        this.xstores = xstores;
      },
      error: (err) => {
        this.snackBar.open(`加载存储节点失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
        this.xstores = [];
      }
    });
  }

  refreshXStores(): void {
    this.loadXStores();
    this.snackBar.open('存储节点列表已刷新', '关闭', { duration: 2000 });
  }

  createNew(): void {
    this.selectedTab = 1;
  }

  viewXStoreDetails(xstore: XStore): void {
    this.snackBar.open(`查看存储节点详情: ${xstore.metadata.name}`, '关闭', { duration: 3000 });
  }

  deleteXStore(xstore: XStore): void {
    if (confirm(`确定删除存储节点 "${xstore.metadata.name}" 吗？`)) {
      this.apiService.deleteXStore(xstore.metadata.namespace || 'default', xstore.metadata.name).subscribe({
        next: () => {
          this.snackBar.open('删除成功!', '关闭', { duration: 3000 });
          this.loadXStores();
        },
        error: (err) => {
          this.snackBar.open(`删除失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
        }
      });
    }
  }

  submit(): void {
    if (this.createForm.invalid) return;
    const v = this.createForm.value;
    const req = {
      name: v.name,
      namespace: v.namespace,
      engine: v.engine,
      nodeCount: Number(v.nodeCount),
      resources: { limits: { cpu: v.cpu, memory: v.memory } },
      storage: { size: v.diskQuota, storageClass: v.storageClass },
      cnReplicas: Number(v.cnReplicas),
      serviceType: v.serviceType
    } as any;
    this.apiService.createXStore(v.namespace, req).subscribe({
      next: () => {
        this.snackBar.open('创建成功！', '关闭', { duration: 3000 });
        this.selectedTab = 0;
        this.loadXStores();
      },
      error: (err) => {
        this.snackBar.open(`创建失败: ${err.error?.message || err.message}`, '关闭', { duration: 5000 });
      }
    });
  }

  getStatusColor(status?: string): string {
    switch (status?.toLowerCase()) {
      case 'running':
        return 'primary';
      case 'ready':
        return 'accent';
      case 'failed':
        return 'warn';
      default:
        return '';
    }
  }

  /**
   * 获取XStore副本数显示文本
   * 优先显示 ready/total，如果没有状态信息则显示规格中的节点数
   */
  getReplicaDisplay(xstore: XStore): string {
    // 优先使用运行时状态信息
    if (xstore.status?.replicaStatus) {
      const ready = xstore.status.replicaStatus.ready ?? 0;
      const total = xstore.status.replicaStatus.total ?? 0;
      return `${ready}/${total}`;
    }
    
    // 如果没有状态信息，使用规格中的节点总数
    if (xstore.spec?.topology?.nodeCount) {
      return `${xstore.spec.topology.nodeCount}`;
    }
    
    // 如果有NodeSets配置，计算总副本数
    if (xstore.spec?.topology?.nodeSets && xstore.spec.topology.nodeSets.length > 0) {
      const totalReplicas = xstore.spec.topology.nodeSets.reduce((sum, nodeSet) => {
        return sum + (nodeSet.replicas || 0);
      }, 0);
      return `${totalReplicas}`;
    }
    
    // 都没有则显示未知
    return 'N/A';
  }
}