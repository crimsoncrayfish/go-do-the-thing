
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

### Compilation issues on a Windows machine

If there are compilation issues one of these is likely to fix it
```cmd
$env:GOTMPDIR = "PATH TO TEMP DIR"
go env -w CGO_ENABLED=1
go env -w CC="zig cc"
```
Ensure ZIG is installed on the pc