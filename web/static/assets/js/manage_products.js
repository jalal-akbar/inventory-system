document.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const filter = urlParams.get('filter');
    if (filter && filter !== '') {
        const searchInput = document.getElementById('productSearch');
        if (searchInput) searchInput.value = ''; // Clear search if filtering
    }
    fetchProducts();
    lucide.createIcons();
    initPriceMasking();

    // Close search results when clicking outside
    document.addEventListener('click', (e) => {
        const results = document.getElementById('search-results');
        const input = document.getElementById('name');
        if (results && !results.contains(e.target) && e.target !== input) {
            results.style.display = 'none';
        }
    });

    // Initialize Bulk Action Toolbar if it doesn't exist
    initBulkToolbar();
});

function initBulkToolbar() {
    if (document.getElementById('bulkActionToolbar')) return;
    const toolbar = document.createElement('div');
    toolbar.id = 'bulkActionToolbar';
    toolbar.className = 'bulk-action-toolbar shadow-lg rounded-4 p-3 bg-dark text-white d-none';
    toolbar.style.position = 'fixed';
    toolbar.style.bottom = '2rem';
    toolbar.style.left = '50%';
    toolbar.style.transform = 'translateX(-50%)';
    toolbar.style.zIndex = '1060';
    toolbar.style.minWidth = '300px';
    
    toolbar.innerHTML = `
        <div class="d-flex align-items-center justify-content-between gap-4">
            <div class="d-flex align-items-center gap-2">
                <span class="badge bg-primary" id="selectedCountText">0</span>
                <span class="small fw-bold">Items Selected</span>
            </div>
            <div class="d-flex gap-2">
                <button class="btn btn-sm btn-outline-light border-0 fw-bold" onclick="bulkToggleStatus('active')">
                    <i data-lucide="power" style="width: 16px; height: 16px;" class="me-1"></i> Activate
                </button>
                <button class="btn btn-sm btn-outline-warning border-0 fw-bold" onclick="bulkToggleStatus('inactive')">
                    <i data-lucide="power-off" style="width: 16px; height: 16px;" class="me-1"></i> Deactivate
                </button>
                <button class="btn btn-sm btn-outline-danger border-0 fw-bold" onclick="bulkDelete()">
                    <i data-lucide="trash-2" style="width: 16px; height: 16px;" class="me-1"></i> Delete
                </button>
                <div class="vr bg-white opacity-25"></div>
                <button class="btn btn-sm btn-link text-white text-decoration-none small" onclick="clearSelection()">Cancel</button>
            </div>
        </div>
    `;
    document.body.appendChild(toolbar);
    if (window.lucide) lucide.createIcons();
}

function toggleAllProducts(master) {
    const checkboxes = document.querySelectorAll('.product-checkbox');
    checkboxes.forEach(cb => cb.checked = master.checked);
    updateBulkActions();
}

function updateBulkActions() {
    const checkboxes = document.querySelectorAll('.product-checkbox:checked');
    const toolbar = document.getElementById('bulkActionToolbar');
    const countText = document.getElementById('selectedCountText');
    const masterCheckbox = document.getElementById('selectAllProducts');
    const allCheckboxes = document.querySelectorAll('.product-checkbox');

    if (checkboxes.length > 0) {
        toolbar.classList.remove('d-none');
        toolbar.classList.add('d-flex');
        countText.innerText = checkboxes.length;
    } else {
        toolbar.classList.add('d-none');
        toolbar.classList.remove('d-flex');
    }

    if (masterCheckbox) {
        masterCheckbox.checked = checkboxes.length === allCheckboxes.length && allCheckboxes.length > 0;
    }
}

function clearSelection() {
    const checkboxes = document.querySelectorAll('.product-checkbox');
    checkboxes.forEach(cb => cb.checked = false);
    const master = document.getElementById('selectAllProducts');
    if (master) master.checked = false;
    updateBulkActions();
}

function exportToExcel() {
    const urlParams = new URLSearchParams(window.location.search);
    window.location.href = `/api/products/export?` + urlParams.toString();
}

