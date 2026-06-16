import { Component } from '@angular/core';
import { ComponentCardComponent } from '../../../shared/components/common/component-card/component-card.component';
// import { AspectRatioVideoComponent } from '../../../shared/components/ui/videos/aspect-ratio-video/aspect-ratio-video.component';
import { SixteenIstoNineComponent } from '../../../shared/components/ui/videos/sixteen-isto-nine/sixteen-isto-nine.component';
import { FourIstoThreeComponent } from '../../../shared/components/ui/videos/four-isto-three/four-isto-three.component';
import { OneIstoOneComponent } from '../../../shared/components/ui/videos/one-isto-one/one-isto-one.component';
import { TwentyoneIstoNineComponent } from "../../../shared/components/ui/videos/twentyone-isto-nine/twentyone-isto-nine.component";

@Component({
  selector: 'app-videos',
  imports: [
    ComponentCardComponent,
    // AspectRatioVideoComponent,
    SixteenIstoNineComponent,
    FourIstoThreeComponent,
    OneIstoOneComponent,
    TwentyoneIstoNineComponent
],
  templateUrl: './videos.component.html',
  styles: ``
})
export class VideosComponent {

}
