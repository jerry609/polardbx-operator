import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSortModule } from '@angular/material/sort';
import { MatTabsModule } from '@angular/material/tabs';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatSnackBarModule, MatSnackBar } from '@angular/material/snack-bar';
import { ApiService } from '../../services/api.service';
import { LoadingService } from '../../services/loading.service';
import { XStoreFollower, XStoreFollowerWithStatus, CreateXStoreFollowerRequest } from '../../models/xstore-follower.model';
import { XStore } from '../../models/xstore.model';
import { MatTableDataSource } from '@angular/material/table';
import { Observable } from 'rxjs';
import { Pod } from '../../models/pod.model';

export interface XStoreFollowerDialogData {
  mode: 'create' | 'edit' | 'view';
  follower?: XStoreFollower;
  namespace?: string;
}

@Component({
  selector: 'app-xstore-follower-management',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatCardModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatTabsModule,
    MatChipsModule,
    MatTooltipModule,
    MatProgressBarModule,
    MatCheckboxModule,
    MatSnackBarModule
  ],
  templateUrl: './xstore-follower-management.component.html',
  styleUrls: ['./xstore-follower-management.component.scss', '../../styles/shared-management.scss']
})
export class XStoreFollowerManagementComponent implements OnInit {
  private fb = inject(FormBuilder);
  private apiService = inject(ApiService);
  private loadingService = inject(LoadingService);
  private snackBar = inject(MatSnackBar);
  
  // Component state
  isLoading = false;
  isProcessing = false;
  selectedTab = 0;
  dialogRef = inject<MatDialogRef<XStoreFollowerManagementComponent>>(MatDialogRef, { optional: true });
  data = inject<XStoreFollowerDialogData>(MAT_DIALOG_DATA, { optional: true }) || {} as XStoreFollowerDialogData;

  // Forms
  followerForm!: FormGroup;
  resourceForm!: FormGroup;

  // Data
  availableXStores: XStore[] = [];
  validSourceXStores: XStore[] = []; // Filtered XStores suitable as backup source
  targetPods: Pod[] = [];
  filteredTargetPods: Pod[] = [];
  followers: XStoreFollowerWithStatus[] = [];
  dataSource = new MatTableDataSource<XStoreFollowerWithStatus>();
  

  
  // Table configuration
  displayedColumns: string[] = [
    'name', 
    'xStoreName', 
    'phase', 
    'progress', 
    'lastActivity', 
    'status',
    'actions'
  ];

  // CPU and Memory options
  cpuOptions = ['0.5', '1', '2', '4', '8'];
  memoryOptions = ['1Gi', '2Gi', '4Gi', '8Gi', '16Gi'];
  storageOptions = ['10Gi', '20Gi', '50Gi', '100Gi', '200Gi'];

  constructor() {
    this.initializeForms();
  }

  ngOnInit(): void {
    this.loadInitialData();
    if (this.data.mode === 'edit' && this.data.follower) {
      this.populateFormFromFollower(this.data.follower);
    }
  }

  private initializeForms(): void {
    // Main follower configuration form
    this.followerForm = this.fb.group({
      name: ['', [
        Validators.required, 
        Validators.pattern(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/)
      ]],
      namespace: [this.data.namespace || 'default', Validators.required],
      xStoreName: ['', Validators.required],
      targetPodName: [''],
      fromXStore: [''],
      fromBackupSet: [''],
      forceRecreate: [false],
      priority: [1, [Validators.min(1), Validators.max(10)]]
    });

    // Resource configuration form
    this.resourceForm = this.fb.group({
      enableResourceLimits: [false],
      requestsCpu: ['1'],
      requestsMemory: ['2Gi'],
      requestsStorage: ['10Gi'],
      limitsCpu: ['2'],
      limitsMemory: ['4Gi'],
      limitsStorage: ['20Gi'],
      nodeSelector: this.fb.group({
        enabled: [false],
        key: [''],
        value: ['']
      })
    });

    // Watch for form changes
    this.resourceForm.get('enableResourceLimits')?.valueChanges.subscribe(enabled => {
      this.updateResourceValidators(enabled);
    });

    // Load target pods when xStoreName changes
    this.followerForm.get('xStoreName')?.valueChanges.subscribe((x: string) => {
      if (x) {
        this.loadTargetPods(x);
      } else {
        this.targetPods = [];
        this.filteredTargetPods = [];
        this.followerForm.patchValue({ targetPodName: '' });
      }
    });
  }