async function bulkToggleStatus(targetStatus) {
    const selectedIds = Array.from(document.querySelectorAll('.product-checkbox:checked')).map(cb => parseInt(cb.value));
    if (selectedIds.length === 0) return;

    if (!confirm(`Are you sure you want to ${targetStatus === 'active' ? 'activate' : 'deactivate'} ${selectedIds.length} products?`)) return;

    try {
        const response = await fetch('/api/products/bulk-status', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('input[name="csrf_token"]')?.value || ''
            },
            body: JSON.stringify({ ids: selectedIds, status: targetStatus })
        });
        const result = await response.json();
        if (result.status === 'success') {
            showToast(result.message, 'success');
            setTimeout(() => window.location.reload(), 1000);
        } else {
            showToast(result.error || 'Failed to update products', 'error');
        }
    } catch (err) {
        showToast('Error communicating with server', 'error');
    }
}

async function bulkDelete() {
    const selectedIds = Array.from(document.querySelectorAll('.product-checkbox:checked')).map(cb => parseInt(cb.value));
    if (selectedIds.length === 0) return;

    if (!confirm(`CRITICAL: Are you sure you want to DELETE ${selectedIds.length} products? This action cannot be undone.`)) return;

    try {
        const response = await fetch('/api/products/bulk-delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('input[name="csrf_token"]')?.value || ''
            },
            body: JSON.stringify({ ids: selectedIds })
        });
        const result = await response.json();
        if (result.status === 'success') {
            showToast(result.message, 'success');
            setTimeout(() => window.location.reload(), 1000);
        } else {
            showToast(result.error || 'Failed to delete products', 'error');
        }
    } catch (err) {
        showToast('Error communicating with server', 'error');
    }
}

const t = (key, replacements = {}) => {
    let text = (window.TRANSLATIONS && window.TRANSLATIONS[key]) ? window.TRANSLATIONS[key] : key;
    for (const [k, v] of Object.entries(replacements)) {
        text = text.replace(`{${k}}`, v);
    }
    return text;
};

function initPriceMasking() {
    const priceInputs = document.querySelectorAll('.price-mask');
    priceInputs.forEach(input => {
        input.addEventListener('input', function (e) {
            // Remove everything except digits
            let value = this.value.replace(/\D/g, '');

            // Format with commas
            this.value = formatWithCommas(value);
        });

        // Prevention for non-digit keyboard entry
        input.addEventListener('keypress', function (e) {
            if (!/\d/.test(e.key)) {
                e.preventDefault();
            }
        });
    });
}

function formatWithCommas(value) {
    if (!value) return "0";
    return parseInt(value).toLocaleString('en-US');
}

function unmaskPrice(value) {
    if (typeof value === 'string') {
        return value.replace(/,/g, '');
    }
    return value;
}

