import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { Observable, Subscription } from 'rxjs';
import { LoadingService, LoadingKeys } from '../../services/loading.service';

@Component({
  selector: 'app-loading-indicator',
  standalone: true,
  imports: [CommonModule, MatProgressSpinnerModule],
  template: `
    <div class="loading-overlay" *ngIf="isLoading$ | async">
      <div class="loading-container">
        <mat-spinner diameter="40"></mat-spinner>
        <div class="loading-text">{{ loadingText }}</div>
      </div>
    </div>
  `,
  styles: [`
    .loading-overlay {
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: rgba(255, 255, 255, 0.7);
      display: flex;
      justify-content: center;
      align-items: center;
      z-index: 10000;
      backdrop-filter: blur(4px);
    }

    .loading-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 20px;
    }

    .loading-text {
      color: var(--primary-color);
      font-size: 16px;
      font-weight: 500;
      text-align: center;
    }
  `]
})
export class LoadingIndicatorComponent implements OnInit, OnDestroy {
  private loadingService = inject(LoadingService);

  isLoading$: Observable<boolean>;
  loadingText = '加载中...';
  private subscription?: Subscription;

  constructor() {
    this.isLoading$ = this.loadingService.getGlobalLoadingState();
  }

  ngOnInit(): void {
    this.subscription = this.loadingService.activeLoadingKeys$.subscribe(keys => {
      this.updateLoadingText(keys);
    });
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }
  
  private updateLoadingText(keys: string[]): void {
    if (keys.length === 0) {
      this.loadingText = '加载中...';
      return;
    }
    
    // 根据活动的加载键设置加载文本
    if (keys.includes(LoadingKeys.CONNECT)) {
      this.loadingText = '正在连接服务器...';
    } else if (keys.includes(LoadingKeys.CLUSTERS_LIST)) {
      this.loadingText = '正在加载集群列表...';
    } else if (keys.includes(LoadingKeys.CLUSTER_DETAIL)) {
      this.loadingText = '正在加载集群详情...';
    } else if (keys.includes(LoadingKeys.PODS_LIST)) {
      this.loadingText = '正在加载 Pod 列表...';
    } else if (keys.includes(LoadingKeys.POD_LOGS)) {
      this.loadingText = '正在加载 Pod 日志...';
    } else if (keys.includes(LoadingKeys.BACKUPS_LIST)) {
      this.loadingText = '正在加载备份列表...';
    } else if (keys.includes(LoadingKeys.CLUSTER_CREATE)) {
      this.loadingText = '正在创建集群...';
    } else if (keys.includes(LoadingKeys.CLUSTER_UPDATE)) {
      this.loadingText = '正在更新集群...';
    } else if (keys.includes(LoadingKeys.CLUSTER_DELETE)) {
      this.loadingText = '正在删除集群...';
    } else {
      this.loadingText = '处理中...';
    }
  }
}