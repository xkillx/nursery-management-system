import { Component, inject } from '@angular/core';
import { ThemeService } from '../../../services/theme.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-theme-toggle-button',
  templateUrl: './theme-toggle-button.component.html',
  imports:[CommonModule]
})
export class ThemeToggleButtonComponent {
  
  private readonly themeService = inject(ThemeService);
  theme$ = this.themeService.theme$;

  toggleTheme() {
    this.themeService.toggleTheme();
  }
}