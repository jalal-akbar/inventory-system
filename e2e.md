# 📋 TESTING MANUAL — Inventory System (E2E)

> **Version:** 1.0  
> **Date:** February 25, 2026  
> **Purpose:** Manual end-to-end testing guide through the browser to ensure every core feature works as expected.  
> **Base URL:** `http://localhost:7070`

---

## 📌 Prerequisites

| #   | Item           | Detail                                                        |
| --- | -------------- | ------------------------------------------------------------- |
| 1   | Server running | Run `make run` — server active on port `7070`                 |
| 2   | Fresh database | Optional: `make reset-db` then `make seed` for clean data     |
| 3   | Admin account  | Username: `admin` / Password: `admin123`                      |
| 4   | Staff account  | Created manually by Admin on **Settings → Manage Users** page |
| 5   | Browser        | Chrome / Firefox latest version                               |

---

## 🗂️ Features Under Test

1. [Authentication (Login & Logout)](#1-authentication-login--logout)
2. [User Management (Admin)](#2-user-management-admin)
3. [New Product Input](#3-new-product-input)
4. [Approval Workflow (Admin)](#4-approval-workflow-admin)
5. [POS / Sales Transaction](#5-pos--sales-transaction)
6. [Stock & Product Status Validation](#6-stock--product-status-validation)
7. [Void Transaction](#7-void-transaction)

---

## 1. Authentication (Login & Logout)

### TC-01 — Admin Login (Happy Path)

|              |                                      |
| ------------ | ------------------------------------ |
| **ID**       | TC-01                                |
| **Priority** | 🔴 High                              |
| **Prereq**   | Admin account exists in the database |

| #   | Step                               | Expected Result                                                                      |
| --- | ---------------------------------- | ------------------------------------------------------------------------------------ |
| 1   | Open `http://localhost:7070/login` | Login page is displayed with username & password fields                              |
| 2   | Enter **Username:** `admin`        | Field is filled                                                                      |
| 3   | Enter **Password:** `admin123`     | Field is filled (characters hidden)                                                  |
| 4   | Click the **Login** button         | Redirect to `/dashboard`                                                             |
| 5   | Check sidebar / topbar             | Username `admin` is displayed; admin menus (Approvals, Settings, Backup) are visible |

**Status:** [ ]

---

### TC-02 — Staff Login (Happy Path)

|              |                                                    |
| ------------ | -------------------------------------------------- |
| **ID**       | TC-02                                              |
| **Priority** | 🔴 High                                            |
| **Prereq**   | Staff account already created by Admin (see TC-06) |

| #   | Step                                     | Expected Result                                                                |
| --- | ---------------------------------------- | ------------------------------------------------------------------------------ |
| 1   | Open `http://localhost:7070/login`       | Login page is displayed                                                        |
| 2   | Enter **Username:** `staff1`             | Field is filled                                                                |
| 3   | Enter **Password:** `(created password)` | Field is filled                                                                |
| 4   | Click the **Login** button               | Redirect to `/dashboard`                                                       |
| 5   | Check sidebar                            | Admin-only menus (**Approvals**, **Settings**, **Backup**) are **NOT** visible |

**Status:** [ ]

---

### TC-03 — Login Failed: Wrong Credentials

|              |           |
| ------------ | --------- |
| **ID**       | TC-03     |
| **Priority** | 🟡 Medium |
| **Prereq**   | —         |

| #   | Step                                 | Expected Result                                                          |
| --- | ------------------------------------ | ------------------------------------------------------------------------ |
| 1   | Open `http://localhost:7070/login`   | Login page is displayed                                                  |
| 2   | Enter **Username:** `admin`          | Field is filled                                                          |
| 3   | Enter **Password:** `wrong_password` | Field is filled                                                          |
| 4   | Click the **Login** button           | **Stays** on the login page, error message "Invalid credentials" appears |

**Status:** [ ]

---

### TC-04 — Login Failed: Account Disabled

|              |                                                                         |
| ------------ | ----------------------------------------------------------------------- |
| **ID**       | TC-04                                                                   |
| **Priority** | 🟡 Medium                                                               |
| **Prereq**   | Staff account has been disabled by Admin (via Settings → Toggle Status) |

| #   | Step                                       | Expected Result                                                       |
| --- | ------------------------------------------ | --------------------------------------------------------------------- |
| 1   | Open `http://localhost:7070/login`         | Login page is displayed                                               |
| 2   | Enter credentials for the disabled account | Fields are filled                                                     |
| 3   | Click the **Login** button                 | **Stays** on the login page, error message "Account disabled" appears |

**Status:** [ ]

---

### TC-05 — Logout

|              |                   |
| ------------ | ----------------- |
| **ID**       | TC-05             |
| **Priority** | 🔴 High           |
| **Prereq**   | User is logged in |

| #   | Step                                                     | Expected Result                                      |
| --- | -------------------------------------------------------- | ---------------------------------------------------- |
| 1   | Click the **Logout** button/link in the sidebar/topbar   | Redirect to `/login`                                 |
| 2   | Try accessing `http://localhost:7070/dashboard` directly | Redirect back to `/login` (session has been cleared) |

**Status:** [ ]

---

## 2. User Management (Admin)

### TC-06 — Admin Creates a New Staff Account

|              |                    |
| ------------ | ------------------ |
| **ID**       | TC-06              |
| **Priority** | 🔴 High            |
| **Prereq**   | Logged in as Admin |

| #   | Step                                                                  | Expected Result                                     |
| --- | --------------------------------------------------------------------- | --------------------------------------------------- |
| 1   | Navigate to **Settings** (`/settings`)                                | Settings page opens                                 |
| 2   | Click the **Manage Users** tab                                        | User management tab is active                       |
| 3   | Fill form: Username = `staff1`, Password = `staff123`, Role = `staff` | Fields are filled                                   |
| 4   | Click **Add User** / **Create User**                                  | Success message appears, new user shown in the list |
| 5   | Logout and login with `staff1` / `staff123`                           | Login successful, redirect to dashboard             |

**Status:** [ ]

---

### TC-07 — Admin Disables a Staff Account

|              |                                                           |
| ------------ | --------------------------------------------------------- |
| **ID**       | TC-07                                                     |
| **Priority** | 🟡 Medium                                                 |
| **Prereq**   | Logged in as Admin, staff account `staff1` already exists |

| #   | Step                                      | Expected Result                             |
| --- | ----------------------------------------- | ------------------------------------------- |
| 1   | Navigate to **Settings → Manage Users**   | User list is displayed                      |
| 2   | Click the status toggle for user `staff1` | Status changes to **inactive**              |
| 3   | Logout and try logging in with `staff1`   | Login fails with "Account disabled" message |

**Status:** [ ]

---

## 3. New Product Input

### TC-08 — Staff Inputs a New Product (Pending Verification)

|              |                    |
| ------------ | ------------------ |
| **ID**       | TC-08              |
| **Priority** | 🔴 High            |
| **Prereq**   | Logged in as Staff |

| #   | Step                                     | Expected Result                                                             |
| --- | ---------------------------------------- | --------------------------------------------------------------------------- |
| 1   | Navigate to **Inventory** (`/inventory`) | Product list page is displayed                                              |
| 2   | Click the **Add Product** button         | Product input modal/form opens                                              |
| 3   | Fill in product data:                    |                                                                             |
|     | - Name: `Paracetamol 500mg`              |                                                                             |
|     | - SKU: `PCT-500`                         |                                                                             |
|     | - Legal Category: `OTC`                  |                                                                             |
|     | - Therapeutic Class: `Analgesics`        |                                                                             |
|     | - Unit: `Box`                            |                                                                             |
|     | - Items per Unit: `10`                   |                                                                             |
|     | - Storage Location: `Rack A1`            |                                                                             |
|     | - Purchase Price: `15000`                |                                                                             |
|     | - Selling Price: `25000`                 |                                                                             |
|     | - Min Stock: `5`                         |                                                                             |
|     | - Batch Number: `B-2026-001`             |                                                                             |
|     | - Expiry Date: `2027-12-31`              |                                                                             |
|     | - Initial Stock: `50`                    |                                                                             |
| 4   | Click **Save** / **Submit**              | Success toast appears: "Product created successfully"                       |
| 5   | Check the product list                   | Product `Paracetamol 500mg` appears with a **pending/unverified** indicator |

**Status:** [ ]

---

### TC-09 — Admin Inputs a New Product (Auto-Verified)

|              |                    |
| ------------ | ------------------ |
| **ID**       | TC-09              |
| **Priority** | 🟡 Medium          |
| **Prereq**   | Logged in as Admin |

| #   | Step                                     | Expected Result                                                                                     |
| --- | ---------------------------------------- | --------------------------------------------------------------------------------------------------- |
| 1   | Navigate to **Inventory** (`/inventory`) | Product list page is displayed                                                                      |
| 2   | Click the **Add Product** button         | Product input modal/form opens                                                                      |
| 3   | Fill in product data:                    |                                                                                                     |
|     | - Name: `Amoxicillin 500mg`              |                                                                                                     |
|     | - SKU: `AMX-500`                         |                                                                                                     |
|     | - Legal Category: `Rx`                   |                                                                                                     |
|     | - Therapeutic Class: `Antibiotics`       |                                                                                                     |
|     | - Unit: `Strip`                          |                                                                                                     |
|     | - Items per Unit: `10`                   |                                                                                                     |
|     | - Purchase Price: `8000`                 |                                                                                                     |
|     | - Selling Price: `15000`                 |                                                                                                     |
|     | - Min Stock: `10`                        |                                                                                                     |
|     | - Batch Number: `B-2026-002`             |                                                                                                     |
|     | - Expiry Date: `2027-06-30`              |                                                                                                     |
|     | - Initial Stock: `100`                   |                                                                                                     |
| 4   | Click **Save**                           | Success toast appears                                                                               |
| 5   | Check the product list                   | Product `Amoxicillin 500mg` appears with **verified** status (auto-approved since Admin created it) |

**Status:** [ ]

---

### TC-10 — Product Input Failed: Empty Name

|              |                             |
| ------------ | --------------------------- |
| **ID**       | TC-10                       |
| **Priority** | 🟡 Medium                   |
| **Prereq**   | Logged in as Admin or Staff |

| #   | Step                                              | Expected Result                                              |
| --- | ------------------------------------------------- | ------------------------------------------------------------ |
| 1   | Open the **Add Product** form                     | Modal opens                                                  |
| 2   | Leave the **Name** field empty, fill other fields | —                                                            |
| 3   | Click **Save**                                    | Error/validation message appears, product is **not** created |

**Status:** [ ]

---

### TC-11 — Product Input Failed: Invalid Unit

|              |                             |
| ------------ | --------------------------- |
| **ID**       | TC-11                       |
| **Priority** | 🟢 Low                      |
| **Prereq**   | Logged in as Admin or Staff |

| #   | Step                                                                           | Expected Result                                                                                            |
| --- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| 1   | Open the **Add Product** form                                                  | Modal opens                                                                                                |
| 2   | Fill all fields correctly, but select/enter an invalid **Unit** (e.g., `Sack`) | —                                                                                                          |
| 3   | Click **Save**                                                                 | Error message: "Invalid unit. Must be one of: Box, Strip, Pcs, Vial, Botol, Tube, Sachet, Ampul, Pot, Dus" |

**Status:** [ ]

---

## 4. Approval Workflow (Admin)

### TC-12 — Admin Views Pending Approval List

|              |                                                               |
| ------------ | ------------------------------------------------------------- |
| **ID**       | TC-12                                                         |
| **Priority** | 🔴 High                                                       |
| **Prereq**   | Logged in as Admin, pending products from Staff exist (TC-08) |

| #   | Step                                                   | Expected Result                                                                                       |
| --- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------- |
| 1   | Check sidebar/topbar                                   | Pending items badge/counter is visible (count > 0)                                                    |
| 2   | Navigate to **Admin → Approvals** (`/admin/approvals`) | Approval dashboard page is displayed                                                                  |
| 3   | Check the pending items list                           | Product input by Staff (e.g., `Paracetamol 500mg`) appears in the list with its batch and stock entry |

**Status:** [ ]

---

### TC-13 — Admin Approves Products Individually

|              |                                                               |
| ------------ | ------------------------------------------------------------- |
| **ID**       | TC-13                                                         |
| **Priority** | 🔴 High                                                       |
| **Prereq**   | Logged in as Admin, pending items exist on the Approvals page |

| #   | Step                                                                | Expected Result                                             |
| --- | ------------------------------------------------------------------- | ----------------------------------------------------------- |
| 1   | Open the **Approvals** page (`/admin/approvals`)                    | Pending list is displayed                                   |
| 2   | Click the **Approve** button on a single item (product/batch/stock) | Item disappears from the pending list                       |
| 3   | Navigate to **Inventory** (`/inventory`)                            | The newly approved product appears with **verified** status |

**Status:** [ ]

---

### TC-14 — Admin Approves Group (Per Product)

|              |                                                                              |
| ------------ | ---------------------------------------------------------------------------- |
| **ID**       | TC-14                                                                        |
| **Priority** | 🟡 Medium                                                                    |
| **Prereq**   | Logged in as Admin, pending product with multiple batch/stock entries exists |

| #   | Step                                             | Expected Result                                                          |
| --- | ------------------------------------------------ | ------------------------------------------------------------------------ |
| 1   | Open the **Approvals** page (`/admin/approvals`) | Pending list is displayed                                                |
| 2   | Click the **Approve Group** button on a product  | All related items (product + batch + stock entries) are approved at once |
| 3   | Check the pending list again                     | The product and all related items are no longer shown                    |

**Status:** [ ]

---

### TC-15 — Admin Approve All

|              |                                                 |
| ------------ | ----------------------------------------------- |
| **ID**       | TC-15                                           |
| **Priority** | 🟡 Medium                                       |
| **Prereq**   | Logged in as Admin, several pending items exist |

| #   | Step                                             | Expected Result                                |
| --- | ------------------------------------------------ | ---------------------------------------------- |
| 1   | Open the **Approvals** page (`/admin/approvals`) | Pending list is displayed                      |
| 2   | Click the **Approve All** button                 | All items are approved at once                 |
| 3   | Check the page again                             | Pending list is empty, badge counter becomes 0 |

**Status:** [ ]

---

### TC-16 — Admin Rejects a Product

|              |                                         |
| ------------ | --------------------------------------- |
| **ID**       | TC-16                                   |
| **Priority** | 🟡 Medium                               |
| **Prereq**   | Logged in as Admin, pending item exists |

| #   | Step                                   | Expected Result                                                                                    |
| --- | -------------------------------------- | -------------------------------------------------------------------------------------------------- |
| 1   | Open the **Approvals** page            | Pending list is displayed                                                                          |
| 2   | Click the **Reject** button on an item | Item disappears from the pending list                                                              |
| 3   | Navigate to **Inventory**              | The rejected product does **not** appear as verified (or shows rejected status per implementation) |

**Status:** [ ]

---

### TC-17 — Staff Cannot Access the Approvals Page

|              |                    |
| ------------ | ------------------ |
| **ID**       | TC-17              |
| **Priority** | 🔴 High            |
| **Prereq**   | Logged in as Staff |

| #   | Step                                                               | Expected Result                                                           |
| --- | ------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| 1   | Directly access `http://localhost:7070/admin/approvals` in browser | Redirect or **Unauthorized / Forbidden** message (not the approvals page) |
| 2   | Check sidebar                                                      | **Approvals** menu is not visible for the Staff role                      |

**Status:** [ ]

---

## 5. POS / Sales Transaction

### TC-18 — Staff Sells an Approved Product (Happy Path)

|              |                                                                                            |
| ------------ | ------------------------------------------------------------------------------------------ |
| **ID**       | TC-18                                                                                      |
| **Priority** | 🔴 High                                                                                    |
| **Prereq**   | Logged in as Staff/Admin, product `Paracetamol 500mg` is **approved** with available stock |

| #   | Step                                                | Expected Result                                                          |
| --- | --------------------------------------------------- | ------------------------------------------------------------------------ |
| 1   | Navigate to **POS** (`/pos`)                        | Cashier page is displayed with the product list                          |
| 2   | Search for `Paracetamol 500mg` using the search bar | Product appears in search results                                        |
| 3   | Click/add the product to the cart                   | Product is added to the shopping cart                                    |
| 4   | Set **Quantity:** `5`                               | Quantity is updated, subtotal auto-calculated (5 × Rp25,000 = Rp125,000) |
| 5   | Enter **Payment Method:** `Cash`                    | Field is filled                                                          |
| 6   | (Optional) Enter **Customer Name**                  | Field is filled                                                          |
| 7   | Click the **Checkout** / **Pay** button             | Success response: "Transaction completed successfully" with `sale_id`    |
| 8   | Navigate to **Inventory**                           | Stock of `Paracetamol 500mg` has decreased by 5 units                    |

**Status:** [ ]

---

### TC-19 — Product Search in POS

|              |                                               |
| ------------ | --------------------------------------------- |
| **ID**       | TC-19                                         |
| **Priority** | 🟡 Medium                                     |
| **Prereq**   | Logged in, several products already available |

| #   | Step                                    | Expected Result                                               |
| --- | --------------------------------------- | ------------------------------------------------------------- |
| 1   | Open the **POS** page (`/pos`)          | Cashier page is displayed                                     |
| 2   | Type `para` in the search bar           | Products containing "para" appear (e.g., Paracetamol 500mg)   |
| 3   | Filter by category (e.g., `Analgesics`) | Only products with the Analgesics therapeutic class are shown |

**Status:** [ ]

---

### TC-20 — Checkout with Transfer Payment Method

|              |                                                  |
| ------------ | ------------------------------------------------ |
| **ID**       | TC-20                                            |
| **Priority** | 🟡 Medium                                        |
| **Prereq**   | Logged in, approved product with available stock |

| #   | Step                                          | Expected Result                                              |
| --- | --------------------------------------------- | ------------------------------------------------------------ |
| 1   | Add a product to the cart on the **POS** page | Product is added to the cart                                 |
| 2   | Set quantity                                  | Subtotal is updated                                          |
| 3   | Select **Payment Method:** `Transfer`         | Field is filled                                              |
| 4   | Click **Checkout**                            | Transaction successful, `payment_method` saved as `Transfer` |

**Status:** [ ]

---

### TC-21 — Checkout with QRIS Payment Method

|              |                                                  |
| ------------ | ------------------------------------------------ |
| **ID**       | TC-21                                            |
| **Priority** | 🟡 Medium                                        |
| **Prereq**   | Logged in, approved product with available stock |

| #   | Step                                          | Expected Result                                          |
| --- | --------------------------------------------- | -------------------------------------------------------- |
| 1   | Add a product to the cart on the **POS** page | Product is added to the cart                             |
| 2   | Select **Payment Method:** `QRIS`             | Field is filled                                          |
| 3   | Click **Checkout**                            | Transaction successful, `payment_method` saved as `QRIS` |

**Status:** [ ]

---

### TC-22 — Checkout with Discount

|              |                                                  |
| ------------ | ------------------------------------------------ |
| **ID**       | TC-22                                            |
| **Priority** | 🟡 Medium                                        |
| **Prereq**   | Logged in, approved product with available stock |

| #   | Step                                          | Expected Result                                                        |
| --- | --------------------------------------------- | ---------------------------------------------------------------------- |
| 1   | Add a product to the cart on the **POS** page | Product is added                                                       |
| 2   | Set quantity: `2`, selling price Rp25,000     | Subtotal: Rp50,000                                                     |
| 3   | Enter **Discount:** `5000`                    | Discount is entered                                                    |
| 4   | Click **Checkout**                            | Transaction successful, total = Rp45,000 (Rp50,000 - Rp5,000 discount) |

**Status:** [ ]

---

## 6. Stock & Product Status Validation

### TC-23 — Selling More Than Available Stock (Edge Case)

|              |                                                                  |
| ------------ | ---------------------------------------------------------------- |
| **ID**       | TC-23                                                            |
| **Priority** | 🔴 High                                                          |
| **Prereq**   | Logged in, approved product with limited stock (e.g., stock = 3) |

| #   | Step                                     | Expected Result                                                |
| --- | ---------------------------------------- | -------------------------------------------------------------- |
| 1   | Open the **POS** page                    | Cashier is displayed                                           |
| 2   | Add a product with stock = 3 to the cart | Product is added to the cart                                   |
| 3   | Set **Quantity:** `10` (exceeds stock)   | —                                                              |
| 4   | Click **Checkout**                       | Error message: insufficient stock. Transaction is **REJECTED** |
| 5   | Check the product list in Inventory      | Stock remains = 3 (unchanged)                                  |

**Status:** [ ]

---

### TC-24 — Selling an Unapproved Product (Edge Case)

|              |                                                                           |
| ------------ | ------------------------------------------------------------------------- |
| **ID**       | TC-24                                                                     |
| **Priority** | 🔴 High                                                                   |
| **Prereq**   | Logged in as Staff, a new product was input but not yet approved by Admin |

| #   | Step                                     | Expected Result                                                                         |
| --- | ---------------------------------------- | --------------------------------------------------------------------------------------- |
| 1   | Input a new product as Staff (see TC-08) | Product is created with **unverified** status                                           |
| 2   | Open the **POS** page (`/pos`)           | Cashier is displayed                                                                    |
| 3   | Search for the newly created product     | Product **does not appear** in POS search results (only verified products shown in POS) |

> **Note:** POS only displays products with batches where `is_verified = 1`, so unverified products are automatically excluded from sale.

**Status:** [ ]

---

### TC-25 — Selling a Product with Stock = 0

|              |                                                 |
| ------------ | ----------------------------------------------- |
| **ID**       | TC-25                                           |
| **Priority** | 🟡 Medium                                       |
| **Prereq**   | Logged in, approved product with depleted stock |

| #   | Step                                   | Expected Result                                                      |
| --- | -------------------------------------- | -------------------------------------------------------------------- |
| 1   | Open the **POS** page                  | Cashier is displayed                                                 |
| 2   | Search/select a product with stock = 0 | Product may not appear, or appears without the option to add to cart |
| 3   | Try adding to cart and checking out    | Error message appears / transaction is rejected                      |

**Status:** [ ]

---

## 7. Void Transaction

### TC-26 — Admin Voids a Transaction Directly

|              |                                                  |
| ------------ | ------------------------------------------------ |
| **ID**       | TC-26                                            |
| **Priority** | 🟡 Medium                                        |
| **Prereq**   | Logged in as Admin, an active transaction exists |

| #   | Step                                             | Expected Result                                                 |
| --- | ------------------------------------------------ | --------------------------------------------------------------- |
| 1   | Navigate to **Sales History** (`/sales-history`) | Sales history list is displayed                                 |
| 2   | Select the transaction to void                   | Transaction details are shown                                   |
| 3   | Click **Void** and enter a reason                | —                                                               |
| 4   | Confirm the void                                 | Transaction status changes to `void`, product stock is restored |

**Status:** [ ]

---

### TC-27 — Staff Requests Transaction Void

|              |                                                  |
| ------------ | ------------------------------------------------ |
| **ID**       | TC-27                                            |
| **Priority** | 🟡 Medium                                        |
| **Prereq**   | Logged in as Staff, an active transaction exists |

| #   | Step                              | Expected Result                                                          |
| --- | --------------------------------- | ------------------------------------------------------------------------ |
| 1   | Navigate to **Sales History**     | Sales history list is displayed                                          |
| 2   | Select the transaction to void    | Transaction details are shown                                            |
| 3   | Click **Void** and enter a reason | —                                                                        |
| 4   | Confirm                           | Transaction status changes to `pending_void`, **not voided immediately** |
| 5   | Log in as **Admin**               | —                                                                        |
| 6   | Open the **Approvals** page       | Staff's void request appears in the pending list                         |
| 7   | **Approve** the void request      | Transaction status changes to `void`, stock is restored                  |

**Status:** [ ]

---

## 📊 E2E Full Workflow Test

### TC-28 — Complete Flow: Input Product → Approve → Sell at Cashier

|              |                                                         |
| ------------ | ------------------------------------------------------- |
| **ID**       | TC-28                                                   |
| **Priority** | 🔴 Critical                                             |
| **Prereq**   | Clean database (optional: `make reset-db && make seed`) |

This is the most critical test case, verifying the entire flow from start to finish:

| #                                  | Step                                                                                                                                                                                                                                                            | Expected Result                                        |
| ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| **Phase 1: Setup**                 |                                                                                                                                                                                                                                                                 |                                                        |
| 1                                  | Log in as **Admin**                                                                                                                                                                                                                                             | Dashboard is displayed                                 |
| 2                                  | Create a new Staff account in **Settings → Manage Users** (username: `kasir1`, password: `kasir123`, role: `staff`)                                                                                                                                             | Account successfully created                           |
| 3                                  | Log out                                                                                                                                                                                                                                                         | Redirect to login                                      |
| **Phase 2: Product Input (Staff)** |                                                                                                                                                                                                                                                                 |                                                        |
| 4                                  | Log in as **Staff** (`kasir1` / `kasir123`)                                                                                                                                                                                                                     | Dashboard displayed without admin menus                |
| 5                                  | Open **Inventory** → **Add Product**                                                                                                                                                                                                                            | Input modal opens                                      |
| 6                                  | Fill in: Name=`Vitamin C 1000mg`, SKU=`VTC-1000`, Legal Category=`OTC`, Therapeutic Class=`Vitamins`, Unit=`Box`, Items per Unit=`12`, Purchase Price=`20000`, Selling Price=`35000`, Min Stock=`10`, Batch=`VTC-B001`, Expiry=`2028-01-31`, Initial Stock=`30` | —                                                      |
| 7                                  | Save                                                                                                                                                                                                                                                            | Product created, status **unverified/pending**         |
| 8                                  | Open **POS** → search `Vitamin C 1000mg`                                                                                                                                                                                                                        | Product **DOES NOT appear** in POS (not yet approved)  |
| 9                                  | Log out                                                                                                                                                                                                                                                         | Redirect to login                                      |
| **Phase 3: Approval (Admin)**      |                                                                                                                                                                                                                                                                 |                                                        |
| 10                                 | Log in as **Admin**                                                                                                                                                                                                                                             | Dashboard displayed, pending badge > 0                 |
| 11                                 | Open **Admin → Approvals**                                                                                                                                                                                                                                      | `Vitamin C 1000mg` appears in the pending list         |
| 12                                 | Click **Approve Group** or approve individually                                                                                                                                                                                                                 | All items (product + batch + stock entry) are approved |
| 13                                 | Log out                                                                                                                                                                                                                                                         | Redirect to login                                      |
| **Phase 4: Sales (Staff)**         |                                                                                                                                                                                                                                                                 |                                                        |
| 14                                 | Log in again as **Staff** (`kasir1`)                                                                                                                                                                                                                            | Dashboard is displayed                                 |
| 15                                 | Open **POS** → search `Vitamin C 1000mg`                                                                                                                                                                                                                        | Product **APPEARS** in POS (approved!) ✅              |
| 16                                 | Add to cart, qty = `3`                                                                                                                                                                                                                                          | Subtotal = 3 × Rp35,000 = Rp105,000                    |
| 17                                 | Select **Cash** payment, click **Checkout**                                                                                                                                                                                                                     | Transaction successful!                                |
| **Phase 5: Verification**          |                                                                                                                                                                                                                                                                 |                                                        |
| 18                                 | Open **Inventory** → search `Vitamin C 1000mg`                                                                                                                                                                                                                  | Remaining stock = **27** (30 - 3) ✅                   |
| 19                                 | Open **Sales History** (`/sales-history`)                                                                                                                                                                                                                       | Latest transaction is displayed with correct details   |

**Status:** [ ]

---

## 📝 Test Case Summary

| ID    | Description                          | Priority    | Status |
| ----- | ------------------------------------ | ----------- | ------ |
| TC-01 | Admin Login (Happy Path)             | 🔴 High     | [ ]    |
| TC-02 | Staff Login (Happy Path)             | 🔴 High     | [ ]    |
| TC-03 | Login Failed: Wrong Credentials      | 🟡 Medium   | [ ]    |
| TC-04 | Login Failed: Account Disabled       | 🟡 Medium   | [ ]    |
| TC-05 | Logout                               | 🔴 High     | [ ]    |
| TC-06 | Admin Creates New Staff Account      | 🔴 High     | [ ]    |
| TC-07 | Admin Disables Staff Account         | 🟡 Medium   | [ ]    |
| TC-08 | Staff Inputs New Product (Pending)   | 🔴 High     | [ ]    |
| TC-09 | Admin Inputs Product (Auto-Verified) | 🟡 Medium   | [ ]    |
| TC-10 | Product Input Failed: Empty Name     | 🟡 Medium   | [ ]    |
| TC-11 | Product Input Failed: Invalid Unit   | 🟢 Low      | [ ]    |
| TC-12 | Admin Views Pending List             | 🔴 High     | [ ]    |
| TC-13 | Admin Approves Product Individually  | 🔴 High     | [ ]    |
| TC-14 | Admin Approves Group                 | 🟡 Medium   | [ ]    |
| TC-15 | Admin Approves All                   | 🟡 Medium   | [ ]    |
| TC-16 | Admin Rejects Product                | 🟡 Medium   | [ ]    |
| TC-17 | Staff Cannot Access Approvals        | 🔴 High     | [ ]    |
| TC-18 | Staff Sells Approved Product         | 🔴 High     | [ ]    |
| TC-19 | Product Search in POS                | 🟡 Medium   | [ ]    |
| TC-20 | Checkout with Transfer               | 🟡 Medium   | [ ]    |
| TC-21 | Checkout with QRIS                   | 🟡 Medium   | [ ]    |
| TC-22 | Checkout with Discount               | 🟡 Medium   | [ ]    |
| TC-23 | Selling Over Available Stock         | 🔴 High     | [ ]    |
| TC-24 | Selling Unapproved Product           | 🔴 High     | [ ]    |
| TC-25 | Selling Product with Zero Stock      | 🟡 Medium   | [ ]    |
| TC-26 | Admin Voids Transaction              | 🟡 Medium   | [ ]    |
| TC-27 | Staff Requests Void                  | 🟡 Medium   | [ ]    |
| TC-28 | Full Workflow E2E                    | 🔴 Critical | [ ]    |

---

> **Total Test Cases:** 28  
> **Passed:** ** / 28  
> **Failed:** ** / 28  
> **Blocked:** \_\_ / 28
