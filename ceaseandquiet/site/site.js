const pages = {home: () => import('./home.js'), scan: () => import('./scan.js'), protection: () => import('./consumer.js')};
const load = pages[document.body.dataset.page];
if (load) { const page = await load(); page.init(); }
