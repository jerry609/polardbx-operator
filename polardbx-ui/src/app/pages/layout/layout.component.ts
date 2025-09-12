import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';

import { NzIconModule, NZ_ICONS, NzIconService } from 'ng-zorro-antd/icon';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import {
  MenuFoldOutline,
  MenuUnfoldOutline,
  AppstoreOutline,
  CloudUploadOutline,
  UndoOutline,
  DatabaseOutline,
  SettingOutline,
  SaveOutline,
  CalendarOutline,
  FileTextOutline,
  HddOutline,
  DeploymentUnitOutline,
  DashboardOutline,
  SlidersOutline,
  ProfileOutline,
  FileSearchOutline,
  HistoryOutline,
  BellOutline,
  ToolOutline,
  AlertOutline,
  NotificationOutline
} from '@ant-design/icons-angular/icons';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { FormsModule } from '@angular/forms';
import { NamespaceService } from '../../services/namespace.service';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { LoginDialogComponent } from '../../components/login-dialog/login-dialog.component';
import { AuthService } from '../../services/auth.service';

const icons = [
  MenuFoldOutline,
  MenuUnfoldOutline,
  AppstoreOutline,
  CloudUploadOutline,
  UndoOutline,
  DatabaseOutline,
  SettingOutline,
  SaveOutline,
  CalendarOutline,
  FileTextOutline,
  HddOutline,
  DeploymentUnitOutline,
  DashboardOutline,
  SlidersOutline,
  ProfileOutline,
  FileSearchOutline,
  HistoryOutline,
  BellOutline,
  ToolOutline,
  AlertOutline,
  NotificationOutline
];

@Component({
  selector: 'app-layout',
  standalone: true,
  imports: [CommonModule, RouterModule, NzIconModule, NzMenuModule, NzLayoutModule, MatSelectModule, MatFormFieldModule, FormsModule, MatDialogModule, MatIconModule, NzButtonModule, NzSelectModule, NzDividerModule],
  templateUrl: './layout.component.html',
  styleUrls: ['./layout.component.scss'],
  providers: [{ provide: NZ_ICONS, useValue: icons }]
})
export class LayoutComponent implements OnInit {
  isCollapsed = false;
  namespaces: string[] = [];
  activeNamespace: string | null = null;

  constructor(private iconService: NzIconService, private ns: NamespaceService, private dialog: MatDialog, private auth: AuthService) {
    // 确保运行时已注册，避免仅依赖providers导致的加载顺序问题
    this.iconService.addIcon(...icons);
  }

  async ngOnInit() {
    await this.ns.init();
    this.ns.namespaces$.subscribe(list => this.namespaces = list || []);
    this.ns.activeNamespace$.subscribe(ns => this.activeNamespace = ns);
    await this.auth.init();
  }

  async openLogin() {
    const s = await this.auth.session$.toPromise();
    if (!s?.enabled) return; // JWT 未启用
    this.dialog.open(LoginDialogComponent, { width: '360px' });
  }

  onChangeNamespace(value: string) {
    this.ns.setActive(value);
  }
} 