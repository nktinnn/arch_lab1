const API_BASE = 'http://localhost:8080';

async function request(method, path, body) {
  const headers = { 'Content-Type': 'application/json' };
  const token = localStorage.getItem('token');
  if (token) headers['Authorization'] = 'Bearer ' + token;

  const res = await fetch(API_BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();
  let data;
  try { data = JSON.parse(text); } catch { data = text; }

  if (!res.ok) {
    const msg = (data && data.error) ? data.error : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return data;
}

const api = {
  // Auth
  register: (username, email, password, role) =>
    request('POST', '/api/auth/register', { username, email, password, role }),
  login: (email, password) =>
    request('POST', '/api/auth/login', { email, password }),

  // Users (admin)
  listUsers: () => request('GET', '/api/users'),
  updateRole: (id, role) => request('PUT', `/api/users/${id}/role`, { role }),
  deleteUser: (id) => request('DELETE', `/api/users/${id}`),

  // Tickets
  createTicket: (title, description, priority) =>
    request('POST', '/api/tickets', { title, description, priority }),
  listTickets: () => request('GET', '/api/tickets'),
  getTicket: (id) => request('GET', `/api/tickets/${id}`),
  updateTicket: (id, data) => request('PUT', `/api/tickets/${id}`, data),
  deleteTicket: (id) => request('DELETE', `/api/tickets/${id}`),

  // Comments
  listComments: (ticketId) => request('GET', `/api/tickets/${ticketId}/comments`),
  addComment: (ticketId, content) =>
    request('POST', `/api/tickets/${ticketId}/comments`, { content }),
  deleteComment: (id) => request('DELETE', `/api/comments/${id}`),
};
