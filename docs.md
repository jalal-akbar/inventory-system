## NT-1: System Core & Safety Foundation

**Scope:** Authentication, Role-Based Access Control (RBAC), Global Settings, and immutable Activity Logging.

### User Stories:

- **As a User**, I want to log in securely so that I can access features based on my role.
- **As an Admin**, I want to manage global business settings (Name, Address, Timezone) and view system activity logs.

### Technical Details:

- **Database Tables:**
  - `users`: `id`, `username`, `password` (BCrypt), `role` (Admin/Staff), `status` (Active/Inactive), `language`.
  - `settings`: `business_name`, `address`, `timezone`.
  - `activity_logs`: `id`, `user_id`, `action` (TEXT), `created_at`.
- **Constraint:** `activity_logs.user_id` uses `ON DELETE SET NULL` to preserve logs even if a user is deleted.
- **Key Logic:** Middleware-based session validation for every request. Admin role required for `/admin/*` routes.

### DoD (Definition of Done):

- **Backend (Go):** JWT-based authentication, structured logging middleware, and timezone-aware timestamps.
- **Frontend (HTMX):** Login page with error handling, Topbar with account menu, and dynamic breadcrumbs.

---

## NT-2: Inventory & Master Data

**Scope:** Management of Products (Drugs), Categories, Units, and Multi-layer Unit Conversion.

### User Stories:

- **As a Staff**, I want to register new products with SKU codes and unit details.
- **As an Admin**, I want to verify new products before they become available for sale.

### Technical Details:

- **Database Tables:**
  - `products`: `sku_code` (Unique), `name`, `category`, `unit` (e.g., Box, Bottle), `items_per_unit` (INT), `min_stock`, `status`.
- **Conversion Logic:**
  - `items_per_unit` defines the multiplier for small units (e.g., 1 Box = 100 Tablets).
  - Stock is tracked in the smallest unit (pcs) at the batch level.
- **Business Rule:** Products created by Staff remain in `is_verified = 0` status until Admin approval.

### DoD (Definition of Done):

- **Backend (Go):** CRUD for products/categories, SKU uniqueness validation, and verification toggle logic.
- **Frontend (HTMX):** Product list with search/filter, and a "Verification" dashboard for Admins.

---

## NT-3: Sales & Transaction Engine

**Scope:** Point of Sale (POS), profit calculation snapshots, Discount management, and Regulatory compliance.

### User Stories:

- **As a Staff**, I want to process sales via a POS interface, applying discounts and capturing customer data.
- **As a Regulator**, I want the system to enforce strict data capture (Doctor name, Prescription #) for Psychotropic drugs.

### Technical Details:

- **Database Tables:**
  - `sales`: `total_amount`, `profit`, `discount`, `payment_method` (Cash/Transfer/QRIS), `customer_name`, `doctor_name`, `prescription_number`, `status` (Active/Void).
  - `sale_items`: `sale_id`, `product_id`, `batch_id`, `quantity`, `price` (Snapshot), `subtotal`.
- **Key Logic:**
  - **Snapshot Pricing:** Selling price is captured from the batch at the moment of sale to prevent retroactive profit changes.
  - **Profit Calculation:** `total_profit = (Sum(item_selling - item_purchase)) - total_discount`.
- **Validation Rule:** If `product.category == 'Psikotropika'`, fields `customer_name`, `doctor_name`, and `prescription_number` are **Mandatory**.

### DoD (Definition of Done):

- **Backend (Go):** Atomic transaction for sales (update stock + create sale + create items), validation hooks for categories.
- **Frontend (HTMX):** Real-time cart calculation, search-as-you-type product picker, and mobile-ready POS.

---

## NT-4: Stock Management & FIFO

**Scope:** Batch-based stock tracking, First-In-First-Out (FIFO) consumption, and stock adjustment approvals.

### User Stories:

- **As a System**, I want to automatically deduct stock from the oldest expiring batch first.
- **As an Admin**, I want to approve or reject stock additions requested by staff.

### Technical Details:

- **Database Tables:**
  - `product_batches`: `product_id`, `batch_number`, `expiry_date`, `current_stock`, `purchase_price`, `selling_price`.
  - `stock_entries`: `product_id`, `batch_id`, `quantity`, `status` (Pending/Approved/Rejected).
- **FIFO Algorithm:**
  - Fetch batches for `product_id` where `current_stock > 0` and `is_verified = 1`.
  - Order by `expiry_date ASC` (Oldest expires first).
  - Loop through batches to fulfill requested quantity, updating `current_stock` incrementally.
- **Stock Restoration:** On Transaction Void, quantity is restored to the **exact batch** from which it was originally taken.

### DoD (Definition of Done):

- **Backend (Go):** FIFO logic implemented in Service layer, Stock adjustment history with approval workflow.
- **Frontend (HTMX):** Stock entry form with batch details, Expiry date highlights (Yellow < 6 mo, Red < 3 mo).

---

## NT-5: Reporting & Analytics

**Scope:** Financial reporting, stock movement audit trails, and expiry monitoring.

### User Stories:

- **As an Admin**, I want to generate Profit-Loss reports and view stock mutation history for a specific date range.

### Technical Details:

- **Key Logic:**
  - **Stock Mutation:** `stok_awal + masuk - keluar = stok_akhir`. Calculated dynamically based on `stock_entries` and `sale_items`.
  - **Expiry Monitoring:** Query `product_batches` where `expiry_date <= CURDATE() + interval` and `stock > 0`.
- **Reporting Engine:** Supports filtering by date range and product category.

### DoD (Definition of Done):

- **Backend (Go):** Reporting services for Financials, Psychotropic logs (legal requirement), and Stock Ledger.
- **Frontend (HTMX):** Data tables with print/export capabilities, and Dashboard Chart.js integration.

---

## NT-6: Admin & System Settings

**Scope:** Bulk approvals, User Management, and Disaster Recovery.

### User Stories:

- **As an Admin**, I want to perform a full database backup and manage user accounts and passwords.

### Technical Details:

- **Backup Logic:** `SHOW CREATE TABLE` and `SELECT * FROM table` to generate a .sql dump file served via download header.
- **Verification Workflow:** Admins can "Bulk Approve" all pending products and stock entries to expedite onboarding.

### DoD (Definition of Done):

- **Backend (Go):** Database utility package for backups (S3 or Local), User management CRUD.
- **Frontend (HTMX):** User management grid, Settings panel, and Backup download button.
