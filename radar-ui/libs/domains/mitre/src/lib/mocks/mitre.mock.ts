import { MitreTactic } from '../interfaces';

export const MITRE_TACTICS: MitreTactic[] = [
    {
        id: 'TA0001',
        name: 'Tactic 1',
        techniques: [
            {
                id: 'TE0011',
                name: 'Technique 1.1'
            },
            {
                id: 'TE0012',
                name: 'Technique 1.2'
            }
        ]
    },
    {
        id: 'TA0002',
        name: 'Tactic 2',
        techniques: [
            {
                id: 'TE0021',
                name: 'Technique 2.1'
            }
        ]
    }
];
