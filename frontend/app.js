const API_URL = `${window.location.protocol}//api.${window.location.hostname.replace(/^(www\.|api\.|dashboard\.|evil\.)/, '')}${window.location.port ? ':' + window.location.port : ''}`;
let TOKEN = null;
let SESSION = null; // { saveCode, sessionId, nickname, character }
let selectedChar = null;

// ── All 23 UMK3 characters ─────────────────────────────────────────
const ALL_CHARS = {
  scorpion:       { label: 'SCORPION', hex: '#f59e0b', rgb: '245,158,11',  img: '/chars/scorpion.gif' },
  subzero:        { label: 'SUB-ZERO', hex: '#3b82f6', rgb: '59,130,246',  img: '/chars/subzero.gif' },
  liukang:        { label: 'LIU KANG', hex: '#ef4444', rgb: '239,68,68',   img: '/chars/liukang.gif' },
  kitana:         { label: 'KITANA', hex: '#8b5cf6', rgb: '139,92,246',  img: '/chars/kitana.gif' },
  raiden:         { label: 'RAIDEN', hex: '#a78bfa', rgb: '167,139,250', img: '/chars/raiden.gif' },
  jax:            { label: 'JAX', hex: '#22c55e', rgb: '34,197,94',   img: '/chars/jax.gif' },
  mileena:        { label: 'MILEENA', hex: '#ec4899', rgb: '236,72,153',  img: '/chars/mileena.gif' },
  kunglao:        { label: 'KUNG LAO', hex: '#84cc16', rgb: '132,204,22',  img: '/chars/kunglao.gif' },
  sonya:          { label: 'SONYA', hex: '#f472b6', rgb: '244,114,182', img: '/chars/sonya.gif' },
  shangtsung:     { label: 'SHANG TSUNG', hex: '#f97316', rgb: '249,115,22',  img: '/chars/shangtsung.gif' },
  kano:           { label: 'KANO', hex: '#78716c', rgb: '120,113,108', img: '/chars/kano.gif' },
  nightwolf:      { label: 'NIGHTWOLF', hex: '#10b981', rgb: '16,185,129',  img: '/chars/nightwolf.gif' },
  cyrax:          { label: 'CYRAX', hex: '#eab308', rgb: '234,179,8',   img: '/chars/cyrax.gif' },
  sektor:         { label: 'SEKTOR', hex: '#dc2626', rgb: '220,38,38',   img: '/chars/sektor.gif' },
  kabal:          { label: 'KABAL', hex: '#64748b', rgb: '100,116,139', img: '/chars/kabal.gif' },
  jade:           { label: 'JADE', hex: '#4ade80', rgb: '74,222,128',  img: '/chars/jade.gif' },
  sindel:         { label: 'SINDEL', hex: '#c026d3', rgb: '192,38,211',  img: '/chars/sindel.gif' },
  ermac:          { label: 'ERMAC', hex: '#b91c1c', rgb: '185,28,28',   img: '/chars/ermac.gif' },
  sheeva:         { label: 'SHEEVA', hex: '#d97706', rgb: '217,119,6',   img: '/chars/sheeva.gif' },
  stryker:        { label: 'STRYKER', hex: '#60a5fa', rgb: '96,165,250',  img: '/chars/stryker.gif' },
  smoke:          { label: 'SMOKE', hex: '#9ca3af', rgb: '156,163,175', img: '/chars/smoke.gif' },
  noobsaibot:     { label: 'NOOB SAIBOT', hex: '#6366f1', rgb: '99,102,241',  img: '/chars/noobsaibot.gif' },
  motaro:         { label: 'MOTARO',      hex: '#c2410c', rgb: '194,65,12',   img: '/chars/motaro.gif' },
  shaokahn:       { label: 'SHAO KAHN',  hex: '#7c3aed', rgb: '124,58,237',  img: '/chars/shaokahn.gif' },
};

