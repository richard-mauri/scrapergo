// Open article links directly in ScraperGo. No transient publisher tabs.
const MENU_ID = 'scrapergo-open-link';
const DEFAULT_READER = 'https://scrapergo.themauris.org/';
function validURL(address) {
  try { return ['http:', 'https:'].includes(new URL(address).protocol); }
  catch { return false; }
}
async function readerOrigin() {
  const value = (await browser.storage.local.get('readerURL')).readerURL || DEFAULT_READER;
  try {
    const u = new URL(value);
    if (!['http:', 'https:'].includes(u.protocol)) throw Error();
    return u.origin + '/';
  } catch { return DEFAULT_READER; }
}
async function openReader(articleURL) {
  const base = await readerOrigin();
  await browser.tabs.create({url: base + '?url=' + encodeURIComponent(articleURL)});
}
function installMenu() {
  browser.contextMenus.removeAll()
    .then(() => browser.contextMenus.create({id: MENU_ID, title: 'Open Link in ScraperGo', contexts: ['link']}))
    .catch(console.error);
}
browser.runtime.onInstalled.addListener(installMenu);
browser.runtime.onStartup.addListener(installMenu);
installMenu();
browser.contextMenus.onClicked.addListener(async info => {
  if (info.menuItemId !== MENU_ID || !validURL(info.linkUrl)) return;
  try { await openReader(info.linkUrl); }
  catch (err) { console.error('ScraperGo could not open reader:', err); }
});
