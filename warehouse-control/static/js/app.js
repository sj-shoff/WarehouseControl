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

            if (res.status === 401 && State.refreshToken && !url.includes('/auth/')) {
                const newToken = await Auth.refreshToken();
                if (newToken) {
                    options.headers.Authorization = `Bearer ${newToken}`;
                    res = await fetch(url, options);
                }
            }

            if (!res.ok) {
                // Пытаемся достать текст ошибки
                let errorMsg = `Ошибка сервера: ${res.status}`;
                try {
                    const errorData = await res.json();
                    errorMsg = errorData.error || errorData.message || errorMsg;
                } catch (e) {
                    const textErr = await res.text();
                    if (textErr) errorMsg = textErr;
                }
                throw new Error(errorMsg);
            }

            if (loader) loader.style.transform = 'scaleX(1)';
            const text = await res.text();
            return text ? JSON.parse(text) : null;
        } catch (err) {
            // ГАРАНТИРУЕМ, что в тост попадет строка
            const finalMsg = err.message || String(err) || "Неизвестная ошибка";
            this.showToast(finalMsg, 'error');
            throw err;
        } finally {
            setTimeout(() => { if (loader) loader.style.transform = 'scaleX(0)'; }, 400);
        }
    },

    showToast(msg, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `fixed bottom-6 right-6 px-6 py-3.5 rounded-2xl shadow-2xl text-white z-[9999] transition-all transform animate-bounce-short ${
            type === 'error' ? 'bg-red-600' : 'bg-emerald-600'
        }`;
        toast.innerHTML = `<i class="fas ${type === 'error' ? 'fa-circle-exclamation' : 'fa-check-circle'} mr-2"></i>${msg}`;
        document.body.appendChild(toast);
        setTimeout(() => toast.remove(), 4000);
    }
};

const Auth = {
    parseJwt(token) {
        try { return JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/'))); } catch (e) { return null; }
    },

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
            const p = this.parseJwt(data.access_token);
            State.token = data.access_token;
            State.refreshToken = data.refresh_token;
            State.user = p?.username || 'User';
            State.role = p?.role || 'viewer';
            localStorage.setItem('token', State.token);
            localStorage.setItem('refresh_token', State.refreshToken);
            localStorage.setItem('username', State.user);
            localStorage.setItem('role', State.role);
            this.initApp();
        } catch (err) { btn.innerHTML = original; }
    },

    async refreshToken() {
        try {
            const data = await HttpClient.request('/auth/refresh', 'POST', { refresh_token: State.refreshToken });
            State.token = data.access_token;
            localStorage.setItem('token', State.token);
            return State.token;
        } catch { return null; }
    },

    logout() { localStorage.clear(); location.reload(); },

    initApp() {
        document.getElementById('login').classList.add('hidden');
        document.getElementById('app').classList.remove('hidden');
        document.getElementById('user-profile').classList.remove('hidden');
        document.getElementById('current-user').textContent = State.user;
        document.getElementById('current-role').textContent = State.role.toUpperCase();

        const script = document.createElement('script');
        script.src = `/static/js/${State.role}.js`;
        script.onload = () => window.RoleModule?.init();
        document.head.appendChild(script);
        App.loadItems();
    }
};

