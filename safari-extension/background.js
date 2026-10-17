const MENU_ID = 'scrapergo-open-link';
const DEFAULT_READER = 'https://scrapergo.themauris.org/';
function newsURL(address) { try { const u = new URL(address); return u.hostname === 'news.google.com' && /^\/(read|articles)\//.test(u.pathname); } catch { return false; } }
function validURL(address) { try { const u = new URL(address); return ['http:', 'https:'].includes(u.protocol); } catch { return false; } }
async function readerOrigin() {
  const value = (await browser.storage.local.get('readerURL')).readerURL || DEFAULT_READER;
  try { const u = new URL(value); if (!['http:', 'https:'].includes(u.protocol)) throw Error(); return u.origin + '/'; }
  catch { return DEFAULT_READER; }
}
async function openReader(articleURL, article = null) {
  const base = await readerOrigin();
  const tab = await browser.tabs.create({url: base + '?url=' + encodeURIComponent(articleURL) + (article ? '&browser=1' : '')});
  if(article){let tries=0;const deliver=async()=>{if(tries++>18)return;try{await browser.tabs.sendMessage(tab.id,{type:'SCRAPERGO_DELIVER',article,reader:base});}catch(_){setTimeout(deliver,250)}};setTimeout(deliver,450);}
}
function installMenu() {
  browser.contextMenus.removeAll().then(() => browser.contextMenus.create({id: MENU_ID, title: 'Open Link in ScraperGo', contexts: ['link']})).catch(console.error);
}
browser.runtime.onInstalled.addListener(installMenu);
browser.runtime.onStartup.addListener(installMenu);
installMenu();
browser.contextMenus.onClicked.addListener(async info => {
  if (info.menuItemId !== MENU_ID || !validURL(info.linkUrl)) return;
  const original = info.linkUrl;
  // Use a temporary Safari tab for ordinary publisher URLs too, so rendered text can be extracted.
  // Let Safari navigate a Google News link. Some Google links remain on Google's
  // own page; after 9 seconds fall back to ScraperGo's server-side resolver.
  let tab;
  try { tab = await browser.tabs.create({url: original, active: false}); }
  catch { await openReader(original); return; }
  let finished = false;
  let timeout;
  const complete = async chosen => {
    if (finished) return;
    finished = true;
    clearTimeout(timeout);
    browser.tabs.onUpdated.removeListener(observe);
    let article=null;
    if(!newsURL(chosen)){
      try { article=await browser.tabs.sendMessage(tab.id,{type:'SCRAPERGO_EXTRACT'}); } catch (_) {}
    }
    try { await browser.tabs.remove(tab.id); } catch (_) {}
    await openReader(chosen,article);
  };
  const observe = (tabID, change, currentTab) => {
    if (tabID !== tab.id) return;
    const candidate = change.url || currentTab.url;
    if (change.status === 'complete' && validURL(candidate) && !newsURL(candidate) && new URL(candidate).hostname !== 'news.google.com') {
      setTimeout(() => { void complete(candidate); }, 700);
    }
  };
  browser.tabs.onUpdated.addListener(observe);
  timeout = setTimeout(() => { void complete(original); }, 9000);
  // Handle redirects that finished before the event listener was attached.
  try { const current = await browser.tabs.get(tab.id); observe(tab.id, {url: current.url, status: current.status}, current); } catch (_) {}
});