const CHARS_LIST = Object.keys(ALL_CHARS);
let MK_COLS = 5;
let mkCursor = 0;

// ── Konami Code easter egg ─────────────────────────────────────────
let _k = [];
const _s = atob('QXJyb3dVcCxBcnJvd1VwLEFycm93RG93bixBcnJvd0Rvd24sQXJyb3dMZWZ0LEFycm93UmlnaHQsQXJyb3dMZWZ0LEFycm93UmlnaHQsYixh').split(',');
let _u = false;
document.addEventListener('keydown', function(e) {
  _k.push(e.key); _k = _k.slice(-10);
  if (_k.join(',') === _s.join(',') && !_u) {
    _u = true;
    const _h = h => h.match(/.{2}/g).map(x => String.fromCharCode(parseInt(x, 16))).join('');
    [
      [_h('f09f8eb020') + _h('4d4f5254414c204b4f4d42415421') + _h('20f09f8eb0'), 'font-size:24px;color:#ff0000;font-weight:bold'],
      [_h('596f75206861766520756e6c6f636b65642074686520616e6369656e74206b6f6d626174207365637265747321'), 'font-size:14px;color:#ffff00'],
      [_h('437265617465206120') + _h('6e6f74652077697468207469746c653a202246494e4953482048494d22'), 'font-size:12px;color:#00ff00'],
      [_h('416e6420636f6e74656e743a2022') + '\u2191\u2191\u2193\u2193\u2190\u2192\u2190\u2192BA"', 'font-size:12px;color:#00ff00']
    ].forEach(([t, s]) => console.log('%c' + t, s));
  }
});

// ── Helpers ────────────────────────────────────────────────────────
function getSessionID() {
  return document.cookie.split(';')
    .map(c => c.trim())
    .find(c => c.startsWith('session_id='))
    ?.split('=')[1] || '';
}

function charBadge(character) {
  const c = ALL_CHARS[character];
  return `<span class="team-badge" style="background:${c?.hex || '#888'}">${(c?.label || character).toUpperCase()}</span>`;
}

function hideAll() {
  ['start-section', 'restore-section', 'savecode-section', 'team-section',
   'login-section', 'app-section'].forEach(id => {
    document.getElementById(id)?.classList.add('hidden');
  });
}

function show(id) {
  hideAll();
  document.getElementById(id)?.classList.remove('hidden');
}

// ── Init ───────────────────────────────────────────────────────────
async function init() {
  renderCharSelect();

  // 1. Valid session cookie → check with server
  const sid = getSessionID();
  if (sid) {
    try {
      const res = await fetch(`${API_URL}/api/session/check`, { credentials: 'include' });
      if (res.ok) {
        SESSION = await res.json();
        const saved = localStorage.getItem('mk_token');
        if (saved) {
          try {
            const payload = JSON.parse(atob(saved.split('.')[1]));
            // Check token not expired
            if (!payload.exp || payload.exp * 1000 > Date.now()) {
              TOKEN = saved;
              document.getElementById('username-display').textContent = payload.username + ' (' + payload.role + ')';
              document.getElementById('app-team-indicator').innerHTML = charBadge(SESSION.character);
              show('app-section');
              loadNotes();
              return;
            }
          } catch (e) {}
          localStorage.removeItem('mk_token');
        }
        showLogin();
        return;
      }
    } catch (e) {}
  }

  // 2. No valid cookie, but save code in localStorage → auto-restore
  const saveCode = localStorage.getItem('mk_save_code');
  if (saveCode) {
    try {
      const res = await fetch(`${API_URL}/api/session/restore`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ saveCode })
      });
      if (res.ok) {
        SESSION = await res.json();
        const saved = localStorage.getItem('mk_token');
        if (saved) {
          try {
            const payload = JSON.parse(atob(saved.split('.')[1]));
            if (!payload.exp || payload.exp * 1000 > Date.now()) {
              TOKEN = saved;
              document.getElementById('username-display').textContent = payload.username + ' (' + payload.role + ')';
              document.getElementById('app-team-indicator').innerHTML = charBadge(SESSION.character);
              show('app-section');
              loadNotes();
              return;
            }
          } catch (e) {}
          localStorage.removeItem('mk_token');
        }
        showLogin();
        return;
      }
    } catch (e) {}
  }

  // 3. Nothing → start screen
  show('start-section');
}

