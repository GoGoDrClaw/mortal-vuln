const baseDomain = window.location.hostname.replace(/^(www\.|api\.|dashboard\.|evil\.)/, '');
const port = window.location.port ? ':' + window.location.port : '';
const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const WS_URL = `${wsProtocol}//api.${baseDomain}${port}/ws`;
const FRONTEND_URL = `${window.location.protocol}//${baseDomain}${port}`;

let ws = null;
let scores = {}; // keyed by sessionId
let audioCtx = null;

function getAudioCtx() {
  if (!audioCtx) audioCtx = new (window.AudioContext || window.webkitAudioContext)();
  if (audioCtx.state === 'suspended') audioCtx.resume();
  return audioCtx;
}

function unlockAudio(btn) {
  try { getAudioCtx().resume().then(() => { btn.textContent = '🔊'; }); } catch (_) {}
}

function makeChars() {
  const base = FRONTEND_URL + '/chars/';
  return {
    scorpion:       { emoji: '🦂', hex: '#f59e0b', rgb: '245,158,11',  img: base + 'scorpion.jpg' },
    subzero:        { emoji: '❄️',  hex: '#3b82f6', rgb: '59,130,246',  img: base + 'subzero.gif' },
    liukang:        { emoji: '🔥', hex: '#ef4444', rgb: '239,68,68',   img: base + 'liukang.gif' },
    kitana:         { emoji: '👸', hex: '#8b5cf6', rgb: '139,92,246',  img: base + 'kitana.jpg' },
    raiden:         { emoji: '⚡', hex: '#a78bfa', rgb: '167,139,250', img: base + 'raiden.jpg' },
    jax:            { emoji: '💪', hex: '#22c55e', rgb: '34,197,94',   img: base + 'jax.gif' },
    mileena:        { emoji: '🎭', hex: '#ec4899', rgb: '236,72,153',  img: base + 'mileena.jpg' },
    kunglao:        { emoji: '🎩', hex: '#84cc16', rgb: '132,204,22',  img: base + 'kunglao.gif' },
    sonya:          { emoji: '🎖️', hex: '#f472b6', rgb: '244,114,182', img: base + 'sonya.gif' },
    shangtsung:     { emoji: '💀', hex: '#f97316', rgb: '249,115,22',  img: base + 'shangtsung.gif' },
    kano:           { emoji: '🔺', hex: '#78716c', rgb: '120,113,108', img: base + 'kano.gif' },
    nightwolf:      { emoji: '🐺', hex: '#10b981', rgb: '16,185,129',  img: base + 'nightwolf.gif' },
    cyrax:          { emoji: '🤖', hex: '#eab308', rgb: '234,179,8',   img: base + 'cyrax.gif' },
    sektor:         { emoji: '🚀', hex: '#dc2626', rgb: '220,38,38',   img: base + 'sektor.gif' },
    kabal:          { emoji: '⚔️',  hex: '#64748b', rgb: '100,116,139', img: base + 'kabal.gif' },
    jade:           { emoji: '💚', hex: '#4ade80', rgb: '74,222,128',  img: base + 'jade.gif' },
    sindel:         { emoji: '👑', hex: '#c026d3', rgb: '192,38,211',  img: base + 'sindel.gif' },
    ermac:          { emoji: '👻', hex: '#b91c1c', rgb: '185,28,28',   img: base + 'ermac.gif' },
    sheeva:         { emoji: '👊', hex: '#d97706', rgb: '217,119,6',   img: base + 'sheeva.gif' },
    stryker:        { emoji: '🚔', hex: '#60a5fa', rgb: '96,165,250',  img: base + 'stryker.gif' },
    classicsubzero: { emoji: '🧊', hex: '#93c5fd', rgb: '147,197,253', img: base + 'classicsubzero.jpg' },
    smoke:          { emoji: '💨', hex: '#9ca3af', rgb: '156,163,175', img: base + 'smoke.gif' },
    noobsaibot:     { emoji: '🌑', hex: '#6366f1', rgb: '99,102,241',  img: base + 'noobsaibot.gif' },
  };
}
const CHARS = makeChars();

