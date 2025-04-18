# A todo app to explore GO, HTMX and Tailwind

https://github.com/user-attachments/assets/0d0cfb7e-ada8-4a50-a080-3283770a717a

[DB MODEL](https://excalidraw.com/#json=tO3xfZeEypuPXJLOuvVhP,FDrih2vpGQ-GoU_99JWUVA)


## Some requirements

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

## Compilation issues on a Windows machine

If there are compilation issues one of these is likely to fix it
```cmd
$env:GOTMPDIR = "PATH TO TEMP DIR"
go env -w CGO_ENABLED=1
go env -w CC="zig cc"
```
Ensure ZIG is installed on the pc

## Generate Public/Private key example
```cmd
ssh-keygen -t rsa -b 4096

--For JWTs (i.e. in keys/)
openssl genpkey -algorithm RSA -out private.key -outform PEM -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in private.key -out public.key

--FOR HTTPS
openssl genpkey -algorithm RSA -out private_key.pem -outform PEM -pkeyopt rsa_keygen_bits:2048
openssl req -new -x509 -key private_key.pem -out certificate.pem -days 365
```

## HTMX is wierd
- Cant process <body></body> as an oob swap
- oob swaps need to be before the main swap if there are 2 things being swopped
- Cant get nested oob swaps to work :(
- oob swaps with rows are [wierd](https://htmx.org/attributes/hx-swap-oob/)

## Goals

- [x] Home page
- [x] Basic tasks (create, edit and list)
- [x] JWT auth from scratch
- [x] Registration
- [x] Login
- [x] Link tasks to users
- [x] [Templ](https://templ.guide/)
- [x] Logout
- [ ] improve task editing security
- [ ] Admin controls to enable users
- [x] Containerize this so i can dev on Windows
- [ ] Projects for task grouping
- [ ] Time logging on tasks
- [ ] Export to csv
- [ ] Dashboard with stats

## Known issues
- [ ] hover on table rows with the z axis translation causes rows to end up "behind" the table
