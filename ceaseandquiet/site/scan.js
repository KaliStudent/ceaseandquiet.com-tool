import {attachAddressSuggestions} from './address.mjs';
import {lookupWebsiteAddress, attachBirthDatePicker} from './form-tools.mjs';
import {FIELDS, encryptProfile, decryptProfile} from './vault.mjs';
import {attachWebsiteVault} from './profile-vault.mjs';
import {attachFormSecurity} from './form-security.mjs';

import {attachScanWorkflow, US_STATES} from './scan-workflow.mjs';
export function init() {
  const state = document.querySelector('[name="region"]');
  for (const [value,label] of US_STATES) { const option=document.createElement('option'); option.value=value;option.textContent=label;state.append(option); }
const form = document.querySelector('#privacy-scan-form');
const formStatus = document.querySelector('#form-status');
const suggestions = attachAddressSuggestions({address: form.elements.namedItem('address'), city: form.elements.namedItem('city'), region: form.elements.namedItem('region'), postalCode: form.elements.namedItem('postalCode'), enabled: document.querySelector('#address-enabled'), results: document.querySelector('#address-results'), status: document.querySelector('#address-status'), lookup: lookupWebsiteAddress});
attachBirthDatePicker(form.elements.namedItem('dateOfBirth'));
const scanWorkflow=attachScanWorkflow({form,fields:FIELDS,button:document.querySelector('#scan-submit'),status:document.querySelector('#scan-status'),resident:document.querySelector('#us-resident')});
void attachFormSecurity({form, container:document.querySelector('#bot-check'), status:document.querySelector('#bot-status'), retry:document.querySelector('#bot-retry'), emailStatus:document.querySelector('#email-verification-status'), code:document.querySelector('#email-code'), send:document.querySelector('#email-send-code'), confirm:document.querySelector('#email-confirm-code'), clearSuggestions:suggestions.clear, onReady:()=>{ scanWorkflow.markHuman(); const address=form.elements.namedItem('address'); if(address.value.trim().length>=4)address.dispatchEvent(new Event('input',{bubbles:true})); }});
let browserStorage;
try { browserStorage = window.localStorage; } catch { /* The vault reports unavailable storage below. */ }
const websiteVault = attachWebsiteVault({form, fields: FIELDS, encrypt: encryptProfile, decrypt: decryptProfile, storage: browserStorage, clearSuggestions: suggestions.clear, controls: {
  passphrase: document.querySelector('#website-passphrase'), mode: document.querySelector('#website-vault-mode'), status: document.querySelector('#website-vault-status'),
  save: document.querySelector('#website-save'), unlock: document.querySelector('#website-unlock'), lock: document.querySelector('#website-lock'), erase: document.querySelector('#website-erase')
}});

let clearTimer;
function clearForm() { form.reset(); }
form.addEventListener('input', () => {
  clearTimeout(clearTimer);
  clearTimer = setTimeout(clearForm, 900000);
  formStatus.textContent = 'Details are only in this page. They clear after 15 minutes of inactivity.';
});
form.addEventListener('reset', () => {
  clearTimeout(clearTimer); suggestions.clear(); document.querySelector('#address-status').textContent = 'Suggestions are enabled. Geoapify receives the partial address you type. Manual entry always works.'; formStatus.textContent = 'Form cleared. Nothing was submitted.';
});
window.addEventListener('pagehide', clearForm);
window.addEventListener('pageshow', event => { if (event.persisted) { clearForm(); websiteVault.lock(); } });


}
