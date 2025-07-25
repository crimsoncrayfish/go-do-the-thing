# A todo app to explore GO, HTMX and Tailwind

## APP Demo
LOGIN:

https://github.com/user-attachments/assets/a649ccc3-6481-4f9c-9339-7399afbbc51b

TASKS:

https://github.com/user-attachments/assets/83962fef-6d57-456a-8b15-bb3b1741efaf

https://github.com/user-attachments/assets/eabc63f6-8925-40fe-a873-24f382ead6eb

RESPONSIVE DESIGN:

https://github.com/user-attachments/assets/019db6ad-4cd1-4796-8665-d72cd45a4e12

https://github.com/user-attachments/assets/6a16eeb7-5d24-4a81-9319-715e0ea25b3c


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
- [x] improve task editing security
- [x] Admin controls to enable users
- [x] Containerize this so i can dev on Windows
- [x] Projects for task grouping
- [ ] Tags on projects and tasks
- [ ] Time logging on tasks
- [ ] Export to csv
- [ ] Dashboard with stats
