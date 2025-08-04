package export

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/shopspring/decimal"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/models"
)

// escapeLatex escapes special LaTeX characters in a string
func escapeLatex(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"$", "\\$",
		"%", "\\%",
		"&", "\\&",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	return replacer.Replace(s)
}

type PaymentMethod struct {
	Type    string
	Details string
}

type InvoiceTemplateData struct {
	Invoice      *models.Invoice
	FromName     string
	FromAddress  string
	FromEmail    string
	ClientAddress string
	ClientEmails []string
	InvoiceDate  string
	DueDate      string
	HasDiscount  bool
	HasTax       bool
	PaymentMethods []PaymentMethod
}

func ExportInvoiceToPDF(invoice *models.Invoice, client *models.Client, cfg *config.Config, exportPath, templatePath string) error {
	// Check if pdflatex is available
	_, err := exec.LookPath("pdflatex")
	if err != nil {
		return fmt.Errorf("pdflatex not found in PATH. Please install LaTeX (e.g., TeX Live, MiKTeX) to export PDFs")
	}
	
	// Ensure export directory exists
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}
	
	// Log where we're saving the file
	fmt.Printf("Saving PDF to directory: %s\n", exportPath)
	
	// Build payment methods from config
	var paymentMethods []PaymentMethod
	
	if cfg.ZelleAccount != "" {
		paymentMethods = append(paymentMethods, PaymentMethod{
			Type:    "Zelle",
			Details: escapeLatex(cfg.ZelleAccount),
		})
	}
	
	if cfg.VenmoAccount != "" {
		paymentMethods = append(paymentMethods, PaymentMethod{
			Type:    "Venmo",
			Details: escapeLatex(cfg.VenmoAccount),
		})
	}
	
	if cfg.BankName != "" && cfg.BankRouting != "" && cfg.BankAccount != "" {
		paymentMethods = append(paymentMethods, PaymentMethod{
			Type:    "Bank Wire",
			Details: escapeLatex(fmt.Sprintf("%s, Routing: %s, Account: %s", cfg.BankName, cfg.BankRouting, cfg.BankAccount)),
		})
	}

	// Prepare template data with escaped strings
	data := InvoiceTemplateData{
		Invoice:       invoice,
		FromName:      escapeLatex(cfg.CompanyName),
		FromAddress:   escapeLatex(cfg.CompanyAddress),
		FromEmail:     escapeLatex(cfg.CompanyEmail),
		ClientAddress: escapeLatex(client.Address),
		ClientEmails:  client.Emails, // Will handle escaping in template
		InvoiceDate:   invoice.Date.Format("January 2, 2006"),
		DueDate:       invoice.DueDate.Format("January 2, 2006"),
		HasDiscount:   invoice.DiscountRate.GreaterThan(decimal.Zero),
		HasTax:        invoice.TaxRate.GreaterThan(decimal.Zero),
		PaymentMethods: paymentMethods,
	}
	
	// Read template
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	
	// Create template with custom functions
	funcMap := template.FuncMap{
		"formatDecimal": func(d decimal.Decimal) string {
			return d.StringFixed(2)
		},
		"escapeLatex": escapeLatex,
	}
	
	tmpl, err := template.New("invoice").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	
	// Create temporary LaTeX file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("invoice_%s.tex", invoice.Number))
	if err := os.WriteFile(tempFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tempFile)
	
	// Run pdflatex with better error handling
	cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-output-directory", tempDir, tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Save the generated .tex file for debugging
		debugFile := filepath.Join(exportPath, fmt.Sprintf("debug_invoice_%s.tex", invoice.Number))
		debugErr := os.WriteFile(debugFile, buf.Bytes(), 0644)
		if debugErr == nil {
			fmt.Printf("DEBUG: Saved generated .tex file to: %s\n", debugFile)
		}
		
		// Parse LaTeX errors from output
		outputStr := string(output)
		if strings.Contains(outputStr, "! LaTeX Error:") {
			// Extract the specific LaTeX error
			lines := strings.Split(outputStr, "\n")
			for i, line := range lines {
				if strings.Contains(line, "! LaTeX Error:") && i+1 < len(lines) {
					return fmt.Errorf("LaTeX error: %s\nFull output saved to debug file", lines[i])
				}
			}
		}
		
		return fmt.Errorf("failed to run pdflatex: %w\nOutput excerpt: %.500s", err, outputStr)
	}
	
	// Move PDF to exports directory
	tempPDF := filepath.Join(tempDir, fmt.Sprintf("invoice_%s.pdf", invoice.Number))
	finalPDF := filepath.Join(exportPath, fmt.Sprintf("invoice_%s.pdf", invoice.Number))
	defer os.Remove(tempPDF)
	
	fmt.Printf("Moving PDF from: %s\n", tempPDF)
	fmt.Printf("Moving PDF to: %s\n", finalPDF)
	
	// Copy the file
	if err := copyFile(tempPDF, finalPDF); err != nil {
		return fmt.Errorf("failed to copy PDF: %w", err)
	}
	
	fmt.Printf("Successfully saved PDF to: %s\n", finalPDF)
	
	// Clean up auxiliary files
	auxFiles := []string{".aux", ".log", ".out"}
	for _, ext := range auxFiles {
		auxFile := filepath.Join(tempDir, fmt.Sprintf("invoice_%s%s", invoice.Number, ext))
		os.Remove(auxFile)
	}
	
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

func GetExportPath(invoice *models.Invoice, exportPath string) string {
	return filepath.Join(exportPath, fmt.Sprintf("invoice_%s.pdf", invoice.Number))
}