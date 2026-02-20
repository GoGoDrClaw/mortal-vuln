const API_URL = `${window.location.protocol}//api.${window.location.hostname.replace(/^(www\.|api\.|dashboard\.|evil\.)/, '')}${window.location.port ? ':' + window.location.port : ''}`;
let TOKEN = null;

// ── All 23 UMK3 characters ─────────────────────────────────────────
const ALL_CHARS = {
  scorpion:       { label: 'SCORPION',         emoji: '🦂', hex: '#f59e0b', rgb: '245,158,11',  img: '/chars/scorpion.jpg' },
  subzero:        { label: 'SUB-ZERO',          emoji: '❄️',  hex: '#3b82f6', rgb: '59,130,246',  img: '/chars/subzero.gif' },
  liukang:        { label: 'LIU KANG',          emoji: '🔥', hex: '#ef4444', rgb: '239,68,68',   img: '/chars/liukang.gif' },
  kitana:         { label: 'KITANA',            emoji: '👸', hex: '#8b5cf6', rgb: '139,92,246',  img: '/chars/kitana.jpg' },
  raiden:         { label: 'RAIDEN',            emoji: '⚡', hex: '#a78bfa', rgb: '167,139,250', img: '/chars/raiden.jpg' },
  jax:            { label: 'JAX',               emoji: '💪', hex: '#22c55e', rgb: '34,197,94',   img: '/chars/jax.gif' },
  mileena:        { label: 'MILEENA',           emoji: '🎭', hex: '#ec4899', rgb: '236,72,153',  img: '/chars/mileena.jpg' },
  kunglao:        { label: 'KUNG LAO',          emoji: '🎩', hex: '#84cc16', rgb: '132,204,22',  img: '/chars/kunglao.gif' },
  sonya:          { label: 'SONYA',             emoji: '🎖️', hex: '#f472b6', rgb: '244,114,182', img: '/chars/sonya.gif' },
  shangtsung:     { label: 'SHANG TSUNG',       emoji: '💀', hex: '#f97316', rgb: '249,115,22',  img: '/chars/shangtsung.gif' },
  kano:           { label: 'KANO',              emoji: '🔺', hex: '#78716c', rgb: '120,113,108', img: '/chars/kano.gif' },
  nightwolf:      { label: 'NIGHTWOLF',         emoji: '🐺', hex: '#10b981', rgb: '16,185,129',  img: '/chars/nightwolf.gif' },
  cyrax:          { label: 'CYRAX',             emoji: '🤖', hex: '#eab308', rgb: '234,179,8',   img: '/chars/cyrax.gif' },
  sektor:         { label: 'SEKTOR',            emoji: '🚀', hex: '#dc2626', rgb: '220,38,38',   img: '/chars/sektor.gif' },
  kabal:          { label: 'KABAL',             emoji: '⚔️',  hex: '#64748b', rgb: '100,116,139', img: '/chars/kabal.gif' },
  jade:           { label: 'JADE',              emoji: '💚', hex: '#4ade80', rgb: '74,222,128',  img: '/chars/jade.gif' },
  sindel:         { label: 'SINDEL',            emoji: '👑', hex: '#c026d3', rgb: '192,38,211',  img: '/chars/sindel.gif' },
  ermac:          { label: 'ERMAC',             emoji: '👻', hex: '#b91c1c', rgb: '185,28,28',   img: '/chars/ermac.gif' },
  sheeva:         { label: 'SHEEVA',            emoji: '👊', hex: '#d97706', rgb: '217,119,6',   img: '/chars/sheeva.gif' },
  stryker:        { label: 'STRYKER',           emoji: '🚔', hex: '#60a5fa', rgb: '96,165,250',  img: '/chars/stryker.gif' },
  smoke:          { label: 'SMOKE',             emoji: '💨', hex: '#9ca3af', rgb: '156,163,175', img: '/chars/smoke.gif' },
  noobsaibot:     { label: 'NOOB SAIBOT',       emoji: '🌑', hex: '#6366f1', rgb: '99,102,241',  img: '/chars/noobsaibot.gif' },
};

// Convenience maps (used by teamBadge etc.)
const COLORS = Object.fromEntries(Object.entries(ALL_CHARS).map(([k, v]) => [k, v.hex]));
const EMOJI  = Object.fromEntries(Object.entries(ALL_CHARS).map(([k, v]) => [k, v.emoji]));

let ACTIVE_TEAMS = [];
let MK_COLS = 3;
let mkCursor = 0;

