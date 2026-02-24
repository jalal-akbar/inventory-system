-- 1. Migrations Tracker
CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration TEXT NOT NULL UNIQUE,
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. Users Table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT CHECK(role IN ('admin', 'staff')) NOT NULL,
    status TEXT CHECK(status IN ('active', 'inactive')) NOT NULL DEFAULT 'active',
    language TEXT DEFAULT 'en',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 3. Settings Table (Dynamic Branding Content)
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    business_name TEXT DEFAULT 'Inventory System',
    address TEXT,
    timezone TEXT DEFAULT 'Asia/Makassar',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 4. Products Table (Renamed from drugs)
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sku_code TEXT UNIQUE DEFAULT NULL,
    name TEXT NOT NULL,
    category TEXT,
    unit TEXT,
    items_per_unit INTEGER DEFAULT 1,
    storage_location TEXT,
    purchase_price REAL DEFAULT 0,
    selling_price REAL DEFAULT 0,
    min_stock INTEGER DEFAULT 10,
    status TEXT CHECK(status IN ('active', 'inactive')) DEFAULT 'active',
    is_verified INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);

-- 5. Product Batches Table (Renamed from drug_batches + Price Freezing)
CREATE TABLE IF NOT EXISTS product_batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    batch_number TEXT NOT NULL,
    expiry_date TEXT NOT NULL,
    current_stock INTEGER DEFAULT 0,
    purchase_price REAL DEFAULT 0, -- Frozen Purchase Price
    selling_price REAL DEFAULT 0,  -- Frozen Selling Price
    is_verified INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
);

-- 6. Sales Table (Refactored for Void Approval System)
CREATE TABLE IF NOT EXISTS sales (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    total_amount REAL NOT NULL,
    profit REAL NOT NULL,
    discount REAL DEFAULT 0,
    payment_method TEXT CHECK(payment_method IN ('Cash', 'Transfer', 'QRIS')) NOT NULL,
    customer_name TEXT,
    doctor_name TEXT,
    prescription_number TEXT,
    status TEXT CHECK(status IN ('active', 'pending_void', 'void')) DEFAULT 'active',
    void_reason TEXT NULL,
    void_requested_by INTEGER NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (void_requested_by) REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status);

-- 7. Sale Items Table
CREATE TABLE IF NOT EXISTS sale_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sale_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    batch_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    price REAL NOT NULL, -- This stores the snapshot price at sale time
    subtotal REAL NOT NULL,
    FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE RESTRICT,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    FOREIGN KEY (batch_id) REFERENCES product_batches(id) ON DELETE RESTRICT
);

-- 8. Stock Entries Table (Unified Approval for Add Stock)
CREATE TABLE IF NOT EXISTS stock_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    batch_id INTEGER NULL,
    quantity INTEGER NOT NULL,
    status TEXT CHECK(status IN ('pending_add', 'approved', 'rejected')) DEFAULT 'pending_add',
    is_verified INTEGER DEFAULT 0,
    requested_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    FOREIGN KEY (requested_by) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_stock_entries_status ON stock_entries(status);

-- 9. Activity Logs Table
CREATE TABLE IF NOT EXISTS activity_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- --------------------------------------------------------
-- SEED INITIAL CONFIGURATION
-- --------------------------------------------------------
INSERT OR IGNORE INTO settings (id, business_name, timezone)
VALUES (1, 'Inventory System', 'Asia/Makassar');

INSERT OR IGNORE INTO users (id, username, password, role, status)
VALUES (1, 'admin', '$2a$10$vEwWavGovVMdN4VmUx5WMO1aYHfbMrzUx4naeajxkOUT0i5LWBrHm', 'admin', 'active');

-- 10. Returns Table
CREATE TABLE IF NOT EXISTS returns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sale_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    total_refund REAL NOT NULL,
    reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE RESTRICT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

-- 11. Return Items Table
CREATE TABLE IF NOT EXISTS return_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    return_id INTEGER NOT NULL,
    sale_item_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    refund_amount REAL NOT NULL,
    condition_status TEXT CHECK(condition_status IN ('good', 'damaged')) DEFAULT 'good',
    FOREIGN KEY (return_id) REFERENCES returns(id) ON DELETE CASCADE,
    FOREIGN KEY (sale_item_id) REFERENCES sale_items(id) ON DELETE RESTRICT
);

-- Triggers for updated_at
CREATE TRIGGER IF NOT EXISTS update_settings_timestamp 
AFTER UPDATE ON settings
FOR EACH ROW
BEGIN
    UPDATE settings SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS update_products_timestamp 
AFTER UPDATE ON products
FOR EACH ROW
BEGIN
    UPDATE products SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;