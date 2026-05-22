import { Component, inject } from '@angular/core';
import { DropdownComponent } from '../../ui/dropdown/dropdown.component';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { DropdownItemTwoComponent } from '../../ui/dropdown/dropdown-item/dropdown-item.component-two';

import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-user-dropdown',
  templateUrl: './user-dropdown.component.html',
  imports:[CommonModule,RouterModule,DropdownComponent,DropdownItemTwoComponent]
})
export class UserDropdownComponent {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  isOpen = false;

  toggleDropdown() {
    this.isOpen = !this.isOpen;
  }

  closeDropdown() {
    this.isOpen = false;
  }

  signOut(): void {
    this.authService.logout().subscribe(() => {
      this.closeDropdown();
      this.router.navigate(['/signin']);
    });
  }

  get displayName(): string {
    const email = this.authService.user()?.email;
    if (!email) {
      return 'Staff user';
    }

    return email;
  }
}
