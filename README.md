# Pharmacy Inventory System (Rabella)

A modern, efficient inventory and Point of Sale (POS) system designed for pharmacies and retail stores. Built with Go, SQLite, and HTMX.

## 🚀 Features

- **RBAC (Role-Based Access Control)**: Secure login for Admin and Staff roles.
- **Inventory Management**: Track products, batches, and stock levels with ease.
- **FIFO/FEFO Stock Logic**: Automatically deducts stock from the oldest expiring batches first.
- **Point of Sale (POS)**: Fast and responsive checkout interface with HTMX.
- **Regulatory Compliance**: Mandatory data capture for psychotropic and restricted medications.
- **Reporting & Analytics**: Comprehensive profit-loss reports, stock movements, and expiry monitoring.
- **Approvals Workflow**: Admin verification for new products and stock entries.
- **Multi-language Support**: Indonesian (ID) and English (EN).

## 🛠 Tech Stack

- **Backend**: [Go](https://go.dev/) (1.25+)
- **Database**: [SQLite](https://sqlite.org/) (via `modernc.org/sqlite`)
- **Frontend**: [HTMX](https://htmx.org/), [Vanilla JS](https://developer.mozilla.org/en-US/docs/Web/JavaScript), [Vanilla CSS](https://developer.mozilla.org/en-US/docs/Web/CSS)
- **Session Management**: [Gorilla Sessions](https://github.com/gorilla/sessions)
- **Internationalization**: Custom i18n implementation in Go.

## 📦 Getting Started

### Prerequisites

- Go 1.25 or higher.
- `make` (optional, for using the Makefile).

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/your-username/inventory-system.git
   cd inventory-system
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   ```

3. **Setup environment variables**:
   Create a `.env` file from the example:

   ```bash
   cp .env.example .env
   ```

   Generate a secure `SESSION_KEY` and add it to your `.env` file.

4. **Initialize and run**:

   ```bash
   make run
   ```

   _The database schema will auto-initialize on the first run._

5. **Seed the database (Optional)**:
   ```bash
   make seed
   ```
   _Initial credentials: `admin` / `admin123` (or check `cmd/seed/main.go`)._

## 📂 Project Structure

- `cmd/`: Application entry points (`main`, `seed`, `inspect`).
- `internal/`: Core business logic.
  - `config/`: Database and application configuration.
  - `domain/`: Data models and interfaces.
  - `handler/`: HTTP request handlers.
  - `service/`: Business services (FIFO logic, sales processing).
  - `repository/`: Database access layer.
  - `middleware/`: Auth, CSRF, and session middlewares.
- `web/`: Web assets.
  - `templates/`: HTML templates (Go `html/template`).
  - `static/`: CSS, JS, and image assets.
- `database/`: SQL schemas and migration files.

## 📜 Makefile Commands

- `make run`: Build and start the server.
- `make build`: Compile the application binary.
- `make build-win`: Compile for Windows (.exe).
- `make seed`: Populate initial data.
- `make clean`: Remove binaries and temporary files.
- `make reset-db`: **Wipe the database** and start fresh.

## 🛡 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
