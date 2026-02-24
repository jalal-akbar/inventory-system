document.addEventListener('DOMContentLoaded', () => {
    fetchInventory();
    lucide.createIcons();
});

const t = (key, replacements = {}) => {
    let text = (window.TRANSLATIONS && window.TRANSLATIONS[key]) ? window.TRANSLATIONS[key] : key;
    for (const [k, v] of Object.entries(replacements)) {
        text = text.replace(`{${k}}`, v);
    }
    return text;
};

function fetchInventory() {
    const searchInput = document.getElementById('inventorySearch');
    const search = searchInput ? searchInput.value : '';

    fetch(`/api/products?search=${encodeURIComponent(search)}`)
        .then(res => res.json())
        .then(data => {
            const tbody = document.getElementById('inventoryTableBody');
            if (!tbody) return;
            tbody.innerHTML = '';

            if (data.data.length === 0) {
                tbody.innerHTML = `<tr><td colspan="6" style="text-align: center; padding: 2rem;">${t('no_products_found')}</td></tr>`;
                return;
            }

            data.data.forEach(product => {
                const tr = document.createElement('tr');

                const total = parseInt(product.total_stock) || 0;
                const size = parseInt(product.items_per_unit) || 1;
                const unit = product.unit || 'Unit';

                let stockDisplay = `<strong>${total} PCS</strong>`;
                if (size > 1) {
                    const bigUnits = Math.floor(total / size);
                    const remainder = total % size;
                    stockDisplay += `<br><small class="text-muted">(${bigUnits} ${unit}, ${remainder} PCS)</small>`;
                }

                let categoryBadge = '';
                switch (product.category) {
                    case 'Obat Bebas':
                        categoryBadge = 'badge-success';
                        break;
                    case 'Obat Keras':
                    case 'Psikotropika':
                        categoryBadge = 'badge-danger';
                        break;
                    default:
                        categoryBadge = 'badge-secondary';
                }

                tr.innerHTML = `
                    <td>
                        <strong>${product.name}</strong>
                    </td>
                    <td><span class="badge ${categoryBadge}">${product.category || '-'}</span></td>
                    <td>${product.storage_location || '-'}</td>
                    <td>${stockDisplay}</td>
                    <td>${product.unit}</td>
                    <td style="text-align: right;">
                        <button class="btn btn-sm btn-outline-info" data-bs-toggle="modal" data-bs-target="#viewBatchesModal" onclick='openViewBatchesModal(${JSON.stringify(product).replace(/'/g, "&#39;")})'>
                            <i data-lucide="layers" style="width: 16px; height: 16px;"></i>
                            Detail
                        </button>
                    </td>
                `;
                tbody.appendChild(tr);
            });
            lucide.createIcons();
        });
}

function openViewBatchesModal(product) {
    document.getElementById('viewBatchesTitle').innerText = `${t('batches_for')} ${product.name}`;
    fetchBatches(product.id);
}

function fetchBatches(productId) {
    fetch(`/api/batches?product_id=${productId}`)
        .then(res => res.json())
        .then(result => {
            const tbody = document.getElementById('batchTableBody');
            if (!tbody) return;
            tbody.innerHTML = '';

            if (result.data.length === 0) {
                tbody.innerHTML = `<tr><td colspan="3" style="text-align: center; padding: 2rem;">${t('no_batches_found')}</td></tr>`;
                return;
            }

            result.data.forEach(batch => {
                const tr = document.createElement('tr');
                const isLow = batch.current_stock > 0 && batch.current_stock < 10;
                tr.innerHTML = `
                    <td><strong>${batch.batch_number}</strong></td>
                    <td>${batch.expiry_date}</td>
                    <td>
                        <span class="badge ${batch.current_stock <= 0 ? 'badge-danger' : (isLow ? 'badge-warning' : 'badge-success')}">
                            ${batch.current_stock}
                        </span>
                    </td>
                `;
                tbody.appendChild(tr);
            });
        });
}
