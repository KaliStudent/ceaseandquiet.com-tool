export const US_STATES=Object.freeze([['AL','Alabama'],['AK','Alaska'],['AZ','Arizona'],['AR','Arkansas'],['CA','California'],['CO','Colorado'],['CT','Connecticut'],['DE','Delaware'],['DC','District of Columbia'],['FL','Florida'],['GA','Georgia'],['HI','Hawaii'],['ID','Idaho'],['IL','Illinois'],['IN','Indiana'],['IA','Iowa'],['KS','Kansas'],['KY','Kentucky'],['LA','Louisiana'],['ME','Maine'],['MD','Maryland'],['MA','Massachusetts'],['MI','Michigan'],['MN','Minnesota'],['MS','Mississippi'],['MO','Missouri'],['MT','Montana'],['NE','Nebraska'],['NV','Nevada'],['NH','New Hampshire'],['NJ','New Jersey'],['NM','New Mexico'],['NY','New York'],['NC','North Carolina'],['ND','North Dakota'],['OH','Ohio'],['OK','Oklahoma'],['OR','Oregon'],['PA','Pennsylvania'],['RI','Rhode Island'],['SC','South Carolina'],['SD','South Dakota'],['TN','Tennessee'],['TX','Texas'],['UT','Utah'],['VT','Vermont'],['VA','Virginia'],['WA','Washington'],['WV','West Virginia'],['WI','Wisconsin'],['WY','Wyoming'],['AS','American Samoa'],['GU','Guam'],['MP','Northern Mariana Islands'],['PR','Puerto Rico'],['VI','US Virgin Islands']]);
export const CONSENT_VERSION='us-scan-removal-v1';
export function submissionPayload(profile,resident){
  if(!resident)throw new Error('Confirm that you are a US resident submitting your own information.');
  if(!US_STATES.some(([code])=>code===profile.region))throw new Error('Select a US state or territory.');
  if(!/^\d{5}(-\d{4})?$/.test(profile.postalCode))throw new Error('Enter a five-digit US ZIP code, with an optional four-digit extension.');
  return {...profile,country:'US',usResident:true,ownInformation:true,authorizeScanAndRemoval:true,consentVersion:CONSENT_VERSION,website:''};
}
export function attachScanWorkflow({form,fields,button,status,resident}){
  let available=false,human=false,busy=false;
  const render=()=>{button.disabled=!available||!human||busy;};
  form.addEventListener('submit',async event=>{
    event.preventDefault();if(busy)return;
    if(!available){status.textContent='Scanning and removal requests are not available yet. Your details have not been submitted.';return;}
    if(!human){status.textContent='Complete the bot check before submitting.';return;}
    for(const name of fields){const field=form.elements.namedItem(name);if(!field.reportValidity())return;}
    busy=true;render();
    try{
      const profile=Object.fromEntries(fields.map(name=>[name,form.elements.namedItem(name).value.trim()]));
      const payload=submissionPayload(profile,resident.checked);payload.website=form.elements.namedItem('website').value;
      const options={method:'POST',mode:'same-origin',credentials:'same-origin',referrerPolicy:'same-origin',cache:'no-store',headers:{'Content-Type':'application/json'},signal:AbortSignal.timeout(15000)};
      const verified=await fetch('/api/v1/form-security/email/status',{...options,body:JSON.stringify({email:profile.email})});
      if(!verified.ok||!(await verified.json()).verified)throw new Error('Verify this email address before submitting.');
      const response=await fetch('/api/v1/privacy-scan',{...options,body:JSON.stringify(payload)});
      if(!response.ok)throw new Error(response.status===503?'The service is not available. No successful submission is being reported.':'The submission was refused. Check your details and verification.');
      const result=await response.json();if(response.status!==202||!result.jobId)throw new Error('The service did not confirm a scan.');
      form.reset();status.textContent='Your authorized scan and removal workflow was accepted. Reference: '+result.jobId;
    }catch(error){status.textContent=error.message||'The submission could not be confirmed.';}
    finally{busy=false;render();}
  });
  void(async()=>{try{const response=await fetch('/api/v1/privacy-scan/status',{cache:'no-store'});if(!response.ok)throw new Error();const config=await response.json();available=config.available===true&&config.country==='US'&&config.consentVersion===CONSENT_VERSION;status.textContent=available?'Submitting authorizes both the scan and removal requests. Verify your email before proceeding.':'The scan and removal service is being prepared for launch. Your details will not be submitted while it is unavailable.';}catch{status.textContent='Service availability could not be checked. Your details have not been submitted.';}render();})();
  render();return{markHuman(){human=true;render();}};
}