const MAX_SCORE = 21;

function charInfo(character) {
  return CHARS[character] || { emoji: '⚔', hex: '#ffdd00', rgb: '255,221,0', img: '' };
}

function connect() {
  ws = new WebSocket(WS_URL);

  ws.onopen = () => {
    const el = document.getElementById('ws-status');
    el.textContent = '⚡ CONNECTED TO OUTWORLD';
    el.style.color = '#22c55e';
  };

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'init') {
      scores = {};
      data.scores.forEach(s => { scores[s.sessionId] = s; });
      renderScoreboard();
      rebuildFeed(data.scores);
    } else if (data.type === 'task_completed') {
      handleTaskCompleted(data.task);
    }
  };

  ws.onerror = () => {
    const el = document.getElementById('ws-status');
    el.textContent = '☠ CONNECTION SEVERED';
    el.style.color = '#ef4444';
  };

  ws.onclose = () => {
    const el = document.getElementById('ws-status');
    el.textContent = '⚡ RECONNECTING TO OUTWORLD...';
    el.style.color = '#f59e0b';
    setTimeout(connect, 3000);
  };
}

function handleTaskCompleted(task) {
  const key = task.sessionId;
  if (!scores[key]) {
    scores[key] = {
      sessionId: key,
      nickname:  task.nickname,
      character: task.character,
      saveCode:  '',
      totalScore: 0,
      completed: []
    };
  }
  const exists = scores[key].completed.find(t => t.taskId === task.taskId);
  if (exists) return;
  scores[key].completed.push(task);
  scores[key].totalScore += task.points;
  renderScoreboard();
  addActivityLog(task);
  playFatalitySound();
}

function renderScoreboard() {
  const container = document.getElementById('scoreboard');
  const sorted = Object.values(scores).sort((a, b) => b.totalScore - a.totalScore);

  if (sorted.length === 0) {
    container.innerHTML = '<div style="color:#332200;font-size:8px;text-align:center;padding:40px;letter-spacing:2px;grid-column:1/-1">AWAITING WARRIORS...</div>';
    return;
  }

  const oldPositions = {};
  container.querySelectorAll('.mk-team-card[data-session]').forEach(el => {
    oldPositions[el.dataset.session] = el.getBoundingClientRect();
  });

  container.innerHTML = sorted.map((s, idx) => {
    const c = charInfo(s.character);
    const pct = Math.min(100, Math.round((s.totalScore / MAX_SCORE) * 100));

    const lastTask = [...(s.completed || [])].sort((a, b) => {
      const ta = a.timestamp || '', tb = b.timestamp || '';
      return ta > tb ? -1 : ta < tb ? 1 : b.taskId - a.taskId;
    })[0];

    const taskItem = lastTask
      ? '<li class="mk-task-item">' +
          '<span class="mk-task-name">TASK ' + lastTask.taskId + '</span>' +
          '<span class="mk-task-pts">+' + lastTask.points + '</span>' +
        '</li>'
      : '';

    return '<div class="mk-team-card" data-session="' + s.sessionId + '" style="--cc:' + c.hex + ';--cr:' + c.rgb + '">' +
      '<div class="mk-card-portrait">' +
        '<span class="mk-card-portrait-emoji">' + c.emoji + '</span>' +
        '<img src="' + c.img + '" alt="' + s.character + '" onerror="this.remove()">' +
        '<div class="mk-rank-badge">' + (idx + 1) + '</div>' +
      '</div>' +
      '<div class="mk-card-body">' +
        '<div class="mk-card-name-row">' +
          '<div class="mk-card-name">' +
            '<div style="color:var(--cc);font-size:10px;margin-bottom:2px">' + escHtml(s.nickname) + '</div>' +
            '<div style="font-size:7px;color:#888;letter-spacing:1px">' + c.emoji + ' ' + s.character.toUpperCase() + '</div>' +
          '</div>' +
          '<div class="mk-card-score">' + s.totalScore + '</div>' +
        '</div>' +
        '<div class="mk-health-wrap">' +
          '<div class="mk-health-label">POWER — ' + pct + '%</div>' +
          '<div class="mk-health-track">' +
            '<div class="mk-health-fill" style="width:' + pct + '%"></div>' +
          '</div>' +
        '</div>' +
        '<ul class="mk-tasks">' +
          (taskItem || '<li class="mk-no-tasks">NO FATALITIES YET</li>') +
        '</ul>' +
      '</div>' +
    '</div>';
  }).join('');

  // FLIP animation
  container.querySelectorAll('.mk-team-card[data-session]').forEach(el => {
    const old = oldPositions[el.dataset.session];
    if (!old) return;
    const cur = el.getBoundingClientRect();
    const dx = old.left - cur.left, dy = old.top - cur.top;
    if (dx === 0 && dy === 0) return;
    el.style.transition = 'none';
    el.style.transform = `translate(${dx}px,${dy}px)`;
    requestAnimationFrame(() => requestAnimationFrame(() => {
      el.style.transition = 'transform 0.5s cubic-bezier(0.25,0.46,0.45,0.94)';
      el.style.transform = '';
    }));
  });
}

