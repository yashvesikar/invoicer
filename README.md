# Invoice Manager

A command-line invoice management tool built with Go and Bubble Tea TUI framework.

## Features

- **Client Management**
  - Add, edit, and delete clients
  - Store client name, address, and multiple email addresses
  - Set default hourly rate per client

- **Invoice Management**
  - Create and manage invoices
  - Automatic invoice numbering (YYYY-##)
  - Line items with quantities and prices
  - Automatic calculation of subtotals, discounts, and taxes
  - Invoice status tracking (draft, sent, paid, overdue)
  - Export invoices to PDF using LaTeX

- **User Interface**
  - Interactive terminal UI using Bubble Tea
  - Keyboard navigation
  - Real-time calculations and updates

## Requirements

- Go 1.18 or higher
- `pdflatex` command available in PATH (for PDF export)
  - On macOS: `brew install --cask mactex` or `brew install basictex`
  - On Ubuntu/Debian: `sudo apt-get install texlive-latex-base`
  - On Windows: Install MiKTeX or TeX Live

## Installation

```bash
go build -o invoicer
```

## Usage

Run the application:

```bash
./invoicer
```

### Navigation

**Main Menu:**
- `↑/k` or `↓/j` - Navigate menu items
- `Enter` - Select menu item
- `q` - Quit application

**Client Management:**
- `a` - Add new client
- `e` - Edit selected client
- `d` - Delete selected client
- `Esc` - Return to main menu

**Client Form:**
- Multiple email addresses can be managed
- Press Enter on "Manage Emails" to add/remove emails
- `+` - Add new email
- `-` - Remove current email
- Default hourly rate can be set per client

**Invoice Management:**
- `a` - Create new invoice
- `e` - Edit selected invoice
- `v` - View invoice details
- `d` - Delete selected invoice
- `Esc` - Return to main menu

**Invoice Details:**
- `p` - Export invoice to PDF
- `e` - Edit invoice
- `Esc` - Return to invoice list

**Forms:**
- `Tab` or `Shift+Tab` - Navigate between fields
- `Enter` - Submit form or select button
- `Esc` - Cancel and return

### Invoice Creation Workflow

1. Select "Manage Invoices" from main menu
2. Press `a` to create a new invoice
3. Select a client (clients must be created first)
4. Set discount percentage, tax percentage, and payment terms
5. Add line items:
   - Enter description, quantity, and unit price
   - Press Enter to add each item
6. Save the invoice

## Data Storage

All data is stored locally in JSON files:
- `./data/clients.json` - Client information
- `./data/invoices.json` - Invoice data

## PDF Export

Exported PDFs are saved to:
- `./exports/invoice_YYYY-##.pdf`

The PDF generation requires `pdflatex` to be installed on your system.

## Invoice Numbering

Invoices are automatically numbered using the format `YYYY-##`, where:
- `YYYY` is the current year
- `##` is a sequential number starting from 01

Examples: 2025-01, 2025-02, 2025-03

## Development

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Shopspring Decimal](https://github.com/shopspring/decimal) - Precise financial calculations