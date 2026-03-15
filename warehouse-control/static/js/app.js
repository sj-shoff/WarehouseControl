const State = {
    token: localStorage.getItem('token'),
    role: localStorage.getItem('role'),
    user: localStorage.getItem('username'),
    lang: localStorage.getItem('lang') || 'ru',
};

const HttpClient = {
    async request(url, method = 'GET', body = null) {
        this.setProgress(30); // Начало загрузки
        
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${State.token}`
            }
        };
        if (body) options.body = JSON.stringify(body);

        try {
            const response = await fetch(url, options);
            this.setProgress(70);

            if (response.status === 401) Auth.logout();
            
            if (!response.ok) {
                const err = await response.json();
                throw new Error(err.error || 'Server error');
            }

            this.setProgress(100);
            return await response.json();
        } catch (err) {
            this.setProgress(0);
            this.showToast(err.message, 'error');
            throw err;
        } finally {
            setTimeout(() => this.setProgress(0), 400);
        }
    },

    setProgress(percent) {
        const loader = document.getElementById('global-loader');
        if (loader) {
            loader.style.transform = `scaleX(${percent / 100})`;
        }
    },

    showToast(msg, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `fixed bottom-5 right-5 px-6 py-3 rounded-2xl shadow-2xl text-white transition-all transform translate-y-20 z-[200] ${
            type === 'error' ? 'bg-red-500' : 'bg-emerald-500'
        }`;
        toast.textContent = msg;
        document.body.appendChild(toast);
        
        setTimeout(() => toast.classList.remove('translate-y-20'), 100);
        setTimeout(() => {
            toast.classList.add('translate-y-20');
            setTimeout(() => toast.remove(), 500);
        }, 3000);
    }
};

const UI = {
    switchTab(index) {
        document.querySelectorAll('.tab').forEach((tab, i) => {
            if (i === index) {
                tab.classList.add('active', 'text-blue-600');
                tab.classList.remove('text-slate-500');
            } else {
                tab.classList.remove('active', 'text-blue-600');
                tab.classList.add('text-slate-500');
            }
        });

        document.getElementById('tab-0').classList.toggle('hidden', index !== 0);
        document.getElementById('tab-1').classList.toggle('hidden', index !== 1);
        
        if (index === 1) App.loadHistory();
    },

    showModal(id) {
        const modal = document.getElementById(id);
        modal.classList.remove('hidden');
        setTimeout(() => modal.querySelector('.bg-white').classList.remove('scale-95'), 10);
    },

    closeModal(id) {
        const modal = document.getElementById(id);
        modal.querySelector('.bg-white').classList.add('scale-95');
        setTimeout(() => modal.classList.add('hidden'), 200);
    }
};

const Auth = {
    async login() {
        const btn = event.target;
        const originalText = btn.innerHTML;
        btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        
        try {
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const data = await HttpClient.request('/auth/login', 'POST', { username, password });
            
            State.token = data.token;
            State.role = data.role;
            State.user = data.username;
            
            localStorage.setItem('token', State.token);
            localStorage.setItem('role', State.role);
            localStorage.setItem('username', State.user);

            this.initApp();
        } catch (e) {
            btn.innerHTML = originalText;
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
        document.getElementById('current-user').textContent = `${State.user} (${State.role})`;
        
        App.loadItems();
        if (State.role === 'viewer') {
            document.getElementById('add-btn').style.display = 'none';
        }
    }
};

const App = {
    async loadItems() {
        const table = document.getElementById('items-table');
        // Показываем шиммер (скелетон) перед загрузкой
        table.innerHTML = '<div class="p-10 space-y-4"><div class="h-8 bg-slate-100 rounded loading-shimmer"></div><div class="h-8 bg-slate-100 rounded loading-shimmer"></div></div>';
        
        const search = document.getElementById('search-items').value;
        const url = `/api/items?limit=10&offset=0${search ? `&search=${search}` : ''}`;
        
        try {
            const data = await HttpClient.request(url);
            this.renderItems(data.items);
        } catch (e) {}
    },

    renderItems(items) {
        const table = document.getElementById('items-table');
        if (!items.length) {
            table.innerHTML = '<div class="p-20 text-center text-slate-400">Inventory is empty</div>';
            return;
        }

        let html = `
            <table class="w-full">
                <thead class="bg-slate-50 border-b border-slate-100">
                    <tr>
                        <th class="p-4 text-xs font-bold text-slate-400 uppercase tracking-wider text-left">Product</th>
                        <th class="p-4 text-xs font-bold text-slate-400 uppercase tracking-wider text-left">SKU</th>
                        <th class="p-4 text-xs font-bold text-slate-400 uppercase tracking-wider text-left">Quantity</th>
                        <th class="p-4 text-xs font-bold text-slate-400 uppercase tracking-wider text-right">Actions</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-50">
                    ${items.map(item => `
                        <tr class="hover:bg-slate-50/50 transition-colors">
                            <td class="p-4">
                                <div class="font-bold text-slate-700">${item.name}</div>
                                <div class="text-xs text-slate-400">${item.category}</div>
                            </td>
                            <td class="p-4 font-mono text-sm">${item.sku}</td>
                            <td class="p-4">
                                <span class="px-3 py-1 rounded-full text-xs font-bold ${item.quantity < 10 ? 'bg-orange-100 text-orange-600' : 'bg-emerald-100 text-emerald-600'}">
                                    ${item.quantity} units
                                </span>
                            </td>
                            <td class="p-4 text-right space-x-2">
                                ${State.role !== 'viewer' ? `
                                    <button onclick="App.editItem(${item.id})" class="text-slate-400 hover:text-blue-600"><i class="fas fa-edit"></i></button>
                                    <button onclick="App.deleteItem(${item.id})" class="text-slate-400 hover:text-red-500"><i class="fas fa-trash"></i></button>
                                ` : ''}
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
        table.innerHTML = html;
    }
};

if (State.token) Auth.initApp();