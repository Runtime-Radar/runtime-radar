import { InjectionToken } from '@angular/core';

export const AVAILABLE_LOCALES = new InjectionToken<string[]>('availableLocales');

export const DEFAULT_LOCALE = new InjectionToken<string>('defaultLocale');
