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

## CTF Format

### CTF Structure

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

Documentation for challenge.yml is currently unavailable. Refer to the provided examples for guidance.

Feel free to improve the documentation for challenge.yml based on the available examples. Your contributions are highly
appreciated!

