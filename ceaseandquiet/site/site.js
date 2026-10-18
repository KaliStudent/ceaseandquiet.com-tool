import {attachAddressSuggestions} from './address.mjs';
import {lookupWebsiteAddress, attachBirthDatePicker} from './form-tools.mjs';
import {FIELDS, encryptProfile, decryptProfile} from './vault.mjs';
import {attachWebsiteVault} from './profile-vault.mjs';
import {attachFormSecurity} from './form-security.mjs';

const search = document.querySelector('#broker-search');
const region = document.querySelector('#region-filter');
const results = document.querySelector('#broker-results');
const status = document.querySelector('#directory-status');
const detail = document.querySelector('#broker-detail');
const previous = document.querySelector('#previous');
const next = document.querySelector('#next');
let offset = 0, query = '', selectedRegion = '', generation = 0, controller;

async function loadBrokers() {
  const request = ++generation;
  controller?.abort(); controller = new AbortController();
  previous.disabled = true; next.disabled = true;
  status.textContent = 'Loading directory…';
  results.replaceChildren(); detail.hidden = true;
  try {
    const params = new URLSearchParams({limit: '12', offset: String(offset), q: query, region: selectedRegion});
    const response = await fetch('/api/v1/brokers?' + params, {cache: 'no-store', signal: controller.signal});
    if (!response.ok) throw new Error('Directory unavailable');
    const data = await response.json();
    if (request !== generation) return;
    for (const item of data.items) {
      const row = document.createElement('li');
      const button = document.createElement('button'); button.type = 'button';
      const name = document.createElement('strong'); name.textContent = item.name;
      const meta = document.createElement('small'); meta.textContent = (item.region || 'region unknown').toUpperCase() + ' · Channel unreviewed';
      button.append(name, meta); button.addEventListener('click', () => showBroker(item));
      row.append(button); results.append(row);
    }
    status.textContent = data.total ? data.total + ' matching broker records · select a record for details' : 'No brokers match. Try a different name or region.';
    document.querySelector('#page-label').textContent = data.total ? (data.offset + 1) + '–' + (data.offset + data.count) + ' of ' + data.total : '0 results';
    previous.disabled = offset === 0; next.disabled = !data.hasMore;
  } catch (error) {
    if (request !== generation || error.name === 'AbortError') return;
    status.textContent = 'Directory unavailable. Try Search again.';
    document.querySelector('#page-label').textContent = '';
  }
}

function showBroker(item) {
  document.querySelector('#detail-name').textContent = item.name;
  const fields = document.querySelector('#detail-fields'); fields.replaceChildren();
  for (const [label, value] of [['Region', item.region || 'Unknown'], ['Category', item.category || 'Uncategorized'], ['Procedure review', 'Pending']]) {
    const term = document.createElement('dt'); term.textContent = label;
    const description = document.createElement('dd'); description.textContent = value;
    fields.append(term, description);
  }
  detail.hidden = false; detail.focus();
}
document.querySelector('#directory-search').addEventListener('submit', event => {
  event.preventDefault(); query = search.value.trim(); selectedRegion = region.value; offset = 0; loadBrokers();
});
previous.addEventListener('click', () => { offset = Math.max(0, offset - 12); loadBrokers(); });
next.addEventListener('click', () => { offset += 12; loadBrokers(); });
document.querySelector('#close-detail').addEventListener('click', () => { detail.hidden = true; search.focus(); });

const form = document.querySelector('#cq-demo-form');
const formStatus = document.querySelector('#form-status');
const suggestions = attachAddressSuggestions({address: form.elements.namedItem('address'), city: form.elements.namedItem('city'), region: form.elements.namedItem('region'), postalCode: form.elements.namedItem('postalCode'), enabled: document.querySelector('#address-enabled'), results: document.querySelector('#address-results'), status: document.querySelector('#address-status'), lookup: lookupWebsiteAddress});
attachBirthDatePicker(form.elements.namedItem('dateOfBirth'));
void attachFormSecurity({form, container:document.querySelector('#bot-check'), status:document.querySelector('#bot-status'), emailStatus:document.querySelector('#email-verification-status'), code:document.querySelector('#email-code'), send:document.querySelector('#email-send-code'), confirm:document.querySelector('#email-confirm-code'), clearSuggestions:suggestions.clear, onReady:()=>{ const address=form.elements.namedItem('address'); if(address.value.trim().length>=4)address.dispatchEvent(new Event('input',{bubbles:true})); }});
let browserStorage;
try { browserStorage = window.localStorage; } catch { /* The vault reports unavailable storage below. */ }
const websiteVault = attachWebsiteVault({form, fields: FIELDS, encrypt: encryptProfile, decrypt: decryptProfile, storage: browserStorage, clearSuggestions: suggestions.clear, controls: {
  passphrase: document.querySelector('#website-passphrase'), mode: document.querySelector('#website-vault-mode'), status: document.querySelector('#website-vault-status'),
  save: document.querySelector('#website-save'), unlock: document.querySelector('#website-unlock'), lock: document.querySelector('#website-lock'), erase: document.querySelector('#website-erase')
}});

let clearTimer;
function clearForm() { form.reset(); }
form.addEventListener('submit', event => { event.preventDefault(); formStatus.textContent = 'This preview never submits. Use Clear form when finished.'; });
form.addEventListener('input', () => {
  clearTimeout(clearTimer);
  clearTimer = setTimeout(clearForm, 120000);
  formStatus.textContent = 'Details are only in this page. They clear after two minutes.';
});
form.addEventListener('reset', () => {
  clearTimeout(clearTimer); suggestions.clear(); document.querySelector('#address-status').textContent = 'Suggestions are enabled. Geoapify receives the partial address you type. Manual entry always works.'; formStatus.textContent = 'Form cleared. Nothing was submitted.';
});
window.addEventListener('pagehide', clearForm);
window.addEventListener('pageshow', event => { if (event.persisted) { clearForm(); websiteVault.lock(); } });
document.addEventListener('visibilitychange', () => { if (document.hidden) clearForm(); });
loadBrokers();
