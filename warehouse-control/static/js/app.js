/**
 * Warehouse Control — Полный Frontend (новый дизайн 2025)
 * Работает с access_token / refresh_token + SPA fallback
 */

const State = {
    token: localStorage.getItem('token') || '',
    refreshToken: localStorage.getItem('refresh_token') || '',
    role: localStorage.getItem('role') || '',
    user: localStorage.getItem('username') || '',
    selectedItems: new Set()
};

const HttpClient = {
    async request(url, method = 'GET', body = null) {
        const loader = document.getElementById('global-loader');
        if (loader) loader.style.transform = 'scaleX(0.3)';

        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${State.token}`
            }
        };
        if (body) options.body = JSON.stringify(body);

        try {
            let res = await fetch(url, options);

            // Автоматический refresh при 401
            if (res.status === 401 && State.refreshToken && !url.includes('/auth/')) {
                const newToken = await Auth.refreshToken();
                if (newToken) {
                    options.headers.Authorization = `Bearer ${newToken}`;
                    res = await fetch(url, options);
                } else {
                    Auth.logout();
                    return null;
                }
            }

            if (!res.ok) {
                const err = await res.json().catch(() => ({}));
                throw new Error(err.error || `HTTP ${res.status}`);
            }

            if (loader) loader.style.transform = 'scaleX(1)';
            const text = await res.text();
            return text ? JSON.parse(text) : null;
        } catch (err) {
            this.showToast(err.message, 'error');
            throw err;
        } finally {
            setTimeout(() => { if (loader) loader.style.transform = 'scaleX(0)'; }, 400);
        }
    },

    showToast(msg, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `fixed bottom-6 right-6 px-6 py-3.5 rounded-2xl shadow-2xl text-white z-[9999] transition-all ${
            type === 'error' ? 'bg-red-600' : 'bg-emerald-600'
        }`;
        toast.innerHTML = `<i class="fas ${type === 'error' ? 'fa-circle-exclamation' : 'fa-check-circle'} mr-2"></i>${msg}`;
        document.body.appendChild(toast);
        setTimeout(() => toast.remove(), 3500);
    }
};

const UI = {
    switchTab(index) {
        document.querySelectorAll('.tab').forEach((tab, i) => {
            tab.classList.toggle('active', i === index);
            tab.classList.toggle('text-blue-600', i === index);
            tab.classList.toggle('text-slate-500', i !== index);
        });
        document.getElementById('tab-0').classList.toggle('hidden', index !== 0);
        document.getElementById('tab-1').classList.toggle('hidden', index !== 1);
        if (index === 1) App.loadHistory();
    },

    showModal(id) {
        document.getElementById(id).classList.remove('hidden');
    },

    closeModal(id) {
        document.getElementById(id).classList.add('hidden');
    }
};

const Auth = {
    async login(e) {
        e.preventDefault();
        const btn = e.target.querySelector('button[type="submit"]');
        const original = btn.innerHTML;
        btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';

        try {
            const data = await HttpClient.request('/auth/login', 'POST', {
                username: document.getElementById('username').value.trim(),
                password: document.getElementById('password').value.trim()
            });

            State.token = data.access_token;
            State.refreshToken = data.refresh_token;
            State.user = data.username || 'user';
            State.role = data.role || 'viewer';

            localStorage.setItem('token', State.token);
            localStorage.setItem('refresh_token', State.refreshToken);
            localStorage.setItem('username', State.user);
            localStorage.setItem('role', State.role);

            this.initApp();
        } catch (err) {
            btn.innerHTML = original;
        }
    },

    async refreshToken() {
        if (!State.refreshToken) return null;
        try {
            const data = await HttpClient.request('/auth/refresh', 'POST', { refresh_token: State.refreshToken });
            State.token = data.access_token;
            if (data.refresh_token) State.refreshToken = data.refresh_token;

            localStorage.setItem('token', State.token);
            localStorage.setItem('refresh_token', State.refreshToken);
            return State.token;
        } catch {
            this.logout();
            return null;
        }
    },

    logout() {
        localStorage.clear();
        location.reload();
    },

    initApp() {
        document.getElementById('login').classList.add('hidden');
        document.getElementById('app').classList.remove('hidden');

        document.getElementById('user-profile').classList.remove('hidden');
        document.getElementById('current-user').textContent = State.user;
        document.getElementById('current-role').textContent = State.role.toUpperCase();

        const canEdit = ['admin', 'manager'].includes(State.role);
        if (canEdit) document.getElementById('add-btn').classList.remove('hidden');
        if (State.role === 'admin') document.getElementById('bulk-delete-btn').classList.remove('hidden');

        App.loadItems();
    }
};

const App = {
    async loadItems() {
        const table = document.getElementById('items-table');
        table.innerHTML = '<div class="p-10"><div class="loading-shimmer h-8 rounded mb-3"></div><div class="loading-shimmer h-8 rounded"></div></div>';

        const search = document.getElementById('search-items').value.trim();
        let url = `/items?limit=15&offset=0`;
        if (search) url += `&search=${encodeURIComponent(search)}`;

        try {
            const data = await HttpClient.request(url);
            this.renderItemsTable(data.items || []);
        } catch (e) {
            table.innerHTML = '<div class="p-20 text-center text-slate-400">Не удалось загрузить товары</div>';
        }
    },

    renderItemsTable(items) {
        const container = document.getElementById('items-table');
        const isAdmin = State.role === 'admin';
        const canEdit = ['admin', 'manager'].includes(State.role);

        let html = `
            <table class="w-full text-left">
                <thead class="bg-slate-50 border-b">
                    <tr>
                        ${isAdmin ? `<th class="p-4 w-10"><input type="checkbox" onchange="App.toggleAll(this)"></th>` : ''}
                        <th class="p-4 text-xs font-bold uppercase">Товар / ID</th>
                        <th class="p-4 text-xs font-bold uppercase">SKU</th>
                        <th class="p-4 text-xs font-bold uppercase">Остаток</th>
                        <th class="p-4 text-xs font-bold uppercase text-right">Действия</th>
                    </tr>
                </thead>
                <tbody class="divide-y">
                    ${items.map(item => `
                        <tr class="hover:bg-slate-50 transition-colors">
                            ${isAdmin ? `<td class="p-4"><input type="checkbox" value="${item.id}" onchange="App.toggleItem(this)" class="item-cb"></td>` : ''}
                            <td class="p-4">
                                <div class="font-bold">#${item.id} ${item.name}</div>
                                <div class="text-xs text-slate-400">${item.category || ''}</div>
                            </td>
                            <td class="p-4 font-mono">${item.sku}</td>
                            <td class="p-4">
                                <span class="px-3 py-1 text-xs font-bold rounded-full ${item.quantity < 10 ? 'bg-red-100 text-red-600' : 'bg-emerald-100 text-emerald-600'}">
                                    ${item.quantity} шт
                                </span>
                            </td>
                            <td class="p-4 text-right space-x-3">
                                <button onclick="App.loadItemHistory(${item.id})" class="w-8 h-8 bg-slate-100 hover:bg-slate-200 rounded-xl"><i class="fas fa-history"></i></button>
                                ${canEdit ? `<button onclick='App.editItem(${JSON.stringify(item)})' class="w-8 h-8 bg-slate-100 hover:bg-slate-200 rounded-xl"><i class="fas fa-pen"></i></button>` : ''}
                                ${canEdit ? `<button onclick="App.deleteItem(${item.id})" class="w-8 h-8 bg-slate-100 hover:bg-red-100 text-red-600 rounded-xl"><i class="fas fa-trash"></i></button>` : ''}
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
        container.innerHTML = html;
    },

    showItemModal() {
        document.getElementById('modal-title').textContent = 'Новый товар';
        document.getElementById('item-form').reset();
        document.getElementById('item-id').value = '';
        UI.showModal('item-modal');
    },

    editItem(item) {
        document.getElementById('modal-title').textContent = `Редактировать #${item.id}`;
        document.getElementById('item-id').value = item.id;
        document.getElementById('item-name').value = item.name;
        document.getElementById('item-sku').value = item.sku;
        document.getElementById('item-quantity').value = item.quantity;
        document.getElementById('item-price').value = item.price;
        document.getElementById('item-category').value = item.category || '';
        document.getElementById('item-location').value = item.location || '';
        UI.showModal('item-modal');
    },

    async saveItem(e) {
        e.preventDefault();
        const id = document.getElementById('item-id').value;
        const payload = {
            name: document.getElementById('item-name').value,
            sku: document.getElementById('item-sku').value,
            quantity: parseInt(document.getElementById('item-quantity').value),
            price: parseFloat(document.getElementById('item-price').value),
            category: document.getElementById('item-category').value,
            location: document.getElementById('item-location').value
        };

        try {
            if (id) {
                await HttpClient.request(`/items/${id}`, 'PUT', payload);
            } else {
                await HttpClient.request('/items', 'POST', payload);
            }
            HttpClient.showToast('Товар сохранён', 'success');
            UI.closeModal('item-modal');
            App.loadItems();
        } catch (e) {}
    },

    async deleteItem(id) {
        if (!confirm('Удалить товар?')) return;
        try {
            await HttpClient.request(`/items/${id}`, 'DELETE');
            App.loadItems();
        } catch (e) {}
    },

    toggleItem(cb) {
        if (cb.checked) State.selectedItems.add(cb.value);
        else State.selectedItems.delete(cb.value);
        const btn = document.getElementById('bulk-delete-btn');
        btn.textContent = `Удалить выбранные (${State.selectedItems.size})`;
        btn.disabled = State.selectedItems.size === 0;
    },

    toggleAll(cb) {
        document.querySelectorAll('.item-cb').forEach(c => {
            c.checked = cb.checked;
            this.toggleItem(c);
        });
    },

    async bulkDelete() {
        const ids = Array.from(State.selectedItems).map(Number);
        if (!ids.length || !confirm(`Удалить ${ids.length} товаров?`)) return;
        try {
            await HttpClient.request('/items/bulk', 'DELETE', { ids });
            State.selectedItems.clear();
            App.loadItems();
        } catch (e) {}
    },

    async loadHistory() {
        const container = document.getElementById('history-table');
        container.innerHTML = '<div class="p-10"><div class="loading-shimmer h-8 rounded"></div></div>';

        const params = new URLSearchParams({
            limit: 20,
            offset: 0
        });
        if (document.getElementById('hist-item').value) params.append('item_id', document.getElementById('hist-item').value);
        if (document.getElementById('hist-action').value) params.append('action', document.getElementById('hist-action').value);
        if (document.getElementById('hist-from').value) params.append('date_from', document.getElementById('hist-from').value + 'T00:00:00Z');
        if (document.getElementById('hist-to').value) params.append('date_to', document.getElementById('hist-to').value + 'T23:59:59Z');

        try {
            const data = await HttpClient.request(`/history?${params}`);
            this.renderHistoryTable(data.records || [], 'history-table', true);
        } catch (e) {}
    },

    async loadItemHistory(id) {
        document.getElementById('history-modal-item-id').textContent = `#${id}`;
        UI.showModal('item-history-modal');
        try {
            const data = await HttpClient.request(`/history/item/${id}`);
            this.renderHistoryTable(data.records || [], 'item-history-content', false);
        } catch (e) {}
    },

    renderHistoryTable(records, containerId, showItem) {
        const container = document.getElementById(containerId);
        container.innerHTML = `
            <table class="w-full text-sm">
                <thead class="bg-slate-50">
                    <tr>
                        <th class="p-4">Дата</th>
                        ${showItem ? '<th class="p-4">ID товара</th>' : ''}
                        <th class="p-4">Действие</th>
                        <th class="p-4">Пользователь</th>
                        <th class="p-4 text-right">Diff</th>
                    </tr>
                </thead>
                <tbody class="divide-y">
                    ${records.map(r => `
                        <tr>
                            <td class="p-4">${new Date(r.changed_at).toLocaleString('ru-RU')}</td>
                            ${showItem ? `<td class="p-4 font-mono">#${r.item_id}</td>` : ''}
                            <td class="p-4 font-bold">${r.action}</td>
                            <td class="p-4">${r.changed_by}</td>
                            <td class="p-4 text-right">
                                <button onclick='App.showDiff(${JSON.stringify(r.old_data || {})}, ${JSON.stringify(r.new_data || {})})' 
                                        class="text-blue-600 hover:underline">Показать diff</button>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
    },

    showDiff(oldData, newData) {
        const diff = Diff.diffLines(
            JSON.stringify(oldData, null, 2),
            JSON.stringify(newData, null, 2)
        );
        let html = '';
        diff.forEach(part => {
            const cls = part.added ? 'text-emerald-400' : part.removed ? 'text-red-400 line-through' : 'text-slate-300';
            html += `<div class="${cls} whitespace-pre">${part.value}</div>`;
        });
        document.getElementById('diff-content').innerHTML = html;
        UI.showModal('diff-modal');
    },

    async exportCSV() {
        try {
            const blob = await HttpClient.request('/history/export', 'GET', null);
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `history_${new Date().toISOString().slice(0,10)}.csv`;
            a.click();
            URL.revokeObjectURL(url);
        } catch (e) {}
    }
};

// Инициализация
if (State.token) Auth.initApp();

// Enter в поиске
document.getElementById('search-items').addEventListener('keypress', e => {
    if (e.key === 'Enter') App.loadItems();
});