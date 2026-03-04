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

// Format PCS to Unit string
function formatPcsToUnitJS(totalPcs, unitName, baseUnitName, itemsPerUnit) {
    if (!baseUnitName) baseUnitName = "Pcs";
    if (totalPcs <= 0) return `0 ${baseUnitName}`;
    if (itemsPerUnit <= 1 || unitName === baseUnitName) return `${totalPcs} ${baseUnitName}`;

    const units = Math.floor(totalPcs / itemsPerUnit);
    const remainingPcs = totalPcs % itemsPerUnit;

    if (units > 0 && remainingPcs > 0) {
        return `${units} ${unitName} ${remainingPcs} ${baseUnitName}`;
    } else if (units > 0) {
        return `${units} ${unitName}`;
    }
    return `${remainingPcs} ${baseUnitName}`;
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
        unit: el.dataset.unit || 'Box',
        baseUnit: el.dataset.baseUnit || 'Pcs',
        stock: parseInt(el.dataset.stock) || 0,
        itemsPerUnit: parseInt(el.dataset.itemsPerUnit) || 1
    };
    addToCart(product);
}

function addToCart(product) {
    if (product.stock <= 0) {
        Swal.fire({
            title: window.posI18n.outOfStock,
            text: window.posI18n.outOfStockDesc,
            icon: 'error',
            confirmButtonColor: '#f97316'
        });
        return;
    }

    const existing = cart.find(item => item.id === product.id);
    if (existing) {
        // Check if adding 1 more exceeds stock
        const currentTotalPcs = existing.selectedUnit === 'UNIT' ? (existing.inputQty * existing.itemsPerUnit) : existing.inputQty;
        const addAmount = existing.selectedUnit === 'UNIT' ? existing.itemsPerUnit : 1;

        if (currentTotalPcs + addAmount > product.stock) {
            Swal.fire({
                title: window.posI18n.stockLimit,
                text: window.posI18n.stockLimitDesc.replace('{name}', product.name).replace('{stock}', product.stock),
                icon: 'warning',
                confirmButtonColor: '#f97316'
            });
            return;
        }
        existing.inputQty += 1;
    } else {
        cart.push({
            id: product.id,
            name: product.name,
            price: product.price,
            sku: product.sku,
            unit: product.unit,
            baseUnit: product.baseUnit,
            itemsPerUnit: product.itemsPerUnit,
            selectedUnit: 'BASEUNIT',
            inputQty: 1,
            stock: product.stock
        });
    }
    renderCart();
    calculateTotal();
}

function changeUnit(id, type) {
    const item = cart.find(item => item.id === id);
    if (!item) return;

    const oldUnit = item.selectedUnit;
    item.selectedUnit = type;

    // Validate if new unit exceeds stock with current inputQty
    if (!validateStockLimit(item)) {
        item.selectedUnit = oldUnit; // Revert if invalid
        return;
    }

    updateCartItemDOM(id);
    calculateTotal();
}

function updateInputQty(id, value) {
    const item = cart.find(item => item.id === id);
    if (!item) return;

    let newQty = parseFloat(value) || 0;
    if (newQty < 0) newQty = 0;

    const originalQty = item.inputQty;
    item.inputQty = newQty;

    if (!validateStockLimit(item)) {
        // validateStockLimit already caps and alerts, so just proceed to DOM update
    }

    // Surgical update for real-time feedback
    updateCartItemDOM(id);
    calculateTotal();
}

function updateQty(id, delta) {
    const item = cart.find(item => item.id === id);
    if (!item) return;

    const originalQty = item.inputQty;
    item.inputQty += delta;

    if (item.inputQty <= 0) {
        removeFromCart(id);
    } else {
        if (!validateStockLimit(item)) {
            item.inputQty = originalQty; // Revert if exceeds stock
            return;
        }
        updateCartItemDOM(id);
        calculateTotal();
    }
}

function removeFromCart(id) {
    cart = cart.filter(i => i.id !== id);
    renderCart();
    calculateTotal();
}

function clearCart() {
    if (cart.length === 0) return;
    Swal.fire({
        title: window.posI18n.clearCartTitle,
        text: window.posI18n.clearCartText,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#f97316',
        cancelButtonText: window.posI18n.cancel,
        confirmButtonText: window.posI18n.clearCartConfirm
    }).then((result) => {
        if (result.isConfirmed) {
            cart = [];
            renderCart();
            calculateTotal();
        }
    });
}

function validateStockLimit(item) {
    const totalPcs = item.selectedUnit === 'UNIT' ? (item.inputQty * item.itemsPerUnit) : item.inputQty;

    if (totalPcs > item.stock) {
        // Cap at maximum possible
        if (item.selectedUnit === 'UNIT') {
            item.inputQty = Math.floor(item.stock / item.itemsPerUnit);
        } else {
            item.inputQty = item.stock;
        }

        Swal.fire({
            title: window.posI18n.stockLimit,
            text: window.posI18n.stockLimitDesc.replace('{name}', item.name).replace('{stock}', item.stock),
            icon: 'warning',
            confirmButtonColor: '#f97316'
        });

        return false;
    }
    return true;
}

