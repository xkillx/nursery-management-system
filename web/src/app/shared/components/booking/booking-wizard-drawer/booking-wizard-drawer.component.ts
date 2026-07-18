import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { DrawerComponent } from '../../ui/modal/drawer.component';
import { CapacityIndicatorComponent } from '../../common/capacity-indicator/capacity-indicator.component';
import { DaySelectorComponent } from '../../form/day-selector/day-selector.component';

export interface BookingWizardStep {
  key: string;
  label: string;
  description: string;
}

export interface BookingData {
  bookingType: 'recurring' | 'adhoc' | 'funded';
  daysOfWeek?: number[];
  specificDate?: string;
  sessionTemplate?: string;
  room?: string;
  startDate?: string;
  endDate?: string;
  fundingAllocation?: string;
  managerOverride?: boolean;
}

@Component({
  selector: 'app-booking-wizard-drawer',
  imports: [CommonModule, FormsModule, DrawerComponent, CapacityIndicatorComponent, DaySelectorComponent],
  templateUrl: './booking-wizard-drawer.component.html',
})
export class BookingWizardDrawerComponent {
  @Input() isOpen = false;
  @Input() currentCapacity = 0;
  @Input() maxCapacity = 0;
  @Input() sessionTemplateOptions: { value: string; label: string }[] = [];
  @Input() roomOptions: { value: string; label: string }[] = [];
  @Input() fundingOptions: { value: string; label: string }[] = [];
  @Input() isManager = false;

  @Output() closed = new EventEmitter<void>();
  @Output() completed = new EventEmitter<BookingData>();

  readonly steps: readonly BookingWizardStep[] = [
    { key: 'booking-type', label: 'Booking Type', description: 'Select booking type' },
    { key: 'schedule', label: 'Schedule', description: 'Configure schedule' },
    { key: 'review', label: 'Review', description: 'Review and submit' },
  ];

  currentStepIndex = 0;
  bookingType: 'recurring' | 'adhoc' | 'funded' | '' = '';
  daysOfWeek: number[] = [];
  specificDate = '';
  sessionTemplate = '';
  room = '';
  startDate = '';
  endDate = '';
  fundingAllocation = '';
  managerOverride = false;

  get currentStep(): BookingWizardStep {
    return this.steps[this.currentStepIndex];
  }

  get isFirstStep(): boolean {
    return this.currentStepIndex === 0;
  }

  get isLastStep(): boolean {
    return this.currentStepIndex === this.steps.length - 1;
  }

  get canProceed(): boolean {
    if (this.currentStepIndex === 0) {
      return this.bookingType !== '';
    }
    if (this.currentStepIndex === 1) {
      if (this.bookingType === 'recurring') {
        return this.daysOfWeek.length > 0 && !!this.startDate && !!this.endDate;
      }
      if (this.bookingType === 'adhoc') {
        return !!this.specificDate;
      }
      if (this.bookingType === 'funded') {
        return !!this.startDate && !!this.fundingAllocation;
      }
    }
    return true;
  }

  get capacityStatusLevel(): 'green' | 'amber' | 'red' {
    if (this.maxCapacity <= 0) return 'green';
    const ratio = this.currentCapacity / this.maxCapacity;
    if (ratio >= 1) return 'red';
    if (ratio >= 0.8) return 'amber';
    return 'green';
  }

  stepIsActive(stepKey: string): boolean {
    return this.steps[this.currentStepIndex].key === stepKey;
  }

  stepIsComplete(stepIndex: number): boolean {
    return stepIndex < this.currentStepIndex;
  }

  nextStep(): void {
    if (this.currentStepIndex < this.steps.length - 1 && this.canProceed) {
      this.currentStepIndex++;
    }
  }

  prevStep(): void {
    if (this.currentStepIndex > 0) {
      this.currentStepIndex--;
    }
  }

  submit(): void {
    const data: BookingData = {
      bookingType: this.bookingType as BookingData['bookingType'],
      daysOfWeek: this.bookingType === 'recurring' ? this.daysOfWeek : undefined,
      specificDate: this.bookingType === 'adhoc' ? this.specificDate : undefined,
      sessionTemplate: this.sessionTemplate || undefined,
      room: this.room || undefined,
      startDate: this.startDate || undefined,
      endDate: this.endDate || undefined,
      fundingAllocation: this.bookingType === 'funded' ? this.fundingAllocation : undefined,
      managerOverride: this.isManager ? this.managerOverride : undefined,
    };
    this.completed.emit(data);
    this.reset();
  }

  reset(): void {
    this.currentStepIndex = 0;
    this.bookingType = '';
    this.daysOfWeek = [];
    this.specificDate = '';
    this.sessionTemplate = '';
    this.room = '';
    this.startDate = '';
    this.endDate = '';
    this.fundingAllocation = '';
    this.managerOverride = false;
  }

  onClose(): void {
    this.reset();
    this.closed.emit();
  }
}