  private updateResourceValidators(enabled: boolean): void {
    const resourceFields = [
      'requestsCpu', 'requestsMemory', 'requestsStorage',
      'limitsCpu', 'limitsMemory', 'limitsStorage'
    ];

    resourceFields.forEach(field => {
      const control = this.resourceForm.get(field);
      if (enabled) {
        control?.setValidators([Validators.required]);
      } else {
        control?.clearValidators();
      }
      control?.updateValueAndValidity();
    });
  }

  private async loadInitialData(): Promise<void> {
    try {
      // Load available XStores
      this.availableXStores = await this.apiService.getXStores(this.data.namespace).toPromise() || [];
      
      // Filter valid source XStores
      await this.filterValidSourceXStores();
      
      // Load existing followers
      await this.loadFollowers();

      // If form already has xStoreName (edit mode), load pods
      const x = this.followerForm.get('xStoreName')?.value;
      if (x) {
        await this.loadTargetPods(x);
      }
    } catch (error) {
      console.error('Failed to load initial data:', error);
    }
  }

  private async loadTargetPods(xstoreName: string): Promise<void> {
    try {
      const ns = this.followerForm.get('namespace')?.value || this.data.namespace || 'default';
      const pods = await this.apiService.getXStorePods(ns, xstoreName).toPromise();
      this.targetPods = pods || [];
      // Prefer running and non-leader
      this.filteredTargetPods = (this.targetPods || []).filter(p => {
        const phase = (p as any)?.status?.phase;
        const role = (p as any)?.metadata?.labels?.['xstore/role'];
        return phase === 'Running' && role !== 'leader';
      });
      if (this.filteredTargetPods.length === 0) {
        this.filteredTargetPods = (this.targetPods || []).filter(p => (p as any)?.status?.phase === 'Running');
      }
      // Auto-select single candidate
      if (this.filteredTargetPods.length === 1 && !this.followerForm.get('targetPodName')?.value) {
        this.followerForm.patchValue({ targetPodName: (this.filteredTargetPods[0] as any)?.metadata?.name });
      }
    } catch (err) {
      console.error('Failed to load target pods for xstore:', xstoreName, err);
      this.targetPods = [];
      this.filteredTargetPods = [];
    }
  }

  private async loadFollowers(): Promise<void> {
    try {
      const followers = await this.apiService.getXStoreFollowers(this.data.namespace).toPromise() || [];
      this.followers = followers.map((follower: XStoreFollower) => this.enrichFollowerWithStatus(follower));
      this.dataSource.data = this.followers;
    } catch (error) {
      console.error('Failed to load XStore followers:', error);
    }
  }