// ── Konami Code easter egg ─────────────────────────────────────────
let _k = [];
const _s = atob('QXJyb3dVcCxBcnJvd1VwLEFycm93RG93bixBcnJvd0Rvd24sQXJyb3dMZWZ0LEFycm93UmlnaHQsQXJyb3dMZWZ0LEFycm93UmlnaHQsYixh').split(',');
let _u = false;

document.addEventListener('keydown', function(e) {
  _k.push(e.key);
  _k = _k.slice(-10);

  if (_k.join(',') === _s.join(',') && !_u) {
    _u = true;
    const _h = h => h.match(/.{2}/g).map(x => String.fromCharCode(parseInt(x, 16))).join('');
    const _m = [
      [_h('f09f8eb020') + _h('4d4f5254414c204b4f4d42415421') + _h('20f09f8eb0'), 'font-size: 24px; color: #ff0000; font-weight: bold; text-shadow: 2px 2px 4px #000;'],
      [_h('596f75206861766520756e6c6f636b65642074686520616e6369656e74206b6f6d626174207365637265747321'), 'font-size: 14px; color: #ffff00;'],
      [_h('437265617465206120') + _h('6e6f74652077697468207469746c653a202246494e4953482048494d22'), 'font-size: 12px; color: #00ff00;'],
      [_h('416e6420636f6e74656e743a2022') + '\u2191\u2191\u2193\u2193\u2190\u2192\u2190\u2192BA"', 'font-size: 12px; color: #00ff00;']
    ];
    _m.forEach(([t, s]) => console.log('%c' + t, s));
    if (document.body) {
      const _f = document.createElement('div');
      _f.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;background:red;opacity:0.5;pointer-events:none;z-index:99999;';
      document.body.appendChild(_f);
      setTimeout(() => _f.remove(), 100);
    }
  }
});

// ── Helpers ────────────────────────────────────────────────────────
function getTeam() {
  return document.cookie.split(';')
    .map(c => c.trim())
    .find(c => c.startsWith('team='))
    ?.split('=')[1] || '';
}

function teamBadge(team) {
  const c = ALL_CHARS[team];
  return `<span class="team-badge" style="background:${c?.hex || '#888'}">${c?.emoji || ''} ${team.toUpperCase()}</span>`;
}

// ── Init ───────────────────────────────────────────────────────────
async function init() {
  // Load active team list from backend
  try {
    const res = await fetch(`${API_URL}/api/teams`);
    ACTIVE_TEAMS = await res.json();
  } catch (e) {
    ACTIVE_TEAMS = ['scorpion', 'subzero', 'liukang', 'kitana', 'raiden', 'jax'];
  }

  renderTeamSelect();

  const t = getTeam();
  if (!t || !ALL_CHARS[t]) return; // no team → show character select

  // Restore session: try localStorage first (survives SameSite cookie issues),
  // then fall back to token cookie
  const saved = localStorage.getItem('mk_token') ||
    document.cookie.split(';').map(c => c.trim()).find(c => c.startsWith('token='))?.split('=')[1];

  if (saved) {
    try {
      const payload = JSON.parse(atob(saved.split('.')[1]));
      TOKEN = saved;
      document.getElementById('username-display').textContent = payload.username + ' (' + payload.role + ')';
      document.getElementById('app-team-indicator').innerHTML = teamBadge(t);
      document.getElementById('team-section').classList.add('hidden');
      document.getElementById('app-section').classList.remove('hidden');
      loadNotes();
      return;
    } catch (e) {
      localStorage.removeItem('mk_token');
    }
  }

  showLogin(t);
}

init();

// ── Team select ────────────────────────────────────────────────────
function renderTeamSelect() {
  const grid = document.querySelector('.mk-select-grid');
  if (!grid) return;

  grid.innerHTML = ACTIVE_TEAMS.map(team => {
    const c = ALL_CHARS[team];
    if (!c) return '';
    return '<div class="mk-cell" data-team="' + team + '" ' +
      'style="--cc:' + c.hex + ';--cr:' + c.rgb + '" ' +
      'onclick="selectTeam(\'' + team + '\')">' +
        '<div class="mk-cell-inner">' +
          '<span class="mk-fallback">' + c.emoji + '</span>' +
          '<img src="' + c.img + '" class="mk-char-img" alt="' + c.label + '" onerror="this.remove()">' +
        '</div>' +
        '<div class="mk-cell-name">' + c.label + '</div>' +
      '</div>';
  }).join('');

  // Compute columns and apply to grid
  const n = ACTIVE_TEAMS.length;
  MK_COLS = Math.ceil(n / 3);
  grid.style.gridTemplateColumns = `repeat(${MK_COLS}, 1fr)`;

  // Compute cell width so all rows fit on screen without scrolling
  const rows = Math.ceil(n / MK_COLS);
  const reservedH = 220; // header + bottom bar + ticker + gaps + padding
  const nameBarH = 22;   // .mk-cell-name height
  const availH = window.innerHeight - reservedH;
  const cellW = Math.round((availH / rows - nameBarH) * 3 / 4); // aspect-ratio 3:4
  const clampedW = Math.max(90, Math.min(cellW, 210));
  grid.style.setProperty('--cell-w', clampedW + 'px');
}

