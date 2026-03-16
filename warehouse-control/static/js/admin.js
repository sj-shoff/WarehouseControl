/**
 * Role Module: ADMIN
 */
window.RoleModule = {
    init() {
        console.log("Admin module active");
        
        // Показываем кнопки управления складом
        document.getElementById('add-btn')?.classList.remove('hidden');
        document.getElementById('bulk-delete-btn')?.classList.remove('hidden');

        // Добавляем кнопку управления сотрудниками в навигацию
        const nav = document.getElementById('role-nav');
        if (nav && !document.getElementById('user-mgmt-btn')) {
            const btn = document.createElement('button');
            btn.id = 'user-mgmt-btn';
            btn.className = "tab px-6 py-2 rounded-lg text-sm font-bold transition-all text-slate-500 hover:text-blue-600";
            btn.innerHTML = '<i class="fas fa-user-shield mr-2"></i>Сотрудники';
            btn.onclick = () => this.showUserModal();
            nav.appendChild(btn);
        }

        if (window.App) App.loadItems();
    },

    showUserModal() {
        // Удаляем старую модалку, если она вдруг осталась
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
                                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-500 transition-all">
                        </div>
                        
                        <div class="space-y-1">
                            <label class="text-xs font-bold text-slate-400 ml-1 uppercase">Пароль</label>
                            <input type="password" id="reg-password" required placeholder="••••••••" 
                                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-500 transition-all">
                        </div>

                        <div class="space-y-1">
                            <label class="text-xs font-bold text-slate-400 ml-1 uppercase">Уровень доступа</label>
                            <select id="reg-role" class="w-full px-5 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl outline-none focus:ring-2 focus:ring-blue-500 transition-all appearance-none cursor-pointer">
                                <option value="viewer">Viewer (Только чтение)</option>
                                <option value="manager">Manager (Склад)</option>
                                <option value="admin">Admin (Полный доступ)</option>
                            </select>
                        </div>

                        <div class="flex gap-3 pt-6">
                            <button type="button" onclick="document.getElementById('reg-modal').remove()" 
                                class="flex-1 px-6 py-4 bg-slate-100 text-slate-600 font-bold rounded-2xl hover:bg-slate-200 transition-colors">
                                Отмена
                            </button>
                            <button type="submit" id="reg-submit-btn" 
                                class="flex-1 px-6 py-4 bg-blue-600 text-white font-bold rounded-2xl hover:bg-blue-700 shadow-lg shadow-blue-200 transition-all active:scale-95">
                                Создать
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        `;

        document.body.insertAdjacentHTML('beforeend', modalHtml);

        const form = document.getElementById('reg-form');
        form.onsubmit = async (e) => {
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
                // Вызываем твой SSO хэндлер
                await HttpClient.request('/auth/register', 'POST', payload);
                
                HttpClient.showToast(`Сотрудник ${payload.username} успешно добавлен!`, 'success');
                document.getElementById('reg-modal').remove();
            } catch (err) {
                console.error("Registration error:", err);
                btn.disabled = false;
                btn.innerText = originalText;
            }
        };
    }
};