const App = {
    // Вызов модалки добавления (то, что у тебя не работало)
    showItemModal() {
        const form = document.getElementById('item-form');
        if (form) form.reset();
        document.getElementById('item-id').value = '';
        document.getElementById('modal-title').innerText = 'Добавить в реестр';
        UI.showModal('item-modal');
    },

    async loadItems() {
        const query = document.getElementById('search-items')?.value || '';
        const items = await HttpClient.request(`/items?search=${query}`);
        this.renderItemsTable(items || []);
    },

    renderItemsTable(items) {
        const container = document.getElementById('items-table');
        if (!items || items.length === 0) {
            container.innerHTML = '<div class="p-20 text-center text-slate-400 font-medium">Склад пуст</div>';
            return;
        }
        let html = `<table class="w-full text-left">
            <thead><tr class="bg-slate-50 border-b border-slate-100">
                <th class="p-4 w-12"><input type="checkbox" onclick="App.toggleAll(this)"></th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">Товар</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">SKU</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">Кол-во</th>
                <th class="p-4"></th>
            </tr></thead><tbody class="divide-y">`;

        items.forEach(item => {
            html += `<tr class="hover:bg-slate-50 group">
                <td class="p-4"><input type="checkbox" value="${item.id}" onchange="App.toggleSelect(${item.id})"></td>
                <td class="p-4"><div class="font-bold">${item.name}</div><div class="text-[10px] text-slate-400">${item.id}</div></td>
                <td class="p-4 text-sm">${item.sku}</td>
                <td class="p-4"><span class="px-2 py-1 rounded-lg text-xs font-bold ${item.quantity < 5 ? 'bg-red-100 text-red-600' : 'bg-blue-100 text-blue-600'}">${item.quantity} шт.</span></td>
                <td class="p-4 text-right">
                    <button onclick="App.editItem(${JSON.stringify(item).replace(/"/g, '&quot;')})" class="p-2 hover:text-emerald-600"><i class="fas fa-pen"></i></button>
                </td>
            </tr>`;
        });
        container.innerHTML = html + '</tbody></table>';
    },

    editItem(item) {
        UI.showModal('item-modal');
        document.getElementById('modal-title').innerText = 'Редактирование';
        document.getElementById('item-id').value = item.id;
        document.getElementById('item-name').value = item.name;
        document.getElementById('item-sku').value = item.sku;
        document.getElementById('item-quantity').value = item.quantity;
    },

    async saveItem(e) {
        e.preventDefault();
        const id = document.getElementById('item-id').value;
        const payload = {
            name: document.getElementById('item-name').value,
            sku: document.getElementById('item-sku').value,
            quantity: parseInt(document.getElementById('item-quantity').value)
        };
        try {
            await HttpClient.request(id ? `/items/${id}` : '/items', id ? 'PUT' : 'POST', payload);
            HttpClient.showToast('Успешно сохранено', 'success');
            UI.closeModal('item-modal');
            this.loadItems();
        } catch (err) {}
    },

    async loadHistory() {
        const history = await HttpClient.request('/history');
        const container = document.getElementById('tab-1');
        if (!history || !history.length) {
            container.innerHTML = '<div class="p-20 text-center text-slate-400">История пуста</div>';
            return;
        }
        container.innerHTML = `<table class="w-full text-left">
            <thead class="bg-slate-50"><tr><th class="p-4">Время</th><th class="p-4">Действие</th><th class="p-4">Кто</th></tr></thead>
            <tbody class="divide-y">${history.map(h => `
                <tr><td class="p-4 text-xs">${new Date(h.changed_at).toLocaleString()}</td><td class="p-4 font-bold">${h.operation}</td><td class="p-4">${h.username || '—'}</td></tr>
            `).join('')}</tbody></table>`;
    },

    toggleSelect(id) { State.selectedItems.has(id) ? State.selectedItems.delete(id) : State.selectedItems.add(id); },
    toggleAll(m) { document.querySelectorAll('#items-table input[type="checkbox"]').forEach(c => c.checked = m.checked); }
};

const UI = {
    switchTab(index) {
        document.querySelectorAll('.tab').forEach((tab, i) => tab.classList.toggle('active', i === index));
        document.getElementById('tab-0').classList.toggle('hidden', index !== 0);
        document.getElementById('tab-1').classList.toggle('hidden', index !== 1);
        index === 0 ? App.loadItems() : App.loadHistory();
    },
    showModal(id) { document.getElementById(id).classList.remove('hidden'); },
    closeModal(id) { document.getElementById(id).classList.add('hidden'); }
};

if (State.token) Auth.initApp();