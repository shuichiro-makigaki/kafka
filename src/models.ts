import {DateTime} from 'luxon';

export class MovieModel {
    id: string;
    file: string;
    title: string;
    lastModifiedTime: DateTime;
}
