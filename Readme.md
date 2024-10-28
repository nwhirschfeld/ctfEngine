# ctfEngine

_ctfEngine_ is a platform designed to host small Capture The Flag (CTF) competitions, making it ideal for onsite events,
classes, and smaller gatherings.

## State of Development

**ctfEngine** is currently in the early stages of development, resembling an alpha version.
It has been tested in events with approximately 10 players.

## Design Goals

**ctfEngine** is built with the following design goals:

- **Easy Deployment and Usage:**
    - Runs seamlessly in a containerized environment.
    - Eliminates dependencies on external Database Management Systems (DBMS).
    - All necessary resources are included in the binary.
    - Combines frontend and backend into a single service.

- **Versatility for Various CTF Projects:**
    - Supports different CTF projects.
    - CTFs are structured in a format that can be stored in a Git repository.
        - CTFs are organized in a folder structure.
        - Utilizes plaintext files where applicable.

## CTF Definition

To use ctfEngine, you need a CTF project in the correct format.
Specify the location of your CTF project while starting ctfEngine using the 
`-l` parameter or place your CTD in `/ctf` (like in the docker container).

To start the example project you could run 
```shell
# go run ctfEngine -l example_ctf

 ┌───────────────────────────────────────────────────┐ 
 │                     ctfEngine                     │ 
 │                   Fiber v2.51.0                   │ 
 │               http://127.0.0.1:3000               │ 
 │       (bound on host 0.0.0.0 and port 3000)       │ 
 │                                                   │ 
 │ Handlers ............ 19  Processes ........... 1 │ 
 │ Prefork ....... Disabled  PID ............. 19503 │ 
 └───────────────────────────────────────────────────┘ 

```

### CTF Structure

The CTF structure consists mainly of `.yml` files and folders.

~~~
├── challenges
│   ├── <challenge id>
│   ├── ...
│   └── <challenge id>
└── ctf.yml
~~~

### Challenge Structure

~~~
<challenge id>
├── files
│   ├── <challenge file>
│   ├── ...
│   └── <challenge file>
├── Dockerfile (optional)
└── challenge.yml
~~~

#### challenge.yml

The `challenge.yml` files should be formatted like the following example:

```yaml
name: Example Challenge
description: >
  This is an example challenge that requires participants to...
value: 100
flag: CTF{example_flag}
category: Web Exploitation
hints:
  - description: "Consider looking into..."
    cost: 10
  - description: "Another hint suggestion..."
    cost: 20
service:
  port: 1337
```

### SignUp Tokens
It is possible to enable a feature, that requires users to provide a single 
use __token__ to sign up.
To enable this feature, _ctf.yml_ needs to be extended by the following line:
```yaml
registrationToken: True
```

Further, you need to add tokens to the database.
```shell
# go run ctfEngine -a "f47ac10b-58cc-4372-a567-0e02b2c3d479" -l example_ctf
added signup token to DB
```

# Contribution

You are welcome to contribute to ctfEngine and enhance its capabilities.

Feel free to:

- Report issues
- Suggest new features
- Submit pull requests

Your contributions help make ctfEngine better for everyone. Thank you for your support!

# Copyright

    ctfEngine is a platform designed to host CTF competitions
    Copyright (C) 2024 Niclas Hirschfeld

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