async function selectTeam(team) {
  document.querySelectorAll('.mk-cell').forEach(el => el.classList.remove('mk-cursor'));
  const cell = document.querySelector(`.mk-cell[data-team="${team}"]`);
  if (cell) cell.classList.add('mk-cursor');

  // Set cookie on frontend domain so document.cookie can read it on refresh
  document.cookie = 'team=' + team + ';path=/';

  document.getElementById('mk-fight-overlay').classList.remove('hidden');

  // API call sets cookie on api.* domain (for CORS/CSRF CTF purposes)
  fetch(`${API_URL}/api/select-team`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ team })
  });

  setTimeout(() => {
    document.getElementById('mk-fight-overlay').classList.add('hidden');
    showLogin(team);
  }, 950);
}

function showLogin(team) {
  document.getElementById('team-section').classList.add('hidden');
  document.getElementById('login-section').classList.remove('hidden');
  document.getElementById('team-indicator').innerHTML = teamBadge(team);
}

function switchTeam() {
  document.cookie = 'team=;Max-Age=0;path=/';
  document.cookie = 'token=;Max-Age=0;path=/';
  localStorage.removeItem('mk_token');
  TOKEN = null;
  document.querySelectorAll('.mk-cell').forEach(el => el.classList.remove('mk-cursor'));
  mkCursor = 0;
  document.getElementById('team-section').classList.remove('hidden');
  document.getElementById('login-section').classList.add('hidden');
  document.getElementById('app-section').classList.add('hidden');
  document.getElementById('login-token').classList.add('hidden');
}

