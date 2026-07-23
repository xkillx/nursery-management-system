import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroPencilSquare,
  heroTrash,
  heroCheck,
  heroXMark,
  heroChevronDown,
  heroChevronUp,
  heroUser,
  heroHome,
  heroMapPin,
} from '@ng-icons/heroicons/outline';
import { ButtonComponent } from '../button/button.component';
import { FormFieldComponent } from '../../form/form-field/form-field.component';
import { InputFieldComponent } from '../../form/input/input-field.component';
import { SelectComponent, Option } from '../../form/select/select.component';
import { SwitchComponent } from '../../form/input/switch.component';

interface LinkedParentEntry {
  id: string;
  parentId: string;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  relationship: string;
  customRelationship: string;
  addressLine1: string;
  addressLine2: string;
  addressCity: string;
  addressPostcode: string;
  hasParentalResponsibility: boolean;
  canPickUp: boolean;
  isEmergencyContact: boolean;
  portalStatus: 'active' | 'none';
  isEditing: boolean;
  isNew: boolean;
}

@Component({
  selector: 'app-parent-card',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    ButtonComponent,
    FormFieldComponent,
    InputFieldComponent,
    SelectComponent,
    SwitchComponent,
  ],
  providers: [
    provideIcons({
      heroPencilSquare,
      heroTrash,
      heroCheck,
      heroXMark,
      heroChevronDown,
      heroChevronUp,
      heroUser,
      heroHome,
      heroMapPin,
    }),
  ],
  template: `
    <div
      class="rounded-xl border bg-white dark:bg-gray-900/40 transition-all duration-200"
      [class.border-brand-200]="parent.isEditing"
      [class.dark:border-brand-800/40]="parent.isEditing"
      [class.border-gray-200]="!parent.isEditing"
      [class.dark:border-gray-800]="!parent.isEditing"
    >
      <!-- Collapsed View -->
      <div class="flex items-center gap-3 p-4">
        <div class="w-10 h-10 rounded-xl bg-brand-50 text-brand-600 font-bold text-sm flex items-center justify-center border border-brand-100 shrink-0 dark:bg-brand-500/10 dark:text-brand-400 dark:border-brand-500/20">
          {{ parent.firstName[0] }}{{ parent.lastName ? parent.lastName[0] : '' }}
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="font-bold text-gray-900 text-sm dark:text-white">
              {{ parent.firstName }} {{ parent.lastName }}
            </span>
            <span class="text-xs text-gray-500 dark:text-gray-400">
              {{ parent.relationship || 'Guardian' }}
            </span>
          </div>
          <div class="flex items-center gap-2 mt-1 flex-wrap">
            @if (parent.phone) {
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ parent.phone }}</span>
            }
            @if (parent.email) {
              <span class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ parent.email }}</span>
            }
          </div>
          <!-- Flag Badges -->
          <div class="flex items-center gap-1.5 mt-2 flex-wrap">
            @if (parent.hasParentalResponsibility) {
              <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-400">
                PR
              </span>
            }
            @if (parent.canPickUp) {
              <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-blue-light-50 text-blue-light-500 dark:bg-blue-light-500/15 dark:text-blue-light-400">
                Pickup
              </span>
            }
            @if (parent.isEmergencyContact) {
              <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-warning-50 text-warning-600 dark:bg-warning-500/15 dark:text-orange-400">
                Emergency
              </span>
            }
            @if (parent.portalStatus === 'active') {
              <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-gray-100 text-gray-600 dark:bg-white/5 dark:text-gray-400">
                Portal Active
              </span>
            }
          </div>
        </div>
        <div class="flex items-center gap-1.5 shrink-0">
          @if (!parent.isEditing) {
            <app-button type="button" variant="ghost" size="xs" className="text-gray-500 hover:text-brand-600" (btnClick)="editRequested.emit()">
              <ng-icon name="heroPencilSquare" size="16" />
              Edit
            </app-button>
          }
        </div>
      </div>

      <!-- Expanded View -->
      @if (parent.isEditing) {
        <div class="border-t border-gray-100 dark:border-gray-800 p-4 space-y-4 animate-fade-in">
          <!-- Basic Info -->
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <app-form-field label="First name" [required]="true">
              <app-input-field
                type="text"
                [(ngModel)]="parent.firstName"
                placeholder="First name"
              />
            </app-form-field>
            <app-form-field label="Last name">
              <app-input-field
                type="text"
                [(ngModel)]="parent.lastName"
                placeholder="Last name"
              />
            </app-form-field>
            <app-form-field label="Email">
              <app-input-field
                type="email"
                [(ngModel)]="parent.email"
                placeholder="email@example.com"
              />
            </app-form-field>
            <app-form-field label="Phone">
              <app-input-field
                type="tel"
                [(ngModel)]="parent.phone"
                placeholder="+44 7000 000000"
              />
            </app-form-field>
          </div>

          <!-- Relationship -->
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <app-form-field label="Relationship" [required]="true">
              <app-select
                [(ngModel)]="parent.relationship"
                [options]="parentRelationshipOptions"
                placeholder="Select relationship"
                [placeholderDisabled]="false"
              />
            </app-form-field>
            @if (parent.relationship === 'Other') {
              <app-form-field label="Custom relationship">
                <app-input-field
                  type="text"
                  [(ngModel)]="parent.customRelationship"
                  placeholder="e.g. Aunt, Uncle, Family Friend"
                />
              </app-form-field>
            }
          </div>

          <!-- Address -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">Address</span>
              @if (childAddress) {
                <div
                  role="button"
                  tabindex="0"
                  class="flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-1.5 text-xs font-medium transition-all duration-200 hover:border-brand-300 hover:text-brand-600"
                  [class.border-brand-500]="useChildAddress"
                  [class.bg-brand-50]="useChildAddress"
                  [class.text-brand-700]="useChildAddress"
                  [class.dark:border-brand-400]="useChildAddress"
                  [class.dark:bg-brand-500/10]="useChildAddress"
                  [class.dark:text-brand-300]="useChildAddress"
                  [class.border-gray-200]="!useChildAddress"
                  [class.text-gray-600]="!useChildAddress"
                  [class.dark:border-gray-700]="!useChildAddress"
                  [class.dark:text-gray-400]="!useChildAddress"
                  (click)="toggleUseChildAddress()"
                  (keydown.enter)="toggleUseChildAddress()"
                  (keydown.space)="toggleUseChildAddress(); $event.preventDefault()"
                >
                  <ng-icon name="heroHome" size="14" aria-hidden="true" />
                  @if (useChildAddress) {
                    Using child's address
                  } @else {
                    Same as child's address
                  }
                </div>
              }
            </div>
            @if (useChildAddress && childAddress) {
              <div class="flex items-start gap-3 rounded-xl border border-brand-200 bg-brand-50/30 p-4 dark:border-brand-800/40 dark:bg-brand-500/5">
                <ng-icon name="heroMapPin" size="18" class="mt-0.5 shrink-0 text-brand-500" aria-hidden="true" />
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">{{ childAddress }}</p>
                  <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">Copied from child's home address</p>
                </div>
                <button
                  type="button"
                  class="shrink-0 text-xs font-medium text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                  (click)="toggleUseChildAddress()"
                >
                  Edit
                </button>
              </div>
            } @else {
              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <app-form-field label="Address line 1">
                  <app-input-field
                    type="text"
                    [(ngModel)]="parent.addressLine1"
                    placeholder="Street address"
                  />
                </app-form-field>
                <app-form-field label="Address line 2">
                  <app-input-field
                    type="text"
                    [(ngModel)]="parent.addressLine2"
                    placeholder="Flat, suite, etc."
                  />
                </app-form-field>
                <app-form-field label="City">
                  <app-input-field
                    type="text"
                    [(ngModel)]="parent.addressCity"
                    placeholder="City"
                  />
                </app-form-field>
                <app-form-field label="Postcode">
                  <app-input-field
                    type="text"
                    [(ngModel)]="parent.addressPostcode"
                    placeholder="SW1A 1AA"
                    className="uppercase"
                  />
                </app-form-field>
              </div>
            }
          </div>

          <!-- Flags -->
          <div class="space-y-3">
            <span class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">Permissions</span>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <div
                role="button"
                tabindex="0"
                class="flex cursor-pointer items-center justify-between gap-3 rounded-xl border p-3 transition-colors"
                [class.border-brand-500]="parent.hasParentalResponsibility"
                [class.bg-brand-50/15]="parent.hasParentalResponsibility"
                [class.dark:border-brand-400]="parent.hasParentalResponsibility"
                [class.dark:bg-brand-500/10]="parent.hasParentalResponsibility"
                [class.border-gray-200]="!parent.hasParentalResponsibility"
                [class.dark:border-gray-800]="!parent.hasParentalResponsibility"
                (click)="parent.hasParentalResponsibility = !parent.hasParentalResponsibility"
                (keydown.enter)="parent.hasParentalResponsibility = !parent.hasParentalResponsibility"
                (keydown.space)="parent.hasParentalResponsibility = !parent.hasParentalResponsibility; $event.preventDefault()"
              >
                <span class="text-xs font-medium text-gray-700 dark:text-gray-300">Parental Responsibility</span>
                <div class="pointer-events-none">
                  <app-switch
                    name="hasParentalResponsibility"
                    [(ngModel)]="parent.hasParentalResponsibility"
                  />
                </div>
              </div>
              <div
                role="button"
                tabindex="0"
                class="flex cursor-pointer items-center justify-between gap-3 rounded-xl border p-3 transition-colors"
                [class.border-blue-light-500]="parent.canPickUp"
                [class.bg-blue-light-50/15]="parent.canPickUp"
                [class.dark:border-blue-light-400]="parent.canPickUp"
                [class.dark:bg-blue-light-500/10]="parent.canPickUp"
                [class.border-gray-200]="!parent.canPickUp"
                [class.dark:border-gray-800]="!parent.canPickUp"
                (click)="parent.canPickUp = !parent.canPickUp"
                (keydown.enter)="parent.canPickUp = !parent.canPickUp"
                (keydown.space)="parent.canPickUp = !parent.canPickUp; $event.preventDefault()"
              >
                <span class="text-xs font-medium text-gray-700 dark:text-gray-300">Can Pick Up</span>
                <div class="pointer-events-none">
                  <app-switch
                    name="canPickUp"
                    [(ngModel)]="parent.canPickUp"
                  />
                </div>
              </div>
              <div
                role="button"
                tabindex="0"
                class="flex cursor-pointer items-center justify-between gap-3 rounded-xl border p-3 transition-colors"
                [class.border-warning-500]="parent.isEmergencyContact"
                [class.bg-warning-50/15]="parent.isEmergencyContact"
                [class.dark:border-orange-400]="parent.isEmergencyContact"
                [class.dark:bg-warning-500/10]="parent.isEmergencyContact"
                [class.border-gray-200]="!parent.isEmergencyContact"
                [class.dark:border-gray-800]="!parent.isEmergencyContact"
                (click)="parent.isEmergencyContact = !parent.isEmergencyContact"
                (keydown.enter)="parent.isEmergencyContact = !parent.isEmergencyContact"
                (keydown.space)="parent.isEmergencyContact = !parent.isEmergencyContact; $event.preventDefault()"
              >
                <span class="text-xs font-medium text-gray-700 dark:text-gray-300">Emergency Contact</span>
                <div class="pointer-events-none">
                  <app-switch
                    name="isEmergencyContact"
                    [(ngModel)]="parent.isEmergencyContact"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center justify-between border-t border-gray-100 dark:border-gray-800 pt-4">
            <app-button type="button" variant="ghost" size="sm" className="text-error-600 hover:bg-error-50 dark:text-error-400 dark:hover:bg-error-500/10" (btnClick)="removeRequested.emit()">
              <ng-icon name="heroTrash" size="16" />
              Remove
            </app-button>
            <div class="flex items-center gap-2">
              <app-button type="button" variant="outline" size="sm" (btnClick)="cancelRequested.emit()">
                Cancel
              </app-button>
              <app-button type="button" variant="primary" size="sm" (btnClick)="saveRequested.emit(parent)">
                <ng-icon name="heroCheck" size="16" />
                Save
              </app-button>
            </div>
          </div>
        </div>
      }
    </div>
  `,
})
export class ParentCardComponent {
  @Input() parent!: LinkedParentEntry;
  @Input() parentRelationshipOptions: Option[] = [];
  @Input() childAddress: string | null = null;

  @Output() editRequested = new EventEmitter<void>();
  @Output() removeRequested = new EventEmitter<void>();
  @Output() saveRequested = new EventEmitter<LinkedParentEntry>();
  @Output() cancelRequested = new EventEmitter<void>();
  @Output() useChildAddressChange = new EventEmitter<boolean>();

  useChildAddress = false;

  toggleUseChildAddress(): void {
    this.useChildAddress = !this.useChildAddress;
    this.useChildAddressChange.emit(this.useChildAddress);
  }
}
