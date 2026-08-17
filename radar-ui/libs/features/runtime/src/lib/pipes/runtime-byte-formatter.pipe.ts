import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'runtimeByteFormatter',
    standalone: false
})
export class RuntimeFeatureByteFormatterPipe implements PipeTransform {
    readonly size = 1024;
    readonly labels = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

    transform(byte: number): string {
        if (!+byte) {
            return `0 ${this.labels[0]}`;
        }

        const value = Math.floor(Math.log(byte) / Math.log(this.size));

        return `${parseFloat((byte / Math.pow(this.size, value)).toFixed(2))} ${this.labels[value]}`;
    }
}
