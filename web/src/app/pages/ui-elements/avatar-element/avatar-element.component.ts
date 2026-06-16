import { Component } from '@angular/core';
import { AvatarComponent } from '../../../shared/components/ui/avatar/avatar.component';
import { ComponentCardComponent } from '../../../shared/components/common/component-card/component-card.component';

@Component({
  selector: 'app-avatar-element',
  imports: [
    AvatarComponent,
    ComponentCardComponent,
  ],
  templateUrl: './avatar-element.component.html',
  styles: ``
})
export class AvatarElementComponent {

}
