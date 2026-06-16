
import { Component } from '@angular/core';
import { BarChartOneComponent } from '../../../shared/components/charts/bar/bar-chart-one/bar-chart-one.component';
import { ComponentCardComponent } from '../../../shared/components/common/component-card/component-card.component';

@Component({
  selector: 'app-bar-chart',
  imports: [
    ComponentCardComponent,
    BarChartOneComponent
],
  templateUrl: './bar-chart.component.html',
  styles: ``
})
export class BarChartComponent {

}
