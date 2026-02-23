// ---- Routing / page switching ----
function showPage(name) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.getElementById('page-' + name).classList.add('active');
  document.querySelectorAll('.nav-links button[data-page]').forEach(b => {
    b.classList.toggle('active', b.dataset.page === name);
  });
}

// ---- Utility ----
function escHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function fmtDate(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  return d.toLocaleString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function statusLabel(s) {
  return { open: 'Открыт', in_progress: 'В работе', resolved: 'Решён', closed: 'Закрыт' }[s] || s;
}

function priorityLabel(p) {
  return { low: 'Низкий', medium: 'Средний', high: 'Высокий', critical: 'Критический' }[p] || p;
}

// ---- Init ----
document.addEventListener('DOMContentLoaded', () => {
  initAuth();
  initCreateTicket();

  // Nav bindings
  document.getElementById('nav-tickets').addEventListener('click', () => {
    showPage('tickets');
    loadTickets();
  });
  document.getElementById('nav-admin').addEventListener('click', () => {
    showPage('admin');
    loadAdminPage();
  });

  // Filter change
  document.getElementById('filter-status').addEventListener('change', loadTickets);
  document.getElementById('filter-priority').addEventListener('change', loadTickets);

  // Restore session
  if (loadSession()) {
    showAppUI();
  } else {
    showAuthPage();
  }
});
