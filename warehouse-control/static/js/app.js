const State = {
    token: localStorage.getItem('token') || '',
    refreshToken: localStorage.getItem('refresh_token') || '',
    role: localStorage.getItem('role') || '',
    user: localStorage.getItem('username') || '',
    selectedItems: new Set(),
    currentItems: [],
    isBulkMode: false
};

const HttpClient = {
    async request(url, method = 'GET', body = null, headers = {}) {
        const loader = document.getElementById('global-loader');
        if (loader) loader.style.transform = 'scaleX(0.3)';
        
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${State.token}`,
                ...headers
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
        try {
            return JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')));
        } catch (e) {
            return null;
        }
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
        } catch (err) {
            btn.innerHTML = original;
        }
    },
    
    async refreshToken() {
        try {
            const data = await HttpClient.request('/auth/refresh', 'POST', {
                refresh_token: State.refreshToken
            });
            State.token = data.access_token;
            localStorage.setItem('token', State.token);
            return State.token;
        } catch {
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
        RoleModule.init(State.role);
        App.loadItems();
    }
};

const RoleModule = {
    init(role) {
        console.log(`Role module active: ${role}`);
        const addBtn = document.getElementById('add-btn');
        const bulkDeleteBtn = document.getElementById('bulk-delete-btn');
        const selectAllBtn = document.getElementById('select-all-btn');
        
        if (addBtn) addBtn.classList.add('hidden');
        if (bulkDeleteBtn) bulkDeleteBtn.classList.add('hidden');
        if (selectAllBtn) selectAllBtn.classList.add('hidden');
        
        if (role === 'admin') {
            if (addBtn) addBtn.classList.remove('hidden');
            if (bulkDeleteBtn) {
                bulkDeleteBtn.classList.remove('hidden');
                bulkDeleteBtn.disabled = true;
            }
            if (selectAllBtn) selectAllBtn.classList.remove('hidden');
            this.addUserMgmtButton();
        }
        else if (role === 'manager') {
            if (addBtn) addBtn.classList.remove('hidden');
            if (selectAllBtn) selectAllBtn.classList.remove('hidden');
            if (bulkDeleteBtn) bulkDeleteBtn.remove();
        }
        else {
            if (addBtn) addBtn.remove();
            if (bulkDeleteBtn) bulkDeleteBtn.remove();
            if (selectAllBtn) selectAllBtn.remove();
            if (window.App) {
                const originalRender = App.renderItemsTable.bind(App);
                App.renderItemsTable = (items) => {
                    originalRender(items);
                    document.querySelectorAll('#items-table tbody tr').forEach(tr => {
                        const actionTd = tr.querySelector('td:last-child');
                        if (actionTd) {
                            actionTd.innerHTML = '<span class="text-xs text-slate-300 italic">Только просмотр</span>';
                        }
                    });
                };
            }
        }
    },
    
    addUserMgmtButton() {
        const nav = document.getElementById('role-nav');
        if (nav && !document.getElementById('user-mgmt-btn')) {
            const btn = document.createElement('button');
            btn.id = 'user-mgmt-btn';
            btn.className = "tab px-6 py-2 rounded-lg text-sm font-bold transition-all text-slate-500 hover:text-blue-600";
            btn.innerHTML = '<i class="fas fa-user-shield mr-2"></i>Сотрудники';
            btn.onclick = () => this.showUserModal();
            nav.appendChild(btn);
        }
    },
    
    showUserModal() {
        document.getElementById('reg-modal')?.remove();
        const modalHtml = `
<div id="reg-modal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-md">
    <div class="bg-white rounded-[2.5rem] p-10 w-full max-w-md shadow-2xl relative border border-slate-100">
        <div class="text-center mb-8">
            <div class="w-16 h-16 bg-blue-50 text-blue-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <i class="fas fa-user-plus fa-xl"></i>
            </div>
            <h2 class="text-2xl font-bold text-slate-800">Новый сотрудник</h2>
            <p class="text-slate-500 text-sm">Создание учетной записи в SSO</p>
        </div>
        <form id="reg-form" class="space-y-4">
            <div class="space-y-1">
                <label class="text-xs font-bold text-slate-400 ml-1 uppercase">Логин</label>
                <input type="text" id="reg-username" required placeholder="ivan_pro"
                    class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-500">
            </div>
            <div class="space-y-1">
                <label class="text-xs font-bold text-slate-400 ml-1 uppercase">Пароль</label>
                <input type="password" id="reg-password" required placeholder="••••••"
                    class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-500">
            </div>
            <div class="space-y-1">
                <label class="text-xs font-bold text-slate-400 ml-1 uppercase">Уровень доступа</label>
                <select id="reg-role" class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="viewer">Viewer</option>
                    <option value="manager">Manager</option>
                    <option value="admin">Admin</option>
                </select>
            </div>
            <div class="flex gap-3 pt-6">
                <button type="button" onclick="document.getElementById('reg-modal').remove()"
                    class="flex-1 px-6 py-4 bg-slate-100 text-slate-600 font-bold rounded-2xl hover:bg-slate-200">
                    Отмена
                </button>
                <button type="submit" id="reg-submit-btn"
                    class="flex-1 px-6 py-4 bg-blue-600 text-white font-bold rounded-2xl hover:bg-blue-700 shadow-lg">
                    Создать
                </button>
            </div>
        </form>
    </div>
</div>`;
        document.body.insertAdjacentHTML('beforeend', modalHtml);
        
        document.getElementById('reg-form').onsubmit = async (e) => {
            e.preventDefault();
            const btn = document.getElementById('reg-submit-btn');
            const originalText = btn.innerText;
            btn.disabled = true;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
            
            const payload = {
                username: document.getElementById('reg-username').value.trim(),
                password: document.getElementById('reg-password').value,
                role: document.getElementById('reg-role').value
            };
            
            try {
                await HttpClient.request('/auth/register', 'POST', payload);
                HttpClient.showToast(`Сотрудник ${payload.username} успешно добавлен!`, 'success');
                document.getElementById('reg-modal').remove();
            } catch (err) {
                btn.disabled = false;
                btn.innerText = originalText;
            }
        };
    }
};

const App = {
    showItemModal() {
        const form = document.getElementById('item-form');
        if (form) form.reset();
        document.getElementById('item-id').value = '';
        document.getElementById('item-name').value = '';
        document.getElementById('item-sku').value = '';
        document.getElementById('item-quantity').value = '';
        document.getElementById('item-price').value = '';
        document.getElementById('item-category').value = '';
        document.getElementById('item-location').value = '';
        document.getElementById('modal-title').innerText = 'Добавить в реестр';
        UI.showModal('item-modal');
    },
    
    async loadItems() {
        const query = document.getElementById('search-items')?.value || '';
        const data = await HttpClient.request(`/items?search=${query}`);
        State.currentItems = data?.items || [];
        this.renderItemsTable(State.currentItems);
    },
    
    renderItemsTable(items) {
        const container = document.getElementById('items-table');
        if (!items || items.length === 0) {
            container.innerHTML = '<div class="p-20 text-center text-slate-400 font-medium">Склад пуст</div>';
            return;
        }
        
        let html = `<table class="w-full text-left">
            <thead><tr class="bg-slate-50 border-b border-slate-100">
                <th class="p-4 w-16"></th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">ID</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">Название</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">SKU</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">Кол-во</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase">Цена</th>
                <th class="p-4 text-xs font-bold text-slate-500 uppercase text-center">Действия</th>
            </tr></thead><tbody class="divide-y">`;
        
        items.forEach(item => {
            const isSelected = State.selectedItems.has(item.id);
            const quantity = item.quantity ?? 0;
            const price = item.price ?? 0;
            
            html += `<tr class="hover:bg-slate-50 group ${isSelected ? 'bg-blue-50' : ''}" data-item-id="${item.id}">
                <td class="p-4">
                    <input type="checkbox" value="${item.id}" onchange="App.toggleSelect(${item.id})" ${isSelected ? 'checked' : ''}>
                </td>
                <td class="p-4 text-sm font-mono text-slate-500">#${item.id}</td>
                <td class="p-4">
                    <div class="font-bold text-slate-800">${item.name}</div>
                </td>
                <td class="p-4 text-sm text-slate-600">${item.sku}</td>
                <td class="p-4">
                    <span class="px-2 py-1 rounded-lg text-xs font-bold ${
                        quantity < 5 ? 'bg-red-100 text-red-600' : 'bg-blue-100 text-blue-600'
                    }">${quantity} шт.</span>
                </td>
                <td class="p-4 text-sm text-slate-600">${price} ₽</td>
                <td class="p-4 text-center">
                    <button onclick="App.showItemHistory(${item.id})" class="p-2 hover:text-blue-600" title="История">
                        <i class="fas fa-history"></i>
                    </button>
                    <button onclick="App.editItem(${item.id})" class="p-2 hover:text-emerald-600" title="Редактировать">
                        <i class="fas fa-pen"></i>
                    </button>
                </td>
            </tr>`;
        });
        
        container.innerHTML = html + '</tbody></table>';
        this.updateBulkDeleteButton();
        this.updateSelectAllButton();
    },
    
    async editItem(id) {
        try {
            const item = await HttpClient.request(`/items/${id}`);
            UI.showModal('item-modal');
            document.getElementById('modal-title').innerText = 'Редактирование';
            document.getElementById('item-id').value = item.id;
            document.getElementById('item-name').value = item.name;
            document.getElementById('item-sku').value = item.sku;
            document.getElementById('item-quantity').value = item.quantity ?? 0;
            document.getElementById('item-price').value = item.price ?? 0;
            document.getElementById('item-category').value = item.category || '';
            document.getElementById('item-location').value = item.location || '';
        } catch (err) {
            console.error('Failed to load item:', err);
        }
    },
    
    async saveItem(e) {
        e.preventDefault();
        
        const quantity = parseInt(document.getElementById('item-quantity').value) || 0;
        if (quantity < 0) {
            HttpClient.showToast('Количество не может быть отрицательным', 'error');
            return;
        }
        
        const price = parseFloat(document.getElementById('item-price').value) || 0;
        if (price < 0) {
            HttpClient.showToast('Цена не может быть отрицательной', 'error');
            return;
        }
        
        const id = document.getElementById('item-id').value;
        const payload = {
            name: document.getElementById('item-name').value,
            sku: document.getElementById('item-sku').value,
            quantity: quantity,
            price: price,
            category: document.getElementById('item-category').value,
            location: document.getElementById('item-location').value
        };
        
        try {
            const response = await HttpClient.request(id ? `/items/${id}` : '/items', id ? 'PUT' : 'POST', payload);
            
            if (id) {
                const itemId = Number(id);
                const index = State.currentItems.findIndex(i => i.id === itemId);
                if (index !== -1) {
                    State.currentItems[index] = response || { ...State.currentItems[index], ...payload };
                }
            } else {
                const newItem = response && response.id 
                    ? { ...response, quantity: response.quantity ?? payload.quantity, price: response.price ?? payload.price }
                    : { ...payload, id: Date.now() };
                State.currentItems.unshift(newItem);
            }
            
            HttpClient.showToast('Успешно сохранено', 'success');
            UI.closeModal('item-modal');
            this.renderItemsTable(State.currentItems);
        } catch (err) {
            console.error('Save error:', err);
        }
    },
    
    async showItemHistory(itemId) {
        try {
            const data = await HttpClient.request(`/history/item/${itemId}`);
            const records = data?.records || data?.history || [];
            const itemName = State.currentItems.find(i => i.id === itemId)?.name || `Товар #${itemId}`;
            this.renderHistoryModal(itemId, itemName, records);
        } catch (err) {
            console.error('Failed to load item history:', err);
            HttpClient.showToast('Не удалось загрузить историю', 'error');
        }
    },
    
    renderHistoryModal(itemId, itemName, records) {
        document.getElementById('history-modal')?.remove();
        
        let historyRows = '';
        if (records.length === 0) {
            historyRows = '<tr><td colspan="6" class="p-8 text-center text-slate-400">История пуста</td></tr>';
        } else {
            records.forEach(rec => {
                const oldData = rec.old_data || rec.oldData || {};
                const newData = rec.new_data || rec.newData || {};
                
                const changes = [];
                if (oldData?.name !== newData?.name) changes.push('Название');
                if (oldData?.sku !== newData?.sku) changes.push('SKU');
                if (oldData?.quantity !== newData?.quantity) changes.push('Кол-во');
                if (oldData?.price !== newData?.price) changes.push('Цена');
                if (oldData?.category !== newData?.category) changes.push('Категория');
                if (oldData?.location !== newData?.location) changes.push('Место');
                
                let changesText = '';
                if (rec.action === 'INSERT') {
                    changesText = '<span class="text-emerald-600 font-bold">Товар создан</span>';
                } else if (rec.action === 'DELETE') {
                    changesText = '<span class="text-red-600 font-bold">Товар удален</span>';
                } else {
                    changesText = changes.length > 0 
                        ? changes.join(', ') 
                        : '<span class="text-slate-400 italic">Без изменений</span>';
                }
                
                historyRows += `
                <tr class="hover:bg-slate-50">
                    <td class="p-4 text-xs">${new Date(rec.changed_at || rec.changed).toLocaleString('ru-RU')}</td>
                    <td class="p-4 text-sm font-mono text-slate-500">#${rec.id}</td>
                    <td class="p-4">
                        <span class="px-2 py-1 rounded text-xs font-bold ${
                            rec.action === 'INSERT' ? 'bg-emerald-100 text-emerald-600' :
                            rec.action === 'UPDATE' ? 'bg-blue-100 text-blue-600' :
                            'bg-red-100 text-red-600'
                        }">${rec.action}</span>
                    </td>
                    <td class="p-4 text-sm">${rec.changed_by || '—'}</td>
                    <td class="p-4 text-sm max-w-md">${changesText}</td>
                    <td class="p-4 text-center">
                        <button onclick="App.showVersionDiff(${rec.id}, ${rec.item_id})" class="text-xs text-blue-600 hover:underline font-bold">
                            <i class="fas fa-columns mr-1"></i>Сравнить
                        </button>
                    </td>
                </tr>`;
            });
        }
        
        const modalHtml = `
<div id="history-modal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-md">
    <div class="bg-white rounded-3xl p-8 w-full max-w-5xl shadow-2xl relative max-h-[90vh] overflow-y-auto">
        <div class="flex justify-between items-center mb-6">
            <div>
                <h2 class="text-xl font-bold">История изменений</h2>
                <p class="text-sm text-slate-500 mt-1">${itemName} (ID: ${itemId})</p>
            </div>
            <button onclick="document.getElementById('history-modal').remove()" class="w-10 h-10 rounded-full bg-slate-100 hover:bg-slate-200">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="overflow-x-auto">
            <table class="w-full text-left">
                <thead class="bg-slate-50">
                    <tr>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Время</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">ID записи</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Действие</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Кто</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Изменения</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase text-center">Действия</th>
                    </tr>
                </thead>
                <tbody class="divide-y">${historyRows}</tbody>
            </table>
        </div>
    </div>
</div>`;
        document.body.insertAdjacentHTML('beforeend', modalHtml);
    },
    
    async showVersionDiff(recordId, itemId) {
        try {
            const data = await HttpClient.request(`/history/diff/${recordId}`);
            this.renderVersionDiffModal(data);
        } catch (err) {
            console.error('Failed to load version diff:', err);
            HttpClient.showToast('Не удалось загрузить сравнение', 'error');
        }
    },
    
    renderVersionDiffModal(data) {
        document.getElementById('version-diff-modal')?.remove();
        
        if (!data || !data.fields || data.fields.length === 0) {
            HttpClient.showToast('Нет данных для сравнения', 'error');
            return;
        }
        
        let rows = '';
        data.fields.forEach(field => {
            let rowClass = '';
            let statusHtml = '';
            
            switch (field.status) {
                case 'ADDED':
                case 'added':
                    rowClass = 'diff-added';
                    statusHtml = '<span class="px-2 py-1 bg-emerald-100 text-emerald-700 rounded text-xs font-bold">ДОБАВЛЕНО</span>';
                    break;
                case 'REMOVED':
                case 'removed':
                    rowClass = 'diff-removed';
                    statusHtml = '<span class="px-2 py-1 bg-red-100 text-red-700 rounded text-xs font-bold">УДАЛЕНО</span>';
                    break;
                case 'CHANGED':
                case 'changed':
                    rowClass = 'diff-changed';
                    statusHtml = '<span class="px-2 py-1 bg-amber-100 text-amber-700 rounded text-xs font-bold">ИЗМЕНЕНО</span>';
                    break;
                default:
                    statusHtml = '<span class="text-slate-400 text-xs">Без изменений</span>';
            }
            
            rows += `
            <tr class="${rowClass} border-b border-slate-100 last:border-0">
                <td class="p-4 font-bold text-slate-600 w-1/4">${field.label}</td>
                <td class="p-4 w-1/4 ${field.status === 'REMOVED' || field.status === 'removed' ? 'text-red-600 line-through' : 'text-slate-600'}">
                    ${field.old !== undefined && field.old !== null ? field.old : '<span class="text-slate-400 italic">—</span>'}
                </td>
                <td class="p-4 w-1/4 ${field.status === 'REMOVED' || field.status === 'removed' ? 'text-slate-400' : 'text-emerald-600 font-bold'}">
                    ${field.new !== undefined && field.new !== null ? field.new : '<span class="text-slate-400 italic">—</span>'}
                </td>
                <td class="p-4 w-1/4 text-center">${statusHtml}</td>
            </tr>`;
        });
        
        const modalHtml = `
<div id="version-diff-modal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-md">
    <div class="bg-white rounded-3xl p-8 w-full max-w-4xl shadow-2xl relative">
        <div class="flex justify-between items-center mb-6">
            <div>
                <h2 class="text-xl font-bold">Сравнение версий</h2>
                <p class="text-sm text-slate-500 mt-1">
                    Запись #${data.record_id} • ${new Date(data.changed_at).toLocaleString('ru-RU')} • ${data.changed_by || '—'}
                </p>
            </div>
            <button onclick="document.getElementById('version-diff-modal').remove()" class="w-10 h-10 rounded-full bg-slate-100 hover:bg-slate-200">
                <i class="fas fa-times"></i>
            </button>
        </div>
        <div class="mb-6">
            <span class="px-3 py-1.5 rounded-lg text-sm font-bold ${
                data.action === 'INSERT' ? 'bg-emerald-100 text-emerald-600' :
                data.action === 'UPDATE' ? 'bg-blue-100 text-blue-600' :
                'bg-red-100 text-red-600'
            }">${data.action}</span>
        </div>
        <div class="overflow-hidden rounded-xl border border-slate-200">
            <table class="w-full">
                <thead class="bg-slate-50">
                    <tr>
                        <th class="p-4 text-left text-xs font-bold text-slate-500 uppercase w-1/4">Поле</th>
                        <th class="p-4 text-left text-xs font-bold text-slate-500 uppercase w-1/4">Было</th>
                        <th class="p-4 text-left text-xs font-bold text-slate-500 uppercase w-1/4">Стало</th>
                        <th class="p-4 text-center text-xs font-bold text-slate-500 uppercase w-1/4">Статус</th>
                    </tr>
                </thead>
                <tbody class="bg-white">${rows}</tbody>
            </table>
        </div>
        <div class="flex gap-4 mt-6 text-xs text-slate-500">
            <div class="flex items-center gap-2">
                <div class="w-4 h-4 rounded bg-emerald-100"></div>
                <span>Добавлено</span>
            </div>
            <div class="flex items-center gap-2">
                <div class="w-4 h-4 rounded bg-red-100"></div>
                <span>Удалено</span>
            </div>
            <div class="flex items-center gap-2">
                <div class="w-4 h-4 rounded bg-amber-100"></div>
                <span>Изменено</span>
            </div>
        </div>
    </div>
</div>`;
        document.body.insertAdjacentHTML('beforeend', modalHtml);
    },
    
    async loadHistory() {
        const limit = document.getElementById('history-limit')?.value || '100';
        
        // ✅ ВАЛИДАЦИЯ ID ТОВАРА (не меньше 1)
        let itemId = document.getElementById('history-item-id')?.value || '';
        if (itemId && parseInt(itemId) < 1) {
            HttpClient.showToast('ID товара должен быть больше 0', 'error');
            return;
        }
        
        const filters = {
            limit: limit,
            item_id: itemId,
            action: document.getElementById('history-action')?.value || '',
            username: document.getElementById('history-username')?.value || '',
            date_from: document.getElementById('history-date-from')?.value || '',
            date_to: document.getElementById('history-date-to')?.value || ''
        };
        
        const queryParams = new URLSearchParams();
        if (filters.limit) queryParams.append('limit', filters.limit);
        if (filters.item_id) queryParams.append('item_id', filters.item_id);
        if (filters.action) queryParams.append('action', filters.action);
        if (filters.username) queryParams.append('username', filters.username);
        if (filters.date_from) queryParams.append('date_from', new Date(filters.date_from).toISOString());
        if (filters.date_to) {
            const dateTo = new Date(filters.date_to);
            dateTo.setHours(23, 59, 59, 999);
            queryParams.append('date_to', dateTo.toISOString());
        }
        
        try {
            const data = await HttpClient.request(`/history?${queryParams.toString()}`);
            const records = data?.records || data?.history || [];
            this.renderHistoryTable(records);
        } catch (err) {
            console.error('Failed to load history:', err);
            HttpClient.showToast('Не удалось загрузить историю', 'error');
        }
    },
    
    renderHistoryTable(records) {
        const container = document.getElementById('history-content');
        if (!container) return;
        
        if (!records || records.length === 0) {
            container.innerHTML = '<div class="p-20 text-center text-slate-400">История пуста</div>';
            return;
        }
        
        let html = `<div class="overflow-x-auto">
            <table class="w-full text-left">
                <thead class="bg-slate-50">
                    <tr>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Время</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">ID товара</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Название</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Действие</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Кто</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase">Изменения</th>
                        <th class="p-4 text-xs font-bold text-slate-500 uppercase text-center">Действия</th>
                    </tr>
                </thead>
                <tbody class="divide-y">`;
        
        records.forEach(rec => {
            const oldData = rec.old_data || rec.oldData || {};
            const newData = rec.new_data || rec.newData || {};
            
            const itemName = newData?.name || oldData?.name || '—';
            
            const changes = [];
            if (oldData?.name !== newData?.name) changes.push('Название');
            if (oldData?.sku !== newData?.sku) changes.push('SKU');
            if (oldData?.quantity !== newData?.quantity) changes.push('Кол-во');
            if (oldData?.price !== newData?.price) changes.push('Цена');
            if (oldData?.category !== newData?.category) changes.push('Категория');
            if (oldData?.location !== newData?.location) changes.push('Место');
            
            let changesText = '';
            if (rec.action === 'INSERT') {
                changesText = '<span class="text-emerald-600 font-bold">Товар создан</span>';
            } else if (rec.action === 'DELETE') {
                changesText = '<span class="text-red-600 font-bold">Товар удален</span>';
            } else {
                changesText = changes.length > 0 
                    ? changes.join(', ') 
                    : '<span class="text-slate-400 italic">—</span>';
            }
            
            html += `
            <tr class="hover:bg-slate-50">
                <td class="p-4 text-xs">${new Date(rec.changed_at || rec.changed).toLocaleString('ru-RU')}</td>
                <td class="p-4 text-sm font-mono text-slate-500">#${rec.item_id || rec.itemId}</td>
                <td class="p-4 text-sm font-bold text-slate-700">${itemName}</td>
                <td class="p-4">
                    <span class="px-2 py-1 rounded text-xs font-bold ${
                        rec.action === 'INSERT' ? 'bg-emerald-100 text-emerald-600' :
                        rec.action === 'UPDATE' ? 'bg-blue-100 text-blue-600' :
                        'bg-red-100 text-red-600'
                    }">${rec.action}</span>
                </td>
                <td class="p-4 text-sm">${rec.changed_by || '—'}</td>
                <td class="p-4 text-sm text-slate-500">
                    ${changesText}
                </td>
                <td class="p-4 text-center">
                    <button onclick="App.showVersionDiff(${rec.id}, ${rec.item_id})" class="text-xs text-blue-600 hover:underline font-bold">
                        <i class="fas fa-columns mr-1"></i>Сравнить
                    </button>
                </td>
            </tr>`;
        });
        
        html += '</tbody></table></div>';
        container.innerHTML = html;
    },
    
    async exportHistory() {
        const limit = document.getElementById('history-limit')?.value || '10000';
        
        let itemId = document.getElementById('history-item-id')?.value || '';
        if (itemId && parseInt(itemId) < 1) {
            HttpClient.showToast('ID товара должен быть больше 0', 'error');
            return;
        }
        
        const filters = {
            limit: limit,
            item_id: itemId,
            action: document.getElementById('history-action')?.value || '',
            username: document.getElementById('history-username')?.value || '',
            date_from: document.getElementById('history-date-from')?.value || '',
            date_to: document.getElementById('history-date-to')?.value || ''
        };
        
        const queryParams = new URLSearchParams();
        if (filters.limit) queryParams.append('limit', filters.limit);
        if (filters.item_id) queryParams.append('item_id', filters.item_id);
        if (filters.action) queryParams.append('action', filters.action);
        if (filters.username) queryParams.append('username', filters.username);
        if (filters.date_from) queryParams.append('date_from', new Date(filters.date_from).toISOString());
        if (filters.date_to) {
            const dateTo = new Date(filters.date_to);
            dateTo.setHours(23, 59, 59, 999);
            queryParams.append('date_to', dateTo.toISOString());
        }
        
        const url = `/history/export?${queryParams.toString()}`;
        
        try {
            const response = await fetch(url, {
                method: 'GET',
                headers: { 'Authorization': `Bearer ${State.token}` }
            });
            
            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText || 'Export failed');
            }
            
            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = `warehouse_history_${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(downloadUrl);
            HttpClient.showToast('Экспорт выполнен успешно', 'success');
        } catch (err) {
            console.error('Export error:', err);
            HttpClient.showToast('Ошибка экспорта: ' + err.message, 'error');
        }
    },
    
    toggleSelect(id) {
        if (State.selectedItems.has(id)) {
            State.selectedItems.delete(id);
        } else {
            State.selectedItems.add(id);
        }
        this.updateBulkDeleteButton();
        this.updateSelectAllButton();
        this.renderItemsTable(State.currentItems);
    },
    
    toggleAllItems() {
        const allCheckboxes = document.querySelectorAll('#items-table tbody input[type="checkbox"]');
        const allChecked = allCheckboxes.length > 0 && Array.from(allCheckboxes).every(c => c.checked);
        
        if (allChecked) {
            State.selectedItems.clear();
        } else {
            State.selectedItems.clear();
            State.currentItems.forEach(item => State.selectedItems.add(item.id));
        }
        this.updateBulkDeleteButton();
        this.updateSelectAllButton();
        this.renderItemsTable(State.currentItems);
    },
    
    updateSelectAllButton() {
        const btn = document.getElementById('select-all-btn');
        const text = document.getElementById('select-all-text');
        if (btn && text) {
            const allCheckboxes = document.querySelectorAll('#items-table tbody input[type="checkbox"]');
            const allChecked = allCheckboxes.length > 0 && Array.from(allCheckboxes).every(c => c.checked);
            text.textContent = allChecked ? 'Снять выделение' : 'Выбрать все';
        }
    },
    
    updateBulkDeleteButton() {
        const btn = document.getElementById('bulk-delete-btn');
        if (btn) {
            btn.disabled = State.selectedItems.size === 0;
            btn.innerHTML = `<i class="fas fa-trash-alt"></i> Удалить выбранные (${State.selectedItems.size})`;
        }
    },
    
    async bulkDelete() {
        if (State.selectedItems.size === 0) return;
        if (!confirm(`Удалить ${State.selectedItems.size} товаров?`)) return;
        
        try {
            await HttpClient.request('/items/bulk', 'DELETE', {
                ids: Array.from(State.selectedItems)
            });
            
            State.currentItems = State.currentItems.filter(item => !State.selectedItems.has(item.id));
            State.selectedItems.clear();
            
            HttpClient.showToast('Товары удалены', 'success');
            this.renderItemsTable(State.currentItems);
        } catch (err) {
            console.error('Bulk delete error:', err);
            HttpClient.showToast('Ошибка удаления', 'error');
        }
    }
};

const UI = {
    switchTab(index) {
        document.querySelectorAll('.tab').forEach((tab, i) => {
            tab.classList.toggle('active', i === index);
            tab.classList.toggle('text-slate-500', i !== index);
        });
        document.getElementById('tab-0').classList.toggle('hidden', index !== 0);
        document.getElementById('tab-1').classList.toggle('hidden', index !== 1);
        
        if (index === 0) {
            App.loadItems();
        } else if (index === 1) {
            App.loadHistory();
        }
    },
    
    showModal(id) {
        document.getElementById(id).classList.remove('hidden');
    },
    
    closeModal(id) { document.getElementById(id).classList.add('hidden'); }
};

if (State.token) Auth.initApp();