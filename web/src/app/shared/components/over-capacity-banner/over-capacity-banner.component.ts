import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

export interface OverCapacityRoom {
  id: string;
  name: string;
  assigned: number;
  capacity: number;
}

@Component({
  selector: 'app-over-capacity-banner',
  imports: [CommonModule],
  templateUrl: './over-capacity-banner.component.html',
})
export class OverCapacityBannerComponent {
  @Input() rooms: OverCapacityRoom[] = [];

  get visible(): boolean {
    return this.rooms.length > 0;
  }

  describe(room: OverCapacityRoom): string {
    return `${room.name} is over capacity (${room.assigned}/${room.capacity})`;
  }
}
