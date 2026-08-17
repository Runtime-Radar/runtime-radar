export enum DetectorType {
    RUNTIME = 'RUNTIME'
}

export interface DetectorMitreTactic {
    id: string;
    techniques: string[];
}

export interface Detector {
    id: string;
    name: string;
    description: string;
    version: number;
    author?: string;
    contact?: string;
    license?: string;
    tactics_covered: DetectorMitreTactic[];
}
