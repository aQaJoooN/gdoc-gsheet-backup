# Google Docs & Sheets Backup Tool

CLI tool to backup Google Docs and Sheets with OAuth authentication.

## Setup

1. **Get Google OAuth credentials:**
   - Place your `credentials.json` file in this folder
   - On first run, you'll authorize the app and `token.json` will be created

2. **Install dependencies:**
```bash
go mod download
```

3. **Create config:**
```bash
cp config.example.yaml config.yaml
```

4. **Edit config.yaml** with your document URLs

5. **Run:**
```bash
go run . config.yaml
```

## Export Formats

**Sheets:** xlsx, csv, pdf, ods, html  
**Docs:** md, docx, pdf, txt, html

## Important Note About Markdown Export

⚠️ **Markdown Limitations**: Google Docs doesn't natively support Markdown export. The app converts HTML to Markdown, but:
- **Tables** may lose formatting (converted to plain text)
- **Complex formatting** may not convert perfectly

**For documents with tables**, we recommend:
- Use `docx` format (preserves all formatting)
- Use `html` format (preserves tables)
- Or manually use Google Docs "Docs to Markdown" add-on

**For simple documents**, `md` format works great!

## Build

```bash
go build -o backup.exe
backup.exe config.yaml
```
