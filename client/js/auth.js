// Auth state
let currentUser = null;

function getUser() { return currentUser; }

function setSession(token, user) {
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
  currentUser = user;
}

function clearSession() {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  currentUser = null;
}

function loadSession() {
  const token = localStorage.getItem('token');
  const raw = localStorage.getItem('user');
  if (token && raw) {
    try { currentUser = JSON.parse(raw); return true; } catch {}
  }
  return false;
}

// ---- UI helpers ----
function showAuthPage() {
  showPage('auth');
  document.getElementById('nav').style.display = 'none';
}

function showAppUI() {
  document.getElementById('nav').style.display = 'flex';
  const u = getUser();
  document.getElementById('nav-username').textContent = u.username + ' (' + u.role + ')';

  // Show/hide admin nav item
  const adminBtn = document.getElementById('nav-admin');
  adminBtn.style.display = (u.role === 'admin') ? '' : 'none';

  showPage('tickets');
  loadTickets();
}

// ---- Auth form handling ----
function initAuth() {
  document.querySelectorAll('.auth-tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.auth-tab').forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      document.getElementById('auth-form-login').style.display =
        tab.dataset.tab === 'login' ? '' : 'none';
      document.getElementById('auth-form-register').style.display =
        tab.dataset.tab === 'register' ? '' : 'none';
      document.getElementById('auth-error').textContent = '';
    });
  });

  document.getElementById('btn-login').addEventListener('click', handleLogin);
  document.getElementById('btn-register').addEventListener('click', handleRegister);
  document.getElementById('btn-logout').addEventListener('click', handleLogout);
}

async function handleLogin() {
  const email = document.getElementById('login-email').value.trim();
  const password = document.getElementById('login-password').value;
  const errEl = document.getElementById('auth-error');
  errEl.textContent = '';
  try {
    const data = await api.login(email, password);
    setSession(data.token, data.user);
    showAppUI();
  } catch (e) {
    errEl.textContent = e.message;
  }
}

async function handleRegister() {
  const username = document.getElementById('reg-username').value.trim();
  const email = document.getElementById('reg-email').value.trim();
  const password = document.getElementById('reg-password').value;
  const role = document.getElementById('reg-role').value;
  const errEl = document.getElementById('auth-error');
  errEl.textContent = '';
  try {
    const data = await api.register(username, email, password, role);
    setSession(data.token, data.user);
    showAppUI();
  } catch (e) {
    errEl.textContent = e.message;
  }
}

function handleLogout() {
  clearSession();
  showAuthPage();
}
