# Go Web Server
This repo is a template web server I use to rapidly create other projects ontop of.

# Disclaimer

Over the years I've found these packages, folder structures, and interface abstractions to intuitively click for me. This repo containers many of the things I find myself having to repeat every project but doesn't mean it's the best for you and your projects. There's almost certainly use cases that would not be compatible with the abstractions and packages I've chosen here. This is just something that works well for me and my uses cases, and maybe someone elses too :)

# Features
 - Multi-stage docker-ized build step
 - Docker compose with reverse proxy, rsync backups, database, prometheus, and grafana
 - Abstracted and 100% tested `routes/` folder to hold business logic
 - `/users` API
    - `PATCH` to edit a user
    - `DELETE` to remove a user
    - `GET` to query users (including fuzzy)
 - OAuth login
 - Database migrations
 - Sane defaults
    - Rate limiting
    - CSRF + XSS middlewares
    - Sane TCP connection timeouts
 - Prometheus metrics