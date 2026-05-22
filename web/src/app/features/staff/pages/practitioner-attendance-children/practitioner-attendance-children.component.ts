import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { StaffApiService } from '../../data/staff-api.service';
import { AttendanceChildRecord } from '../../models/attendance-child.models';

@Component({
  selector: 'app-practitioner-attendance-children',
  imports: [CommonModule],
  templateUrl: './practitioner-attendance-children.component.html',
})
export class PractitionerAttendanceChildrenComponent {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  children: AttendanceChildRecord[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  ngOnInit(): void {
    this.loadChildren();
  }

  loadChildren(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi.listAttendanceChildren().subscribe({
      next: (children) => {
        this.children = children;
        this.isLoading = false;
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = mapped.requestId
          ? `${mapped.message} (Request: ${mapped.requestId})`
          : mapped.message;
      },
    });
  }
}
