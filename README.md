# asu-java-build

`asu-java-build` is a lightweight Java build tool written in Go.
It compiles Java source files, collects dependencies, and packages them into either a standard JAR or a self-contained fat-JAR. 
It also supports building **native executables** (via GraalVM Native Image, using Docker or local tooling).

---

## Features

- **Automatic dependency collection**
  - Scans `lib/**/*.jar` under the project
  - Reads additional dependencies from a `dep.lib` file
  - Supports relative paths resolved against `~/.m2/repository`
- **Standard JAR mode**
  - Generates `META-INF/dependencies.lib` inside the JAR
  - Writes a `Class-Path` entry in `MANIFEST.MF`
- **Static (fat-JAR) mode**
  - Bundles all dependencies into a single executable JAR
  - Generates `META-INF/dependencies-static.lib`
  - Merges `META-INF/services/*` files across dependencies
- **Export dependencies**
  - With `-explib`, copies dependencies into an output `lib/` directory
- **Native image build**
  - Build GraalVM native executables via Docker or local GraalVM
- **Cross-platform support**
  - Works on Windows, Linux, and macOS
  - Handles path separators automatically
- **UTF-8 by default**
  - All file I/O uses UTF-8 encoding

---

## Installation

Requirements:

- Go 1.18+
- JDK (with `javac` and `jar` available in PATH)
- For native builds:
  - GraalVM `native-image` locally, or
  - Docker (using `ghcr.io/graalvm/native-image-community` images)

Build from source:

```bash
git clone https://github.com/bruceasu/asu-java-build.git
cd asu-java-build
go build -o asu-java-build main.go
````

On Windows the output will be `asu-java-build.exe`.

---

## Usage

```
Usage of asu-java-build:
  -docker-image string
        Docker image for linux native build (default "ghcr.io/graalvm/native-image-community:21-ol9")
  -explib
        Export dependencies to lib/ folder next to the JAR
  -jar string
        Use an existing JAR instead of compiling sources (skip javac/jar)
  -jarname string
        Runnable JAR name when main class is known (default "app.jar")
  -main string
        Main class FQCN (e.g. com.example.Main). If omitted and only one source exists, it will be inferred.
  -native
        Build a native image after JAR is ready
  -native-args string
        Extra args to pass to native-image (e.g. '-Ob -march=native')
  -native-driver string
        Native builder: docker | local (default "docker")
  -native-name string
        Native binary name (default: app-<os>-amd64)
  -native-target string
        Target OS for native when using docker: linux | windows (default "linux")
  -nojar
        Do not create JAR (will disable native build)
  -out string
        Output directory for .class/JAR/native (default "bin")
  -src string
        Java source directory (recursive) (default "src")
  -static
        Package dependencies inside JAR (fat-jar, like shade plugin)
```

---

## Examples

### Standard JAR

```bash
asu-java-build -src src -out bin -main com.example.Main
```

Produces:

```
bin/app.jar
META-INF/dependencies.lib (inside the JAR)
```

### Fat-JAR

```bash
asu-java-build -src src -out bin -main com.example.Main -static
```

Produces:

```
bin/app.jar (includes all dependencies)
META-INF/dependencies-static.lib (inside the JAR)
```

### Export dependencies

```bash
asu-java-build -src src -out bin -main com.example.Main -explib
```

Produces:

```
bin/lib/a.jar
bin/lib/b.jar
```

### Build a native executable (via Docker)

```bash
asu-java-build -src src -out bin -main com.example.Main -native
```

Produces:

```
bin/app-linux-amd64   (native binary)
```

### Build a native executable (local GraalVM)

```bash
asu-java-build -src src -out bin -main com.example.Main -native -native-driver local
```

Pass extra options to GraalVM:

```bash
asu-java-build -src src -out bin -main com.example.Main -native -native-args "-Ob -march=native"
```

---

## Launcher

This project only builds JARs / native executables.
For runtime launching of JARs, see the companion project **[asu-run-jar](https://github.com/bruceasu/asu-run-jar)**:

* If the JAR contains `dependencies-static.lib` → run with `java -jar`
* If the JAR contains `dependencies.lib` → constructs the correct classpath and runs
* Otherwise → falls back to `java -jar`

---

## Development Notes

* `collectDeps()` → gathers dependencies from `lib/` and `dep.lib`
* `javacCompile()` → invokes `javac` with proper `-cp`
* `buildJar()` / `buildFatJar()` → package into standard or fat-JAR
* `copyExplib()` → export dependencies into `lib/` folder
* In fat-JAR mode:

  * Keep only the **main JAR’s** `META-INF/MANIFEST.MF`
  * Skip `MANIFEST.MF` from dependencies
  * Merge `META-INF/services/*` files

Contributions and PRs are welcome.

---

## License

```
Apache License 2.0

Copyright 2025 BruceAsu

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

