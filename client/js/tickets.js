// ---- Ticket list ----
async function loadTickets() {
  const container = document.getElementById('tickets-list');
  container.innerHTML = '<div class="spinner">Загрузка...</div>';

  const filterStatus = document.getElementById('filter-status').value;
  const filterPriority = document.getElementById('filter-priority').value;

  try {
    let tickets = await api.listTickets();
    if (filterStatus) tickets = tickets.filter(t => t.status === filterStatus);
    if (filterPriority) tickets = tickets.filter(t => t.priority === filterPriority);

    if (!tickets.length) {
      container.innerHTML = '<div class="empty-state">Нет обращений</div>';
      return;
    }

    container.innerHTML = tickets.map(t => `
      <div class="ticket-item priority-${t.priority}" onclick="openTicket(${t.id})">
        <div class="ticket-info">
          <div class="ticket-title">#${t.id} ${escHtml(t.title)}</div>
          <div class="ticket-meta">
            Автор: ${escHtml(t.author_name)} &bull;
            ${fmtDate(t.created_at)}
            ${t.assignee_name ? ' &bull; Исполнитель: ' + escHtml(t.assignee_name) : ''}
          </div>
        </div>
        <div class="ticket-badges">
          <span class="badge badge-${t.status}">${statusLabel(t.status)}</span>
          <span class="badge badge-${t.priority}">${priorityLabel(t.priority)}</span>
        </div>
      </div>
    `).join('');
  } catch (e) {
    container.innerHTML = `<div class="empty-state">${e.message}</div>`;
  }
}

// ---- Create ticket modal ----
function initCreateTicket() {
  document.getElementById('btn-new-ticket').addEventListener('click', () => {
    document.getElementById('modal-create').classList.add('open');
  });
  document.getElementById('modal-create-close').addEventListener('click', () => {
    document.getElementById('modal-create').classList.remove('open');
  });
  document.getElementById('btn-create-submit').addEventListener('click', handleCreateTicket);
}

async function handleCreateTicket() {
  const title = document.getElementById('ct-title').value.trim();
  const description = document.getElementById('ct-desc').value.trim();
  const priority = document.getElementById('ct-priority').value;
  const errEl = document.getElementById('ct-error');
  errEl.textContent = '';

  if (!title || !description) {
    errEl.textContent = 'Заполните заголовок и описание';
    return;
  }
  try {
    await api.createTicket(title, description, priority);
    document.getElementById('modal-create').classList.remove('open');
    document.getElementById('ct-title').value = '';
    document.getElementById('ct-desc').value = '';
    loadTickets();
  } catch (e) {
    errEl.textContent = e.message;
  }
}

// ---- Ticket detail ----
async function openTicket(id) {
  showPage('ticket-detail');
  const root = document.getElementById('ticket-detail-root');
  root.innerHTML = '<div class="spinner">Загрузка...</div>';

  try {
    const [ticket, comments] = await Promise.all([
      api.getTicket(id),
      api.listComments(id),
    ]);
    renderTicketDetail(ticket, comments);
  } catch (e) {
    root.innerHTML = `<div class="empty-state">${e.message}</div>`;
  }
}

