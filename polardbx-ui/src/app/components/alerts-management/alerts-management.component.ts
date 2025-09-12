import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBarModule, MatSnackBar } from '@angular/material/snack-bar';
import { ApiService } from '../../services/api.service';

@Component({
  selector: 'app-alerts-management',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatCardModule, MatTableModule, MatButtonModule, MatIconModule, MatFormFieldModule, MatInputModule, MatSelectModule, MatSnackBarModule],
  template: `
    <div class="alerts">
      <mat-card>
        <mat-card-header>
          <mat-card-title>
            <mat-icon>alert</mat-icon>
            告警管理
          </mat-card-title>
          <mat-card-subtitle>规则包 / 路由 / 静默</mat-card-subtitle>
        </mat-card-header>
        <mat-card-content>
          <div class="toolbar">
            <button mat-raised-button color="primary" (click)="reload()"><mat-icon>refresh</mat-icon>刷新</button>
            <button mat-raised-button (click)="createProfile()"><mat-icon>add</mat-icon>新建规则包</button>
          </div>

          <h3>规则包</h3>
          <table mat-table [dataSource]="profiles" class="mat-elevation-z1">
            <ng-container matColumnDef="name">
              <th mat-header-cell *matHeaderCellDef>名称</th>
              <td mat-cell *matCellDef="let p">{{ p }}</td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef>操作</th>
              <td mat-cell *matCellDef="let p">
                <button mat-icon-button (click)="viewProfile(p)"><mat-icon>visibility</mat-icon></button>
                <button mat-icon-button (click)="editProfile(p)"><mat-icon>edit</mat-icon></button>
                <button mat-icon-button color="warn" (click)="deleteProfile(p)"><mat-icon>delete</mat-icon></button>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="['name','actions']"></tr>
            <tr mat-row *matRowDef="let row; columns: ['name','actions']"></tr>
          </table>

          <h3>静默</h3>
          <table mat-table [dataSource]="silences" class="mat-elevation-z1">
            <ng-container matColumnDef="id">
              <th mat-header-cell *matHeaderCellDef>ID</th>
              <td mat-cell *matCellDef="let s">{{ s.id }}</td>
            </ng-container>
            <ng-container matColumnDef="matchers">
              <th mat-header-cell *matHeaderCellDef>匹配</th>
              <td mat-cell *matCellDef="let s">{{ s.matchers?.length || 0 }}</td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef>操作</th>
              <td mat-cell *matCellDef="let s">
                <button mat-icon-button color="warn" (click)="deleteSilence(s.id)"><mat-icon>delete</mat-icon></button>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="silenceCols"></tr>
            <tr mat-row *matRowDef="let row; columns: silenceCols"></tr>
          </table>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .alerts { display: grid; gap: 16px; }
    .toolbar { margin: 8px 0; display: flex; gap: 8px; }
  `]
})
export class AlertsManagementComponent {
  profiles: string[] = [];
  silences: any[] = [];
  silenceCols = ['id','matchers','actions'];

  constructor(private api: ApiService, private snack: MatSnackBar, private fb: FormBuilder) {}

  reload() {
    this.api.listAlertProfiles().subscribe(r => this.profiles = r?.names || []);
    this.api.listSilences().subscribe(r => this.silences = r || []);
  }

  createProfile() {
    const name = prompt('输入规则包名称');
    if (!name) return;
    const content = 'groups: []\n';
    this.api.createAlertProfile(name, content).subscribe(() => {
      this.snack.open('创建成功', '关闭', { duration: 2000 });
      this.reload();
    });
  }

  viewProfile(name: string) {
    this.api.getAlertProfile(name).subscribe(p => {
      alert(`规则包: ${p.name}\n\n${p.content?.slice(0, 500)}`);
    });
  }

  editProfile(name: string) {
    this.api.getAlertProfile(name).subscribe(p => {
      const content = prompt(`编辑规则包 ${name}`, p.content || '');
      if (content == null) return;
      this.api.updateAlertProfile(name, content).subscribe(() => {
        this.snack.open('已更新', '关闭', { duration: 2000 });
        this.reload();
      });
    });
  }

  deleteProfile(name: string) {
    if (!confirm(`删除规则包 ${name} ?`)) return;
    this.api.deleteAlertProfile(name).subscribe(() => {
      this.snack.open('已删除', '关闭', { duration: 2000 });
      this.reload();
    });
  }

  deleteSilence(id: string) {
    if (!confirm(`删除静默 ${id} ?`)) return;
    this.api.deleteSilence(id).subscribe(() => {
      this.snack.open('已删除', '关闭', { duration: 2000 });
      this.reload();
    });
  }
}