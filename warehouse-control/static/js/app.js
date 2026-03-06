let token = localStorage.getItem('token');
const api = (url, method = 'GET', body = null) => {
  const opts = {
    method,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  };
  if (body) opts.body = JSON.stringify(body);
  return fetch(url, opts);
};
async function login() {
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;
  const res = await fetch('/auth/login', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({username, password})
  });
  if (res.ok) {
    const data = await res.json();
    token = data.token;
    localStorage.setItem('token', token);
    document.getElementById('login').classList.add('hidden');
    document.getElementById('app').classList.remove('hidden');
    loadItems();
  } else {
    alert('Ошибка входа');
  }
}
function logout() {
  localStorage.removeItem('token');
  location.reload();
}
async function loadItems() {
  const res = await api('/items');
  const data = await res.json();
  let html = `<table class="w-full"><thead><tr class="bg-gray-100"><th class="p-4 text-left">ID</th><th class="p-4 text-left">Название</th><th class="p-4 text-left">SKU</th><th class="p-4 text-left">Кол-во</th><th class="p-4 text-left">Цена</th><th class="p-4 text-left">Действия</th></tr></thead><tbody>`;
  data.items.forEach(item => {
    html += `<tr class="border-b"><td class="p-4">${item.id}</td><td class="p-4">${item.name}</td><td class="p-4">${item.sku}</td><td class="p-4">${item.quantity}</td><td class="p-4">${item.price}</td><td class="p-4"><button onclick="deleteItem(${item.id})" class="text-red-600">Удалить</button></td></tr>`;
  });
  html += `</tbody></table>`;
  document.getElementById('items-table').innerHTML = html;
}
async function deleteItem(id) {
  if (confirm('Удалить товар?')) {
    await api(`/items/${id}`, 'DELETE');
    loadItems();
  }
}
async function exportCSV() {
  const res = await api('/history/export');
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'history.csv';
  a.click();
}
// Инициализация
if (token) {
  document.getElementById('login').classList.add('hidden');
  document.getElementById('app').classList.remove('hidden');
  loadItems();
}