function renderTicketDetail(ticket, comments) {
  const user = getUser();
  const root = document.getElementById('ticket-detail-root');

  const canEdit = user.role === 'admin' || user.role === 'operator' ||
    (user.role === 'user' && ticket.author_id === user.id && ticket.status === 'open');
  const canChangeStatus = user.role === 'admin' || user.role === 'operator';
  const canDelete = user.role === 'admin';
  const canAssign = user.role === 'admin' || user.role === 'operator';

  const statusOptions = ['open', 'in_progress', 'resolved', 'closed']
    .map(s => `<option value="${s}" ${ticket.status === s ? 'selected' : ''}>${statusLabel(s)}</option>`)
    .join('');

  const priorityOptions = ['low', 'medium', 'high', 'critical']
    .map(p => `<option value="${p}" ${ticket.priority === p ? 'selected' : ''}>${priorityLabel(p)}</option>`)
    .join('');

  root.innerHTML = `
    <button class="btn btn-ghost btn-sm back-btn" onclick="showPage('tickets'); loadTickets()">← Назад</button>
    <div class="card ticket-detail-header">
      <div style="display:flex;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;gap:8px">
        <h2>#${ticket.id} ${escHtml(ticket.title)}</h2>
        ${canDelete ? `<button class="btn btn-danger btn-sm" onclick="deleteTicket(${ticket.id})">Удалить</button>` : ''}
      </div>
      <div class="detail-meta">
        <span class="badge badge-${ticket.status}">${statusLabel(ticket.status)}</span>
        <span class="badge badge-${ticket.priority}">${priorityLabel(ticket.priority)}</span>
        <span style="color:var(--text-muted);font-size:0.83rem">Автор: ${escHtml(ticket.author_name)}</span>
        ${ticket.assignee_name ? `<span style="color:var(--text-muted);font-size:0.83rem">Исполнитель: ${escHtml(ticket.assignee_name)}</span>` : ''}
        <span style="color:var(--text-muted);font-size:0.83rem">${fmtDate(ticket.created_at)}</span>
      </div>
      <div class="detail-description">${escHtml(ticket.description)}</div>

      ${canEdit ? `
      <div class="detail-actions" style="flex-wrap:wrap;gap:8px">
        ${canChangeStatus ? `
        <select id="td-status">${statusOptions}</select>
        ` : ''}
        <select id="td-priority">${priorityOptions}</select>
        <button class="btn btn-primary btn-sm" onclick="saveTicketChanges(${ticket.id})">Сохранить</button>
      </div>
      ` : ''}
    </div>

    <div class="card comments-section" style="margin-top:16px">
      <h3>Комментарии (${comments.length})</h3>
      <div class="comment-list" id="comment-list">
        ${comments.length ? comments.map(c => renderComment(c, user)).join('') : '<div style="color:var(--text-muted);font-size:0.9rem">Комментариев пока нет</div>'}
      </div>
      <div class="comment-form">
        <textarea id="new-comment" placeholder="Добавить комментарий..."></textarea>
        <button class="btn btn-primary btn-sm" style="align-self:flex-end" onclick="submitComment(${ticket.id})">Отправить</button>
      </div>
      <div class="error-msg" id="comment-error"></div>
    </div>
  `;
}

function renderComment(c, user) {
  const canDel = user.role === 'admin' || user.role === 'operator';
  return `
    <div class="comment-item" id="comment-${c.id}">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div>
          <span class="comment-author">${escHtml(c.username)}</span>
          <span class="comment-date"> &bull; ${fmtDate(c.created_at)}</span>
        </div>
        ${canDel ? `<button class="btn btn-ghost btn-sm" onclick="deleteComment(${c.id})">✕</button>` : ''}
      </div>
      <div class="comment-content">${escHtml(c.content)}</div>
    </div>
  `;
}

async function saveTicketChanges(id) {
  const statusEl = document.getElementById('td-status');
  const priorityEl = document.getElementById('td-priority');
  const data = { priority: priorityEl.value };
  if (statusEl) data.status = statusEl.value;
  try {
    await api.updateTicket(id, data);
    openTicket(id);
  } catch (e) {
    alert(e.message);
  }
}

async function deleteTicket(id) {
  if (!confirm('Удалить тикет?')) return;
  try {
    await api.deleteTicket(id);
    showPage('tickets');
    loadTickets();
  } catch (e) {
    alert(e.message);
  }
}

async function submitComment(ticketId) {
  const content = document.getElementById('new-comment').value.trim();
  const errEl = document.getElementById('comment-error');
  errEl.textContent = '';
  if (!content) { errEl.textContent = 'Введите текст комментария'; return; }
  try {
    await api.addComment(ticketId, content);
    document.getElementById('new-comment').value = '';
    openTicket(ticketId);
  } catch (e) {
    errEl.textContent = e.message;
  }
}

async function deleteComment(id) {
  if (!confirm('Удалить комментарий?')) return;
  try {
    await api.deleteComment(id);
    document.getElementById('comment-' + id)?.remove();
  } catch (e) {
    alert(e.message);
  }
}