  private async filterValidSourceXStores(): Promise<void> {
    this.validSourceXStores = [];
    
    console.log(`Starting XStore filtering. Total XStores: ${this.availableXStores.length}`, this.availableXStores);
    
    for (const xstore of this.availableXStores) {
      const isValid = await this.isValidSourceXStore(xstore);
      console.log(`XStore ${xstore.metadata.name}: ${isValid ? 'VALID' : 'INVALID'} (phase: ${xstore.status?.phase})`);
      
      if (isValid) {
        this.validSourceXStores.push(xstore);
      }
    }
    
    console.log(`✅ Filtered ${this.validSourceXStores.length} valid source XStores:`, this.validSourceXStores.map(x => x.metadata.name));
    
    // Fallback: if no valid sources found, include all Running XStores
    if (this.validSourceXStores.length === 0) {
      console.warn('⚠️ No valid sources found, falling back to all Running XStores');
      this.validSourceXStores = this.availableXStores.filter(x => x.status?.phase === 'Running');
      console.log(`Fallback: Added ${this.validSourceXStores.length} Running XStores:`, this.validSourceXStores.map(x => x.metadata.name));
    }
    
    // Auto-select first valid source if available
    if (this.validSourceXStores.length > 0 && this.data.mode === 'create') {
      const autoSelected = this.validSourceXStores[0].metadata.name;
      
      // 🎯 Set auto-selected value with explicit control update
      this.followerForm.patchValue({ fromXStore: autoSelected });
      
      // 🔍 Force update and verify
      this.followerForm.get('fromXStore')?.setValue(autoSelected);
      this.followerForm.get('fromXStore')?.markAsDirty();
      this.followerForm.get('fromXStore')?.updateValueAndValidity();
      
      console.log(`🎯 Auto-selected source XStore: ${autoSelected}`);
      console.log(`✅ Form control value after auto-selection:`, this.followerForm.get('fromXStore')?.value);
    }
  }

  private async isValidSourceXStore(xstore: XStore): Promise<boolean> {
    try {
      // 1. Check XStore phase - must be Running
      if (xstore.status?.phase !== 'Running') {
        console.log(`XStore ${xstore.metadata.name} is not Running (phase: ${xstore.status?.phase})`);
        return false;
      }

      // 2. Check replica status if available (safe access)
      const replicaStatus = (xstore.status as any)?.replicaStatus;
      if (replicaStatus) {
        const readyReplicas = replicaStatus.ready || replicaStatus.available || 0;
        const totalReplicas = replicaStatus.total || 0;
        
        if (readyReplicas === 0 || totalReplicas === 0) {
          console.log(`XStore ${xstore.metadata.name} has no ready replicas (${readyReplicas}/${totalReplicas})`);
          return false;
        }
        
        console.log(`XStore ${xstore.metadata.name} appears valid (Running, ${readyReplicas}/${totalReplicas} ready)`);
        return true;
      }

      // 3. Fallback: if we can't get detailed status, assume valid if Running
      // This is safe for basic validation
      console.log(`XStore ${xstore.metadata.name} appears valid (Running with replicas)`);
      return true;
      
    } catch (error) {
      console.error(`Error checking XStore ${xstore.metadata.name}:`, error);
      return false;
    }
  }

  private enrichFollowerWithStatus(follower: XStoreFollower): XStoreFollowerWithStatus {
    const status = follower.status;
    const phase = status?.phase || '';
    
    // Map actual phase values to recovery states
    const isRecovering = ['FollowerPhaseCheck', 'FollowerPhaseBackupPrepare', 'FollowerPhaseBackupStart', 
                         'FollowerPhaseBackup', 'FollowerPhaseLoggerRebuild', 'FollowerPhaseMonitorBackup',
                         'FollowerPhaseBeforeRestore', 'FollowerPhaseRestore', 'FollowerPhaseAfterRestore',
                         'FollowerPhaseLoggerCreate', 'FollowerCreateRemotePod'].includes(phase);
    const isHealthy = phase === 'FollowerPhaseSuccess';
    const hasFailed = phase === 'FollowerPhaseFailed';
    
    return {
      ...follower,
      isRecovering,
      isHealthy,
      hasFailures: hasFailed,
      displayStatus: this.getDisplayStatus(phase),
      lastActivity: follower.metadata.creationTimestamp // Use creation time as last activity
    };
  }

  private getDisplayStatus(phase?: string): string {
    if (!phase) return '已创建';
    
    const statusMap: { [key: string]: string } = {
      '': '已创建',
      'FollowerPhaseCheck': '检查中',
      'FollowerPhaseBackupPrepare': '备份准备',
      'FollowerPhaseBackupStart': '开始备份',
      'FollowerPhaseBackup': '备份中',
      'FollowerPhaseLoggerRebuild': '重建日志',
      'FollowerPhaseMonitorBackup': '监控备份',
      'FollowerPhaseBeforeRestore': '准备恢复',
      'FollowerPhaseRestore': '恢复中',
      'FollowerPhaseAfterRestore': '完成恢复',
      'FollowerPhaseSuccess': '成功',
      'FollowerPhaseWaitSwitch': '等待切换',
      'FollowerPhaseFailed': '失败',
      'FollowerPhaseLoggerCreate': '创建日志器',
      'FollowerCreateRemotePod': '创建远程Pod',
      'FollowerPhaseDeleting': '删除中'
    };

    return statusMap[phase] || phase;
  }