init();

// ── Start screen ───────────────────────────────────────────────────
function startNewGame() {
  selectedChar = null;
  document.getElementById('nickname-section').classList.add('hidden');
  document.querySelector('.mk-grid-frame')?.classList.remove('hidden');
  document.querySelectorAll('.mk-cell').forEach(el => el.classList.remove('mk-cursor'));
  show('team-section');
}

function startContinue() {
  const saved = localStorage.getItem('mk_save_code') || '';
  document.getElementById('restore-code').value = saved;
  document.getElementById('restore-error').classList.add('hidden');
  show('restore-section');
}

function backToStart() {
  show('start-section');
}

// ── Restore session ────────────────────────────────────────────────
async function doRestore() {
  const code = document.getElementById('restore-code').value.trim();
  if (!code) return;

  const errEl = document.getElementById('restore-error');
  errEl.classList.add('hidden');

  const res = await fetch(`${API_URL}/api/session/restore`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ saveCode: code })
  });
  const d = await res.json();

  if (!res.ok) {
    errEl.textContent = d.error || 'Invalid save code';
    errEl.classList.remove('hidden');
    return;
  }

  SESSION = d;
  localStorage.removeItem('mk_token');
  TOKEN = null;

  document.getElementById('savecode-display').textContent = d.saveCode;
  document.getElementById('restore-info').textContent =
    ` Welcome back, ${d.nickname}! ${d.tasksCompleted} tasks completed, ${d.totalScore} pts.`;
  document.getElementById('restore-info').classList.remove('hidden');

  show('savecode-section');
}

// ── Character select ───────────────────────────────────────────────
function renderCharSelect() {
  const grid = document.querySelector('.mk-select-grid');
  if (!grid) return;

  grid.innerHTML = CHARS_LIST.map(char => {
    const c = ALL_CHARS[char];
    return '<div class="mk-cell" data-team="' + char + '" ' +
      'style="--cc:' + c.hex + ';--cr:' + c.rgb + '" ' +
      'onclick="selectChar(\'' + char + '\')">' +
        '<div class="mk-cell-inner">' +
          '<img src="' + c.img + '" class="mk-char-img" alt="' + c.label + '" onerror="this.remove()">' +
        '</div>' +
        '<div class="mk-cell-name">' + c.label + '</div>' +
      '</div>';
  }).join('');

  const n = CHARS_LIST.length;
  MK_COLS = Math.ceil(n / 4); // 4 rows
  grid.style.gridTemplateColumns = `repeat(${MK_COLS}, 1fr)`;

  const rows = Math.ceil(n / MK_COLS);
  const reservedH = 280;
  const nameBarH = 22;
  const availH = window.innerHeight - reservedH;
  const cellW = Math.round((availH / rows - nameBarH) * 3 / 4);
  const clamped = Math.max(70, Math.min(cellW, 180));
  grid.style.setProperty('--cell-w', clamped + 'px');
}

function selectChar(char) {
  selectedChar = char;
  const c = ALL_CHARS[char];

  document.getElementById('selected-char-label').innerHTML = c.label;
  document.getElementById('nickname-input').value = '';
  document.getElementById('nickname-error').classList.add('hidden');

  // Swap: hide grid, show nickname form
  document.querySelector('.mk-grid-frame')?.classList.add('hidden');
  document.getElementById('nickname-section').classList.remove('hidden');
  document.getElementById('nickname-input').focus();
}

