export function newsArticles(items) {
  if (!Array.isArray(items)) return [];
  return items.filter(item => {
    try { const url=new URL(item.url); return typeof item.title==='string'&&item.title.length>0&&item.title.length<=300&&url.protocol==='https:'&&['eff.org','www.eff.org'].includes(url.host)&&!url.username&&!url.password&&!Number.isNaN(Date.parse(item.published)); }
    catch {return false;}
  }).slice(0,6);
}
export async function init() {
  const status=document.querySelector('#news-status'), list=document.querySelector('#news-articles');
  try {
    const response=await fetch('/api/v1/privacy-news',{cache:'no-store',signal:AbortSignal.timeout(8000)});
    if(!response.ok)throw new Error('News unavailable');
    const data=await response.json(),items=newsArticles(data.items);
    if(!items.length)throw new Error('No recent privacy headlines');
    for(const item of items){
      const card=document.createElement('article');card.className='news-card';
      const source=document.createElement('p');source.className='resource-label';source.textContent='EFF · PRIVACY NEWS';
      const date=document.createElement('time');date.dateTime=item.published;date.textContent=new Intl.DateTimeFormat('en-US',{dateStyle:'medium'}).format(new Date(item.published));
      const title=document.createElement('h3'),link=document.createElement('a');link.href=item.url;link.target='_blank';link.rel='noopener noreferrer';link.textContent=item.title;title.append(link);
      const more=document.createElement('span');more.className='news-arrow';more.textContent='Read the story ↗';more.setAttribute('aria-hidden','true');
      card.append(source,date,title,more);list.append(card);
    }
    status.textContent=data.stale?'Latest saved headlines. The publisher feed is temporarily unavailable.':'Recent privacy headlines from the Electronic Frontier Foundation.';
  } catch {status.textContent='The live news feed is temporarily unavailable. Use the publisher link below for the latest stories.';}
}
