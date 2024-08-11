# A todo app to explore GO, HTMX and Tailwind

![demo](https://github.com/user-attachments/assets/b7878ba7-3ec4-45ea-8d48-1a2cc8728cc6)

### Some requirements

```cmd
npm install htmx.org@2.0.1
```
for hot reloading
```cmd
--installation
go install github.com/air-verse/air@latest
--running
air
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
- oob swaps need to be before the main swap if there are 2 things being swopped
- Cant get nested oob swaps to work :(
- oob swaps with rows are [wierd](https://htmx.org/attributes/hx-swap-oob/)