function cancelCharSelect() {
  selectedChar = null;
  document.getElementById('nickname-section').classList.add('hidden');
  document.querySelector('.mk-grid-frame')?.classList.remove('hidden');
  document.querySelectorAll('.mk-cell').forEach(el => el.classList.remove('mk-cursor'));
}

// ── New game ───────────────────────────────────────────────────────
async function doNewGame() {
  if (!selectedChar) return;
  const nickname = document.getElementById('nickname-input').value.trim();
  const errEl = document.getElementById('nickname-error');
  errEl.classList.add('hidden');

  if (!nickname || nickname.length > 50) {
    errEl.textContent = 'Enter a nickname (1–50 characters)';
    errEl.classList.remove('hidden');
    return;
  }

  // Fight animation — show for at least 2s regardless of API speed
  document.getElementById('mk-fight-overlay').classList.remove('hidden');

  const [res] = await Promise.all([
    fetch(`${API_URL}/api/session/new`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ nickname, character: selectedChar })
    }),
    new Promise(r => setTimeout(r, 2000))
  ]);
  const d = await res.json();

  document.getElementById('mk-fight-overlay').classList.add('hidden');

  if (!res.ok) {
    errEl.textContent = d.error || 'Failed to create session';
    errEl.classList.remove('hidden');
    return;
  }

  SESSION = d;
  localStorage.removeItem('mk_token');
  TOKEN = null;

  document.getElementById('savecode-display').textContent = d.saveCode;
  document.getElementById('restore-info').classList.add('hidden');
  show('savecode-section');
}

// ── After save code shown ──────────────────────────────────────────
function proceedToLogin() {
  showLogin();
}

function showLogin() {
  if (!SESSION) return;
  hideAll();
  document.getElementById('login-section').classList.remove('hidden');
  document.getElementById('team-indicator').innerHTML = charBadge(SESSION.character) +
    `<span style="color:#888;font-size:9px;margin-left:8px">${SESSION.nickname}</span>`;
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
  localStorage.setItem('mk_token', TOKEN);


  const payload = JSON.parse(atob(TOKEN.split('.')[1]));
  document.getElementById('username-display').textContent = payload.username + ' (' + payload.role + ')';
  document.getElementById('app-team-indicator').innerHTML = charBadge(SESSION.character);

  document.getElementById('login-section').classList.add('hidden');
  document.getElementById('app-section').classList.remove('hidden');
  loadNotes();
}

function goToMainMenu() {
  TOKEN = null;
  SESSION = null;
  localStorage.removeItem('mk_token');
  // Keep session_id cookie and mk_save_code intact
  show('start-section');
}

function doLogout() {
  TOKEN = null;
  SESSION = null;
  localStorage.removeItem('mk_token');
  // Clear cookies but keep mk_save_code for auto-restore
  document.cookie = 'session_id=;Max-Age=0;path=/';
  document.cookie = 'token=;Max-Age=0;path=/';
  init();
}

async function resetDB() {
  if (!confirm('Reset your database to its initial state?')) return;
  await fetch(`${API_URL}/api/reset`, { method: 'POST', credentials: 'include' });
  alert(' Database reset!');
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

  // Intentionally vulnerable to XSS — CTF task
  el.innerHTML = notes.map(n =>
    '<div class="note" id="note-card-' + n.id + '">' +
      '<h3>' + n.title + '</h3>' +
      '<p id="note-content-' + n.id + '">' + n.content.slice(0, 40) + (n.content.length > 40 ? '…' : '') + '</p>' +
      '<div class="note-meta">' +
        'ID: ' + n.id + ' &nbsp;|&nbsp; ' + (n.created || '') +
        ' <button class="btn-primary" style="margin-left:8px;padding:3px 8px;" onclick="expandNote(' + n.id + ')">EXPAND</button>' +
        ' <button class="btn-danger" style="margin-left:4px;padding:3px 8px;" onclick="deleteNote(' + n.id + ')">DESTROY</button>' +
      '</div>' +
    '</div>'
  ).join('');
}

