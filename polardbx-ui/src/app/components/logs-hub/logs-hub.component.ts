import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router, ActivatedRoute, NavigationEnd } from '@angular/router';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { filter } from 'rxjs/operators';

@Component({
  selector: 'app-logs-hub',
  standalone: true,
  imports: [CommonModule, RouterModule, NzTabsModule, NzIconModule],
  template: `
    <div class="logs-hub">
      <nz-tabset nzType="card" class="tabs" [nzTabBarGutter]="8"
                 [nzSelectedIndex]="selectedIndex"
                 (nzSelectedIndexChange)="onTabChange($event)">
        <nz-tab nzTitle="采集器"></nz-tab>
        <nz-tab nzTitle="ILM"></nz-tab>
        <nz-tab nzTitle="Logstash/日志"></nz-tab>
      </nz-tabset>

      <div class="outlet">
        <router-outlet></router-outlet>
      </div>
    </div>
  `,
  styles: [`
    .logs-hub { padding: 8px 16px; background: #ffffff; }
    .tabs { background: #fff; margin-bottom: 8px; }
    .outlet { background: #fff; border: 1px solid #e0e0e0; border-radius: 8px; padding: 12px; }
  `]
})
export class LogsHubComponent implements OnInit {
  selectedIndex = 0;
  private paths = ['collectors', 'ilm', 'cluster'];

  constructor(private router: Router, private route: ActivatedRoute) {}

  ngOnInit(): void {
    this.updateSelectedFromUrl();
    this.router.events.pipe(filter(e => e instanceof NavigationEnd)).subscribe(() => this.updateSelectedFromUrl());
  }

  onTabChange(idx: number): void {
    const p = this.paths[idx] || 'collectors';
    this.router.navigate([p], { relativeTo: this.route });
  }

  private updateSelectedFromUrl(): void {
    const child = this.route.firstChild;
    const seg = child?.snapshot?.url?.[0]?.path || 'collectors';
    const idx = this.paths.indexOf(seg);
    this.selectedIndex = idx >= 0 ? idx : 0;
  }
}


