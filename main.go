package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bruceasu/asu-java-build/pkg"
	"github.com/bruceasu/asu-java-build/pkg/ide"
)

var (
	// 基本编译参数
	srcDir   = flag.String("src", "src", "Java source directory (recursive)")
	outDir   = flag.String("out", "bin", "Output directory for .class/JAR/native")
	mainFQCN = flag.String("main", "", "Main class FQCN (e.g. com.example.Main). If omitted and only one source exists, it will be inferred.")
	jarName  = flag.String("jarname", "app.jar", "Runnable JAR name when main class is known")
	// 依赖管理
	staticMode = flag.Bool("static", false, "Package dependencies inside JAR (fat-jar, like shade plugin)")
	expLib     = flag.Bool("explib", false, "Export dependencies to lib/ folder next to the JAR")
	skipJar    = flag.Bool("nojar", false, "Do not create JAR (will disable native build)")
	// 直接使用已有 JAR 做 native（跳过 javac/jar 阶段）
	inputJar = flag.String("jar", "", "Use an existing JAR instead of compiling sources (skip javac/jar)")
	// Native 构建开关与模式
	nativeEnable = flag.Bool("native", false, "Build a native image after JAR is ready")
	nativeDriver = flag.String("native-driver", "docker", "native builder: docker | local")
	nativeTarget = flag.String("native-target", "linux", "target OS for native when using docker: linux | windows (windows image rarely used)")
	nativeName   = flag.String("native-name", "", "native binary name (default: app-<os>-amd64)")
	nativeArgs   = flag.String("native-args", "", "extra args to pass to native-image (e.g. '-Ob -march=native')")
	dockerImage  = flag.String("docker-image", "ghcr.io/graalvm/native-image-community:21-ol9", "Docker image for linux native build")
	// IDE 项目设置生成
	vscFlag     = flag.String("vsc", "", "Generate VS Code Java project in specified folder")
	eclipseFlag = flag.String("eclipse", "", "Generate Eclipse Java project in specified folder")
	ideaFlag    = flag.String("idea", "", "Generate IntelliJ IDEA Java project in specified folder")
	mavenFlag   = flag.String("maven", "", "Generate Maven Java project in specified folder")
	springFlag  = flag.String("spring", "", "Generate Spring Boot Java project in specified folder")
)

// ---------- 工具函数 ----------

// 这些常用函数的别名,方便 main.go 内部使用
var (
	info    = pkg.Info
	fail    = pkg.Fail
	mustAbs = pkg.MustAbs
)

func main() {
	flag.Parse()

	// IDE 项目生成
	if *vscFlag != "" {
		if err := ide.GenerateVSC(*vscFlag); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}
	if *eclipseFlag != "" {
		if err := ide.GenerateEclipse(*eclipseFlag); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}
	if *ideaFlag != "" {
		if err := ide.GenerateIdea(*ideaFlag); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}
	if *mavenFlag != "" {
		if err := ide.GenerateMaven(*mavenFlag, pkg.FindMavenCommand); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}
	if *springFlag != "" {
		if err := ide.GenerateSpring(*springFlag); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}

	var jarPath string

	// 检测是否为 Maven 项目
	isMavenProject := pkg.DetectMavenProject()

	if strings.TrimSpace(*inputJar) != "" {
		// 直接使用外部 JAR
		jarPath = mustAbs(*inputJar)
		// native-image 可从 JAR manifest 读 Main-Class；留空也可
		info("[JAR] Using existing JAR: %s", jarPath)
		if *nativeEnable && *skipJar {
			fail(errors.New("-nojar conflicts with -jar input; remove -nojar when using -jar"))
		}
	} else if isMavenProject {
		// Maven 项目构建流程
		var err error
		var mainClass string
		jarPath, mainClass, err = pkg.BuildMavenProject(
			*mainFQCN, *outDir, *jarName,
			*skipJar, *staticMode, *expLib,
		)
		if err != nil {
			fail(err)
		}
		_ = mainClass // mainClass may be used for native build or left empty
	} else {
		// 正常流程：收集源码 -> 编译 -> 打 JAR
		var err error
		var mainClass string
		jarPath, mainClass, err = pkg.BuildJavaProject(
			*srcDir, *outDir, *jarName, *mainFQCN,
			*skipJar, *staticMode, *expLib,
		)
		if err != nil {
			fail(err)
		}
		_ = mainClass // mainClass may be used for native build or left empty
	}

	// native 构建
	if *nativeEnable {
		if err := pkg.BuildNative(jarPath, *outDir, *nativeName, *nativeArgs, *dockerImage, *nativeDriver, *nativeTarget); err != nil {
			fail(err)
		}
	}
}