async function createNote() {
  const t = document.getElementById('note-title').value;
  const c = document.getElementById('note-content').value;
  await fetch(`${API_URL}/api/notes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + TOKEN },
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

// ── Victory screen ────────────────────────────────────────────────
const MAIN_TASKS = [0,1,2,3,4,5,6,7,8,9];

function showVictory() {
  const c = ALL_CHARS[SESSION?.character] || {};
  const overlay = document.getElementById('victory-overlay');
  document.getElementById('vic-portrait').src = c.img || '';
  document.getElementById('vic-nickname').textContent = SESSION?.nickname || '';
  overlay.classList.remove('hidden');
}

// ── Progress ───────────────────────────────────────────────────────
async function loadProgress() {
  const el = document.getElementById('progress-list');
  el.innerHTML = '<p style="color:#64748b;font-size:9px">LOADING...</p>';

  const res = await fetch(`${API_URL}/api/session/progress`, { credentials: 'include' });
  if (!res.ok) { el.innerHTML = '<p style="color:#ef4444">Failed to load</p>'; return; }
  const d = await res.json();

  const completed = d.completed || [];
  if (completed.length === 0) {
    el.innerHTML = '<p style="color:#64748b;font-size:9px;letter-spacing:2px">NO FATALITIES YET</p>';
    return;
  }

  const c = ALL_CHARS[SESSION?.character] || {};
  const doneIds = new Set(completed.map(t => t.taskId));
  const allDone = MAIN_TASKS.every(id => doneIds.has(id));

  el.innerHTML =
    `<div style="margin-bottom:10px;font-size:9px;letter-spacing:2px;color:${c.hex||'#ffdd00'}">
       SCORE: ${d.totalScore} PTS &nbsp;|&nbsp; ${completed.length} TASK${completed.length!==1?'S':''} COMPLETED
     </div>` +
    completed.map(t =>
      `<div class="note" style="border-left:3px solid ${c.hex||'#ffdd00'};margin-bottom:6px;padding:8px 10px">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span style="color:${c.hex||'#ffdd00'};font-size:10px;letter-spacing:1px">TASK ${t.taskId}${t.taskName ? ' — ' + t.taskName : ''}</span>
          <span style="font-size:10px;font-weight:bold">
            <span style="color:#22c55e">+${t.points} PT</span>
            ${t.firstBlood ? '<span style="color:#ef4444;margin-left:6px;font-size:8px">FIRST BLOOD</span>' : ''}
            ${t.comboBonus > 0 ? `<span style="color:#f97316;margin-left:6px;font-size:8px">COMBO x${t.comboBonus + 1}</span>` : ''}
          </span>
        </div>
        <div style="color:#64748b;font-size:8px;margin-top:4px;letter-spacing:1px">${t.timestamp || '—'}</div>
      </div>`
    ).join('');

  if (allDone) setTimeout(showVictory, 400);
}

// ── Keyboard navigation (char select) ─────────────────────────────
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
  // If nickname input is focused, don't intercept
  if (document.activeElement === document.getElementById('nickname-input')) return;

  let moved = false;
  switch (e.key) {
    case 'ArrowRight': mkCursor = (mkCursor + 1) % CHARS_LIST.length; moved = true; break;
    case 'ArrowLeft':  mkCursor = (mkCursor - 1 + CHARS_LIST.length) % CHARS_LIST.length; moved = true; break;
    case 'ArrowDown':  mkCursor = (mkCursor + MK_COLS) % CHARS_LIST.length; moved = true; break;
    case 'ArrowUp':    mkCursor = (mkCursor - MK_COLS + CHARS_LIST.length) % CHARS_LIST.length; moved = true; break;
    case 'Enter':
    case ' ':
      selectChar(CHARS_LIST[mkCursor]);
      moved = true;
      break;
  }
  if (moved) { e.preventDefault(); mkUpdateCursor(); }
});
