/**
 * Role Module: VIEWER
 */
window.RoleModule = {
    init() {
        console.log("Viewer module active (Read-only)");
        
        // Прячем все кнопки управления
        document.getElementById('add-btn')?.remove();
        document.getElementById('bulk-delete-btn')?.remove();

        // Перехватываем рендер таблицы, чтобы убрать колонку действий или кнопки в ней
        if (window.App) {
            const originalRender = App.renderItemsTable.bind(App);
            App.renderItemsTable = (items) => {
                originalRender(items);
                // Находим все ячейки с кнопками и чистим их
                document.querySelectorAll('#items-table tr').forEach(tr => {
                    const actionTd = tr.querySelector('td:last-child');
                    if (actionTd) actionTd.innerHTML = '<span class="text-xs text-slate-300 italic">Только просмотр</span>';
                });
            };
            App.loadItems();
        }
    }
};