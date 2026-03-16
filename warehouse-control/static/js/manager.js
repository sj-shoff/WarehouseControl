/**
 * Role Module: MANAGER
 */
window.RoleModule = {
    init() {
        console.log("Manager module active");
        
        // Менеджеру можно добавлять
        document.getElementById('add-btn')?.classList.remove('hidden');
        
        // Но нельзя массово удалять (скрываем кнопку)
        document.getElementById('bulk-delete-btn')?.remove();

        if (window.App) App.loadItems();
    }
};