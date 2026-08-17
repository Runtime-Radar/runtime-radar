import { I18nLocale } from '@cs/i18n';

import { MITRE_TACTICS } from '../mocks/mitre.mock';
import { MitreEntityState } from '../interfaces';
import { adapter } from './mitre-reducer.store';
import { getMitreTactic, getMitreTechnique } from './mitre-selector.store';

describe('MitreDomainReducer', () => {
    let entityState: MitreEntityState = adapter.getInitialState();

    beforeEach(() => {
        entityState = adapter.setAll(MITRE_TACTICS, adapter.getInitialState());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('getMitreTactic', () => {
        const locale = I18nLocale.EN;
        const tactic = MITRE_TACTICS[0];

        it('should return tactic', () => {
            expect(getMitreTactic(tactic.id, locale).projector(entityState)).toEqual(tactic);
        });

        it('should return undefined when tactic is not found', () => {
            expect(getMitreTactic('TA0003', locale).projector(entityState)).toBeUndefined();
        });
    });

    describe('getMitreTechnique', () => {
        const locale = I18nLocale.EN;
        const tactic = MITRE_TACTICS[0];

        it('should return technique', () => {
            expect(getMitreTechnique(tactic.techniques[0].id, tactic.id, locale).projector(entityState)).toEqual(
                tactic.techniques[0]
            );
        });

        it('should return undefined when technique is not found', () => {
            expect(getMitreTechnique('TE0013', tactic.id, locale).projector(entityState)).toBeUndefined();
            expect(getMitreTechnique(tactic.techniques[0].id, 'TA0003', locale).projector(entityState)).toBeUndefined();
        });
    });
});
