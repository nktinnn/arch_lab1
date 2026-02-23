// ---- Admin: user management ----
async function loadAdminPage() {
  const tbody = document.getElementById('users-tbody');
  tbody.innerHTML = '<tr><td colspan="5" class="spinner">Загрузка...</td></tr>';
  try {
    const users = await api.listUsers();
    const currentId = getUser().id;
    tbody.innerHTML = users.map(u => `
      <tr>
        <td>${u.id}</td>
        <td>${escHtml(u.username)}</td>
        <td>${escHtml(u.email)}</td>
        <td>
          <select class="role-select" data-id="${u.id}" ${u.id === currentId ? 'disabled' : ''}>
            <option value="user" ${u.role === 'user' ? 'selected' : ''}>user</option>
            <option value="operator" ${u.role === 'operator' ? 'selected' : ''}>operator</option>
            <option value="admin" ${u.role === 'admin' ? 'selected' : ''}>admin</option>
          </select>
        </td>
        <td>
          <button class="btn btn-ghost btn-sm"
            onclick="saveRole(${u.id})"
            ${u.id === currentId ? 'disabled' : ''}>
            Сохранить
          </button>
          <button class="btn btn-danger btn-sm" style="margin-left:6px"
            onclick="deleteUser(${u.id})"
            ${u.id === currentId ? 'disabled' : ''}>
            Удалить
          </button>
        </td>
      </tr>
    `).join('');
  } catch (e) {
    tbody.innerHTML = `<tr><td colspan="5" style="color:var(--danger)">${e.message}</td></tr>`;
  }
}

async function saveRole(userId) {
  const sel = document.querySelector(`select.role-select[data-id="${userId}"]`);
  try {
    await api.updateRole(userId, sel.value);
    alert('Роль обновлена');
  } catch (e) {
    alert(e.message);
  }
}

async function deleteUser(userId) {
  if (!confirm('Удалить пользователя?')) return;
  try {
    await api.deleteUser(userId);
    loadAdminPage();
  } catch (e) {
    alert(e.message);
  }
}