function renderCart() {
    const container = document.getElementById('cart-items');
    const checkoutBtn = document.getElementById('checkout-btn');
    if (!container) return;

    if (cart.length === 0) {
        container.innerHTML = `
            <div class="empty-cart animate-fade-in">
                <i data-lucide="shopping-basket" style="width: 80px; height: 80px;"></i>
                <h6 class="text-dark fw-bold mb-2">${window.posI18n.emptyCartTitle}</h6>
                <p class="small text-muted mb-0">${window.posI18n.emptyCartDesc}</p>
            </div>`;
        initIcons();
        if (checkoutBtn) {
            checkoutBtn.disabled = true;
            checkoutBtn.classList.remove('animate-pulse');
        }
        updateCartCount(0);
        return;
    }

    if (checkoutBtn) {
        checkoutBtn.disabled = false;
        checkoutBtn.classList.add('animate-pulse');
    }

    container.innerHTML = cart.map(item => {
        const displayPrice = item.selectedUnit === 'UNIT' ? (item.price * item.itemsPerUnit) : item.price;
        const lineTotal = displayPrice * item.inputQty;

        return `
        <div class="cart-item animate-fade-in" id="cart-item-${item.id}">
            <div class="cart-item-info">
                <div class="cart-item-title">${item.name}</div>
                <button class="btn-remove-item" onclick="removeFromCart(${item.id})" title="${window.posI18n.remove}">
                    <i data-lucide="x" style="width: 14px; height: 14px;"></i>
                </button>
            </div>
            <div class="cart-item-actions">
                <div class="qty-pill">
                    <button class="qty-btn" onclick="updateQty(${item.id}, -1)">
                        <i data-lucide="minus" style="width: 12px;"></i>
                    </button>
                    <input type="number" id="qty-input-${item.id}" class="cart-qty-input" value="${item.inputQty}" oninput="updateInputQty(${item.id}, this.value)">
                    <button class="qty-btn" onclick="updateQty(${item.id}, 1)">
                        <i data-lucide="plus" style="width: 12px;"></i>
                    </button>
                </div>
                <select class="unit-selector" onchange="changeUnit(${item.id}, this.value)">
                    <option value="BASEUNIT" ${item.selectedUnit === 'BASEUNIT' ? 'selected' : ''}>${item.baseUnit}</option>
                    ${item.unit && item.unit !== item.baseUnit ? `<option value="UNIT" ${item.selectedUnit === 'UNIT' ? 'selected' : ''}>${item.unit}</option>` : ''}
                </select>
                <div class="cart-item-price">
                    <div class="unit-price" id="unit-price-${item.id}">${formatRupiah(displayPrice)}</div>
                    <div class="line-total" id="line-total-${item.id}">${formatRupiah(lineTotal)}</div>
                </div>
            </div>
        </div>`;
    }).join('');
    initIcons();
    updateCartCount(cart.length);
}

function updateCartItemDOM(id) {
    const item = cart.find(item => item.id === id);
    if (!item) return;

    const qtyInput = document.getElementById(`qty-input-${id}`);
    const unitPriceEl = document.getElementById(`unit-price-${id}`);
    const lineTotalEl = document.getElementById(`line-total-${id}`);

    if (qtyInput && document.activeElement !== qtyInput) {
        qtyInput.value = item.inputQty;
    }

    const displayPrice = item.selectedUnit === 'UNIT' ? (item.price * item.itemsPerUnit) : item.price;
    const lineTotal = displayPrice * item.inputQty;

    if (unitPriceEl) unitPriceEl.innerText = formatRupiah(displayPrice);
    if (lineTotalEl) lineTotalEl.innerText = formatRupiah(lineTotal);
}

function updateCartCount(count) {
    const badge = document.getElementById('cart-count-badge');
    if (!badge) return;
    if (count > 0) {
        badge.innerText = count;
        badge.classList.remove('d-none');
    } else {
        badge.classList.add('d-none');
    }
}

