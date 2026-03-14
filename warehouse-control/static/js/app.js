let token = localStorage.getItem('token');
let role = localStorage.getItem('role');
let currentLang = 'ru';

const translations = {
  ru: {
    title: 'Warehouse Control',
    login: 'Войти',
    warehouse: 'Склад',
    logout: 'Выход',
    items: 'Товары',
    history: 'История',
    add_item: 'Добавить товар',
    search: 'Поиск',
    show_history: 'Показать историю',
    export_csv: 'Экспорт CSV',
    diff_view: 'Просмотр различий'
  },
  en: {
    title: 'Warehouse Control',
    login: 'Login',
    warehouse: 'Warehouse',
    logout: 'Logout',
    items: 'Items',
    history: 'History',
    add_item: 'Add Item',
    search: 'Search',
    show_history: 'Show History',
    export_csv: 'Export CSV',
    diff_view: 'Diff View'
  }
};

function translate() {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    el.textContent = translations[currentLang][key] || key;
  });
}

document.getElementById('lang-switch').addEventListener('change', (e) => {
  currentLang = e.target.value;
  translate();
});

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
    role = data.role;
    localStorage.setItem('token', token);
    localStorage.setItem('role', role);
    document.getElementById('login').classList.add('hidden');
    document.getElementById('app').classList.remove('hidden');
    document.getElementById('current-user').textContent = `${data.username} (${data.role})`;
    toggleButtonsByRole();
    loadItems();
    translate();
  } else {
    alert('Ошибка входа');
  }
}

function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('role');
  location.reload();
}

function toggleButtonsByRole() {
  if (role === 'viewer') {
    document.getElementById('add-btn').style.display = 'none';
  } else {
    document.getElementById('add-btn').style.display = 'flex';
  }
}

let currentItemId = null;

function showAddModal() {
  currentItemId = null;
  document.getElementById('modal-title').textContent = 'Add Item';
  clearModalFields();
  document.getElementById('item-modal').classList.remove('hidden');
}

function showEditModal(item) {
  currentItemId = item.id;
  document.getElementById('modal-title').textContent = 'Edit Item';
  document.getElementById('item-name').value = item.name;
  document.getElementById('item-sku').value = item.sku;
  document.getElementById('item-quantity').value = item.quantity;
  document.getElementById('item-price').value = item.price;
  document.getElementById('item-category').value = item.category;
  document.getElementById('item-location').value = item.location;
  document.getElementById('item-modal').classList.remove('hidden');
}

function clearModalFields() {
  document.getElementById('item-name').value = '';
  document.getElementById('item-sku').value = '';
  document.getElementById('item-quantity').value = '';
  document.getElementById('item-price').value = '';
  document.getElementById('item-category').value = '';
  document.getElementById('item-location').value = '';
}

async function saveItem() {
  const item = {
    name: document.getElementById('item-name').value,
    sku: document.getElementById('item-sku').value,
    quantity: parseInt(document.getElementById('item-quantity').value),
    price: parseFloat(document.getElementById('item-price').value),
    category: document.getElementById('item-category').value,
    location: document.getElementById('item-location').value
  };
  let res;
  if (currentItemId) {
    res = await api(`/items/${currentItemId}`, 'PUT', item);
  } else {
    res = await api('/items', 'POST', item);
  }
  if (res.ok) {
    closeModal();
    loadItems();
  } else {
    alert('Error saving item');
  }
}

function closeModal() {
  document.getElementById('item-modal').classList.add('hidden');
}

async function loadItems(page = 0) {
  const limit = 10;
  const offset = page * limit;
  const search = document.getElementById('search-items').value;
  let url = `/items?limit=${limit}&offset=${offset}`;
  if (search) url += `&search=${search}`;
  const res = await api(url);
  const data = await res.json();
  let html = `<table class="w-full"><thead><tr class="bg-gray-100"><th class="p-4 text-left">ID</th><th class="p-4 text-left">Name</th><th class="p-4 text-left">SKU</th><th class="p-4 text-left">Quantity</th><th class="p-4 text-left">Price</th><th class="p-4 text-left">Actions</th></tr></thead><tbody>`;
  data.items.forEach(item => {
    html += `<tr class="border-b"><td class="p-4">${item.id}</td><td class="p-4">${item.name}</td><td class="p-4">${item.sku}</td><td class="p-4">${item.quantity}</td><td class="p-4">${item.price}</td><td class="p-4">`;
    if (role !== 'viewer') {
      html += `<button onclick="showEditModal(${JSON.stringify(item)})" class="text-blue-600 mr-2">Edit</button><button onclick="deleteItem(${item.id})" class="text-red-600">Delete</button>`;
    }
    html += `</td></tr>`;
  });
  html += `</tbody></table>`;
  document.getElementById('items-table').innerHTML = html;
  renderPagination('items-pagination', data.total, limit, page, loadItems);
}

