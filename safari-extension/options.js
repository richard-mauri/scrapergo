const input=document.getElementById('readerURL');const status=document.getElementById('status');
browser.storage.local.get('readerURL').then(x => { input.value=x.readerURL || 'https://scrapergo.themauris.org/'; });
document.getElementById('save').onclick=async () => { try { const u=new URL(input.value);if(!['https:','http:'].includes(u.protocol))throw Error('Enter an HTTP or HTTPS URL');await browser.storage.local.set({readerURL:u.origin+'/'});status.textContent='Saved. New right-click actions will use this address.';}catch(e){status.textContent='Invalid URL: '+e.message;} };
