# Localization (i18n) Strategy

This document details the multi-language (i18n) strategy implemented in **kebaikanku.id**. 

The platform is designed to be globally ready, supporting multiple languages while prioritizing **Indonesian (id)** as the main locale, and **English (en)** as the secondary locale.

---

## 🎨 Architectural Overview

To keep the application highly modular and compatible with Cloudflare Pages (which hosts the static pre-rendered pages), we implement a hybrid localization architecture:

1.  **Frontend Localization**:
    *   Uses a lightweight, compile-time/client-side dictionary structure.
    *   Avoids heavy third-party libraries to ensure the landing page bundle remains extremely lightweight and fast to load.
    *   Saves the language preference in the local storage (`localStorage`) or URL path parameters (e.g., `/en/privacy`).
2.  **Backend Error & API Localization**:
    *   Go Backend sends standardized error codes.
    *   Frontend reads these codes and maps them to localized UI error messages.
    *   For emails and transaction updates, the backend reads the client's `Accept-Language` header to deliver localized emails/SMS.

---

## 📁 Frontend Translation Files Structure

Locales are stored as standard JSON key-value pairs in the frontend projects under `src/lib/locales/`:

```
frontend/landing/src/lib/locales/
├── id.json                     # Indonesian Translations (Primary)
└── en.json                     # English Translations (Secondary)
```

### Example: `id.json`
```json
{
  "common": {
    "appName": "kebaikanku.id",
    "donateNow": "Donasi Sekarang",
    "learnMore": "Pelajari Selengkapnya"
  },
  "hero": {
    "title": "Donasi Transparan & Bebas Potongan Biaya Platform",
    "subtitle": "Platform crowdfunding open-source yang dirancang untuk lembaga zakat dan kemanusiaan terpercaya."
  }
}
```

### Example: `en.json`
```json
{
  "common": {
    "appName": "kebaikanku.id",
    "donateNow": "Donate Now",
    "learnMore": "Learn More"
  },
  "hero": {
    "title": "Transparent Donations & Zero Platform Fee Deductions",
    "subtitle": "An open-source crowdfunding platform built for trusted zakat and humanitarian institutions."
  }
}
```

---

## ⚙️ SvelteKit Lightweight i18n Implementation

To maintain zero-dependency pre-rendering on Cloudflare Pages, we use a simple Svelte Store for localization:

```javascript
// frontend/landing/src/lib/i18n.js
import { writable, derived } from 'svelte/store';
import idTranslations from './locales/id.json';
import enTranslations from './locales/en.json';

const translations = {
  id: idTranslations,
  en: enTranslations
};

// Default to Indonesian ('id')
export const locale = writable('id');

export const t = derived(locale, ($locale) => {
  return (key) => {
    const keys = key.split('.');
    let current = translations[$locale] || translations['id'];
    
    for (const k of keys) {
      if (current[k] === undefined) {
        return key; // Fallback to key itself
      }
      current = current[k];
    }
    return current;
  };
});
```

### Usage in Svelte Component:
```html
<script>
  import { t, locale } from '$lib/i18n';
</script>

<h1>{$t('hero.title')}</h1>
<p>{$t('hero.subtitle')}</p>

<button on:click={() => locale.set('en')}>English</button>
<button on:click={() => locale.set('id')}>Bahasa Indonesia</button>
```

---

## 🏢 Backend Accept-Language Header

When sending notification templates (e.g., Zakat Receipt, Payment Instructions) directly from the Go server via Email, the backend uses the standard HTTP `Accept-Language` header to determine the correct language.

### Go Language Detection Example
```go
func GetPreferredLanguage(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if strings.HasPrefix(acceptLang, "en") {
		return "en"
	}
	return "id" // default fallback
}
```
Based on this language code, the backend loads the corresponding HTML email template (e.g., `receipt_id.html` or `receipt_en.html`).