  // Tooltip for phase chip
  getPhaseTooltip(follower: XStoreFollowerWithStatus): string {
    const s = follower.status as any;
    const phase = s?.phase || 'Unknown';
    const task = s?.currentJobTask ? `，任务：${s.currentJobTask}` : '';
    return `阶段：${this.getDisplayStatus(phase)}（${phase}）${task}`;
  }

  getStatusChipClass(phase?: string): string {
    switch (phase) {
      case 'FollowerPhaseReady':
      case 'FollowerPhaseRestored':
        return 'success';
      case 'FollowerPhaseFailed':
        return 'error';
      case 'FollowerPhaseRestore':
      case 'FollowerPhaseCheck':
      case 'FollowerPhaseMonitorBackup':
        return 'info';
      default:
        return 'pending';
    }
  }

  resetForm(): void {
    this.followerForm.reset();
    this.initializeForms();
  }

  save(): void {
    if (this.followerForm.valid) {
      this.isLoading = true;
      // 实现保存逻辑
      const formValue = this.followerForm.value;
      const request: CreateXStoreFollowerRequest = {
        name: formValue.name,
        xStoreName: formValue.xStoreName,
        targetPodName: formValue.targetPodName,
        fromXStore: formValue.fromXStore
      };

      this.apiService.createXStoreFollower('default', request).subscribe({
        next: (result) => {
          this.snackBar.open('XStore Follower 创建成功', '关闭', { duration: 3000 });
          this.isLoading = false;
          this.refreshFollowers();
          this.selectedTab = 0; // 切换到列表页面
        },
        error: (error) => {
          this.snackBar.open('创建失败: ' + error.message, '关闭', { duration: 5000 });
          this.isLoading = false;
        }
      });
    }
  }

  private populateFormFromFollower(follower: XStoreFollower): void {
    this.followerForm.patchValue({
      name: follower.metadata.name,
      namespace: follower.metadata.namespace,
      xStoreName: follower.spec.xStoreName,
      targetPodName: (follower.spec as any)?.targetPodName || '',
      fromXStore: follower.spec.fromXStore,
      fromBackupSet: follower.spec.fromBackupSet,
      forceRecreate: follower.spec.forceRecreate,
      priority: follower.spec.priority || 1
    });

    // Populate resource form if resources are defined
    const resources = follower.spec.resources;
    if (resources) {
      this.resourceForm.patchValue({
        enableResourceLimits: true,
        requestsCpu: resources.requests?.cpu || '1',
        requestsMemory: resources.requests?.memory || '2Gi',
        requestsStorage: resources.requests?.storage || '10Gi',
        limitsCpu: resources.limits?.cpu || '2',
        limitsMemory: resources.limits?.memory || '4Gi',
        limitsStorage: resources.limits?.storage || '20Gi'
      });
    }

    // Populate node selector if defined
    const nodeSelector = follower.spec.nodeSelector;
    if (nodeSelector && Object.keys(nodeSelector).length > 0) {
      const firstKey = Object.keys(nodeSelector)[0];
      this.resourceForm.patchValue({
        nodeSelector: {
          enabled: true,
          key: firstKey,
          value: nodeSelector[firstKey]
        }
      });
    }
  }

