export function partnerLinks(items) {
  if(!Array.isArray(items))return [];
  return items.filter(item=>{
    try{const url=new URL(item.url);return url.protocol==='https:'&&!url.username&&!url.password&&typeof item.name==='string'&&item.name.length>0&&item.name.length<=80&&typeof item.description==='string'&&item.description.length<=400;}
    catch{return false;}
  }).slice(0,8);
}
export async function init(){
  try{
    const response=await fetch('/affiliate-links.json',{cache:'no-store'});if(!response.ok)return;
    const items=partnerLinks(await response.json());if(!items.length)return;
    const section=document.querySelector('#partner-resources'),list=document.querySelector('#partner-cards');
    for(const item of items){const card=document.createElement('article');card.className='panel resource-card';const title=document.createElement('h3');title.textContent=item.name;const copy=document.createElement('p');copy.textContent=item.description;const link=document.createElement('a');link.className='glow-button secondary';link.href=item.url;link.target='_blank';link.rel='sponsored noopener noreferrer';link.textContent='Visit '+item.name+' ↗';card.append(title,copy,link);list.append(card);}
    section.hidden=false;
  }catch{/* The official free resources remain usable independently. */}
}
