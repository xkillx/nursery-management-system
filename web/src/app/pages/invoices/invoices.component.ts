import { Component } from '@angular/core';
import { InvoiceSidebarComponent } from '../../shared/components/invoice/invoice-sidebar/invoice-sidebar.component';
import { InvoiceMainComponent } from '../../shared/components/invoice/invoice-main/invoice-main.component';

@Component({
  selector: 'app-invoices',
  imports: [
    InvoiceSidebarComponent,
    InvoiceMainComponent
  ],
  templateUrl: './invoices.component.html',
  styles: ``
})
export class InvoicesComponent {

}
