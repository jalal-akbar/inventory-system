/**
 * POS Logic - Apotek Desktop
 * Handles cart management, searching, shortcuts, and UI interactions.
 */

let cart = [];

// Initialize Lucide Icons
function initIcons() {
    if (window.lucide) {
        lucide.createIcons();
    }
}

// Format Number as IDR
function formatNumber(num) {
    return new Intl.NumberFormat('id-ID').format(num);
}

// Format Rupiah with Prefix
function formatRupiah(num) {
    return 'Rp ' + formatNumber(num);
}

// ---------------------------------------------------------
// Cart Management
// ---------------------------------------------------------

function addCardToCart(el) {
    const product = {
        id: parseInt(el.dataset.id),
        name: el.dataset.name,
        price: parseFloat(el.dataset.price),
        sku: el.dataset.sku,
        unit: el.dataset.unit || 'Pcs',
        itemsPerUnit: parseInt(el.dataset.itemsPerUnit) || 1
    };
    addToCart(product);
}

function addToCart(product) {
    const existing = cart.find(item => item.id === product.id);
    if (existing) {
        existing.qty += 1;
    } else {
        cart.push({
            id: product.id,
            name: product.name,
            price: product.price,
            sku: product.sku,
            unit: product.unit,
            itemsPerUnit: product.itemsPerUnit,
            qty: 1
        });
    }
    renderCart();
    calculateTotal();
}

function updateQty(id, delta) {
    const item = cart.find(item => item.id === id);
    if (!item) return;
    
    item.qty += delta;
    if (item.qty <= 0) {
        cart = cart.filter(i => i.id !== id);
    }
    renderCart();
    calculateTotal();
}

function clearCart() {
    if (cart.length === 0) return;
    Swal.fire({
        title: 'Clear cart?',
        text: 'All items will be removed',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#f97316',
        confirmButtonText: 'Yes, clear it'
    }).then((result) => {
        if (result.isConfirmed) {
            cart = [];
            renderCart();
            calculateTotal();
        }
    });
}

function renderCart() {
    const container = document.getElementById('cart-items');
    if (!container) return;

    if (cart.length === 0) {
        container.innerHTML = `
            <div class="text-center py-5 text-muted h-100 d-flex flex-column align-items-center justify-content-center">
                <i data-lucide="shopping-basket" class="mb-3 opacity-20" style="width: 64px; height: 64px;"></i>
                <p>Your cart is empty</p>
            </div>`;
        initIcons();
        document.getElementById('checkout-btn').disabled = true;
        return;
    }

    document.getElementById('checkout-btn').disabled = false;
    container.innerHTML = cart.map(item => `
        <div class="cart-item animate-fade-in">
            <div class="flex-grow-1">
                <div class="fw-bold text-dark small">${item.name}</div>
                <div class="smaller text-muted">${item.sku}</div>
                <div class="d-flex align-items-center gap-2 mt-1">
                    <div class="fw-bold text-primary small">${formatRupiah(item.price)}</div>
                    <div class="smaller text-muted">/ ${item.unit}</div>
                </div>
            </div>
            <div class="qty-controls">
                <button class="qty-btn" onclick="updateQty(${item.id}, -1)">
                    <i data-lucide="minus" style="width: 14px;"></i>
                </button>
                <div class="text-center" style="min-width: 30px">
                    <span class="small fw-bold px-1">${item.qty}</span>
                    <div class="smaller text-muted" style="font-size: 0.65rem">${item.unit}</div>
                </div>
                <button class="qty-btn" onclick="updateQty(${item.id}, 1)">
                    <i data-lucide="plus" style="width: 14px;"></i>
                </button>
            </div>
        </div>
    `).join('');
    initIcons();
}

function calculateTotal() {
    const subtotal = cart.reduce((sum, item) => sum + (item.price * item.qty), 0);
    const discountInput = document.getElementById('discount-input');
    const discount = parseFloat(discountInput ? discountInput.value : 0) || 0;
    const total = subtotal - discount;

    const subtotalDisplay = document.getElementById('subtotal-display');
    const totalDisplay = document.getElementById('total-display');

    if (subtotalDisplay) subtotalDisplay.innerText = formatRupiah(subtotal);
    if (totalDisplay) totalDisplay.innerText = formatRupiah(total < 0 ? 0 : total);
}

// ---------------------------------------------------------
// Search & Filter
// ---------------------------------------------------------