function renderPagination(id, total, limit, currentPage, loadFunc) {
  const pages = Math.ceil(total / limit);
  let html = '';
  for (let i = 0; i < pages; i++) {
    html += `<button onclick="${loadFunc.name}(${i})" class="px-4 py-2 ${i === currentPage ? 'bg-blue-600 text-white' : 'bg-gray-200'} rounded mx-1">${i+1}</button>`;
  }
  document.getElementById(id).innerHTML = html;
}

async function deleteItem(id) {
  if (confirm('Delete item?')) {
    await api(`/items/${id}`, 'DELETE');
    loadItems();
  }
}

async function bulkDelete(ids) {
  if (confirm('Delete selected items?')) {
    await api('/items/bulk', 'DELETE', {ids});
    loadItems();
  }
}

async function loadHistory(page = 0) {
  const limit = 10;
  const offset = page * limit;
  let url = '/history?limit=' + limit + '&offset=' + offset;
  const itemId = document.getElementById('history-item-id').value;
  const action = document.getElementById('history-action').value;
  const username = document.getElementById('history-username').value;
  const dateFrom = document.getElementById('history-date-from').value;
  const dateTo = document.getElementById('history-date-to').value;
  if (itemId) url += '&item_id=' + itemId;
  if (action) url += '&action=' + action;
  if (username) url += '&username=' + username;
  if (dateFrom) url += '&date_from=' + dateFrom + 'T00:00:00Z';
  if (dateTo) url += '&date_to=' + dateTo + 'T23:59:59Z';
  const res = await api(url);
  const data = await res.json();
  let html = `<table class="w-full"><thead><tr class="bg-gray-100"><th class="p-4 text-left">ID</th><th class="p-4 text-left">Item ID</th><th class="p-4 text-left">Action</th><th class="p-4 text-left">Changed By</th><th class="p-4 text-left">Changed At</th><th class="p-4 text-left">Actions</th></tr></thead><tbody>`;
  data.records.forEach(rec => {
    html += `<tr class="border-b"><td class="p-4">${rec.id}</td><td class="p-4">${rec.item_id}</td><td class="p-4">${rec.action}</td><td class="p-4">${rec.changed_by}</td><td class="p-4">${new Date(rec.changed_at).toLocaleString()}</td><td class="p-4"><button onclick="showDiff(${JSON.stringify(rec)})" class="text-blue-600">Diff</button></td></tr>`;
  });
  html += `</tbody></table>`;
  document.getElementById('history-table').innerHTML = html;
  renderPagination('history-pagination', data.total, limit, page, loadHistory);
}

function showDiff(rec) {
  const oldData = rec.old_data ? JSON.stringify(rec.old_data, null, 2) : '';
  const newData = rec.new_data ? JSON.stringify(rec.new_data, null, 2) : '';
  const diff = Diff.diffLines(oldData, newData);
  let html = '';
  diff.forEach(part => {
    const color = part.added ? 'green' : part.removed ? 'red' : 'gray';
    html += `<span style="color: ${color};">${part.value}</span>`;
  });
  document.getElementById('diff-content').innerHTML = html;
  document.getElementById('diff-modal').classList.remove('hidden');
}

function closeDiffModal() {
  document.getElementById('diff-modal').classList.add('hidden');
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

function switchTab(tab) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  document.querySelector(`[data-tab="${tab}"]`).classList.add('active');
  document.querySelectorAll('[id^=tab-]').forEach(d => d.classList.add('hidden'));
  document.getElementById(`tab-${tab}`).classList.remove('hidden');
}

if (token) {
  document.getElementById('login').classList.add('hidden');
  document.getElementById('app').classList.remove('hidden');
  toggleButtonsByRole();
  loadItems();
  translate();
}