function fetchProducts() {
    const searchInput = document.getElementById('productSearch');
    const search = searchInput ? searchInput.value : '';
    const urlParams = new URLSearchParams(window.location.search);
    const filter = urlParams.get('filter') || '';

    fetch(`/api/products?search=${encodeURIComponent(search)}&filter=${filter}`)
        .then(res => res.json())
        .then(data => {
            const tbody = document.getElementById('productTableBody');
            if (!tbody) return;
            tbody.innerHTML = '';

            if (data.data.length === 0) {
                tbody.innerHTML = `<tr><td colspan="${USER_ROLE === 'admin' ? 8 : 7}" style="text-align: center; padding: 2rem;">${t('no_products_found')}</td></tr>`;
                return;
            }

            data.data.forEach(product => {
                const tr = document.createElement('tr');
                if (product.is_verified == 0) tr.classList.add('row-unverified');

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
                        <div class="d-flex align-items-center">
                            <span class="indicator-dot ${product.category === 'Obat Bebas' ? 'bg-success' : (product.category === 'Obat Keras' || product.category === 'Psikotropika' ? 'bg-danger' : 'bg-secondary')}" 
                                  style="width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-right: 8px;" 
                                  title="${product.category}"></span>
                            <div>
                                <strong>${product.name}</strong>
                                ${product.is_verified == 0 ? '<span class="badge badge-warning ms-1" style="font-size: 10px;">Pending</span>' : ''}
                            </div>
                        </div>
                    </td>
                    <td><span class="badge ${categoryBadge}">${product.category || '-'}</span></td>
                    <td>${product.unit}</td>
                    <td>${product.items_per_unit}</td>
                    <td>${stockDisplay}</td>
                    ${USER_ROLE === 'admin' ? `<td>${new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(product.selling_price)}</td>` : ''}
                    <td><span class="badge ${product.status === 'active' ? 'badge-success' : 'badge-danger'}">${product.status}</span></td>
                    <td style="text-align: right;">
                        <div class="d-flex gap-2 justify-content-end">
                            ${(product.is_verified == 0 && USER_ROLE === 'admin') ? `
                                <button class="btn btn-sm btn-success" onclick="verifyProduct(${product.id})" title="Verify Product">
                                    <i data-lucide="check-circle" style="width: 16px; height: 16px;"></i>
                                </button>
                            ` : ''}
                            <button class="btn btn-sm btn-outline-primary" data-bs-toggle="modal" data-bs-target="#batchModal" onclick='openBatchModal(${JSON.stringify(product).replace(/'/g, "&#39;")})' title="Add Stock">
                                <i data-lucide="plus-circle" style="width: 16px; height: 16px;"></i>
                            </button>
                            ${USER_ROLE === 'admin' ? `
                            <button class="btn btn-sm btn-outline-secondary" data-bs-toggle="modal" data-bs-target="#productModal" onclick='openEditProductModal(${JSON.stringify(product).replace(/'/g, "&#39;")})' title="Edit Product">
                                <i data-lucide="edit-3" style="width: 16px; height: 16px;"></i>
                            </button>
                            ` : ''}
                            <button class="btn btn-sm btn-outline-info" data-bs-toggle="modal" data-bs-target="#viewBatchesModal" onclick='openViewBatchesModal(${JSON.stringify(product).replace(/'/g, "&#39;")})' title="View Batches">
                                <i data-lucide="layers" style="width: 16px; height: 16px;"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-success" onclick="openPrintLabel(${product.id})" title="Print Label">
                                <i data-lucide="printer" style="width: 16px; height: 16px;"></i>
                            </button>
                        </div>
                    </td>
                `;
                tbody.appendChild(tr);
            });
            lucide.createIcons();
        });
}

function openPrintLabel(id) {
    window.open(`print-label?id=${id}`, '_blank', 'width=600,height=500');
}

function openAddProductModal() {
    document.getElementById('modalTitle').innerText = window.TRANSLATIONS ? window.TRANSLATIONS['add_new_product_label'] || 'Add New Product' : 'Add New Product'; // Need to ensure label exists or use t
    // Actually, modalTitle is already set in PHP for initial load, but JS might change it.
    // Let's use t()
    document.getElementById('modalTitle').innerText = window.TRANSLATIONS ? window.TRANSLATIONS['create_product'] || 'Add New Product' : 'Add New Product';
    document.getElementById('productId').value = '';
    document.getElementById('productForm').reset();
    document.getElementById('totalCalcDisplay').innerText = `${t('total_to_be_added')}: 0 pcs`;

    // Reset toggle fields
    const itemsPerUnitGroup = document.getElementById('itemsPerUnitGroup');
    if (itemsPerUnitGroup) itemsPerUnitGroup.style.display = 'none';
    document.getElementById('itemsPerUnitHelp').innerText = '';
    document.getElementById('minStockHelp').innerText = '';
    const batchCalcDisplay = document.getElementById('batchCalcDisplay');
    if (batchCalcDisplay) {
        batchCalcDisplay.classList.remove('badge-success');
        batchCalcDisplay.classList.add('badge-info');
    }

    // Clear search results
    selectedProductId = null;
    document.getElementById('search-results').style.display = 'none';
    enableProductFields(true);

    // Auto-focus first input
    setTimeout(() => {
        document.getElementById('name').focus();
    }, 500);
}

function handleUnitChange() {
    const unit = document.getElementById('unit').value;
    const itemsPerUnitGroup = document.getElementById('itemsPerUnitGroup');
    const itemsPerUnitInput = document.getElementById('items_per_unit');

    if (unit === 'Box' || unit === 'Strip') {
        itemsPerUnitGroup.style.display = 'block';
    } else {
        itemsPerUnitGroup.style.display = 'none';
        itemsPerUnitInput.value = 1;
    }
    updateItemsPerUnitHelp();
    calculateInitialTotal();
    updateMinStockHelp();
}

function updateItemsPerUnitHelp() {
    const unit = document.getElementById('unit').value;
    const itemsInput = document.getElementById('items_per_unit');
    const items = itemsInput ? itemsInput.value : 0;
    const help = document.getElementById('itemsPerUnitHelp');
    if (unit && unit !== 'Pcs') {
        help.innerText = `💡 1 ${unit} = ${items} PCS`;
    } else {
        help.innerText = '';
    }
}

function updateMinStockHelp() {
    const minStockInput = document.getElementById('min_stock');
    const minStock = parseInt(minStockInput ? minStockInput.value : 0) || 0;
    const itemsPerUnitInput = document.getElementById('items_per_unit');
    const itemsPerUnit = parseInt(itemsPerUnitInput ? itemsPerUnitInput.value : 1) || 1;
    const unit = document.getElementById('unit').value || 'Unit';
    const help = document.getElementById('minStockHelp');

    if (itemsPerUnit > 1) {
        const units = (minStock / itemsPerUnit).toFixed(1).replace(/\.0$/, '');
        help.innerText = t('min_stock_help', { units: units, unit: unit });
    } else {
        help.innerText = t('min_stock_help', { units: minStock, unit: 'PCS' });
    }
}

function calculateInitialTotal() {
    const qtyEl = document.getElementById('initial_stock');
    const itemsEl = document.getElementById('items_per_unit');
    const calcDisplay = document.getElementById('totalCalcDisplay');

    if (!qtyEl || !itemsEl || !calcDisplay) return;

    const quantity = parseFloat(qtyEl.value) || 0;
    const itemsPerUnit = parseFloat(itemsEl.value) || 1;
    const total = Math.round(quantity * itemsPerUnit);

    updateItemsPerUnitHelp();
    updateMinStockHelp();

    if (total > 0) {
        calcDisplay.innerText = t('warehouse_confirm', { total: total });
        calcDisplay.classList.remove('badge-info');
        calcDisplay.classList.add('badge-success');
    } else {
        calcDisplay.innerText = `${t('total_to_be_added')}: 0 pcs`;
        calcDisplay.classList.remove('badge-success');
        calcDisplay.classList.add('badge-info');
    }
}

function showToast(message, type = 'success') {
    let container = document.getElementById('toast-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toast-container';
        document.body.appendChild(container);
    }

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;

    const icon = type === 'success' ? 'check-circle' : 'alert-circle';

    toast.innerHTML = `
        <i data-lucide="${icon}" style="width: 20px; height: 20px;"></i>
        <span>${message}</span>
    `;

    container.appendChild(toast);
    if (window.lucide) lucide.createIcons();

    setTimeout(() => {
        toast.style.animation = 'toastFadeOut 0.3s ease-in forwards';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function openEditProductModal(product) {
    document.getElementById('modalTitle').innerText = t('edit_product');
    document.getElementById('productId').value = product.id;
    document.getElementById('name').value = product.name;
    document.getElementById('category').value = product.category;
    document.getElementById('unit').value = product.unit;
    document.getElementById('items_per_unit').value = product.items_per_unit;
    document.getElementById('storage_location').value = product.storage_location;
    document.getElementById('min_stock').value = product.min_stock;

    if (document.getElementById('purchase_price')) {
        document.getElementById('purchase_price').value = formatWithCommas(product.purchase_price);
        document.getElementById('selling_price').value = formatWithCommas(product.selling_price);
    }

    enableProductFields(true);
    handleUnitChange(); // Ensure toggle state is correct
    calculateInitialTotal();
}

function saveProduct(e) {
    e.preventDefault();
    setModalLoading('productModalLoading', 'productSubmitBtn', true);
    const id = document.getElementById('productId').value;
    const url = id ? '/api/products/update' : '/api/products/create';

    const data = {
        id: id,
        name: document.getElementById('name').value,
        category: document.getElementById('category').value,
        unit: document.getElementById('unit').value,
        items_per_unit: document.getElementById('items_per_unit').value,
        storage_location: document.getElementById('storage_location').value,
        min_stock: document.getElementById('min_stock').value,
        purchase_price: unmaskPrice(document.getElementById('purchase_price').value),
        selling_price: unmaskPrice(document.getElementById('selling_price').value),
        sku_code: document.getElementById('sku_code').value,
        batch_number: document.getElementById('initial_batch').value,
        expiry_date: document.getElementById('initial_expiry').value,
        initial_stock: document.getElementById('initial_stock').value,
        existing_product_id: selectedProductId
    };

    if (!data.sku_code || data.sku_code.trim() === '') {
        data.sku_code = null;
    }

    let finalUrl = url;
    let finalData = data;

    if (!id && selectedProductId) {
        finalUrl = '/api/batches/create';
        finalData = {
            product_id: selectedProductId,
            batch_number: data.batch_number,
            expiry_date: data.expiry_date,
            package_count: data.initial_stock,
            items_per_unit: data.items_per_unit
        };
    }

    fetch(finalUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(finalData)
    })
        .then(res => res.json())
        .then(result => {
            if (result.success) {
                showToast(result.message, 'success');
                closeModal('productModal');
                fetchProducts();
                if (window.updatePendingCount) updatePendingCount();
            } else {
                showToast(result.error || "Error: Please check your input!", 'error');
            }
        })
        .finally(() => {
            setModalLoading('productModalLoading', 'productSubmitBtn', false, id ? t('save_changes') : t('save_changes'));
        });
}

function setModalLoading(overlayId, btnId, isLoading, originalText) {
    const overlay = document.getElementById(overlayId);
    const btn = document.getElementById(btnId);

    if (isLoading) {
        if (overlay) overlay.classList.add('show');
        if (btn) {
            btn.disabled = true;
            btn.innerHTML = `<span class="spinner-border-sm"></span> ${t('js_saving')}`;
            btn.classList.add('muted');
        }
    } else {
        if (overlay) overlay.classList.remove('show');
        if (btn) {
            btn.disabled = false;
            btn.innerHTML = originalText;
            btn.classList.remove('muted');
        }
    }
}

function verifyProduct(id) {
    if (!confirm(t('verify_confirm'))) return;

    fetch('/api/products/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id })
    })
        .then(res => res.json())
        .then(result => {
            if (result.success) {
                showToast('Product verified successfully');
                fetchProducts();
                if (window.updatePendingCount) updatePendingCount();
            } else {
                showToast(result.error || 'Failed to verify product', 'error');
            }
        });
}

function deleteProduct(id, name) {
    Swal.fire({
        title: t('confirm_delete_title'),
        text: t('confirm_delete_text', { name: name }),
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#ef4444',
        confirmButtonText: t('yes_inactivate')
    }).then((result) => {
        if (result.isConfirmed) {
            fetch('/api/products/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: id })
            })
                .then(res => res.json())
                .then(result => {
                    if (result.success) {
                        Swal.fire(t('inactivated'), result.message, 'success');
                        fetchProducts();
                    } else {
                        Swal.fire(t('error'), result.error, 'error');
                    }
                });
        }
    });
}

function openBatchModal(product) {
    document.getElementById('batchProductId').value = product.id;
    document.getElementById('batchProductName').value = product.name;
    document.getElementById('batchForm').reset();

    const unit = product.unit.toLowerCase();
    let defaultItemsPerUnit = product.items_per_unit || 1;

    if (unit === 'vial' || unit === 'botol') {
        defaultItemsPerUnit = 1;
    }

    document.getElementById('batch_items_per_unit').value = defaultItemsPerUnit;

    // Reset the display to default state
    const calcDisplay = document.getElementById('batchCalcDisplay');
    if (calcDisplay) {
        calcDisplay.innerText = `${t('total_to_be_added')}: 0 pcs`;
        calcDisplay.classList.remove('badge-success');
        calcDisplay.classList.add('badge-info');
    }
}

function calculateTotalStock() {
    const qtyEl = document.getElementById('quantity');
    const itemsEl = document.getElementById('batch_items_per_unit');
    const calcDisplay = document.getElementById('batchCalcDisplay');

    if (!qtyEl || !itemsEl || !calcDisplay) return;

    const quantity = parseFloat(qtyEl.value) || 0;
    const itemsPerUnit = parseFloat(itemsEl.value) || 1;
    const total = Math.round(quantity * itemsPerUnit);

    if (total > 0) {
        calcDisplay.innerText = `Konfirmasi: Anda akan memasukkan ${total} pcs ke dalam gudang.`;
        calcDisplay.classList.remove('badge-info');
        calcDisplay.classList.add('badge-success');
    } else {
        calcDisplay.innerText = `${t('total_to_be_added')}: 0 pcs`;
        calcDisplay.classList.remove('badge-success');
        calcDisplay.classList.add('badge-info');
    }
}

let selectedProductId = null;
let searchTimeout = null;

let lastInputTime = 0;

function highlightMatch(text, query) {
    if (!query) return text;
    const regex = new RegExp(`(${query})`, 'gi');
    return text.replace(regex, '<span class="highlight">$1</span>');
}

function searchProducts(query) {
    const resultsContainer = document.getElementById('search-results');
    const now = Date.now();
    const isRapid = (now - lastInputTime) < 50;
    lastInputTime = now;

    if (!query || query.length < 2) {
        resultsContainer.innerHTML = '';
        resultsContainer.style.display = 'none';
        if (!query) {
            selectedProductId = null;
            enableProductFields(true);
        }
        return;
    }

    document.getElementById('name').onkeydown = function (e) {
        const results = resultsContainer.querySelectorAll('.search-result-item');

        if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
            if (results.length === 0) return;
            e.preventDefault();
            let activeIndex = -1;
            results.forEach((el, i) => { if (el.classList.contains('selected')) activeIndex = i; });

            if (e.key === 'ArrowDown') {
                if (activeIndex < results.length - 1) {
                    if (activeIndex >= 0) results[activeIndex].classList.remove('selected');
                    results[activeIndex + 1].classList.add('selected');
                    results[activeIndex + 1].scrollIntoView({ block: 'nearest' });
                }
            } else {
                if (activeIndex > 0) {
                    results[activeIndex].classList.remove('selected');
                    results[activeIndex - 1].classList.add('selected');
                    results[activeIndex - 1].scrollIntoView({ block: 'nearest' });
                }
            }
        } else if (e.key === 'Enter') {
            let activeIndex = -1;
            results.forEach((el, i) => { if (el.classList.contains('selected')) activeIndex = i; });
            if (activeIndex >= 0) {
                e.preventDefault();
                results[activeIndex].click();
            }
        }
    };

    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        fetch(`/api/products/search?q=${encodeURIComponent(query)}`)
            .then(res => res.json())
            .then(data => {
                if (isRapid && data.length === 1 && (data[0].sku_code === query || data[0].name === query)) {
                    selectProduct(data[0]);
                    return;
                }

                resultsContainer.innerHTML = '';
                if (data.length === 0) {
                    resultsContainer.innerHTML = `
                        <div class="search-result-item" onclick="selectedProductId=null; enableProductFields(true); document.getElementById('search-results').style.display='none';">
                            <span class="product-name text-primary">${t('js_register_new', { query: query })}</span>
                        </div>
                    `;
                    resultsContainer.style.display = 'block';
                    return;
                }

                data.forEach((product, index) => {
                    const div = document.createElement('div');
                    div.className = 'search-result-item' + (index === 0 ? ' selected' : '');
                    div.innerHTML = `
                        <span class="product-stock">${product.total_stock} <span class="stock-label">Pcs</span></span>
                        <span class="product-name">${highlightMatch(product.name, query)} 
                            ${product.sku_code ? `<small class="text-muted ml-2">[${highlightMatch(product.sku_code, query)}]</small>` : ''}
                        </span>
                        <span class="product-info">${product.category} | ${product.unit}</span>
                    `;
                    div.onclick = () => selectProduct(product);
                    resultsContainer.appendChild(div);
                });

                resultsContainer.style.display = 'block';
            });
    }, 300);
}

function selectProduct(product) {
    selectedProductId = product.id;
    document.getElementById('name').value = product.name;
    document.getElementById('sku_code').value = product.sku_code || '';
    document.getElementById('category').value = product.category || '';
    document.getElementById('unit').value = product.unit || '';
    document.getElementById('items_per_unit').value = product.items_per_unit || 1;

    const sellingPriceInput = document.getElementById('selling_price');
    if (sellingPriceInput) {
        sellingPriceInput.value = formatWithCommas(product.selling_price);
    }

    document.getElementById('search-results').style.display = 'none';
    enableProductFields(false);
    handleUnitChange();

    setTimeout(() => {
        const initialStockInput = document.getElementById('initial_stock');
        if (initialStockInput) {
            initialStockInput.focus();
            initialStockInput.select();
        }
    }, 100);
}

function enableProductFields(enabled) {
    const fields = ['sku_code', 'category', 'unit', 'items_per_unit', 'selling_price', 'purchase_price'];
    fields.forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            el.disabled = !enabled;
        }
    });
}

function generateBatchNumber() {
    const date = new Date();
    const dateStr = date.getFullYear() +
        String(date.getMonth() + 1).padStart(2, '0') +
        String(date.getDate()).padStart(2, '0');
    const rand = Math.random().toString(36).substring(2, 6).toUpperCase();
    document.getElementById('batch_number').value = `BCH-${dateStr}-${rand}`;
}

function saveBatch(e) {
    e.preventDefault();
    const data = {
        product_id: document.getElementById('batchProductId').value,
        batch_number: document.getElementById('batch_number').value,
        expiry_date: document.getElementById('expiry_date').value,
        package_count: document.getElementById('quantity').value,
        items_per_unit: document.getElementById('batch_items_per_unit').value
    };

    setModalLoading('batchModalLoading', 'batchSubmitBtn', true);
    fetch('/api/batches/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
        .then(res => res.json())
        .then(result => {
            if (result.success) {
                Swal.fire(t('success'), result.message, 'success');
                closeModal('batchModal');
                fetchProducts();
            } else {
                Swal.fire(t('error'), result.error, 'error');
            }
        })
        .finally(() => {
            setModalLoading('batchModalLoading', 'batchSubmitBtn', false, t('add_stock'));
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
                tbody.innerHTML = `<tr><td colspan="4" style="text-align: center; padding: 2rem;">${t('no_batches_found')}</td></tr>`;
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
                    <td style="text-align: right;">
                        ${USER_ROLE === 'admin' ? `
                        <button class="btn btn-sm btn-outline-danger" onclick='openAdjustmentModal(${JSON.stringify(batch).replace(/'/g, "&#39;")})' title="Adjust/Return">
                            <i data-lucide="minus-circle" style="width: 16px; height: 16px;"></i>
                        </button>
                        ` : ''}
                    </td>
                `;
                tbody.appendChild(tr);
            });
            lucide.createIcons();
        });
}

