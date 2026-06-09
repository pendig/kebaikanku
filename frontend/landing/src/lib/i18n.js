import { writable, derived } from 'svelte/store';
import idTranslations from './locales/id.json';
import enTranslations from './locales/en.json';

const translations = {
	id: idTranslations,
	en: enTranslations
};

// Check if browser environment to retrieve saved language
const isBrowser = typeof window !== 'undefined';
const savedLocale = isBrowser ? localStorage.getItem('lang') : 'id';
const initialLocale = (savedLocale === 'en' || savedLocale === 'id') ? savedLocale : 'id';

export const locale = writable(initialLocale);

// Subscribe to locale changes to persist preference in browser
if (isBrowser) {
	locale.subscribe((value) => {
		localStorage.setItem('lang', value);
	});
}

export const t = derived(locale, ($locale) => {
	return (key) => {
		const keys = key.split('.');
		let current = translations[$locale] || translations['id'];

		for (const k of keys) {
			if (current === undefined || current[k] === undefined) {
				// Fallback to English if not found in current locale
				let fallback = translations['en'];
				for (const fk of keys) {
					if (fallback === undefined || fallback[fk] === undefined) {
						return key; // Return raw key if not in English either
					}
					fallback = fallback[fk];
				}
				return fallback;
			}
			current = current[k];
		}
		return current;
	};
});
