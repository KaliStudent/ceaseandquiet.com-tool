import test from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {submissionPayload,US_STATES} from '../site/scan-workflow.mjs';
import {partnerLinks} from '../site/consumer.js';
import {newsArticles} from '../site/home.js';

test('consumer website has three distinct pages and privacy news follows the service explanation',async()=>{
  const home=await readFile(new URL('../site/index.html',import.meta.url),'utf8');
  const scan=await readFile(new URL('../site/privacy-scan.html',import.meta.url),'utf8');
  const protection=await readFile(new URL('../site/consumer-protection.html',import.meta.url),'utf8');
  for(const html of [home,scan,protection]){
    for(const path of ['/privacy-scan/','/consumer-protection/'])assert.ok(html.includes('href="'+path+'"'));
    const ids=[...html.matchAll(/\bid="([^"]+)"/g)].map(match=>match[1]);assert.equal(new Set(ids).size,ids.length);
    assert.match(html,/US RESIDENTS ONLY/);
  }
  assert.ok(home.indexOf('id="privacy-news"')>home.indexOf('id="service"'));
  assert.doesNotMatch(home,/id="privacy-scan-form"|Local vault|Open the extension/);
  assert.match(scan,/Submitting this form authorizes both actions/);assert.match(scan,/What is a data broker/);
  assert.match(protection,/AnnualCreditReport/);assert.match(protection,/Experian/);assert.match(protection,/Have I Been Pwned/);assert.match(protection,/Affiliate disclosure/);
});
test('scan payload requires US residency and authorizes the scan and removal together',()=>{
  const profile={region:'CA',postalCode:'00000'};
  assert.throws(()=>submissionPayload(profile,false),/US resident/);
  assert.throws(()=>submissionPayload({...profile,region:'ON'},true),/US state/);
  assert.throws(()=>submissionPayload({...profile,postalCode:'non-US'},true),/ZIP/);
  const input=submissionPayload(profile,true);assert.equal(input.authorizeScanAndRemoval,true);assert.equal(input.country,'US');assert.equal(input.ownInformation,true);
  assert.equal(US_STATES.length,56);
});
test('affiliate links remain exact, are user configured, and cannot execute scripts',async()=>{
  const link={name:'Example partner',url:'https://example.test/offer?affiliate=synthetic',description:'Example only'};
  assert.equal(partnerLinks([link])[0].url,link.url);
  assert.equal(partnerLinks([{...link,url:'javascript:alert(1)'}]).length,0);
  assert.deepEqual(JSON.parse(await readFile(new URL('../site/affiliate-links.json',import.meta.url),'utf8')),[]);
});
test('news cards allow only publisher HTTPS links and valid publication dates',()=>{
  const item={title:'Privacy news',url:'https://www.eff.org/deeplinks/example',published:'2026-10-19T00:00:00Z'};
  assert.equal(newsArticles([item]).length,1);assert.equal(newsArticles([{...item,url:'https://unrelated.example/'}]).length,0);assert.equal(newsArticles([{...item,published:'invalid'}]).length,0);
});
