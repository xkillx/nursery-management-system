import { Component, Input } from '@angular/core';
import { BadgeComponent } from '../../ui/badge/badge.component';

@Component({
  selector: 'app-attendance-summary-bar',
  imports: [BadgeComponent],
  template: `
    <div class="flex flex-wrap items-center gap-2">
      <app-badge size="sm" color="light">
        Total: {{ total }}
      </app-badge>
      <app-badge size="sm" color="success">
        Present: {{ present }}
      </app-badge>
      <app-badge size="sm" color="error">
        Absent: {{ absent }}
      </app-badge>
      <app-badge size="sm" color="warning">
        Late: {{ late }}
      </app-badge>
    </div>
  `,
})
export class AttendanceSummaryBarComponent {
  @Input() total = 0;
  @Input() present = 0;
  @Input() absent = 0;
  @Input() late = 0;
}