  // Create or update XStore follower
  async saveFollower(): Promise<void> {
    if (!this.followerForm.valid) {
      return;
    }

    this.isProcessing = true;

    try {
      const formValue = this.followerForm.value;
      const resourceValue = this.resourceForm.value;

      // 🔍 Debug: Log current form values
      console.log('📝 Form values before submission:', formValue);
      console.log('📝 fromXStore value:', formValue.fromXStore);
      console.log('📝 Form control value:', this.followerForm.get('fromXStore')?.value);

      const request: CreateXStoreFollowerRequest = {
        name: formValue.name, // Add user-provided name
        xStoreName: formValue.xStoreName,
        targetPodName: formValue.targetPodName && formValue.targetPodName.trim() !== '' ? formValue.targetPodName : undefined,
        fromXStore: formValue.fromXStore && formValue.fromXStore.trim() !== '' ? formValue.fromXStore : undefined,
        fromBackupSet: formValue.fromBackupSet && formValue.fromBackupSet.trim() !== '' ? formValue.fromBackupSet : undefined,
        forceRecreate: formValue.forceRecreate,
        priority: formValue.priority,
        resources: resourceValue.enableResourceLimits ? {
          requests: {
            cpu: resourceValue.requestsCpu,
            memory: resourceValue.requestsMemory,
            storage: resourceValue.requestsStorage
          },
          limits: {
            cpu: resourceValue.limitsCpu,
            memory: resourceValue.limitsMemory,
            storage: resourceValue.limitsStorage
          }
        } : undefined,
        nodeSelector: resourceValue.nodeSelector.enabled ? {
          [resourceValue.nodeSelector.key]: resourceValue.nodeSelector.value
        } : undefined
      };

      // 🔍 Debug: Log final request
      console.log('📤 Final request to backend:', request);

      if (this.data.mode === 'create') {
        await this.apiService.createXStoreFollower(formValue.namespace, request).toPromise();
      } else if (this.data.mode === 'edit' && this.data.follower) {
        // For edit mode, we need to update the existing follower
        const updatedFollower: XStoreFollower = {
          ...this.data.follower,
          spec: {
            ...this.data.follower.spec,
            xStoreName: request.xStoreName,
            fromXStore: request.fromXStore,
            fromBackupSet: request.fromBackupSet,
            forceRecreate: request.forceRecreate,
            priority: request.priority,
            resources: request.resources,
            nodeSelector: request.nodeSelector
          }
        };
        
        await this.apiService.updateXStoreFollower(formValue.namespace, updatedFollower).toPromise();
      }

      // Show success message
      const action = this.data.mode === 'create' ? '创建' : '更新';
      this.snackBar.open(`XStore Follower ${action}成功！`, '关闭', {
        duration: 3000,
        horizontalPosition: 'center',
        verticalPosition: 'top'
      });

      // Reload the followers list if we're in management mode
      if (this.selectedTab === 0) {
        await this.loadFollowers();
      }

      this.dialogRef?.close(true);
    } catch (error) {
      console.error('Failed to save XStore follower:', error);
    } finally {
      this.isProcessing = false;
    }
  }

  // Delete XStore follower
  async deleteFollower(follower: XStoreFollowerWithStatus): Promise<void> {
    if (!confirm(`确定要删除 XStore Follower "${follower.metadata.name}" 吗？此操作不可撤销。`)) {
      return;
    }

    try {
      await this.apiService.deleteXStoreFollower(
        follower.metadata.namespace, 
        follower.metadata.name
      ).toPromise();
      
      await this.loadFollowers();
    } catch (error) {
      console.error('Failed to delete XStore follower:', error);
    }
  }

  // Edit existing follower
  editFollower(follower: XStoreFollowerWithStatus): void {
    this.data.mode = 'edit';
    this.data.follower = follower;
    this.populateFormFromFollower(follower);
    this.selectedTab = 1; // Switch to form tab
  }

