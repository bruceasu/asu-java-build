# asu-java-build

[English](#english) | [中文](#中文)

---

## 中文

### 简介

`asu-java-build` 是一个多功能的 Java 项目构建和管理工具，用 Go 语言编写。它可以：

- ✅ 编译 Java 源代码（支持递归扫描源文件）
- ✅ **自动检测和构建 Maven 项目**
- ✅ 打包可运行的 JAR 文件（普通/Fat JAR/依赖导出）
- ✅ 使用 GraalVM Native Image 构建原生二进制文件
- ✅ 支持 Docker 容器构建 Linux 原生镜像（跨平台）
- ✅ 支持本地 GraalVM 构建原生镜像
- ✅ **快速生成 IDE 项目**（VS Code/Eclipse/IntelliJ IDEA/Maven/Spring Boot）

### 特性

- **零配置入口推断**：单文件项目自动推断主类
- **Maven 项目支持**：
  - 自动检测 `pom.xml` 并使用 Maven 构建
  - 优先使用 Maven Wrapper（`mvnw`/`mvnw.cmd`）
  - 支持从 `pom.xml` 提取主类配置
  - 使用 Maven 插件管理依赖
  - 自动复制依赖到输出目录
- **灵活构建模式**：支持从源码、现有 JAR 文件构建
- **依赖管理**：
  - 自动扫描 `lib/` 目录和 `dep.lib` 文件收集依赖
  - 支持 Fat JAR 构建（类似 Maven Shade Plugin）
  - 支持导出依赖到 `lib/` 目录
  - 自动合并 `META-INF/services/` 文件（SPI 支持）
  - Maven 坐标自动解析和下载（Maven Central）
- **跨平台原生构建**：通过 Docker 在 Windows/Mac 上构建 Linux 原生镜像
- **本地原生构建**：支持使用本机 GraalVM 构建原生镜像
- **IDE 项目生成**：
  - VS Code Java 项目（带 settings.json 和示例代码）
  - Eclipse 项目（.project 和 .classpath）
  - IntelliJ IDEA 项目（.idea 和 .iml）
  - Maven 项目（通过 maven-archetype-quickstart）
  - Spring Boot 项目（通过 spring initializr CLI）
  - 自动初始化 Git 仓库和 .gitignore
- **自定义参数**：支持传递额外参数给 `native-image`

### 前置要求

#### 基础构建（Java 编译 + JAR）
- JDK 8 或更高版本（需要 `javac` 和 `jar` 命令在 PATH 中）

#### Docker 原生构建
- Docker Desktop（用于 `-native -native-driver docker`）

#### 本地原生构建
- GraalVM（需要安装 `native-image` 组件）

### 快速开始

#### 1. 构建工具本身

```bash
# Windows
go build -o asu-java-build.exe .

# Linux/Mac
go build -o asu-java-build .

# 交叉编译（在 Windows 上构建 Linux 版本）
set GOOS=linux
set GOARCH=amd64
go build -o asu-java-build-linux .
```

#### 2. 基本用法

```bash
# 场景 A：快速生成项目
asu-java-build -vsc              # 生成 VS Code Java 项目
asu-java-build -idea             # 生成 IntelliJ IDEA 项目
asu-java-build -maven            # 生成 Maven 项目
asu-java-build -spring           # 生成 Spring Boot 项目

# 场景 B：单文件项目（自动推断主类）
asu-java-build

# 场景 C：多文件项目（指定主类）
asu-java-build -main com.example.Main

# 场景 D：使用现有 JAR 构建原生镜像
asu-java-build -jar target/myapp.jar -native

# 场景 E：从源码一步到位构建原生镜像（Linux via Docker）
asu-java-build -main com.example.Main -native

# 场景 F：本地构建原生镜像（需要本机安装 GraalVM）
asu-java-build -main com.example.Main -native -native-driver local
```

### 命令行参数

#### IDE 项目生成参数
| 参数 | 说明 |
|------|------|
| `-vsc` | 生成 VS Code Java 项目（包含 settings.json 和示例 Main.java） |
| `-eclipse` | 生成 Eclipse 项目（包含 .project 和 .classpath） |
| `-idea` | 生成 IntelliJ IDEA 项目（包含 .idea 目录和 .iml 文件） |
| `-maven` | 使用 maven-archetype-quickstart 生成 Maven 项目 |
| `-spring` | 使用 spring CLI 生成 Spring Boot 项目 |

**注意**：所有 IDE 项目生成都会自动：
- 初始化 Git 仓库（如果尚未初始化）
- 创建 `.gitignore` 文件（排除 IDE 配置、编译输出等）
- 创建基本的目录结构

#### 基础编译参数
| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-src` | `src` | Java 源码目录（递归扫描） |
| `-out` | `bin` | 输出目录（.class/JAR/native） |
| `-main` | 自动推断 | 主类全限定名（如 `com.example.Main`） |
| `-jarname` | `app.jar` | 可运行 JAR 文件名 |
| `-nojar` | `false` | 不创建 JAR（将禁用原生构建） |

#### 依赖管理参数
| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-static` | `false` | 将依赖打包到 JAR 内（Fat JAR，类似 Maven Shade Plugin） |
| `-explib` | `false` | 将依赖导出到 JAR 同目录的 `lib/` 文件夹 |

#### 使用现有 JAR
| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-jar` | 空 | 使用现有 JAR 文件（跳过 javac/jar 阶段） |

#### 原生构建参数
| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-native` | `false` | 在 JAR 准备好后构建原生镜像 |
| `-native-driver` | `docker` | 构建方式：`docker` 或 `local` |
| `-native-target` | `linux` | Docker 构建目标系统：`linux` 或 `windows` |
| `-native-name` | `app-linux-amd64` | 原生二进制文件名 |
| `-native-args` | 空 | 传递给 `native-image` 的额外参数 |
| `-docker-image` | `ghcr.io/graalvm/native-image-community:21-ol9` | Docker 镜像 |

### 使用示例

#### 示例 1：编译单个 Java 文件

```bash
# 项目结构
src/
  com/
    example/
      Main.java

# 构建
asu-java-build

# 产物
bin/
  com/example/Main.class
  app.jar
```

#### 示例 2：多文件项目构建 JAR

```bash
asu-java-build -main com.example.Main -jarname myapp.jar
```

#### 示例 3：构建 Linux 原生镜像（通过 Docker）

```bash
asu-java-build -main com.example.Main -native

# 产物：bin/app-linux-amd64
# 在 WSL2/Linux 中运行：
chmod +x bin/app-linux-amd64
./bin/app-linux-amd64
```

#### 示例 4：使用自定义 Docker 镜像和参数

```bash
asu-java-build -main com.example.Main -native \
  -docker-image ghcr.io/graalvm/native-image-community:21-ol9-20250121 \
  -native-name myapp \
  -native-args "-Ob --no-fallback -march=native"
```

#### 示例 5：本地构建原生镜像

```bash
# 需要先安装 GraalVM 并将 native-image 添加到 PATH
asu-java-build -main com.example.Main -native -native-driver local

# Windows 产物：bin/app-windows-amd64.exe
# Linux 产物：bin/app-linux-amd64
```

#### 示例 6：使用现有 JAR 构建原生镜像

```bash
# 假设你用 Maven 构建了 target/myapp.jar
asu-java-build -jar target/myapp.jar -native
```

#### 示例 7：构建 Fat JAR（包含所有依赖）

```bash
# 项目结构
lib/
  commons-lang3-3.12.0.jar
  gson-2.8.9.jar
src/
  com/example/Main.java

# 构建 Fat JAR（所有依赖打包在 JAR 内）
asu-java-build -main com.example.Main -static

# 产物：bin/app.jar（包含所有依赖类）
```

#### 示例 8：导出依赖到 lib 目录

```bash
# 构建普通 JAR 并将依赖复制到 lib/ 目录
asu-java-build -main com.example.Main -explib

# 产物：
# bin/app.jar（MANIFEST.MF 包含 Class-Path: lib/*.jar）
# bin/lib/commons-lang3-3.12.0.jar
# bin/lib/gson-2.8.9.jar
```

#### 示例 9：使用 dep.lib 文件管理依赖

```bash
# 创建 dep.lib 文件（项目根目录）
# dep.lib 内容（支持 Maven 坐标格式）：
# com.google.code.gson:gson:2.10.1
# org.apache.commons:commons-lang3:3.12.0

# 构建（自动解析 Maven 坐标并下载缺失的依赖）
asu-java-build -main com.example.Main

# 或构建 Fat JAR
asu-java-build -main com.example.Main -static
```

#### 示例 10：混合使用多种依赖来源

```bash
# dep.lib 内容：
# # Maven 坐标（自动下载）
# com.google.code.gson:gson:2.10.1
# 
# # 本地绝对路径
# C:\custom-libs\my-lib.jar
# 
# # 相对路径
# libs/local-dep.jar

# lib/ 目录：
# lib/
#   another-dep.jar

# 构建时会自动合并所有来源的依赖
asu-java-build -main com.example.Main -static
```

### IDE 项目生成示例

#### 示例 11：快速创建 VS Code Java 项目

```bash
# 创建项目目录
mkdir my-project
cd my-project

# 生成 VS Code Java 项目
asu-java-build -vsc

# 项目结构：
# .vscode/
#   settings.json          # Java 配置（源码路径、输出路径等）
# src/
#   Main.java              # 示例 Hello World 代码
# .gitignore               # 排除 .vscode/、bin/ 等
# .git/                    # Git 仓库

# 用 VS Code 打开
code .
```

#### 示例 12：生成 Eclipse 项目

```bash
mkdir my-eclipse-project
cd my-eclipse-project
asu-java-build -eclipse

# 项目结构：
# .project                 # Eclipse 项目配置
# .classpath               # Eclipse classpath 配置
# src/                     # 源码目录
# bin/                     # 输出目录
# .gitignore
# .git/

# 在 Eclipse 中导入：File → Import → Existing Projects into Workspace
```

#### 示例 13：生成 IntelliJ IDEA 项目

```bash
mkdir my-idea-project
cd my-idea-project
asu-java-build -idea

# 项目结构：
# .idea/                   # IDEA 配置目录
#   modules.xml
#   misc.xml
# my-idea-project.iml      # IDEA 模块文件
# src/                     # 源码目录
# bin/                     # 输出目录
# .gitignore
# .git/

# 用 IntelliJ IDEA 打开：File → Open → 选择项目目录
```

#### 示例 14：生成 Maven 项目

```bash
# 生成 Maven 项目（交互式）
asu-java-build -maven

# 会提示输入：
# - groupId (默认: me.asu)
# - artifactId (默认: 当前目录名)
# - version (默认: 1.0.0-SNAPSHOT)
# - Java version (8/17/21, 默认: 21)

# 项目结构：
# pom.xml                  # Maven 配置
# src/
#   main/
#     java/
#       <package>/
#         App.java         # Maven archetype 生成的主类
#   test/
#     java/
#       <package>/
#         AppTest.java
# .gitignore
# .git/

# 构建和运行
mvn compile
mvn exec:java
```

#### 示例 15：生成 Spring Boot 项目

```bash
# 生成 Spring Boot 项目（交互式）
asu-java-build -spring

# 会提示输入：
# - groupId (默认: com.example)
# - artifactId (默认: 当前目录名)
# - version (默认: 1.0.0-SNAPSHOT)
# - Java version (8/17/21, 默认: 17)
# - package name (默认: groupId)
# - dependencies (默认: web，可用: web,lombok,actuator,validation,websocket,cache,data-jpa,security)

# 项目结构：
# pom.xml                  # Spring Boot Maven 配置
# src/
#   main/
#     java/
#       <package>/
#         Application.java # Spring Boot 主类
#     resources/
#       application.properties
#   test/
# .gitignore
# .git/

# 构建和运行
mvn spring-boot:run
```

### Maven 项目构建示例

#### 示例 16：Maven 项目 - 普通 JAR

```bash
cd my-maven-project
asu-java-build -main com.example.Main

# 自动执行：
# 1. 检测 pom.xml
# 2. 运行 mvn compile -DskipTests
# 3. 从 target/classes 打包 JAR
# 产物：bin/app.jar（不包含依赖，但包含 META-INF/dependencies.lib）
```

#### 示例 17：Maven 项目 - 导出依赖到 lib/

```bash
cd my-maven-project
asu-java-build -main com.example.Main -explib

# 自动执行：
# 1. mvn compile -DskipTests
# 2. mvn dependency:copy-dependencies（复制到 target/lib）
# 3. 将依赖复制到 bin/lib/
# 4. 打包 JAR（MANIFEST.MF 包含 Class-Path）
# 产物：bin/app.jar + bin/lib/*.jar
```

#### 示例 18：Maven 项目 - Fat JAR

```bash
cd my-maven-project
asu-java-build -main com.example.Main -static

# 自动执行：
# 1. mvn compile -DskipTests
# 2. mvn dependency:copy-dependencies
# 3. 合并所有依赖到单个 JAR
# 产物：bin/app.jar（包含所有依赖）
```

#### 示例 19：Maven 项目 - 构建原生镜像

```bash
cd my-maven-project
asu-java-build -main com.example.Main -static -native

# 先构建 Fat JAR，再构建原生镜像
# 产物：bin/app-linux-amd64
```

#### 示例 20：从 pom.xml 自动读取主类

```xml
<!-- pom.xml 配置 -->
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-jar-plugin</artifactId>
            <configuration>
                <archive>
                    <manifest>
                        <mainClass>com.example.Main</mainClass>
                    </manifest>
                </archive>
            </configuration>
        </plugin>
    </plugins>
</build>
```

```bash
# 不需要指定 -main，工具会自动从 pom.xml 读取
asu-java-build -static
```

### Maven 项目详细说明

#### 自动检测

工具会自动检测项目根目录的 `pom.xml` 文件，如果存在则使用 Maven 构建流程。

#### Maven 命令选择

优先级顺序：
1. Maven Wrapper（`mvnw.cmd` 或 `mvnw`）
2. 系统 Maven（`mvn`）

#### 构建模式对比

| 模式 | 命令 | 输出 | 使用场景 |
|------|------|------|---------|
| **普通 JAR** | `asu-java-build -main com.example.Main` | `bin/app.jar`<br>`bin/lib/*.jar` | 开发测试 |
| **导出依赖** | `asu-java-build -main com.example.Main -explib` | `bin/app.jar`<br>`bin/lib/*.jar` | 生产部署（多文件） |
| **Fat JAR** | `asu-java-build -main com.example.Main -static` | `bin/app.jar` | 生产部署（单文件）<br>容器化 |
| **原生镜像** | `asu-java-build -main com.example.Main -static -native` | `bin/app-linux-amd64` | 无 JVM 部署<br>高性能场景 |

#### 普通 JAR 模式

```bash
asu-java-build -main com.example.Main
```

执行流程：
- `mvn compile -DskipTests` - 编译项目
- `mvn dependency:copy-dependencies` - 复制依赖到 target/lib
- 复制依赖到 bin/lib（与 JAR 同目录）
- 从 `target/classes` 打包 JAR，MANIFEST.MF 包含 `Class-Path: lib/*.jar`
- 输出：`bin/app.jar` + `bin/lib/*.jar`

运行方式：
```bash
java -jar bin/app.jar  # Class-Path 自动包含 lib/*.jar
```

#### 导出依赖模式（`-explib`）

```bash
asu-java-build -main com.example.Main -explib
```

执行流程与普通 JAR 模式相同，主要区别是会明确提示依赖已导出。

#### Fat JAR 模式（`-static`）

```bash
asu-java-build -main com.example.Main -static
```

执行流程：
- `mvn compile -DskipTests` - 编译项目
- `mvn dependency:copy-dependencies` - 复制依赖到 target/lib
- 合并所有依赖到单个 JAR（类似 Maven Shade Plugin）
- 自动合并 `META-INF/services/` 文件（SPI 支持）
- 输出：`bin/app.jar`（包含所有依赖）

优点：
- 单文件部署
- 无需外部依赖
- 适合容器化和原生镜像构建

#### 主类配置

**方法 1：命令行指定**

```bash
asu-java-build -main com.example.Main
```

**方法 2：从 pom.xml 读取**

在 `pom.xml` 中配置：

```xml
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-jar-plugin</artifactId>
            <version>3.3.0</version>
            <configuration>
                <archive>
                    <manifest>
                        <mainClass>com.example.Main</mainClass>
                    </manifest>
                </archive>
            </configuration>
        </plugin>
    </plugins>
</build>
```

然后直接运行：
```bash
asu-java-build -static  # 无需 -main 参数
```

#### 与标准 Maven 的区别

| 功能 | 标准 Maven | asu-java-build |
|------|-----------|----------------|
| 编译 | `mvn compile` | 自动调用 `mvn compile -DskipTests` |
| 打包 | `mvn package` | 自动打包到 `bin/` |
| Fat JAR | 需要配置 Shade Plugin | 使用 `-static` 参数 |
| 依赖导出 | 需要配置 Dependency Plugin | 自动复制到 `bin/lib/` |
| 原生镜像 | 需要配置 Native Maven Plugin | 使用 `-native` 参数 |
| 速度 | 完整生命周期 | 仅编译和打包（跳过测试）|

### 依赖管理

工具支持三种方式管理依赖：

#### 1. 扫描 `lib/` 目录

将 JAR 依赖放在项目根目录的 `lib/` 目录下（支持子目录）：

```
project/
├── lib/
│   ├── commons-lang3-3.12.0.jar
│   ├── gson-2.8.9.jar
│   └── external/
│       └── custom-lib.jar
├── src/
│   └── com/example/Main.java
```

工具会自动递归扫描 `lib/` 目录下的所有 `.jar` 文件。

#### 2. 使用 `dep.lib` 文件

在项目根目录创建 `dep.lib` 文件，支持三种格式：

**a. Maven 坐标格式（推荐）**

直接使用 Maven 坐标 `groupId:artifactId:version`，工具会自动：
- 查找本地 Maven 仓库（`$HOME/.m2/repository/`）
- 如果本地不存在，自动从 Maven 中央仓库下载

```
# dep.lib
commons-io:commons-io:2.11.0
com.google.code.gson:gson:2.10.1
org.apache.commons:commons-lang3:3.12.0
```

**b. 绝对路径（Windows/Linux 通用）**

```
# dep.lib

# Windows 路径（支持反斜杠和正斜杠）
C:\libs\commons-lang3-3.12.0.jar
D:\projects\libs\gson-2.8.9.jar
C:/libs/another-lib.jar

# Linux/Mac 路径
/home/user/libs/custom-lib.jar
/usr/local/share/java/my-lib.jar
```

**c. 相对路径（跨平台）**

```
# dep.lib
libs/commons-lang3-3.12.0.jar
../shared-libs/custom.jar
```

**路径兼容性说明：**
- ✅ 自动识别 Windows 绝对路径（`C:\` 或 `D:\` 开头）
- ✅ 自动识别 Linux/Mac 绝对路径（`/` 开头）
- ✅ 支持 Windows 正斜杠路径（`C:/path/to/file.jar`）
- ✅ 支持相对路径（`libs/file.jar`、`../shared/file.jar`）
- ✅ 智能区分 Maven 坐标和文件路径

支持注释（以 `#` 开头的行会被忽略）：

```
# dep.lib - 项目依赖列表

# JSON 处理
com.google.code.gson:gson:2.10.1

# 工具类
org.apache.commons:commons-lang3:3.12.0
```

#### 3. 依赖处理模式

- **普通 JAR**（默认）：在 `MANIFEST.MF` 中添加 `Class-Path`，指向 `lib/` 目录
- **Fat JAR** (`-static`)：将所有依赖解压并合并到一个 JAR 中
  - 自动合并 `META-INF/services/` 文件（支持 Java SPI）
  - 生成 `META-INF/dependencies-static.lib` 记录依赖来源
- **导出模式** (`-explib`)：将依赖复制到输出目录的 `lib/` 文件夹
  - 生成 `META-INF/dependencies.lib` 记录依赖列表

### 项目结构

```
asu-java-build/
├── main.go              # 主程序
├── go.mod               # Go 模块定义
├── build.bat            # Windows 构建脚本
├── README.md            # 本文档
└── simple-standalone/   # 示例 Java 项目
    ├── src/
    │   └── com/
    │       └── example/
    │           └── Main.java
    └── bin/             # 编译输出目录
```

### 与 Maven 结合使用

```bash
# 1. 使用 Maven 构建 JAR
mvn clean package

# 2. 使用本工具构建原生镜像
asu-java-build -jar target/myapp.jar -native
```

### 与 GraalVM Agent 配合

如果需要生成反射/资源配置：

```bash
# 1. 在 Windows/Mac 上使用 GraalVM Agent 运行
java -agentlib:native-image-agent=config-output-dir=src/main/resources/META-INF/native-image \
  -jar app.jar

# 2. 配置文件会生成在 src/main/resources/META-INF/native-image/
# 3. 使用本工具构建原生镜像时会自动拾取这些配置
asu-java-build -jar app.jar -native
```

### 故障排查

#### 问题：`javac` 或 `jar` 命令未找到
**解决方案**：确保 JDK 的 `bin` 目录在系统 PATH 中

```bash
# Windows
set PATH=%JAVA_HOME%\bin;%PATH%

# Linux/Mac
export PATH=$JAVA_HOME/bin:$PATH
```

#### 问题：Docker 构建失败
**解决方案**：
- 确保 Docker Desktop 正在运行
- 检查网络连接（需要拉取 Docker 镜像）
- 尝试手动拉取镜像：`docker pull ghcr.io/graalvm/native-image-community:21-ol9`

#### 问题：本地原生构建失败
**解决方案**：
- 确保已安装 GraalVM
- 安装 `native-image` 组件：`gu install native-image`
- 验证：`native-image --version`

#### 问题：Maven 依赖下载失败
**解决方案**：
- 检查网络连接（需要访问 Maven 中央仓库：https://repo1.maven.org）
- 验证 Maven 坐标格式是否正确：`groupId:artifactId:version`
- 检查依赖是否存在于 Maven 中央仓库
- 如果使用代理，确保代理设置正确
- 手动下载后放到：`$HOME/.m2/repository/groupId/artifactId/version/`

#### 问题：依赖冲突或类找不到
**解决方案**：
- 使用 `-static` 构建 Fat JAR 以包含所有依赖
- 检查 `dep.lib` 和 `lib/` 目录中的依赖版本
- 确保所有传递依赖都已包含

#### 问题：Maven 项目生成失败
**解决方案**：
- 确保 Maven 已安装并在 PATH 中
- 检查网络连接（Maven 从 Maven Central 下载 archetype）
- 验证 Maven 版本：`mvn --version`

#### 问题：Spring Boot 项目生成失败
**解决方案**：
- 确保已安装 Spring CLI：`spring --version`
- 安装 Spring CLI：从 https://docs.spring.io/spring-boot/docs/current/reference/html/getting-started.html#getting-started.installing.cli 下载
- 检查网络连接（Spring initializer 需要互联网访问）

#### 问题：Git 初始化失败
**解决方案**：
- 确保已安装 Git：`git --version`
- 如果不需要 Git，工具仍会成功创建项目结构
- 从 https://git-scm.com/downloads 安装 Git

### 许可证

本项目使用的许可证请查看项目根目录。

---

## English

### Introduction

`asu-java-build` is a multi-purpose Java project build and management tool written in Go. It can:

- ✅ Compile Java source code (with recursive source file scanning)
- ✅ **Automatically detect and build Maven projects**
- ✅ Package runnable JAR files (Normal/Fat JAR/Export Dependencies)
- ✅ Build native binaries using GraalVM Native Image
- ✅ Support Docker-based Linux native image building (cross-platform)
- ✅ Support local GraalVM native image building
- ✅ **Quickly generate IDE projects** (VS Code/Eclipse/IntelliJ IDEA/Maven/Spring Boot)

### Features

- **Zero-config Main Class Inference**: Automatically infers main class for single-file projects
- **Maven Project Support**:
  - Automatically detects `pom.xml` and uses Maven build
  - Prioritizes Maven Wrapper (`mvnw`/`mvnw.cmd`)
  - Extracts main class configuration from `pom.xml`
  - Manages dependencies using Maven plugins
  - Automatically copies dependencies to output directory
- **Flexible Build Modes**: Build from source code or existing JAR files
- **Dependency Management**:
  - Automatically scans `lib/` directory and `dep.lib` file for dependencies
  - Supports Fat JAR builds (similar to Maven Shade Plugin)
  - Supports exporting dependencies to `lib/` directory
  - Automatically merges `META-INF/services/` files (SPI support)
  - Maven coordinate auto-resolution and download (Maven Central)
- **Cross-platform Native Builds**: Build Linux native images on Windows/Mac via Docker
- **Local Native Builds**: Support building native images using local GraalVM
- **IDE Project Generation**:
  - VS Code Java projects (with settings.json and sample code)
  - Eclipse projects (.project and .classpath)
  - IntelliJ IDEA projects (.idea and .iml)
  - Maven projects (via maven-archetype-quickstart)
  - Spring Boot projects (via spring initializr CLI)
  - Auto-initialize Git repository and .gitignore
- **Custom Arguments**: Pass extra arguments to `native-image`

### Prerequisites

#### Basic Build (Java Compilation + JAR)
- JDK 8 or higher (requires `javac` and `jar` commands in PATH)

#### Docker Native Build
- Docker Desktop (for `-native -native-driver docker`)

#### Local Native Build
- GraalVM (requires `native-image` component installed)

### Quick Start

#### 1. Build the Tool Itself

```bash
# Windows
go build -o asu-java-build.exe .

# Linux/Mac
go build -o asu-java-build .

# Cross-compile (build Linux version on Windows)
set GOOS=linux
set GOARCH=amd64
go build -o asu-java-build-linux .
```

#### 2. Basic Usage

```bash
# Scenario A: Quickly generate projects
asu-java-build -vsc              # Generate VS Code Java project
asu-java-build -idea             # Generate IntelliJ IDEA project
asu-java-build -maven            # Generate Maven project
asu-java-build -spring           # Generate Spring Boot project

# Scenario B: Single-file project (auto-infer main class)
asu-java-build

# Scenario C: Multi-file project (specify main class)
asu-java-build -main com.example.Main

# Scenario D: Build native image from existing JAR
asu-java-build -jar target/myapp.jar -native

# Scenario E: Build native image from source (Linux via Docker)
asu-java-build -main com.example.Main -native

# Scenario F: Local native build (requires GraalVM installed)
asu-java-build -main com.example.Main -native -native-driver local
```

### Command Line Arguments

#### IDE Project Generation Arguments
| Flag | Description |
|------|-------------|
| `-vsc` | Generate VS Code Java project (with settings.json and sample Main.java) |
| `-eclipse` | Generate Eclipse project (with .project and .classpath) |
| `-idea` | Generate IntelliJ IDEA project (with .idea directory and .iml file) |
| `-maven` | Generate Maven project using maven-archetype-quickstart |
| `-spring` | Generate Spring Boot project using spring CLI |

**Note**: All IDE project generation will automatically:
- Initialize Git repository (if not already initialized)
- Create `.gitignore` file (excluding IDE configs, build outputs, etc.)
- Create basic directory structure

#### Basic Compilation Arguments
| Flag | Default | Description |
|------|---------|-------------|
| `-src` | `src` | Java source directory (recursive scan) |
| `-out` | `bin` | Output directory (.class/JAR/native) |
| `-main` | Auto-inferred | Main class FQCN (e.g. `com.example.Main`) |
| `-jarname` | `app.jar` | Runnable JAR file name |
| `-nojar` | `false` | Do not create JAR (disables native build) |

#### Dependency Management Arguments
| Flag | Default | Description |
|------|---------|-------------|
| `-static` | `false` | Package dependencies inside JAR (Fat JAR, like Maven Shade Plugin) |
| `-explib` | `false` | Export dependencies to `lib/` folder next to the JAR |

#### Using Existing JAR
| Flag | Default | Description |
|------|---------|-------------|
| `-jar` | Empty | Use existing JAR file (skip javac/jar phase) |

#### Native Build Arguments
| Flag | Default | Description |
|------|---------|-------------|
| `-native` | `false` | Build native image after JAR is ready |
| `-native-driver` | `docker` | Build method: `docker` or `local` |
| `-native-target` | `linux` | Target OS for Docker build: `linux` or `windows` |
| `-native-name` | `app-linux-amd64` | Native binary name |
| `-native-args` | Empty | Extra arguments to pass to `native-image` |
| `-docker-image` | `ghcr.io/graalvm/native-image-community:21-ol9` | Docker image |

### Usage Examples

#### Example 1: Compile Single Java File

```bash
# Project structure
src/
  com/
    example/
      Main.java

# Build
asu-java-build

# Output
bin/
  com/example/Main.class
  app.jar
```

#### Example 2: Multi-file Project Build JAR

```bash
asu-java-build -main com.example.Main -jarname myapp.jar
```

#### Example 3: Build Linux Native Image (via Docker)

```bash
asu-java-build -main com.example.Main -native

# Output: bin/app-linux-amd64
# Run in WSL2/Linux:
chmod +x bin/app-linux-amd64
./bin/app-linux-amd64
```

#### Example 4: Custom Docker Image and Arguments

```bash
asu-java-build -main com.example.Main -native \
  -docker-image ghcr.io/graalvm/native-image-community:21-ol9-20250121 \
  -native-name myapp \
  -native-args "-Ob --no-fallback -march=native"
```

#### Example 5: Local Native Build

```bash
# Requires GraalVM installed with native-image in PATH
asu-java-build -main com.example.Main -native -native-driver local

# Windows output: bin/app-windows-amd64.exe
# Linux output: bin/app-linux-amd64
```

#### Example 6: Build Native Image from Existing JAR

```bash
# Assuming you built target/myapp.jar with Maven
asu-java-build -jar target/myapp.jar -native
```

#### Example 7: Build Fat JAR (with all dependencies)

```bash
# Project structure
lib/
  commons-lang3-3.12.0.jar
  gson-2.8.9.jar
src/
  com/example/Main.java

# Build Fat JAR (all dependencies packaged inside JAR)
asu-java-build -main com.example.Main -static

# Output: bin/app.jar (contains all dependency classes)
```

#### Example 8: Export Dependencies to lib Directory

```bash
# Build normal JAR and copy dependencies to lib/ directory
asu-java-build -main com.example.Main -explib

# Output:
# bin/app.jar (MANIFEST.MF contains Class-Path: lib/*.jar)
# bin/lib/commons-lang3-3.12.0.jar
# bin/lib/gson-2.8.9.jar
```

#### Example 9: Manage Dependencies with dep.lib File

```bash
# Create dep.lib file (in project root)
# dep.lib content (supports Maven coordinate format):
# com.google.code.gson:gson:2.10.1
# org.apache.commons:commons-lang3:3.12.0

# Build (automatically resolves Maven coordinates and downloads missing dependencies)
asu-java-build -main com.example.Main

# Or build Fat JAR
asu-java-build -main com.example.Main -static
```

#### Example 10: Mix Multiple Dependency Sources

```bash
# dep.lib content:
# # Maven coordinates (auto-download)
# com.google.code.gson:gson:2.10.1
# 
# # Local absolute path
# C:\custom-libs\my-lib.jar
# 
# # Relative path
# libs/local-dep.jar

# lib/ directory:
# lib/
#   another-dep.jar

# All dependency sources are automatically merged during build
asu-java-build -main com.example.Main -static
```

### IDE Project Generation Examples

#### Example 11: Quickly Create VS Code Java Project

```bash
# Create project directory
mkdir my-project
cd my-project

# Generate VS Code Java project
asu-java-build -vsc

# Project structure:
# .vscode/
#   settings.json          # Java configuration (source paths, output paths, etc.)
# src/
#   Main.java              # Sample Hello World code
# .gitignore               # Excludes .vscode/, bin/, etc.
# .git/                    # Git repository

# Open with VS Code
code .
```

#### Example 12: Generate Eclipse Project

```bash
mkdir my-eclipse-project
cd my-eclipse-project
asu-java-build -eclipse

# Project structure:
# .project                 # Eclipse project configuration
# .classpath               # Eclipse classpath configuration
# src/                     # Source directory
# bin/                     # Output directory
# .gitignore
# .git/

# Import in Eclipse: File → Import → Existing Projects into Workspace
```

#### Example 13: Generate IntelliJ IDEA Project

```bash
mkdir my-idea-project
cd my-idea-project
asu-java-build -idea

# Project structure:
# .idea/                   # IDEA configuration directory
#   modules.xml
#   misc.xml
# my-idea-project.iml      # IDEA module file
# src/                     # Source directory
# bin/                     # Output directory
# .gitignore
# .git/

# Open with IntelliJ IDEA: File → Open → Select project directory
```

#### Example 14: Generate Maven Project

```bash
# Generate Maven project (interactive)
asu-java-build -maven

# Prompts for:
# - groupId (default: me.asu)
# - artifactId (default: current directory name)
# - version (default: 1.0.0-SNAPSHOT)
# - Java version (8/17/21, default: 21)

# Project structure:
# pom.xml                  # Maven configuration
# src/
#   main/
#     java/
#       <package>/
#         App.java         # Main class generated by Maven archetype
#   test/
#     java/
#       <package>/
#         AppTest.java
# .gitignore
# .git/

# Build and run
mvn compile
mvn exec:java
```

#### Example 15: Generate Spring Boot Project

```bash
# Generate Spring Boot project (interactive)
asu-java-build -spring

# Prompts for:
# - groupId (default: com.example)
# - artifactId (default: current directory name)
# - version (default: 1.0.0-SNAPSHOT)
# - Java version (8/17/21, default: 17)
# - package name (default: groupId)
# - dependencies (default: web, available: web,lombok,actuator,validation,websocket,cache,data-jpa,security)

# Project structure:
# pom.xml                  # Spring Boot Maven configuration
# src/
#   main/
#     java/
#       <package>/
#         Application.java # Spring Boot main class
#     resources/
#       application.properties
#   test/
# .gitignore
# .git/

# Build and run
mvn spring-boot:run
```

### Maven Project Build Examples

#### Example 16: Maven Project - Regular JAR

```bash
cd my-maven-project
asu-java-build -main com.example.Main

# Automatically executes:
# 1. Detect pom.xml
# 2. Run mvn compile -DskipTests
# 3. Package JAR from target/classes
# Output: bin/app.jar + bin/lib/*.jar
```

#### Example 17: Maven Project - Export Dependencies to lib/

```bash
cd my-maven-project
asu-java-build -main com.example.Main -explib

# Automatically executes:
# 1. mvn compile -DskipTests
# 2. mvn dependency:copy-dependencies (to target/lib)
# 3. Copy dependencies to bin/lib/
# 4. Package JAR (MANIFEST.MF contains Class-Path)
# Output: bin/app.jar + bin/lib/*.jar
```

#### Example 18: Maven Project - Fat JAR

```bash
cd my-maven-project
asu-java-build -main com.example.Main -static

# Automatically executes:
# 1. mvn compile -DskipTests
# 2. mvn dependency:copy-dependencies
# 3. Merge all dependencies into single JAR
# Output: bin/app.jar (includes all dependencies)
```

#### Example 19: Maven Project - Build Native Image

```bash
cd my-maven-project
asu-java-build -main com.example.Main -static -native

# Build Fat JAR first, then native image
# Output: bin/app-linux-amd64
```

#### Example 20: Auto-read Main Class from pom.xml

```xml
<!-- pom.xml configuration -->
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-jar-plugin</artifactId>
            <configuration>
                <archive>
                    <manifest>
                        <mainClass>com.example.Main</mainClass>
                    </manifest>
                </archive>
            </configuration>
        </plugin>
    </plugins>
</build>
```

```bash
# No need to specify -main, tool reads from pom.xml automatically
asu-java-build -static
```

### Maven Project Details

#### Auto-detection

The tool automatically detects `pom.xml` in the project root and uses Maven build process if found.

#### Maven Command Selection

Priority order:
1. Maven Wrapper (`mvnw.cmd` or `mvnw`)
2. System Maven (`mvn`)

#### Build Mode Comparison

| Mode | Command | Output | Use Case |
|------|---------|--------|----------|
| **Regular JAR** | `asu-java-build -main com.example.Main` | `bin/app.jar`<br>`bin/lib/*.jar` | Development/Testing |
| **Export Dependencies** | `asu-java-build -main com.example.Main -explib` | `bin/app.jar`<br>`bin/lib/*.jar` | Production (multi-file) |
| **Fat JAR** | `asu-java-build -main com.example.Main -static` | `bin/app.jar` | Production (single file)<br>Containerization |
| **Native Image** | `asu-java-build -main com.example.Main -static -native` | `bin/app-linux-amd64` | No JVM deployment<br>High performance |

#### Regular JAR Mode

```bash
asu-java-build -main com.example.Main
```

Execution flow:
- `mvn compile -DskipTests` - Compile project
- `mvn dependency:copy-dependencies` - Copy dependencies to target/lib
- Copy dependencies to bin/lib (same directory as JAR)
- Package JAR from `target/classes`, MANIFEST.MF includes `Class-Path: lib/*.jar`
- Output: `bin/app.jar` + `bin/lib/*.jar`

Run:
```bash
java -jar bin/app.jar  # Class-Path automatically includes lib/*.jar
```

#### Export Dependencies Mode (`-explib`)

```bash
asu-java-build -main com.example.Main -explib
```

Execution flow is the same as Regular JAR mode, with explicit confirmation that dependencies are exported.

#### Fat JAR Mode (`-static`)

```bash
asu-java-build -main com.example.Main -static
```

Execution flow:
- `mvn compile -DskipTests` - Compile project
- `mvn dependency:copy-dependencies` - Copy dependencies to target/lib
- Merge all dependencies into single JAR (similar to Maven Shade Plugin)
- Automatically merge `META-INF/services/` files (SPI support)
- Output: `bin/app.jar` (includes all dependencies)

Benefits:
- Single-file deployment
- No external dependencies required
- Perfect for containerization and native image builds

#### Main Class Configuration

**Method 1: Command Line**

```bash
asu-java-build -main com.example.Main
```

**Method 2: Read from pom.xml**

Configure in `pom.xml`:

```xml
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-jar-plugin</artifactId>
            <version>3.3.0</version>
            <configuration>
                <archive>
                    <manifest>
                        <mainClass>com.example.Main</mainClass>
                    </manifest>
                </archive>
            </configuration>
        </plugin>
    </plugins>
</build>
```

Then run:
```bash
asu-java-build -static  # No -main flag needed
```

#### Comparison with Standard Maven

| Feature | Standard Maven | asu-java-build |
|---------|---------------|----------------|
| Compile | `mvn compile` | Auto-runs `mvn compile -DskipTests` |
| Package | `mvn package` | Auto-packages to `bin/` |
| Fat JAR | Requires Shade Plugin config | Use `-static` flag |
| Export Dependencies | Requires Dependency Plugin config | Auto-copies to `bin/lib/` |
| Native Image | Requires Native Maven Plugin config | Use `-native` flag |
| Speed | Full lifecycle | Compile and package only (skip tests) |

### Dependency Management

The tool supports three ways to manage dependencies:

#### 1. Scan `lib/` Directory

Place JAR dependencies in the `lib/` directory at the project root (subdirectories supported):

```
project/
├── lib/
│   ├── commons-lang3-3.12.0.jar
│   ├── gson-2.8.9.jar
│   └── external/
│       └── custom-lib.jar
├── src/
│   └── com/example/Main.java
```

The tool will automatically scan all `.jar` files recursively in the `lib/` directory.

#### 2. Use `dep.lib` File

Create a `dep.lib` file in the project root with three supported formats:

**a. Maven Coordinate Format (Recommended)**

Use Maven coordinates `groupId:artifactId:version` directly. The tool will automatically:
- Search the local Maven repository (`$HOME/.m2/repository/`)
- Download from Maven Central if not found locally

```
# dep.lib
commons-io:commons-io:2.11.0
com.google.code.gson:gson:2.10.1
org.apache.commons:commons-lang3:3.12.0
```

**b. Absolute Paths (Windows/Linux Compatible)**

```
# dep.lib

# Windows paths (supports both backslash and forward slash)
C:\libs\commons-lang3-3.12.0.jar
D:\projects\libs\gson-2.8.9.jar
C:/libs/another-lib.jar

# Linux/Mac paths
/home/user/libs/custom-lib.jar
/usr/local/share/java/my-lib.jar
```

**c. Relative Paths (Cross-platform)**

```
# dep.lib
libs/commons-lang3-3.12.0.jar
../shared-libs/custom.jar
```

**Path Compatibility:**
- ✅ Automatically recognizes Windows absolute paths (starting with `C:\` or `D:\`)
- ✅ Automatically recognizes Linux/Mac absolute paths (starting with `/`)
- ✅ Supports Windows forward slash paths (`C:/path/to/file.jar`)
- ✅ Supports relative paths (`libs/file.jar`, `../shared/file.jar`)
- ✅ Intelligently distinguishes between Maven coordinates and file paths

Comments are supported (lines starting with `#` are ignored):

```
# dep.lib - Project Dependencies

# JSON processing
com.google.code.gson:gson:2.10.1

# Utilities
org.apache.commons:commons-lang3:3.12.0
```

#### 3. Dependency Processing Modes

- **Normal JAR** (default): Adds `Class-Path` to `MANIFEST.MF` pointing to `lib/` directory
- **Fat JAR** (`-static`): Extracts and merges all dependencies into one JAR
  - Automatically merges `META-INF/services/` files (Java SPI support)
  - Generates `META-INF/dependencies-static.lib` to record dependency sources
- **Export Mode** (`-explib`): Copies dependencies to `lib/` folder in output directory
  - Generates `META-INF/dependencies.lib` to record dependency list

### Project Structure

```
asu-java-build/
├── main.go              # Main program
├── go.mod               # Go module definition
├── build.bat            # Windows build script
├── README.md            # This document
└── simple-standalone/   # Example Java project
    ├── src/
    │   └── com/
    │       └── example/
    │           └── Main.java
    └── bin/             # Compilation output directory
```

### Integration with Maven

```bash
# 1. Build JAR with Maven
mvn clean package

# 2. Build native image with this tool
asu-java-build -jar target/myapp.jar -native
```

### Working with GraalVM Agent

To generate reflection/resource configurations:

```bash
# 1. Run with GraalVM Agent on Windows/Mac
java -agentlib:native-image-agent=config-output-dir=src/main/resources/META-INF/native-image \
  -jar app.jar

# 2. Configuration files will be generated in src/main/resources/META-INF/native-image/
# 3. These configurations will be automatically picked up when building native image
asu-java-build -jar app.jar -native
```

### Troubleshooting

#### Issue: `javac` or `jar` command not found
**Solution**: Ensure JDK's `bin` directory is in system PATH

```bash
# Windows
set PATH=%JAVA_HOME%\bin;%PATH%

# Linux/Mac
export PATH=$JAVA_HOME/bin:$PATH
```

#### Issue: Docker build fails
**Solution**:
- Ensure Docker Desktop is running
- Check network connection (required to pull Docker image)
- Try manually pulling the image: `docker pull ghcr.io/graalvm/native-image-community:21-ol9`

#### Issue: Local native build fails
**Solution**:
- Ensure GraalVM is installed
- Install `native-image` component: `gu install native-image`
- Verify: `native-image --version`

#### Issue: Maven dependency download fails
**Solution**:
- Check network connection (requires access to Maven Central: https://repo1.maven.org)
- Verify Maven coordinate format is correct: `groupId:artifactId:version`
- Check if dependency exists in Maven Central Repository
- If using a proxy, ensure proxy settings are correct
- Manually download and place in: `$HOME/.m2/repository/groupId/artifactId/version/`

#### Issue: Dependency conflicts or ClassNotFoundException
**Solution**:
- Use `-static` to build Fat JAR with all dependencies included
- Check dependency versions in `dep.lib` and `lib/` directory
- Ensure all transitive dependencies are included

#### Issue: Maven project generation fails
**Solution**:
- Ensure Maven is installed and in PATH
- Check network connection (Maven downloads archetypes from Maven Central)
- Verify Maven version: `mvn --version`

#### Issue: Spring Boot project generation fails
**Solution**:
- Ensure Spring CLI is installed: `spring --version`
- Install Spring CLI: Download from https://docs.spring.io/spring-boot/docs/current/reference/html/getting-started.html#getting-started.installing.cli
- Check network connection (Spring initializer requires internet access)

#### Issue: Git initialization fails
**Solution**:
- Ensure Git is installed: `git --version`
- If Git is not needed, the tool will still create the project structure successfully
- Install Git from: https://git-scm.com/downloads

### License

Please check the project root directory for license information.