function openAdjustmentModal(batch) {
    if (batch.current_stock <= 0) {
        Swal.fire(t('error'), t('js_adj_zero_stock'), 'error');
        return;
    }
    document.getElementById('adjBatchId').value = batch.id;
    document.getElementById('adjBatchNumber').value = batch.batch_number;
    document.getElementById('adjCurrentStock').value = batch.current_stock;
    const adjQuantityInput = document.getElementById('adjQuantity');
    if (adjQuantityInput) adjQuantityInput.value = '';
    const adjNoteInput = document.getElementById('adjNote');
    if (adjNoteInput) adjNoteInput.value = '';

    const modal = new bootstrap.Modal(document.getElementById('adjustmentModal'));
    modal.show();
}

function saveAdjustment(e) {
    e.preventDefault();
    const data = {
        batch_id: document.getElementById('adjBatchId').value,
        quantity: document.getElementById('adjQuantity').value,
        reason: document.getElementById('adjReason').value,
        note: document.getElementById('adjNote').value
    };

    if (parseInt(data.quantity) > parseInt(document.getElementById('adjCurrentStock').value)) {
        Swal.fire(t('error'), t('js_adj_exceed_stock'), 'error');
        return;
    }

    fetch('/api/inventory/adjust', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
        .then(res => res.json())
        .then(result => {
            if (result.success) {
                Swal.fire(t('success'), result.message, 'success');
                closeModal('adjustmentModal');
                fetchProducts();
                closeModal('viewBatchesModal');
            } else {
                Swal.fire(t('error'), result.error, 'error');
            }
        });
}

function closeModal(id) {
    const el = document.getElementById(id);
    if (!el) return;
    const modal = bootstrap.Modal.getInstance(el);
    if (modal) modal.hide();
    else el.style.display = 'none';
}
