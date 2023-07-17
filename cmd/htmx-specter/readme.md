# HTMX, Golang, SpecterCSS Boilerplate

## KR notes

- <https://github.com/akkcheung/go-jokes>
- move to postgres for db
- see need for cors
- use .env file

## Overview

Built using Golang, [htmx](https://htmx.org) and [Spectre.css Framework](https://picturepan2.github.io/spectre/) only.

General features:

- Sign In
- Sign Up for new user
- add your own joke(s)
- get a random joke

Application design:

- MVC design
- 0% Javascript means no build
- Just Golang, no web framework is used
- Render html using Golang html/template

Installation (local):

1. docker build -t go-jokes .
2. docker run --rm -p 5000:5000 go-jokes

Reference(s):

- [Interview with Senior JS Developer in 2022](https://www.youtube.com/watch?v=Uo3cL4nrGOk&t=169s)
- [The Hypermedia Drive Application(HDA) architecture](https://htmx.org/essays/hypermedia-driven-applications/)
- [Deploy goland webapp on heroku](https://dzone.com/articles/deploying-a-simple-golang-webapp-on-heroku)
