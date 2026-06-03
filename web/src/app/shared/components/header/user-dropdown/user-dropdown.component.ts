import { Component, inject } from '@angular/core';
import { DropdownComponent } from '../../ui/dropdown/dropdown.component';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';

import { ROLES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-user-dropdown',
  templateUrl: './user-dropdown.component.html',
  imports:[CommonModule,RouterModule,DropdownComponent]
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
      return 'User';
    }

    return email;
  }

  get sessionLabel(): string {
    const role = this.authService.currentRole();
    switch (role) {
      case ROLES.manager:
        return 'Manager session';
      case ROLES.practitioner:
        return 'Practitioner session';
      case ROLES.parent:
        return 'Parent session';
      default:
        return 'Session';
    }
  }
}
