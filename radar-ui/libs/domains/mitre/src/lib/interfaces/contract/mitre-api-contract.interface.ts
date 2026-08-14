import { MitreTactic } from './mitre-contract.interface';

export interface GetMitreMatrixResponse {
    matrix_version: string;
    tactics: MitreTactic[];
}
