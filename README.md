# Google Docs & Sheets Backup Tool

Automated CLI tool to backup Google Docs, Sheets, and Drive files with OAuth/Service Account authentication and Git integration.

[![Auto Release](https://github.com/aQaJoooN/gdoc-gsheet-backup/actions/workflows/release.yml/badge.svg)](https://github.com/aQaJoooN/gdoc-gsheet-backup/actions/workflows/release.yml)
[![Test](https://github.com/aQaJoooN/gdoc-gsheet-backup/actions/workflows/test.yml/badge.svg)](https://github.com/aQaJoooN/gdoc-gsheet-backup/actions/workflows/test.yml)

## Features

- ✅ Backup Google Sheets (xlsx, csv, pdf, ods, html)
- ✅ Backup Google Docs (docx, pdf, txt, html, md)
- ✅ Download Google Drive files (any file type: PDFs, images, videos, etc.)
- ✅ OAuth or Service Account authentication
- ✅ Local or Git repository storage
- ✅ Auto-generated README.md with backup inventory
- ✅ Azure DevOps on-premise support with PAT authentication
- ✅ GitHub integration with Personal Access Token
- ✅ Multi-platform binaries (Linux, macOS, Windows)
- ✅ Automated CI/CD with GitHub Actions

## Quick Start

### Download Pre-built Binary

Download the latest release for your platform from the [Releases page](https://github.com/aQaJoooN/gdoc-gsheet-backup/releases).

**Linux:**
```bash
wget https://github.com/aQaJoooN/gdoc-gsheet-backup/releases/latest/download/gdoc-gsheet-backup-linux-amd64
chmod +x gdoc-gsheet-backup-linux-amd64
./gdoc-gsheet-backup-linux-amd64 config.yaml
```

**macOS:**
```bash
wget https://github.com/aQaJoooN/gdoc-gsheet-backup/releases/latest/download/gdoc-gsheet-backup-darwin-arm64
chmod +x gdoc-gsheet-backup-darwin-arm64
./gdoc-gsheet-backup-darwin-arm64 config.yaml
```

**Windows:**
Download `gdoc-gsheet-backup-windows-amd64.exe` and run:
```cmd
gdoc-gsheet-backup-windows-amd64.exe config.yaml
```

### Build from Source

1. **Install Go** (1.21 or later): https://go.dev/download/

2. **Clone and build:**
```bash
git clone https://github.com/aQaJoooN/gdoc-gsheet-backup.git
cd gdoc-gsheet-backup
go mod download
go build -o gdoc-gsheet-backup
```

## Setup

1. **Get Google OAuth credentials:**
   - Create a project in [Google Cloud Console](https://console.cloud.google.com/)
   - Enable Google Drive API
   - Create OAuth 2.0 credentials (Desktop app) or Service Account
   - Download `credentials.json` and place it in the project folder

2. **Create config:**
```bash
cp config.example.yaml config.yaml
```

3. **Edit config.yaml** with your document URLs and settings

4. **Run:**
```bash
./gdoc-gsheet-backup config.yaml
```

## Configuration

See `config.example.yaml` for a complete example.

```yaml
google_account:
  credentials_file: "credentials.json"
  token_file: "token.json"

google_sheets:
  - url: "https://docs.google.com/spreadsheets/d/YOUR_SHEET_ID/edit"
    export_format: "xlsx"
    name: "my-spreadsheet"

google_docs:
  - url: "https://docs.google.com/document/d/YOUR_DOC_ID/edit"
    export_format: "docx"
    name: "my-document"

google_drive_files:
  - url: "https://drive.google.com/file/d/YOUR_FILE_ID/view"
    name: "presentation.pdf"
  - url: "https://drive.google.com/file/d/ANOTHER_FILE_ID/view"
    name: "diagram.png"

backup:
  type: "git"  # or "local"
  local_path: "./backups"
  git:
    repository: "https://github.com/username/backup-repo.git"
    branch: "main"
    credentials:
      username: "your-username"
      token: "your-personal-access-token"
    commit_message: "Backup: {timestamp}"
```

## Export Formats

**Sheets:** xlsx, csv, pdf, ods, html  
**Docs:** docx, pdf, txt, html, md  
**Drive Files:** Any file type (original format preserved)

## Important Note About Markdown Export

⚠️ **Markdown Limitations**: Google Docs doesn't natively support Markdown export. The app converts HTML to Markdown, but:
- **Tables** may lose formatting (converted to plain text)
- **Complex formatting** may not convert perfectly

**For documents with tables**, we recommend:
- Use `docx` format (preserves all formatting)
- Use `html` format (preserves tables)

## Google Drive Files Download

Download any file type from Google Drive (PDFs, images, videos, presentations, etc.):

```yaml
google_drive_files:
  - url: "https://drive.google.com/file/d/YOUR_FILE_ID/view"
    name: "my-file.pdf"  # Custom filename
  
  - url: "https://drive.google.com/file/d/ANOTHER_FILE_ID/view"
    name: "image.png"
```

The tool will download files in their original format and save them with the specified name (or original filename if not specified).

## Git Integration

### GitHub
```yaml
backup:
  type: "git"
  git:
    repository: "https://github.com/username/backup-repo.git"
    branch: "main"
    credentials:
      username: "your-username"
      token: "ghp_your_personal_access_token"
```

### Azure DevOps (On-Premise)
```yaml
backup:
  type: "git"
  git:
    repository: "https://your-azure-devops.com/tfs/Project/_git/repo"
    branch: "main"
    credentials:
      username: "your-username"
      token: "your-pat-token"
```

The tool uses Base64-encoded PAT authentication for Azure DevOps compatibility.

## Auto-Generated README

When using git backup, the tool automatically generates a `README.md` in your backup repository with:
- List of all backed up Sheets, Docs, and Drive files
- Links to source Google Docs/Sheets/Drive
- File formats and sizes
- Backup timestamp
- Summary statistics

Example generated README structure:
```
# Google Docs & Sheets Backup

Last backup: 2026-02-24 10:30:00

## 📊 Google Sheets
| File Name | Format | Source URL |
|-----------|--------|------------|
| spreadsheet.xlsx | xlsx | [Open in Google Sheets](...) |

## 📝 Google Docs
| File Name | Format | Source URL |
|-----------|--------|------------|
| document.docx | docx | [Open in Google Docs](...) |

## 📁 Google Drive Files
| File Name | Source URL |
|-----------|------------|
| presentation.pdf | [Open in Google Drive](...) |

## 📋 Backup Summary
- Total Sheets: 1
- Total Docs: 1
- Total Drive Files: 1
- Total Files: 3
```

## Development

### Running Tests
```bash
go test ./...
```

### Building for Multiple Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o gdoc-gsheet-backup-linux-amd64

# macOS
GOOS=darwin GOARCH=arm64 go build -o gdoc-gsheet-backup-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o gdoc-gsheet-backup-windows-amd64.exe
```

## Releasing

The project uses GitHub Actions for automatic releases. To create a new release:

1. Commit your changes with conventional commit messages:
   - `feat:` for new features (minor version bump)
   - `fix:` for bug fixes (patch version bump)
   - `BREAKING CHANGE:` for breaking changes (major version bump)

2. Push to main branch:
```bash
git push origin main
```

3. GitHub Actions will automatically:
   - Create a new version tag
   - Build binaries for all platforms
   - Create a GitHub release with binaries and checksums

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