// ── Auth ───────────────────────────────────────────────────────────
async function doLogin() {
  const u = document.getElementById('login-user').value;
  const p = document.getElementById('login-pass').value;

  const res = await fetch(`${API_URL}/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ username: u, password: p })
  });

  const d = await res.json();

  if (!res.ok) {
    document.getElementById('login-error').textContent = d.error || JSON.stringify(d);
    document.getElementById('login-error').classList.remove('hidden');
    return;
  }

  TOKEN = d.token;
  localStorage.setItem('mk_token', TOKEN); // persist across refreshes

  document.getElementById('login-token').classList.remove('hidden');
  document.getElementById('token-text').textContent = TOKEN;

  const payload = JSON.parse(atob(TOKEN.split('.')[1]));
  document.getElementById('username-display').textContent = payload.username + ' (' + payload.role + ')';
  document.getElementById('app-team-indicator').innerHTML = teamBadge(getTeam());

  document.getElementById('login-section').classList.add('hidden');
  document.getElementById('app-section').classList.remove('hidden');

  loadNotes();
}

function doLogout() {
  TOKEN = null;
  localStorage.removeItem('mk_token');
  document.cookie = 'token=;Max-Age=0;path=/';
  document.getElementById('login-section').classList.remove('hidden');
  document.getElementById('login-section').querySelector('#login-token').classList.add('hidden');
  document.getElementById('app-section').classList.add('hidden');
  document.getElementById('team-indicator').innerHTML = teamBadge(getTeam());
}

async function resetDB() {
  if (!confirm('Reset your team database to its initial state?')) return;
  await fetch(`${API_URL}/api/reset`, {
    method: 'POST',
    credentials: 'include'
  });
  alert('✅ Database reset!');
  loadNotes();
}

// ── Tabs ───────────────────────────────────────────────────────────
function switchTab(n, el) {
  document.querySelectorAll('[id^=tab-]').forEach(e => e.classList.add('hidden'));
  document.getElementById('tab-' + n).classList.remove('hidden');
  document.querySelectorAll('.tab').forEach(e => e.classList.remove('active'));
  el.classList.add('active');
}

// ── Notes ──────────────────────────────────────────────────────────
async function loadNotes() {
  const res = await fetch(`${API_URL}/api/notes`, {
    headers: { Authorization: 'Bearer ' + TOKEN },
    credentials: 'include'
  });
  const notes = await res.json();
  const el = document.getElementById('notes-list');

  if (!notes || notes.length === 0) {
    el.innerHTML = '<p style="color:#64748b">No notes</p>';
    return;
  }

  // Vulnerable to XSS - intentional for CTF
  el.innerHTML = notes.map(n =>
    '<div class="note" id="note-card-' + n.id + '">' +
      '<h3>' + n.title + '</h3>' +
      '<p id="note-content-' + n.id + '">' + n.content.slice(0, 40) + (n.content.length > 40 ? '…' : '') + '</p>' +
      '<div class="note-meta">' +
        '▸ ID: ' + n.id + ' &nbsp;|&nbsp; ' + (n.created || '') +
        ' <button class="btn-primary" style="margin-left:8px;padding:3px 8px;" onclick="expandNote(' + n.id + ')">▶ EXPAND</button>' +
        ' <button class="btn-danger" style="margin-left:4px;padding:3px 8px;" onclick="deleteNote(' + n.id + ')">💀 DESTROY</button>' +
      '</div>' +
    '</div>'
  ).join('');
}

async function createNote() {
  const t = document.getElementById('note-title').value;
  const c = document.getElementById('note-content').value;

  await fetch(`${API_URL}/api/notes`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: 'Bearer ' + TOKEN
    },
    credentials: 'include',
    body: JSON.stringify({ title: t, content: c })
  });

  document.getElementById('note-title').value = '';
  document.getElementById('note-content').value = '';
  document.querySelector('.tab').click();
  loadNotes();
}

async function deleteNote(id) {
  if (!confirm('Delete this note?')) return;

  await fetch(`${API_URL}/api/notes/delete/${id}`, {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + TOKEN },
    credentials: 'include'
  });

  loadNotes();
}

async function expandNote(id) {
  const btn = document.querySelector('#note-card-' + id + ' .btn-primary');
  if (btn) btn.textContent = '…';

  const res = await fetch(`${API_URL}/api/notes/${id}`, {
    headers: { Authorization: 'Bearer ' + TOKEN },
    credentials: 'include'
  });
  const d = await res.json();

  const contentEl = document.getElementById('note-content-' + id);
  if (contentEl) contentEl.innerHTML = d.content;
  if (btn) btn.style.display = 'none';
}

async function lookupNote() {
  const id = document.getElementById('lookup-id').value;
  const res = await fetch(`${API_URL}/api/notes/${id}`, {
    headers: { Authorization: 'Bearer ' + TOKEN },
    credentials: 'include'
  });
  const d = await res.json();
  const el = document.getElementById('lookup-result');

  if (!res.ok) {
    el.innerHTML = '<p class="error">' + (d.error || 'Error') + '</p>';
    return;
  }

  el.innerHTML = `<div class="note" style="margin-top:10px">
    <h3>${d.title}</h3>
    <p>${d.content}</p>
    <div class="note-meta">▸ USER_ID: ${d.user_id}</div>
  </div>`;
}

// ── MK3 Keyboard navigation ────────────────────────────────────────
function mkUpdateCursor() {
  document.querySelectorAll('.mk-cell').forEach((el, i) => {
    el.classList.toggle('mk-cursor', i === mkCursor);
  });
}

document.addEventListener('keydown', function(e) {
  const ts = document.getElementById('team-section');
  if (!ts || ts.classList.contains('hidden')) return;
  const overlay = document.getElementById('mk-fight-overlay');
  if (overlay && !overlay.classList.contains('hidden')) return;

  let moved = false;
  switch (e.key) {
    case 'ArrowRight': mkCursor = (mkCursor + 1) % ACTIVE_TEAMS.length;                        moved = true; break;
    case 'ArrowLeft':  mkCursor = (mkCursor - 1 + ACTIVE_TEAMS.length) % ACTIVE_TEAMS.length;  moved = true; break;
    case 'ArrowDown':  mkCursor = (mkCursor + MK_COLS) % ACTIVE_TEAMS.length;                  moved = true; break;
    case 'ArrowUp':    mkCursor = (mkCursor - MK_COLS + ACTIVE_TEAMS.length) % ACTIVE_TEAMS.length; moved = true; break;
    case 'Enter':
    case ' ':
      selectTeam(ACTIVE_TEAMS[mkCursor]);
      moved = true;
      break;
  }
  if (moved) {
    e.preventDefault();
    mkUpdateCursor();
  }
});
