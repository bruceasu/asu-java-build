# JBang feature comparison

`asu-java-build` already handles recursive Java source builds, Maven projects,
fat JARs, `jlink`/`jpackage` runtime images, and IDE project
scaffolding. The most useful JBang gap for this project is lightweight Maven
dependency declaration for ordinary Java source projects.

## Added: source-level dependencies and BOMs

For a non-Maven project, declare dependencies in the main Java source file
using `//DEPS` lines. The directives can appear anywhere in that file, and a
line can contain multiple space-separated dependencies:

```java
//DEPS io.quarkus:quarkus-bom:3.2.0@pom
//DEPS io.quarkus:quarkus-core
//DEPS ./vendor/codec.jar ./vendor/plugins/
```

The BOM supplies the version for the versionless dependency. Maven resolves its
runtime dependency graph, including transitive dependencies. Maven must be
available on `PATH` (or through the project's Maven Wrapper) for transitive
resolution, BOMs, and versionless coordinates.

Local JAR paths, directories, and filesystem globs are also accepted. Relative
paths are resolved from the project working directory. A directory is scanned
recursively for JAR files. This lets a project keep its dependency declarations
with its main source and avoid requiring `dep.lib` or a conventional `lib/`
folder.

`dep.lib` and `lib/` are still accepted for existing projects, but are optional
when dependencies are declared with `//DEPS`. The same Maven coordinate forms
remain accepted in `dep.lib`:

```text
io.quarkus:quarkus-bom:3.2.0@pom
io.quarkus:quarkus-core
```

Existing exact coordinates such as `org.slf4j:slf4j-api:2.0.17` remain
supported. If Maven is unavailable, these retain the previous direct-JAR
download behavior, with a warning that transitive dependencies were not
resolved. BOMs and versionless coordinates require Maven and fail with an
actionable error when it is missing.

For deployable output, use `-static` to merge dependencies into one JAR or
`-explib` to copy them beside the application JAR.

## Remaining JBang gaps

- Run a Java file directly as a script, including JBang's compact source mode.
- Select or download a requested JDK version per script.
- Declare custom repositories and richer dependency rules such as exclusions.
- Include explicit extra source/resource files through `//SOURCES` and `//FILES`;
  this tool currently compiles all Java files under its source directory.
- Install scripts as commands and manage aliases/catalogs for local or remote
  scripts.
- Export a script into a Maven/Gradle project. The builder can generate IDE
  projects, but does not perform this script-to-project export workflow.

These are lower priority than dependency graph support because `asu-java-build`
already covers project builds, packaging, runtime images, and
IDE scaffolding.
