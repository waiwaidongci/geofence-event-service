const byId = id => document.getElementById(id);
const loadJSON = async path => {
  const response = await fetch(path);
  if (!response.ok) throw new Error(`${response.status}`);
  return response.json();
};
const count = async (path, target) => {
  const payload = await loadJSON(path);
  byId(target).textContent = payload.total ?? payload.items?.length ?? 0;
  return payload.items ?? [];
};
async function refresh() {
  try {
    await loadJSON('/healthz');
    byId('health').classList.add('ready');
    byId('health').lastChild.textContent = 'Online';
    await Promise.all([
      count('/api/v1/terminals?limit=1', 'terminalCount'),
      count('/api/v1/geofences', 'geofenceCount')
    ]);
    const events = await count('/api/v1/events?status=open&limit=8', 'eventCount');
    byId('events').innerHTML = events.length ? events.map(event => `<li><div class="event-title"><span>${event.type}</span><span>${event.status}</span></div><div class="event-meta">${event.terminal_id} / ${event.geofence_id}</div></li>`).join('') : '<li class="empty">No open events</li>';
  } catch (error) {
    byId('health').classList.remove('ready');
    byId('health').lastChild.textContent = 'Unavailable';
  }
}
byId('refresh').addEventListener('click', refresh);
refresh();
