const search = document.querySelector('#broker-search');
const results = document.querySelector('#broker-results');
const status = document.querySelector('#directory-status');
async function loadBrokers() {
  status.textContent = 'Loading directory…';
  results.replaceChildren();
  try {
    const response = await fetch(`/api/v1/brokers?limit=25&q=${encodeURIComponent(search.value.trim())}`, {cache:'no-store'});
    if (!response.ok) throw new Error('Directory unavailable');
    const data = await response.json();
    for (const item of data.items) {
      const row = document.createElement('li');
      const name = document.createElement('strong'); name.textContent = item.name;
      const detail = document.createElement('small'); detail.textContent = `${item.region || 'region unknown'} · ${item.channelStatus}`;
      row.append(name,detail); results.append(row);
    }
    status.textContent = `${data.count} shown · Contact channels unverified`;
  } catch { status.textContent = 'Directory unavailable in this preview.'; }
}
document.querySelector('#search-button').addEventListener('click',loadBrokers);
search.addEventListener('keydown',event=>{if(event.key==='Enter'){event.preventDefault();loadBrokers();}});
loadBrokers();