function filterCategory(el, cat) {
    document.querySelectorAll('.category-pill').forEach(p => p.classList.remove('active'));
    el.classList.add('active');
    
    const searchInput = document.getElementById('product-search');
    // Using HTMX ajax to update results
    if (window.htmx) {
        htmx.ajax('GET', `/pos/search?category=${encodeURIComponent(cat)}&q=${encodeURIComponent(searchInput ? searchInput.value : '')}`, {
            target: '#products-display'
        });
    }
}

// ---------------------------------------------------------
// Checkout Process
// ---------------------------------------------------------

async function processCheckout() {
    if (cart.length === 0) return;

    const customerName = document.getElementById('customer_name').value;
    const paymentMethodEl = document.querySelector('input[name="payment_method"]:checked');
    const paymentMethod = paymentMethodEl ? paymentMethodEl.value : 'cash';
    const discountInput = document.getElementById('discount-input');
    const discount = parseFloat(discountInput ? discountInput.value : 0) || 0;

    const data = {
        items: cart.map(i => ({ id: i.id, qty: i.qty })),
        payment_method: paymentMethod,
        customer_name: customerName,
        discount: discount
    };

    const checkoutBtn = document.getElementById('checkout-btn');
    checkoutBtn.disabled = true;
    const originalText = checkoutBtn.innerText;
    checkoutBtn.innerText = 'Processing...';

    try {
        const response = await fetch('/pos/checkout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            Swal.fire({
                title: 'Success!',
                text: result.message || 'Transaction completed',
                icon: 'success',
                confirmButtonColor: '#f97316'
            }).then(() => {
                cart = [];
                document.getElementById('customer_name').value = '';
                if (discountInput) discountInput.value = 0;
                renderCart();
                calculateTotal();
            });
        } else {
            throw new Error(result.error || 'Checkout failed');
        }
    } catch (error) {
        Swal.fire({
            title: 'Error',
            text: error.message,
            icon: 'error',
            confirmButtonColor: '#f97316'
        });
    } finally {
        checkoutBtn.disabled = false;
        checkoutBtn.innerText = originalText;
    }
}

// ---------------------------------------------------------
// UI Enhancements: Fullscreen, Clock, Shortcuts
// ---------------------------------------------------------

function toggleFullScreen() {
    const fsIcon = document.getElementById('fs-icon');
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen();
        if (fsIcon) fsIcon.setAttribute('data-lucide', 'minimize');
    } else {
        if (document.exitFullscreen) {
            document.exitFullscreen();
            if (fsIcon) fsIcon.setAttribute('data-lucide', 'maximize');
        }
    }
    initIcons();
}

function initClock() {
    const clockContainer = document.getElementById('clock-container');
    const clockTime = document.getElementById('global-live-clock');
    const clockDate = document.getElementById('global-live-date');

    if (clockContainer && clockTime && clockDate) {
        const timezone = clockContainer.getAttribute('data-timezone') || 'UTC';
        const lang = clockContainer.getAttribute('data-lang') || 'en';

        function updateClock() {
            const now = new Date();
            try {
                const timeFormatter = new Intl.DateTimeFormat(lang, {
                    hour: '2-digit', minute: '2-digit', second: '2-digit',
                    hour12: false, timeZone: timezone
                });
                const dateFormatter = new Intl.DateTimeFormat(lang, {
                    day: '2-digit', month: 'short', year: 'numeric', timeZone: timezone
                });
                clockTime.textContent = timeFormatter.format(now);
                clockDate.textContent = dateFormatter.format(now);
            } catch (e) {
                clockTime.textContent = now.toLocaleTimeString();
                clockDate.textContent = now.toLocaleDateString();
            }
        }
        updateClock();
        setInterval(updateClock, 1000);
    }
}

// Keyboard Shortcuts
function initShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Focus search with '/'
        if (e.key === '/' && document.activeElement.tagName !== 'INPUT') {
            e.preventDefault();
            const searchInput = document.getElementById('product-search');
            if (searchInput) searchInput.focus();
        }
        
        // Clear cart with 'Escape' if search not focused
        if (e.key === 'Escape' && document.activeElement.tagName !== 'INPUT') {
            clearCart();
        }
    });
}

// ---------------------------------------------------------
// Init
// ---------------------------------------------------------

document.addEventListener('DOMContentLoaded', () => {
    initIcons();
    initClock();
    initShortcuts();
    renderCart(); // Show empty state if needed
});
