





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

```cmd
npm install htmx.org@2.0.1
```