  // View follower details
  viewFollower(follower: XStoreFollowerWithStatus): void {
    // Extract source XStore name from fromPodName (e.g., "quick-start-minimal-j7sv-dn-2-single-0" -> "quick-start-minimal-j7sv-dn-2")
    const getSourceXStoreName = (fromPodName: string): string => {
      if (!fromPodName) return '未指定';
      // Remove common pod suffixes (-single-0, -candidate-0, etc.)
      const match = fromPodName.match(/^(.+?)-(single|candidate|follower)-\d+$/);
      return match ? match[1] : fromPodName;
    };

    // Create a detailed view dialog
    const details = {
      基本信息: [
        { label: '名称', value: follower.metadata.name },
        { label: '命名空间', value: follower.metadata.namespace },
        { label: '目标 XStore', value: follower.spec.xStoreName },
        { label: '创建时间', value: follower.metadata.creationTimestamp }
      ],
      配置信息: [
        { label: '目标 XStore', value: follower.spec.xStoreName },
        { label: 'UID', value: (follower.spec as any).xStoreUid || '未指定' },
        { label: '源 XStore', value: getSourceXStoreName((follower.spec as any).fromPodName) },
        { label: '源 Pod', value: (follower.spec as any).fromPodName || '未指定' },
        { label: '备份集', value: (follower.spec as any).fromBackupSet || '未指定' },
        { label: '强制重建', value: (follower.spec as any).forceRecreate ? '是' : '否' },
        { label: '优先级', value: (follower.spec as any).priority?.toString() || '未设置' }
      ],
      状态信息: [
        { label: '当前阶段', value: follower.displayStatus },
        { label: '阶段代码', value: follower.status?.phase || '无' },
        { label: '当前任务', value: (follower.status as any)?.currentJobTask || '无' },
        { label: '备份 Job', value: (follower.status as any)?.backupJobName || '无' },
        { label: '恢复 Job', value: (follower.status as any)?.restoreJobName || '无' },
        { label: '当前 Job', value: (follower.status as any)?.currentJobName || '无' },
        { label: '是否恢复中', value: follower.isRecovering ? '是' : '否' },
        { label: '是否健康', value: follower.isHealthy ? '是' : '否' },
        { label: '有失败记录', value: follower.hasFailures ? '是' : '否' }
      ]
    };

    // Format details as readable text
    let message = '';
    Object.entries(details).forEach(([section, items]) => {
      message += `【${section}】\n`;
      items.forEach(item => {
        message += `${item.label}: ${item.value}\n`;
      });
      message += '\n';
    });

    alert(message);
  }

  // Refresh followers list
  async refreshFollowers(): Promise<void> {
    await this.loadFollowers();
  }

  // Refresh XStore list and revalidate sources
  async refreshXStores(): Promise<void> {
    try {
      this.availableXStores = await this.apiService.getXStores(this.data.namespace).toPromise() || [];
      await this.filterValidSourceXStores();
      
      this.snackBar.open('XStore 列表已刷新', '关闭', {
        duration: 2000,
        horizontalPosition: 'center',
        verticalPosition: 'top'
      });
    } catch (error) {
      console.error('Failed to refresh XStores:', error);
      this.snackBar.open('刷新 XStore 列表失败', '关闭', {
        duration: 3000,
        horizontalPosition: 'center',
        verticalPosition: 'top'
      });
    }
  }

  // Get status chip color
  getStatusColor(phase?: string): string {
    if (!phase) return 'basic';
    
    // Map XStoreFollower phases to colors
    if (phase === 'FollowerPhaseSuccess') return 'primary';
    if (phase === 'FollowerPhaseFailed') return 'warn';
    if (['FollowerPhaseCheck', 'FollowerPhaseBackupPrepare', 'FollowerPhaseBackupStart', 
         'FollowerPhaseBackup', 'FollowerPhaseLoggerRebuild', 'FollowerPhaseMonitorBackup',
         'FollowerPhaseBeforeRestore', 'FollowerPhaseRestore', 'FollowerPhaseAfterRestore',
         'FollowerPhaseLoggerCreate', 'FollowerCreateRemotePod'].includes(phase)) {
      return 'accent'; // Processing phases
    }
    
    return 'basic'; // Default/unknown phases
  }

