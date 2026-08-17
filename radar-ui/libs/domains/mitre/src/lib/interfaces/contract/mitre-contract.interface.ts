export interface MitreTactic {
    id: string;
    name: string;
    techniques: MitreTechnique[];
}

export interface MitreTechnique {
    id: string;
    name: string;
}