/*
Maven 项目用法示例

A. Maven 项目 - 普通 JAR
cd my-maven-project
asu-java-build -main com.example.Main

B. Maven 项目 - Fat JAR
cd my-maven-project
asu-java-build -main com.example.Main -static

C. Maven 项目 - 导出依赖到 lib/
cd my-maven-project
asu-java-build -main com.example.Main -explib

D. Maven 项目 - 原生镜像
cd my-maven-project
asu-java-build -main com.example.Main -static -native

说明 / 小贴士
	• 自动检测：工具会自动检测 pom.xml 并使用 Maven 构建流程
	• Maven Wrapper：优先使用 mvnw/mvnw.cmd，如不存在则使用系统 mvn
	• 依赖管理：
		○ 普通模式：不包含依赖（需要运行时提供）
		○ -explib：调用 Maven 插件复制依赖到 target/lib，然后复制到输出目录
		○ -static：先复制依赖，再合并成 Fat JAR
	• 主类推断：工具会尝试从 pom.xml 读取 <mainClass> 配置
*/

/*
用法示例
A. 只有一个源文件（自动推断入口）→ JAR
# 编译 Go 工具
go build -o jbuild .
# Java：src 下只有 Hello.java
./jbuild
# 产物：bin/*.class + bin/app.jar
B. 多源文件（指定入口）→ JAR
./jbuild -main com.example.Main
C. 直接用现成 JAR（跳过 javac/jar）→ Native (Linux via Docker)
# 比如你用 Maven 打好了 target/app.jar
./jbuild -jar target/app.jar -native
# 产物：bin/app-linux-amd64
D. 源码→JAR→Native（Linux via Docker，一步到位）
./jbuild -main com.example.Main -native
# 可加优化参数传给 native-image：
./jbuild -main com.example.Main -native -native-args "-Ob --no-fallback"
# 指定镜像或二进制名：
./jbuild -native -docker-image ghcr.io/graalvm/native-image-community:21-ol9-20250121 -native-name myapp
E. 本地构建 native（需要本机已安装 GraalVM + native-image）
# 生成当前平台二进制（Windows 变 .exe；Linux/Mac 生成本地可执行）
./jbuild -main com.example.Main -native -native-driver local

说明 / 小贴士
	• 跨平台产物：你可以用 GOOS/GOARCH 交叉编译这个 Go 工具本身，例如：
		○ GOOS=windows GOARCH=amd64 go build -o jbuild.exe .
		○ GOOS=linux GOARCH=amd64 go build -o jbuild-linux .
	• 与 GraalVM Agent 配合：照你之前的做法，在 Windows 上跑 Agent，把生成的配置落在 src/main/resources/META-INF/native-image/。本工具在容器里 native-image 时会自动拾取这些配置。
	• Maven 结合：如果你用 Maven 打包出 target/xxx.jar，用 -jar 指过去，直接走 native 即可；无需重复编译。
	• WSL2 测试：用 Docker 构建后，产物在 bin/app-linux-amd64（或你指定的名字）。在 WSL2：
cd /mnt/c/Users/<你>/path/to/project/bin
chmod +x app-linux-amd64
./app-linux-amd64

*/