  // Get progress percentage based on phase
  getProgressPercentage(follower: XStoreFollowerWithStatus): number {
    const phase = follower.status?.phase || '';
    
    // Estimate progress based on phase
    const progressMap: { [key: string]: number } = {
      '': 0,
      'FollowerPhaseCheck': 10,
      'FollowerPhaseBackupPrepare': 20,
      'FollowerPhaseBackupStart': 30,
      'FollowerPhaseBackup': 50,
      'FollowerPhaseLoggerRebuild': 60,
      'FollowerPhaseMonitorBackup': 65,
      'FollowerPhaseBeforeRestore': 70,
      'FollowerPhaseRestore': 85,
      'FollowerPhaseAfterRestore': 95,
      'FollowerPhaseSuccess': 100,
      'FollowerPhaseWaitSwitch': 90,
      'FollowerPhaseFailed': 0,
      'FollowerPhaseLoggerCreate': 40,
      'FollowerCreateRemotePod': 25,
      'FollowerPhaseDeleting': 50
    };
    
    return progressMap[phase] || 0;
  }

  // Check if form is valid
  isFormValid(): boolean {
    return this.followerForm.valid && 
           (!this.resourceForm.get('enableResourceLimits')?.value || this.resourceForm.valid);
  }

  // Cancel operation
  cancel(): void {
    if (this.dialogRef) {
      this.dialogRef.close();
    } else {
      this.selectedTab = 0;
    }
  }

  // Get field error message
  getFieldError(formGroup: FormGroup, fieldName: string): string {
    const field = formGroup.get(fieldName);
    if (field?.errors && field.touched) {
      if (field.errors['required']) {
        return '此字段为必填项';
      }
      if (field.errors['pattern']) {
        return '格式不正确，请使用小写字母、数字和连字符';
      }
      if (field.errors['min']) {
        return `最小值为 ${field.errors['min'].min}`;
      }
      if (field.errors['max']) {
        return `最大值为 ${field.errors['max'].max}`;
      }
    }
    return '';
  }

  // Switch to create mode
  async createNew(): Promise<void> {
    this.data.mode = 'create';
    this.followerForm.reset({
      namespace: this.data.namespace || 'default',
      priority: 1,
      forceRecreate: false
    });
    this.resourceForm.reset({
      enableResourceLimits: false,
      requestsCpu: '1',
      requestsMemory: '2Gi',
      requestsStorage: '10Gi',
      limitsCpu: '2',
      limitsMemory: '4Gi',
      limitsStorage: '20Gi',
      nodeSelector: {
        enabled: false,
        key: '',
        value: ''
      }
    });
    
    // 🎯 Auto-select appropriate source XStore
    console.log('🚀 Triggering auto-selection for new XStoreFollower');
    await this.filterValidSourceXStores();
    
    this.selectedTab = 1; // Switch to form tab
  }

  // Get title based on mode
  getTitle(): string {
    switch (this.data.mode) {
      case 'create': return '创建 XStore Follower';
      case 'edit': return '编辑 XStore Follower';
      case 'view': return 'XStore Follower 详情';
      default: return 'XStore Follower 管理';
    }
  }

  // Get XStore ready status for display
  getXStoreReadyStatus(xstore: XStore): string {
    try {
      // Try to get replica status first
      const replicaStatus = (xstore.status as any)?.replicaStatus;
      if (replicaStatus) {
        const ready = replicaStatus.ready || replicaStatus.available || 0;
        const total = replicaStatus.total || 0;
        return `${ready}/${total} Ready`;
      }

      // Fallback to readyStatus if available (this might be the correct field)
      const readyStatus = (xstore.status as any)?.readyStatus;
      if (readyStatus && typeof readyStatus === 'string') {
        return `${readyStatus}`;
      }

      // Check detailedStatus
      const detailedStatus = (xstore.status as any)?.detailedStatus;
      if (detailedStatus && detailedStatus.replicaStatus) {
        const ready = detailedStatus.replicaStatus.ready || detailedStatus.replicaStatus.available || 0;
        const total = detailedStatus.replicaStatus.total || 0;
        return `${ready}/${total} Ready`;
      }

      // Final fallback - show as available since it passed filtering
      return 'Available';
    } catch (error) {
      return 'Unknown';
    }
  }
}