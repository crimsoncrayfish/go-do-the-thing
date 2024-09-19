# A todo app to explore GO, HTMX and Tailwind

![demo](https://github.com/user-attachments/assets/b7878ba7-3ec4-45ea-8d48-1a2cc8728cc6)

### Some requirements

for htmx
```cmd
npm install htmx.org@2.0.1
```
for hot reloading
```cmd
--installation
go install github.com/air-verse/air@latest
--running
alias air='~/{go directory here}/air'
air
```
for templ
```cmd
    go install github.com/a-h/templ/cmd/templ@latest
```

### Compilation issues on a Windows machine

If there are compilation issues one of these is likely to fix it
```cmd
$env:GOTMPDIR = "PATH TO TEMP DIR"
go env -w CGO_ENABLED=1
go env -w CC="zig cc"
```
Ensure ZIG is installed on the pc

### Generate Public/Private key example
```cmd
ssh-keygen -t rsa -b 4096
```

### HTMX is wierd
- Cant process <body></body> as an oob swap
- oob swaps need to be before the main swap if there are 2 things being swopped
- Cant get nested oob swaps to work :(
- oob swaps with rows are [wierd](https://htmx.org/attributes/hx-swap-oob/)

### Goals

- ✔️ Home page
- ✔️ Basic tasks (create, edit and list)
- ✔️ JWT auth from scratch
- ✔️ Registration
- ✔️ Login
- ✔️ Link tasks to users
- Admin controls to enable users
- [Templ](https://templ.guide/)
- Containerize this so i can dev on Windows
- Logout
- Projects for task grouping
- Time logging on tasks
- Export to csv
- Dashboard with stats