function calculateTotal() {
    const subtotal = cart.reduce((sum, item) => {
        const displayPrice = item.selectedUnit === 'UNIT' ? (item.price * item.itemsPerUnit) : item.price;
        return sum + (displayPrice * item.inputQty);
    }, 0);
    const discountInput = document.getElementById('discount-input');
    let discount = parseFloat(discountInput ? discountInput.value : 0) || 0;

    // Sanitize and cap discount
    if (discount < 0) {
        discount = 0;
        if (discountInput) discountInput.value = 0;
    }
    if (discount > subtotal) {
        discount = subtotal;
        if (discountInput) discountInput.value = subtotal;
    }

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

    // Sync with hidden input for search inclusion
    const activeCatInput = document.getElementById('active-category');
    if (activeCatInput) activeCatInput.value = cat;

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

    const checkoutBtn = document.getElementById('checkout-btn');
    if (checkoutBtn.disabled) return; // Prevent multiple submissions

    const customerName = document.getElementById('customer_name').value;
    const paymentMethodEl = document.querySelector('input[name="payment_method"]:checked');
    const paymentMethod = paymentMethodEl ? paymentMethodEl.value : 'cash';
    const discountInput = document.getElementById('discount-input');
    const discount = parseFloat(discountInput ? discountInput.value : 0) || 0;

    const data = {
        items: cart.map(i => ({
            id: i.id,
            qty: i.selectedUnit === 'UNIT' ? (i.inputQty * i.itemsPerUnit) : i.inputQty,
            unit: i.selectedUnit === 'UNIT' ? i.unit : i.baseUnit
        })),
        payment_method: paymentMethod,
        customer_name: customerName,
        discount: discount
    };

    checkoutBtn.disabled = true;
    const originalHTML = checkoutBtn.innerHTML;
    // Replace only the text span, keep F8 shortcut hint if possible or just use innerHTML carefully
    checkoutBtn.innerHTML = `<span>${window.posI18n.processing}</span> <span class="kbd-hint-btn">F8</span>`;

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
            await Swal.fire({
                title: window.posI18n.successTitle,
                text: result.message || window.posI18n.transactionCompleted,
                icon: 'success',
                confirmButtonColor: '#f97316'
            });

            // Reduce stock in DOM to prevent page reload
            cart.forEach(item => {
                const qtyToDeduct = item.selectedUnit === 'UNIT' ? (item.inputQty * item.itemsPerUnit) : item.inputQty;
                const productCards = document.querySelectorAll(`.product-card[data-id="${item.id}"]`);

                productCards.forEach(card => {
                    let currentStock = parseInt(card.dataset.stock) || 0;
                    let newStock = currentStock - qtyToDeduct;
                    if (newStock < 0) newStock = 0;

                    card.dataset.stock = newStock;

                    const stockBadge = card.querySelector('.stock-badge');
                    if (stockBadge) {
                        const unitLabel = card.dataset.unit || item.unit;
                        const baseUnitLabel = card.dataset.baseUnit || item.baseUnit;
                        const itemsPerUnit = parseInt(card.dataset.itemsPerUnit) || 1;
                        stockBadge.innerText = formatPcsToUnitJS(newStock, unitLabel, baseUnitLabel, itemsPerUnit);

                        stockBadge.classList.remove('critical-stock', 'low-stock', 'in-stock');
                        if (newStock <= 0) stockBadge.classList.add('critical-stock');
                        else if (newStock < 5) stockBadge.classList.add('low-stock');
                        else stockBadge.classList.add('in-stock');
                    }

                    if (newStock <= 0) {
                        card.classList.add('out-of-stock');
                        card.removeAttribute('onclick');

                        if (!card.querySelector('.out-of-stock-overlay')) {
                            const overlay = document.createElement('div');
                            overlay.className = 'out-of-stock-overlay';
                            overlay.innerHTML = `<div class="out-of-stock-label">${window.posI18n.outOfStockLabel}</div>`;
                            card.insertBefore(overlay, card.firstChild);
                        }
                    }
                });
            });

            cart = [];
            document.getElementById('customer_name').value = '';
            if (discountInput) discountInput.value = 0;
            renderCart();
            calculateTotal();
        } else {
            throw new Error(result.error || 'Checkout failed');
        }
    } catch (error) {
        await Swal.fire({
            title: window.posI18n.errorTitle,
            text: error.message,
            icon: 'error',
            confirmButtonColor: '#f97316'
        });
    } finally {
        // Only re-enable if cart still has items (unlikely on success, but good for error scenarios)
        if (cart.length > 0) {
            checkoutBtn.disabled = false;
            checkoutBtn.innerHTML = originalHTML;
        } else {
            // successful checkout - keep disabled and restore generic text if needed or let renderCart handle it
            checkoutBtn.disabled = true;
            checkoutBtn.innerHTML = originalHTML; // Restore HTML but keep disabled
        }
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
        const timezone = clockContainer.getAttribute('data-timezone') ||
            Intl.DateTimeFormat().resolvedOptions().timeZone;
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

        // Checkout with 'F8'
        if (e.key === 'F8') {
            e.preventDefault();
            const checkoutBtn = document.getElementById('checkout-btn');
            if (checkoutBtn && !checkoutBtn.disabled) {
                processCheckout();
            }
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

    // Fix: Stop propagation on user dropdown clicks so it doesn't toggle off when clicking inside
    const userDropdown = document.getElementById('userDropdownContent');
    if (userDropdown) {
        userDropdown.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    }
});
