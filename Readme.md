# A todo app to explore GO, HTMX and Tailwind

![demo](https://github.com/user-attachments/assets/b7878ba7-3ec4-45ea-8d48-1a2cc8728cc6)

## Scripts for initializing the DB

CREATE TODO items
```SQLite
CREATE TABLE IF NOT EXISTS todo (
    [id] INTEGER PRIMARY KEY,
    [description] TEXT,
    [status] INTEGER DEFAULT 0,
    [assigned_to] TEXT,
    [due_date] TEXT,
    [created_by] TEXT,
    [create_date] TEXT,
    [is_deleted] INTEGER default 0
);
```

### Some requirements

```cmd
npm install htmx.org@2.0.1
```
for hot reloading
```cmd
--installation
go install github.com/cosmtrek/air@latest
--running

```

### Compilation issues on a Windows machine

If there are compilation issues one of these is likely to fix it
```cmd
$env:GOTMPDIR = "PATH TO TEMP DIR"
go env -w CGO_ENABLED=1
go env -w CC="zig cc"
```
Ensure ZIG is installed on the pc

### HTMX is wierd
- Cant process <body></body> as an oob swap
- oob swaps before main swap
- oob swaps with rows are [wierd](https://htmx.org/attributes/hx-swap-oob/)
