/**
 * Dashboard Global Search Implementation
 */

document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('global-dashboard-search');
    const resultsContainer = document.getElementById('global-search-results');
    let searchTimeout = null;

    if (!searchInput || !resultsContainer) return;

    searchInput.addEventListener('input', (e) => {
        const query = e.target.value.trim();

        if (query.length < 2) {
            resultsContainer.innerHTML = '';
            resultsContainer.classList.add('d-none');
            return;
        }

        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            fetch(`/api/products/search?q=${encodeURIComponent(query)}`)
                .then(res => res.json())
                .then(data => {
                    renderResults(data, query);
                })
                .catch(err => console.error('Search failed:', err));
        }, 300);
    });

    function renderResults(products, query) {
        if (products.length === 0) {
            resultsContainer.innerHTML = `
                <div class="p-4 text-center">
                    <p class="text-muted mb-0">Produk tidak ditemukan.</p>
                    <a href="/products" class="btn btn-link btn-sm text-primary">Ke Inventori</a>
                </div>
            `;
            resultsContainer.classList.remove('d-none');
            return;
        }

        const html = products.map(product => {
            const stockIcon = product.total_stock > 10 ? 'check-circle-2' : 'alert-circle';
            const iconColor = product.total_stock > 10 ? 'text-success' : 'text-warning';

            return `
                <a href="/products?search=${encodeURIComponent(product.name)}" class="d-flex align-items-center p-3 text-decoration-none border-bottom hover-bg-light">
                    <div class="bg-light p-2 rounded-3 me-3">
                        <i data-lucide="package" style="width: 20px; height: 20px; color: #64748b;"></i>
                    </div>
                    <div class="flex-grow-1">
                        <div class="d-flex justify-content-between align-items-center mb-1">
                            <span class="fw-bold text-dark">${highlightMatch(product.name, query)}</span>
                            <span class="badge bg-light text-dark fw-normal border">${product.category}</span>
                        </div>
                        <div class="d-flex align-items-center gap-3 small">
                            <span class="d-flex align-items-center gap-1">
                                <i data-lucide="${stockIcon}" class="${iconColor}" style="width: 14px; height: 14px;"></i>
                                <span class="text-muted">Stok: <strong>${product.total_stock}</strong> ${product.unit}</span>
                            </span>
                            <span class="text-muted">|</span>
                            <span class="text-muted"><i data-lucide="map-pin" class="me-1" style="width: 14px; height: 14px;"></i> ${product.storage_location || 'Tanpa Lokasi'}</span>
                            <span class="text-muted">|</span>
                            <span class="fw-bold text-primary">${formatRupiah(product.selling_price)}</span>
                        </div>
                    </div>
                </a>
            `;
        }).join('');

        resultsContainer.innerHTML = html;
        resultsContainer.classList.remove('d-none');

        // Re-initialize Lucide icons for the results
        if (window.lucide) lucide.createIcons(resultsContainer);
    }

    function highlightMatch(text, query) {
        if (!query) return text;
        const regex = new RegExp(`(${query})`, 'gi');
        return text.replace(regex, '<mark class="p-0 bg-warning bg-opacity-25">$1</mark>');
    }

    function formatRupiah(amount) {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            maximumFractionDigits: 0
        }).format(amount);
    }

    // Close on click outside
    document.addEventListener('click', (e) => {
        if (!searchInput.contains(e.target) && !resultsContainer.contains(e.target)) {
            resultsContainer.classList.add('d-none');
        }
    });
});

// Add hover style
const style = document.createElement('style');
style.textContent = `
    .hover-bg-light:hover { background-color: #f8fafc; }
`;
document.head.appendChild(style);
