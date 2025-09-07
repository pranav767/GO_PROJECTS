# Markdown Notes API

A Go-based REST API for managing Markdown notes, with endpoints for uploading, listing, grammar checking, and rendering notes as HTML.

## Features

- Upload notes as Markdown text or file
- List all saved notes
- Render Markdown notes as HTML
- Grammar check for note content (using LanguageTool API)

## Project Structure

```
cmd/server/main.go           # Server entry point
internal/controllers/        # API controllers
internal/services/           # Business logic (note, markdown, grammar)
models/note.go               # Note model
routes/router.go             # API routes
notes/                       # Saved Markdown files
utils/helpers.go             # Utility functions (ID generation)
```

## Endpoints

### Upload Markdown as File

```bash
curl -F "file=@your_note.md" http://localhost:8080/upload
```

### Upload Markdown as JSON

```bash
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"content": "# My Note"}'
```

### List Notes

```bash
curl -X GET http://localhost:8080/notes
```

### Render Note as HTML

```bash
curl http://localhost:8080/notes/<note_id>/html
```
Or open in browser:
```
http://localhost:8080/notes/<note_id>/html
```


### Grammar Check

```bash
curl -X GET http://localhost:8080/notes/<note_id>/grammar
```

Example response:

```
{"grammar_errors":{"matches":[{"message":"Possible spelling mistake found.","offset":221,"length":2},{"message":"This sentence does not start with an uppercase letter.","offset":242,"length":4},{"message":"Two consecutive dots","offset":246,"length":2},{"message":"This sentence does not start with an uppercase letter.","offset":249,"length":4},{"message":"Two consecutive commas","offset":301,"length":2},{"message":"Possible typo: you repeated a whitespace","offset":303,"length":2},{"message":"Possible spelling mistake found.","offset":471,"length":13}]}}
```

## How to Run

1. Install dependencies:
   - Go modules (see `go.mod`)
   - For Markdown rendering: `github.com/gomarkdown/markdown`
   - For grammar: uses LanguageTool API (no local install needed)

2. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

3. Use the endpoints above with `curl` or your browser.
