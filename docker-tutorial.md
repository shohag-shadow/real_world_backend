# Docker From Zero: A Hands-On Tutorial

## About this tutorial

This tutorial takes you from "what is Docker?" to running a small multi-container application the way real teams do. It is written as a series of short lessons. Each one has the same shape: a concept explained in plain English, a hands-on section where you type real commands, a short recap, and a few questions to check yourself.

**What you will build.** A small Go web API, packaged as a Docker image, talking to a Postgres database, started with one command using Docker Compose, with the data kept safe in a volume.

**Prerequisites**

- You can use a terminal and run basic commands.
- You know roughly what a web server and a database are.
- You do not need to know Docker. You do not need Go installed either, because Docker will compile the code for you.

**How to use it**

- Type the commands yourself instead of pasting them. It sticks better.
- Do the lessons in order. Each one builds on the last.
- Commands are shown for bash or zsh. On Windows PowerShell they work the same unless a note says otherwise.
- Expect about six to eight hours in total.

## Table of contents

1. [Why Docker exists](#lesson-1-why-docker-exists)
2. [Installing Docker and checking it works](#lesson-2-installing-docker-and-checking-it-works)
3. [Core vocabulary](#lesson-3-core-vocabulary)
4. [Running your first containers](#lesson-4-running-your-first-containers)
5. [Containers vs virtual machines](#lesson-5-containers-vs-virtual-machines)
6. [Building your first image](#lesson-6-building-your-first-image)
7. [Layers and the build cache](#lesson-7-layers-and-the-build-cache)
8. [Ports](#lesson-8-ports)
9. [Environment variables](#lesson-9-environment-variables)
10. [Volumes and persistent data](#lesson-10-volumes-and-persistent-data)
11. [Networking between containers](#lesson-11-networking-between-containers)
12. [Docker Compose](#lesson-12-docker-compose)
13. [Smaller images: .dockerignore and multi-stage builds](#lesson-13-smaller-images-dockerignore-and-multi-stage-builds)
14. [Registries and tags](#lesson-14-registries-and-tags)
15. [Debugging and cleanup](#lesson-15-debugging-and-cleanup)
16. [Good practices and security basics](#lesson-16-good-practices-and-security-basics)
17. [Final project: Go API with Postgres](#lesson-17-final-project-go-api-with-postgres)

Appendices: [A. Cheat sheet](#appendix-a-cheat-sheet), [B. Troubleshooting](#appendix-b-troubleshooting), [C. Answer key](#appendix-c-answer-key), [D. Where to go next](#appendix-d-where-to-go-next)

---

## Lesson 1: Why Docker exists

### What you will learn

The problem Docker solves and what it does about it.

### Concept

You build an app on your laptop and it works. You send it to a teammate or deploy it to a server, and it breaks. They have a different language version, a missing library, a different operating system, or a setting you forgot you changed. This is the classic "it works on my machine" problem.

The cause is simple: your app does not run alone. It depends on a whole environment around it, and that environment is different on every machine.

Docker fixes this by packaging the app together with its environment, so it runs the same way everywhere.

The best real-life comparison is the shipping container. Before standard containers, cargo was loaded loose onto ships. Every port handled it differently, and loading was slow and error-prone. Once everything went into one standard box, any crane, ship, or truck could move it without caring what was inside. The box is the same on the outside no matter what it holds.

Docker does that for software. Your app goes into a standard box. Any machine running Docker can run that box without caring what is inside.

What this gives you in practice:

- **Consistency.** Development, testing, and production run the same package.
- **Easy onboarding.** A new teammate runs one command instead of following a ten-page setup guide.
- **Isolation.** Two apps that need different versions of the same tool can live on one machine without fighting.
- **Simple deployment.** You ship the box, not a list of instructions.

What Docker is not:

- It is not a virtual machine (Lesson 5 covers the difference).
- It does not fix bugs in your code. It only makes the environment predictable.

### Recap

Docker packages an app with everything it needs so it runs the same everywhere, like a standard shipping container.

---

## Lesson 2: Installing Docker and checking it works

### What you will learn

How to install Docker and what happens when you run your first command.

### Hands-on

**Windows and macOS.** Install Docker Desktop from the official Docker website. On Windows it uses the WSL 2 backend. If it complains about virtualization, enable virtualization in your BIOS or UEFI settings.

**Linux.** Install Docker Engine by following the official instructions for your distribution at docs.docker.com/engine/install. Then allow your user to run Docker without `sudo`:

```bash
sudo usermod -aG docker $USER
```

Log out and back in for this to take effect.

**Verify the install.**

```bash
docker version
docker run hello-world
```

If you see a friendly message beginning with "Hello from Docker!", everything works.

### What just happened

Docker has three moving parts:

```
docker CLI  ->  Docker daemon  ->  Registry (Docker Hub)
```

- The **CLI** is the `docker` command you type.
- The **daemon** is a background service that does the real work: building images, running containers.
- A **registry** is a server that stores images. Docker Hub is the default public one.

When you ran `hello-world`, the CLI asked the daemon to run that image. The daemon did not find it on your machine, so it downloaded it from Docker Hub, created a container from it, ran it, printed the message, and the container exited.

### Recap

You have Docker installed. The CLI talks to a daemon, and the daemon pulls images from a registry when it needs them.

---

## Lesson 3: Core vocabulary

### What you will learn

The handful of words that everything else in Docker is built on.

### Concept

| Term | Meaning | Analogy |
| --- | --- | --- |
| **Image** | A read-only package containing your app, its dependencies, and the minimal system files it needs | A class in programming, or a recipe |
| **Container** | A running (or stopped) instance created from an image | An object created from the class, or the meal cooked from the recipe |
| **Dockerfile** | A text file with step-by-step instructions to build an image | The written recipe itself |
| **Registry** | A server that stores and distributes images | A library or app store |
| **Tag** | A label for a version of an image, like `postgres:16` | A version number |
| **Layer** | One step of an image build, stored and cached separately | A page in a stack of pages |
| **Volume** | Storage that lives outside a container and survives its removal | A rented storage unit |
| **Network** | A virtual network that lets containers talk to each other | A private office phone system |

The most important pair is image and container. An image never changes and never runs. A container is what you get when you run it. You can start five containers from one image, and each is isolated from the others, just like you can bake five cakes from one recipe.

The overall flow:

```
Dockerfile  ->  (build)  ->  Image  ->  (run)  ->  Container
```

### Check yourself

1. What is the difference between an image and a container?
2. If you start three containers from the same image, how many images do you have?

---

## Lesson 4: Running your first containers

### What you will learn

The everyday commands for starting, inspecting, and stopping containers.

### Concept

`docker run` does several things in one step: it pulls the image if you do not have it, creates a container, and starts it.

One idea matters a lot here: **a container lives only as long as its main process.** `hello-world` printed a message and finished, so its container stopped immediately. A web server keeps running, so its container keeps running. When the main process ends, the container ends.

### Hands-on

Run a web server in the background:

```bash
docker run -d -p 8080:80 --name web nginx
```

- `-d` runs it in the background (detached).
- `-p 8080:80` connects port 8080 on your machine to port 80 in the container. Lesson 8 covers this properly.
- `--name web` gives the container a name so you do not have to use the generated ID.

Open `http://localhost:8080` in your browser. You should see the nginx welcome page.

Now inspect it:

```bash
docker ps
docker logs web
docker exec -it web sh
```

- `docker ps` lists running containers. Add `-a` to include stopped ones.
- `docker logs web` shows what the container printed. Add `-f` to follow along live.
- `docker exec -it web sh` opens a shell inside the running container. Try `ls /usr/share/nginx/html`, then type `exit`.

Stop and remove it:

```bash
docker stop web
docker ps -a
docker start web
docker stop web
docker rm web
```

Notice that a stopped container still exists until you remove it with `docker rm`.

One more experiment. Start an interactive Ubuntu container:

```bash
docker run -it ubuntu bash
```

Inside, run `cat /etc/os-release` and `ls /`. You are in a separate little world with its own filesystem. Type `exit` to leave. Because bash was the main process, the container stops when you exit.

### Recap

`run` creates and starts, `ps` lists, `logs` shows output, `exec` opens a door inside, `stop` halts, `rm` deletes. A container ends when its main process ends.

### Check yourself

1. Why did the `hello-world` container stop on its own while `nginx` kept running?
2. What is the difference between `docker stop` and `docker rm`?

---

## Lesson 5: Containers vs virtual machines

### What you will learn

How containers differ from virtual machines, and why that matters.

### Concept

A **virtual machine** (VM) emulates a whole computer. Each VM carries its own full operating system on top of a layer called a hypervisor. It is like an apartment building where every unit has its own foundation, plumbing, and electrical system. Strong separation, but heavy and slow to start.

A **container** shares the host machine's operating system kernel and only packages what the app needs on top. It is like apartments that share the building's foundation and utilities while each unit stays private. Much lighter.

| | Virtual machine | Container |
| --- | --- | --- |
| Includes a full OS | Yes | No, shares the host kernel |
| Typical size | Gigabytes | Megabytes to a few hundred MB |
| Startup time | Minutes | Seconds or less |
| Isolation | Very strong | Strong, but weaker than a VM |
| Good for | Running different operating systems, strong isolation | Packaging and running apps |

A detail worth knowing: containers share the host kernel, so Linux containers need a Linux kernel. On Windows and macOS, Docker Desktop quietly runs a small Linux VM underneath, and your containers run inside it.

### Recap

VMs virtualize hardware and carry a full OS. Containers virtualize the operating system and share the kernel, which makes them small and fast.

### Check yourself

1. Why do containers start faster than virtual machines?
2. Why does Docker Desktop on Windows run a Linux VM?

---

## Lesson 6: Building your first image

### What you will learn

How to write a Dockerfile and build your own image.

### Hands-on

Create a folder called `demo` and add three files.

**`go.mod`**

```
module demo

go 1.23
```

**`main.go`**

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "docker-demo"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s\n", name)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	log.Printf("%s listening on :%s", name, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

**`Dockerfile`** (no file extension)

```dockerfile
FROM golang:1.23
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o server .
EXPOSE 8080
CMD ["./server"]
```

What each instruction does:

| Instruction | Meaning |
| --- | --- |
| `FROM` | The base image to start from. Here, one that already has Go installed |
| `WORKDIR` | Sets the working folder inside the image. Created if missing |
| `COPY` | Copies files from your machine into the image |
| `RUN` | Runs a command while the image is being built |
| `EXPOSE` | Documents which port the app listens on |
| `CMD` | The command that runs when a container starts |

Build and run it:

```bash
docker build -t demo:1.0 .
docker run -d -p 8080:8080 --name demo demo:1.0
curl http://localhost:8080
docker logs demo
```

- `-t demo:1.0` tags the image with a name and version.
- The final `.` is the **build context**: the folder whose files Docker may copy from.
- On PowerShell, use `curl.exe` instead of `curl`, or just open the URL in a browser.

You should see `Hello from docker-demo`. Check the image:

```bash
docker images
```

Note the size. It is large, because the base image contains the entire Go toolchain. Lesson 13 shows how to shrink it.

### Recap

A Dockerfile lists the steps to build an image. `docker build` turns it into an image, and `docker run` turns the image into a container.

### Check yourself

1. What does the `.` at the end of `docker build -t demo:1.0 .` mean?
2. What is the difference between `RUN` and `CMD`?

---

## Lesson 7: Layers and the build cache

### What you will learn

Why Dockerfile instruction order affects build speed.

### Concept

Every instruction in a Dockerfile creates a **layer**. Docker caches each layer. On the next build, if an instruction and its inputs have not changed, Docker reuses the cached layer instead of running it again.

The rule that follows: **once a layer changes, every layer after it is rebuilt.**

So you should put things that change rarely near the top and things that change often near the bottom. Dependencies change rarely. Your source code changes constantly. That is why the Dockerfile copies `go.mod` and downloads dependencies before copying the rest of the code.

Think of it as packing a suitcase. Things you never unpack go in first. Things you grab every day go on top.

### Hands-on

Build again without changing anything:

```bash
docker build -t demo:1.0 .
```

It finishes almost instantly, and the output marks steps as `CACHED`.

Now edit `main.go` and change `"Hello from %s\n"` to `"Greetings from %s\n"`. Rebuild:

```bash
docker build -t demo:1.0 .
```

Watch which steps are cached. The `go mod download` step stays cached because `go.mod` did not change. The `COPY . .` step and everything after it run again, because your source changed.

Look at the layers:

```bash
docker history demo:1.0
```

**Thought exercise.** Imagine the Dockerfile did `COPY . .` first and `RUN go mod download` second. What would happen on every small code edit? (The answer is in Appendix C.)

### Recap

Each instruction is a cached layer. A changed layer invalidates everything after it, so order instructions from least to most frequently changing.

### Check yourself

1. Why is `go.mod` copied before the rest of the source code?
2. If you change only the `CMD` line, which layers get rebuilt?

---

## Lesson 8: Ports

### What you will learn

How traffic from your machine reaches a container.

### Concept

A container has its own isolated network. A server listening on port 8080 inside it is not automatically reachable from outside. You have to **publish** the port.

```
-p HOST_PORT:CONTAINER_PORT
```

The left side is the port on your machine. The right side is the port inside the container. Think of a building with a front desk: visitors call the main number (host port), and the desk forwards the call to the right room (container port).

### Hands-on

Run the same image twice on different host ports:

```bash
docker rm -f demo
docker run -d -p 8080:8080 --name demo1 demo:1.0
docker run -d -p 8081:8080 --name demo2 demo:1.0
curl http://localhost:8080
curl http://localhost:8081
```

Both respond, and both apps believe they are on port 8080. The host ports are different, so there is no clash.

To make a port reachable only from your own machine, bind it to localhost:

```bash
docker run -d -p 127.0.0.1:8082:8080 --name demo3 demo:1.0
```

Clean up:

```bash
docker rm -f demo1 demo2 demo3
```

`docker rm -f` stops and removes in one step.

Things to remember:

- `EXPOSE` in a Dockerfile is documentation only. It does not publish anything. Only `-p` does.
- If you see "port is already allocated", something on your machine already uses that host port. Pick another one.
- Inside the container, the app must listen on all interfaces. In Go, `":8080"` does this. Listening on `127.0.0.1` inside a container makes it unreachable from outside.

### Recap

`-p HOST:CONTAINER` publishes a port. Host ports must be unique, and container ports can repeat across containers.

### Check yourself

1. What does `-p 9000:3000` mean?
2. Does `EXPOSE 8080` in a Dockerfile make the port reachable from your browser?

---

## Lesson 9: Environment variables

### What you will learn

How to configure a container without rebuilding the image.

### Concept

An image should be the same everywhere, but its settings should not be. The database address on your laptop is different from the one in production. The standard solution is **environment variables**: you pass settings in when the container starts.

The app from Lesson 6 already reads `APP_NAME` and `PORT` from its environment.

### Hands-on

```bash
docker run -d -p 8081:8080 -e APP_NAME=orders --name orders demo:1.0
curl http://localhost:8081
```

You should see `Hello from orders`. Same image, different behavior.

Check what the container sees:

```bash
docker exec orders env
```

You can also set a default inside the Dockerfile with `ENV APP_NAME=docker-demo`. A value passed with `-e` at run time overrides it.

**A warning about secrets.** Do not put passwords or API keys in a Dockerfile with `ENV` or `COPY`. They become part of the image, and anyone with the image can read them using `docker history` or by inspecting the layers. Pass secrets in at run time instead.

Clean up with `docker rm -f orders`.

### Recap

Use `-e NAME=value` to configure containers at run time. Keep configuration and secrets out of the image.

### Check yourself

1. Why is it better to pass a database password at run time than to write it in the Dockerfile?
2. If the Dockerfile sets `ENV APP_NAME=a` and you run with `-e APP_NAME=b`, which value wins?

---

## Lesson 10: Volumes and persistent data

### What you will learn

How to keep data after a container is gone.

### Concept

The filesystem inside a container is temporary. When you remove the container, everything written inside it disappears. That is fine for a stateless web app and a disaster for a database.

A **volume** is storage that lives outside the container. You attach it, and the data survives no matter what happens to the container. Think of a storage unit: you can replace the truck that visits it as often as you like, and your belongings stay put.

There are two kinds you will use:

| | Named volume | Bind mount |
| --- | --- | --- |
| What it is | Storage managed by Docker | A folder from your machine mapped into the container |
| Best for | Databases and other app data | Development, where you edit files and see changes instantly |
| Syntax | `-v mydata:/data` | `-v "$PWD/site":/usr/share/nginx/html` |

### Hands-on

**1. See the data disappear.**

```bash
docker run --name scratch alpine sh -c "echo hi > /note.txt"
docker rm scratch
docker run --rm alpine cat /note.txt
```

The last command fails with "No such file or directory". The file died with the first container. (`--rm` removes a container automatically when it exits.)

**2. Use a named volume.**

```bash
docker run --rm -v mydata:/data alpine sh -c "echo hello > /data/note.txt"
docker run --rm -v mydata:/data alpine cat /data/note.txt
docker volume ls
docker volume inspect mydata
```

The second container is brand new, yet it prints `hello`. The data lives in the volume. Remove it when you are done:

```bash
docker volume rm mydata
```

**3. Use a bind mount.**

```bash
mkdir site
echo "<h1>Hello from my laptop</h1>" > site/index.html
docker run -d -p 8080:80 -v "$PWD/site":/usr/share/nginx/html:ro --name site nginx
```

Open `http://localhost:8080`. Now edit `site/index.html` and refresh the page. The change shows up immediately, because the container is reading your actual folder. The `:ro` at the end makes the mount read-only inside the container.

On Windows `cmd`, use `%cd%` instead of `$PWD`. Clean up with `docker rm -f site`.

### Recap

Container storage is disposable. Use named volumes for data that must survive, and bind mounts to share files between your machine and a container while developing.

### Check yourself

1. What happens to files written inside a container when you remove it?
2. Would you use a named volume or a bind mount for a database's data files? Why?

---

## Lesson 11: Networking between containers

### What you will learn

How containers find and talk to each other.

### Concept

Picture an office with a private phone system. Colleagues dial each other using short internal extensions. Outsiders need the public number, and only the numbers you choose to publish ring through.

Docker works the same way:

- Containers on the same **user-defined network** can reach each other by **name**. Docker provides the lookup for you.
- Published ports (`-p`) are the public numbers. Publish only what the outside world must reach. A web API usually needs one. Its database should not have one.

One mistake catches nearly everyone: inside a container, `localhost` means **that container itself**. If your API container tries to reach a database at `localhost:5432`, it is knocking on its own door. Use the database container's name instead.

### Hands-on

```bash
docker network create appnet
docker run -d --name web --network appnet nginx
docker run -it --rm --network appnet alpine sh
```

You are now inside a second container on the same network. Run:

```bash
wget -qO- http://web
```

You get nginx's welcome page, even though `web` published no port to your machine. Type `exit`, then clean up:

```bash
docker rm -f web
docker network rm appnet
```

Useful commands: `docker network ls` lists networks, and `docker network inspect appnet` shows which containers are attached.

Docker Compose creates a network for your project automatically, which is why service names just work there (next lesson).

### Recap

Put containers on the same user-defined network and they reach each other by name. Publish ports only for traffic coming from outside. `localhost` inside a container is the container itself.

### Check yourself

1. Why would an API container fail to connect to a database using `localhost`?
2. How does one container address another on the same user-defined network?

---

## Lesson 12: Docker Compose

### What you will learn

How to describe and run several containers together from one file.

### Concept

Real applications have multiple parts: an API, a database, maybe a cache. Starting each with a long `docker run` command, plus the right networks, volumes, and variables, gets tedious and easy to get wrong.

**Docker Compose** lets you write all of that in one YAML file and start everything with one command. The file is also documentation of how your system fits together, and you can commit it to version control.

### Hands-on

In your `demo` folder from Lesson 6, create `compose.yaml`:

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      APP_NAME: compose-demo
```

Reading it:

- `services` lists the containers. Here there is one, called `app`.
- `build: .` builds the image from the Dockerfile in this folder.
- `ports` and `environment` are the same ideas as `-p` and `-e`.

Run it:

```bash
docker compose up -d
docker compose ps
docker compose logs -f
```

Press Ctrl+C to stop following logs. Visit `http://localhost:8080`, then shut it down:

```bash
docker compose down
```

After changing your code, rebuild and restart:

```bash
docker compose build
docker compose up -d
```

The commands you will use most:

| Command | What it does |
| --- | --- |
| `docker compose up -d` | Create and start everything in the background |
| `docker compose down` | Stop and remove the containers and network |
| `docker compose down -v` | Same, and also delete the named volumes |
| `docker compose ps` | Show the project's containers |
| `docker compose logs -f api` | Follow logs for the `api` service |
| `docker compose exec api sh` | Open a shell in the `api` service |
| `docker compose build` | Rebuild images |

Three notes:

- Compose accepts `compose.yaml` or `docker-compose.yml` as the file name.
- Older tutorials start with a `version:` line. You no longer need it.
- Service names become hostnames. In the final project, the API reaches the database at the host name `db`.

### Recap

Compose describes a multi-container setup in one file. `up -d` starts it and `down` stops it.

### Check yourself

1. Which command starts all services in the background?
2. What does `docker compose down -v` do that plain `down` does not?

---

## Lesson 13: Smaller images: .dockerignore and multi-stage builds

### What you will learn

How to make images smaller, faster to ship, and safer.

### Concept

The image from Lesson 6 contains the whole Go toolchain, even though it is only needed to compile the program. Shipping a compiler to production is wasteful and widens the attack surface.

**Multi-stage builds** fix this. You use one stage to build and a second, tiny stage to run. Only the finished program is copied across. Think of a restaurant: the messy prep kitchen stays in the back, and the customer only receives the finished plate.

A second tool is the **`.dockerignore`** file. It works like `.gitignore` and keeps unwanted files out of the build context, which speeds up builds and prevents accidents like copying secrets into an image.

### Hands-on

Create `.dockerignore` in `demo`:

```
.git
.env
*.md
```

Replace your `Dockerfile` with this:

```dockerfile
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/server ./server
EXPOSE 8080
CMD ["./server"]
```

What is new:

- `AS build` names the first stage.
- `CGO_ENABLED=0` makes Go produce a fully static binary that needs no system libraries, so it can run on a minimal image.
- The second `FROM` starts a fresh, tiny image. Everything from the first stage is thrown away except what you copy over.
- `COPY --from=build` pulls the compiled binary out of the first stage.

Build it under a new tag and compare:

```bash
docker build -t demo:small .
docker images
```

The old image is several hundred megabytes. The new one is a small fraction of that, typically around ten to fifteen. It runs identically:

```bash
docker run -d -p 8080:8080 --name small demo:small
curl http://localhost:8080
docker rm -f small
```

### Recap

Build in one stage, run in another. Use `.dockerignore` to keep the build context clean. Smaller images pull faster and expose less.

### Check yourself

1. Why is a smaller image better beyond saving disk space?
2. What does `COPY --from=build` do?

---

## Lesson 14: Registries and tags

### What you will learn

How images are named, versioned, and shared.

### Concept

A full image name has up to four parts:

```
registry/namespace/repository:tag
docker.io/yourname/demo:1.0
```

- **Registry** is where the image lives. It defaults to Docker Hub if omitted. Others include GitHub Container Registry and the registries from the big cloud providers.
- **Namespace** is usually a user or organization name.
- **Repository** is the image's name.
- **Tag** is the version label. If you leave it out, Docker uses `latest`.

A common misunderstanding: `latest` is just a tag that happens to be the default. It does not mean "the newest version." It means whatever the publisher last labeled that way, and it can change under you. For anything that matters, use a specific version like `postgres:16` or `demo:1.0`.

### Hands-on

You need a free Docker Hub account for this. Replace `yourname` with your username.

```bash
docker login
docker tag demo:small yourname/demo:1.0
docker push yourname/demo:1.0
```

`docker tag` gives an existing image a second name. `docker push` uploads it. Now anyone can run it:

```bash
docker pull yourname/demo:1.0
docker run -d -p 8080:8080 yourname/demo:1.0
```

Stop and remove that container when you are done.

### Recap

Images are named `registry/namespace/repository:tag`. Push to share, pull to fetch. Pin specific tags instead of relying on `latest`.

### Check yourself

1. Does the `latest` tag always point to the newest version of an image?
2. What are the parts of the name `ghcr.io/acme/billing:2.3`?

---

## Lesson 15: Debugging and cleanup

### What you will learn

How to find out why a container misbehaves, and how to reclaim disk space.

### Concept

Most Docker problems fall into a few buckets: the container will not start, it starts and exits, it runs but cannot be reached, or it runs but behaves wrongly. A calm, repeatable routine finds most of them.

### Hands-on

**The debugging routine**

```bash
docker ps -a
docker logs <container>
docker inspect <container>
docker exec -it <container> sh
docker stats
```

1. `docker ps -a` shows every container, including ones that exited, with their status.
2. `docker logs` is the first place to look. Most errors are printed there.
3. `docker inspect` dumps the full configuration as JSON: environment, mounts, networks, ports.
4. `docker exec -it ... sh` lets you look around inside a running container.
5. `docker stats` shows live CPU and memory use.

If a container exits immediately and you cannot exec into it, start a fresh one with a shell instead of the usual command:

```bash
docker run -it demo:small sh
```

**Exit codes worth recognizing**

| Code | Usual meaning |
| --- | --- |
| 0 | The process finished normally |
| 1 | The application hit an error |
| 137 | The process was killed, often because it ran out of memory |
| 139 | The process crashed with a memory fault |

You can see a stopped container's code in `docker ps -a`.

**Cleanup**

Docker keeps stopped containers, old images, and unused volumes until you remove them.

```bash
docker system df
docker container prune
docker image prune
docker system prune
```

- `docker system df` shows how much space Docker is using.
- `prune` commands remove unused things after asking for confirmation.
- `docker volume prune` removes unused volumes. Be careful: that deletes data permanently.

### Recap

Check status, then logs, then inspect, then exec in. Clean up with the `prune` commands, and think twice before pruning volumes.

### Check yourself

1. A container keeps exiting right after it starts. What are the first two commands you run?
2. What does exit code 137 usually point to?

---

## Lesson 16: Good practices and security basics

### What you will learn

Habits that make your images smaller, safer, and more predictable.

### Concept and checklist

1. **Pin versions.** Use `golang:1.23` and `postgres:16`, not `latest`. Builds stay reproducible.
2. **Start from small base images.** Alpine or slim variants carry less software, so there is less to patch and less to attack.
3. **Do not run as root.** If someone exploits your app, a non-root user limits the damage. Add this to the final stage of a Dockerfile:

   ```dockerfile
   RUN adduser -D appuser
   USER appuser
   ```

   On Alpine, `adduser -D` creates a user without a password. Other base images use different commands.
4. **One main job per container.** Run the API in one container and the database in another. This makes each easier to restart, scale, and replace.
5. **Keep secrets out of images.** No passwords or keys in `ENV`, `COPY`, or build arguments. Pass them at run time.
6. **Use `.dockerignore`.** Keep `.git`, `.env`, and local clutter out of the build context.
7. **Add health checks.** A health check lets Docker and Compose know when a service is actually ready, not just started. The final project uses one.
8. **Set a restart policy** for long-running services, for example `restart: unless-stopped` in Compose.
9. **Order Dockerfile steps for caching.** Rarely changing steps first.
10. **Scan images.** Tools such as Docker Scout and Trivy check images for known vulnerabilities. Rebuild regularly to pick up patched base images.

**Keeping passwords out of the Compose file.** Compose automatically reads a file named `.env` in the same folder and lets you reference its values:

```
# .env
DB_PASSWORD=choose-something-better
```

```yaml
environment:
  POSTGRES_PASSWORD: ${DB_PASSWORD}
```

Add `.env` to `.gitignore` so it never gets committed. For real production systems, use your platform's secret management instead of plain files.

### Check yourself

1. Why should a container process run as a non-root user?
2. Why avoid `latest` in production?

---

## Lesson 17: Final project: Go API with Postgres

### What you will build

A notes API with two endpoints, backed by Postgres, started with a single command. You will use almost everything from this tutorial: a multi-stage Dockerfile, a non-root user, Compose, networking by name, a volume, a health check, and environment variables.

```
 your machine                 Docker network (created by Compose)
 ------------                 ------------------------------------
 curl :8080  ->  [ api container ]  ->  [ db container ]
                                         |
                                    named volume: pgdata
```

### Step 1: Create the project

Make a folder called `notes` and add the files below.

**`main.go`**

```go
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func connect() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return db
		}
		log.Printf("database not ready, retrying: %v", err)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("could not connect to database: %v", err)
	return nil
}

func main() {
	db := connect()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		text TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok\n"))
	})

	http.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var n Note
			if err := json.NewDecoder(r.Body).Decode(&n); err != nil || n.Text == "" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			err := db.QueryRow(
				`INSERT INTO notes (text) VALUES ($1) RETURNING id`, n.Text,
			).Scan(&n.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(n)

		case http.MethodGet:
			rows, err := db.Query(`SELECT id, text FROM notes ORDER BY id`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			notes := []Note{}
			for rows.Next() {
				var n Note
				if err := rows.Scan(&n.ID, &n.Text); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				notes = append(notes, n)
			}
			json.NewEncoder(w).Encode(notes)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Generate `go.mod` and `go.sum`.** If Go is installed on your machine:

```bash
go mod init notes
go mod tidy
```

If it is not, let a container do it for you. This also uses a bind mount from Lesson 10:

```bash
docker run --rm -v "$PWD":/app -w /app golang:1.23 sh -c "go mod init notes && go mod tidy"
```

On Linux the generated files may end up owned by root. If that bothers you, run `sudo chown $USER go.mod go.sum`.

**`Dockerfile`**

```dockerfile
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server .

FROM alpine:3.20
RUN adduser -D appuser
WORKDIR /app
COPY --from=build /out/server ./server
USER appuser
EXPOSE 8080
CMD ["./server"]
```

**`.dockerignore`**

```
.git
.env
*.md
```

**`compose.yaml`**

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:secret@db:5432/notes?sslmode=disable
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: notes
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d notes"]
      interval: 5s
      timeout: 3s
      retries: 10

volumes:
  pgdata:
```

Take a moment to connect this file to earlier lessons:

- The API reaches the database at the host name `db` (Lesson 11).
- The database port is **not** published, so only the API can reach it (Lesson 11).
- `pgdata` is a named volume holding Postgres's files (Lesson 10).
- The health check makes Compose wait until Postgres is accepting connections before starting the API. The retry loop in `main.go` is a second line of defense, since real systems should not assume startup order.
- The passwords here are plain text for learning only. Lesson 16 shows how to move them to `.env`.

### Step 2: Start everything

```bash
docker compose up -d
docker compose ps
docker compose logs -f api
```

Press Ctrl+C to stop following the logs. Wait until `db` shows as healthy and `api` is running.

### Step 3: Try it

```bash
curl -X POST http://localhost:8080/notes -H "Content-Type: application/json" -d '{"text":"first note"}'
curl -X POST http://localhost:8080/notes -H "Content-Type: application/json" -d '{"text":"second note"}'
curl http://localhost:8080/notes
curl http://localhost:8080/health
```

On PowerShell, quoting JSON is awkward. Use `curl.exe` with escaped quotes (`-d "{\"text\":\"first note\"}"`), or send the requests from Postman or any REST client.

### Step 4: Prove the data persists

```bash
docker compose down
docker compose up -d
curl http://localhost:8080/notes
```

Your notes are still there. The containers were destroyed and recreated, but the volume was untouched.

Now delete the volume too:

```bash
docker compose down -v
docker compose up -d
curl http://localhost:8080/notes
```

The list is empty. This is the exact difference between `down` and `down -v`.

When you are finished, run `docker compose down -v` to clean up.

### Step 5: Challenges

Try these without looking anything up first, then check the docs.

1. Move the database password into a `.env` file and reference it in `compose.yaml`.
2. Add `restart: unless-stopped` to both services. Then stop the `db` container manually and watch what happens.
3. Add a third service, `adminer` (image `adminer`, published on host port 8081), and use it to browse your `notes` table. Which host name do you enter as the server?
4. Shrink the image further and compare sizes with `docker images`.
5. Add a `GET /notes/{id}` endpoint, rebuild with `docker compose build`, and restart.

---

## Appendix A: Cheat sheet

**Images**

| Command | Purpose |
| --- | --- |
| `docker build -t name:tag .` | Build an image from the Dockerfile in this folder |
| `docker images` | List local images |
| `docker pull name:tag` | Download an image |
| `docker tag src dst` | Give an image another name |
| `docker push name:tag` | Upload an image |
| `docker rmi name:tag` | Remove an image |
| `docker history name:tag` | Show an image's layers |

**Containers**

| Command | Purpose |
| --- | --- |
| `docker run -d -p 8080:80 --name web nginx` | Run in background with a published port and a name |
| `docker run -it image sh` | Run interactively with a shell |
| `docker run -e KEY=value image` | Set an environment variable |
| `docker run -v name:/path image` | Attach a named volume |
| `docker run --rm image` | Remove the container when it exits |
| `docker ps` / `docker ps -a` | List running / all containers |
| `docker logs -f name` | Follow a container's logs |
| `docker exec -it name sh` | Open a shell in a running container |
| `docker stop name` / `docker start name` | Stop / start a container |
| `docker rm name` / `docker rm -f name` | Remove a container / force remove |
| `docker inspect name` | Show full details as JSON |
| `docker stats` | Live resource usage |

**Volumes and networks**

| Command | Purpose |
| --- | --- |
| `docker volume ls` | List volumes |
| `docker volume inspect name` | Show volume details |
| `docker volume rm name` | Remove a volume |
| `docker network create name` | Create a network |
| `docker network ls` | List networks |

**Compose**

| Command | Purpose |
| --- | --- |
| `docker compose up -d` | Start everything in the background |
| `docker compose down` | Stop and remove containers and network |
| `docker compose down -v` | Same, plus delete named volumes |
| `docker compose ps` | List project containers |
| `docker compose logs -f service` | Follow one service's logs |
| `docker compose exec service sh` | Shell into a service |
| `docker compose build` | Rebuild images |

**Cleanup**

| Command | Purpose |
| --- | --- |
| `docker system df` | Show disk usage |
| `docker container prune` | Remove stopped containers |
| `docker image prune` | Remove unused images |
| `docker system prune` | Remove stopped containers, unused networks, and dangling images |
| `docker volume prune` | Remove unused volumes (deletes data) |

---

## Appendix B: Troubleshooting

| Symptom | Likely cause and fix |
| --- | --- |
| "Cannot connect to the Docker daemon" | Docker is not running. Start Docker Desktop, or on Linux run `sudo systemctl start docker`. |
| "permission denied" on the Docker socket (Linux) | Your user is not in the `docker` group. See Lesson 2, then log out and back in. |
| "port is already allocated" | Another program or container uses that host port. Pick a different left-hand number in `-p`. |
| Container exits immediately | Read `docker logs`. Remember a container lives only while its main process runs. |
| "connection refused" between containers | You used `localhost`, or the containers are on different networks. Use the service or container name. Also check the app listens on all interfaces. |
| Code change does not show up | You did not rebuild. Run `docker build` or `docker compose build`, then restart. |
| Data vanished | There was no volume, or you ran `down -v` or `volume prune`. |
| `COPY` says file not found | The file is outside the build context or excluded by `.dockerignore`. |
| "exec format error" | The image was built for a different CPU architecture, for example ARM vs x86. Rebuild for the right platform. |
| "no space left on device" | Run `docker system df`, then prune what you no longer need. |
| Changed `POSTGRES_PASSWORD` but nothing happened | Postgres reads it only when first initializing an empty volume. Remove the volume with `docker compose down -v` to start fresh. |

---

## Appendix C: Answer key

**Lesson 3**
1. An image is a read-only blueprint. A container is a running instance created from it.
2. One image. Three containers.

**Lesson 4**
1. `hello-world` runs a program that finishes immediately, so its container ends. `nginx` runs a server that keeps going.
2. `stop` halts a running container. `rm` deletes the container.

**Lesson 5**
1. Containers do not boot an operating system. They start a process on the host's existing kernel.
2. Linux containers need a Linux kernel, and Windows and macOS do not have one, so Docker Desktop provides a small Linux VM.

**Lesson 6**
1. It is the build context: the folder Docker may copy files from when building.
2. `RUN` executes during the build and its result is stored in the image. `CMD` defines what runs when a container starts.

**Lesson 7**
- Thought exercise: every code edit would invalidate the `COPY . .` layer, and since the dependency download comes after it, dependencies would be re-downloaded on every build.
1. Dependencies change rarely. Copying `go.mod` first lets the download layer stay cached when only source code changes.
2. Only the `CMD` layer changes, and nothing else needs rebuilding.

**Lesson 8**
1. Host port 9000 forwards to port 3000 inside the container.
2. No. `EXPOSE` is documentation. Only `-p` publishes a port.

**Lesson 9**
1. Anyone who gets the image can read what is baked into it, and changing the password would require a rebuild.
2. `b`. The value passed at run time overrides the default.

**Lesson 10**
1. They are deleted along with the container.
2. A named volume. Docker manages it, it survives container removal, and no particular folder on your machine is involved.

**Lesson 11**
1. Inside the API container, `localhost` is the API container itself, not the database container.
2. By its name, for example `db`.

**Lesson 12**
1. `docker compose up -d`
2. It also deletes the named volumes defined in the project, removing the stored data.

**Lesson 13**
1. Faster pulls and deploys, and less software inside the image, which means fewer things to attack or patch.
2. It copies a file from the earlier build stage into the current stage.

**Lesson 14**
1. No. It is just the default tag, and whatever the publisher last labeled that way.
2. Registry `ghcr.io`, namespace `acme`, repository `billing`, tag `2.3`.

**Lesson 15**
1. `docker ps -a` to see the status, then `docker logs <container>` to read the error.
2. The process was killed, often because it ran out of memory.

**Lesson 16**
1. If the app is compromised, a non-root user limits what the attacker can do.
2. The tag can change, so builds are not reproducible and a rebuild may silently pull a different version.

---

## Appendix D: Where to go next

You now understand the core of Docker. Here is a sensible path forward:

1. **Environment-specific Compose files.** Learn how to override settings for development and production using multiple Compose files.
2. **CI/CD.** Build and push images automatically from GitHub Actions or another pipeline whenever you merge code.
3. **Image scanning.** Add Docker Scout or Trivy to your pipeline.
4. **Hot reload in development.** Combine bind mounts with a tool like Air for Go so code changes restart the app inside the container without rebuilding.
5. **Deployment targets.** Run your containers on a single server first. Then look at managed container services from the cloud providers.
6. **Kubernetes.** When one machine is not enough, Kubernetes runs and manages containers across many. Everything you learned here carries over: images, containers, volumes, networks, and environment variables all have direct counterparts.

The best practice from here is repetition. Containerize your next project from the start, and the commands will become second nature.