function rebuildFeed(sessions) {
  const all = [];
  sessions.forEach(s => {
    (s.completed || []).forEach(t => { all.push({ ...t, sessionId: s.sessionId }); });
  });
  all.sort((a, b) => {
    const ta = a.timestamp || '', tb = b.timestamp || '';
    return ta < tb ? -1 : ta > tb ? 1 : a.taskId - b.taskId;
  });
  const activityDiv = document.getElementById('activity');
  activityDiv.innerHTML = '';
  if (all.length === 0) {
    activityDiv.innerHTML = '<div class="mk-empty-feed">AWAITING KOMBAT...</div>';
    return;
  }
  all.forEach(task => addActivityLog(task, false));
}

function addActivityLog(task, isNew = true) {
  const c = charInfo(task.character);
  const ts = task.timestamp
    ? (task.timestamp.split('T')[1]?.slice(0, 5) || task.timestamp.slice(0, 5))
    : '';

  const entry = document.createElement('div');
  entry.className = isNew ? 'log-entry new' : 'log-entry';
  entry.style.borderColor = c.hex;
  entry.innerHTML =
    '<div class="log-left">' +
      '<span class="log-team" style="color:' + c.hex + '">' +
        c.emoji + ' ' + escHtml(task.nickname) +
        '<span style="font-size:7px;color:#666;margin-left:4px">(' + task.character + ')</span>' +
      '</span>' +
      '<span class="log-desc">COMPLETED TASK ' + task.taskId + '</span>' +
    '</div>' +
    '<span class="log-pts" style="color:' + c.hex + '">+' + task.points + ' PT</span>' +
    '<span class="log-time">' + ts + '</span>';

  const activityDiv = document.getElementById('activity');
  const empty = activityDiv.querySelector('.mk-empty-feed');
  if (empty) empty.remove();
  activityDiv.insertBefore(entry, activityDiv.firstChild);
  if (isNew) setTimeout(() => entry.classList.remove('new'), 1000);
  while (activityDiv.children.length > 50) activityDiv.removeChild(activityDiv.lastChild);
}

function escHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function playFatalitySound() {
  try {
    const ctx = getAudioCtx();
    [880, 440].forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.connect(gain); gain.connect(ctx.destination);
      osc.frequency.value = freq; osc.type = 'square';
      const t = ctx.currentTime + i * 0.12;
      gain.gain.setValueAtTime(0.15, t);
      gain.gain.exponentialRampToValueAtTime(0.001, t + 0.1);
      osc.start(t); osc.stop(t + 0.1);
    });
  } catch (_) {}
}

connect();
