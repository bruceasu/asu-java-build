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
	srcDir   = flag.String("src", "src", "Java source directory (recursive)")
	outDir   = flag.String("out", "bin", "Output directory for .class/JAR/runtime app-image")
	mainFQCN = flag.String("main", "", "Main class FQCN (e.g. com.example.Main). If omitted and only one source exists, it will be inferred.")
	jarName  = flag.String("jarname", "app.jar", "Runnable JAR name when main class is known")

	staticMode = flag.Bool("static", false, "Package dependencies inside JAR (fat-jar, like shade plugin)")
	expLib     = flag.Bool("explib", false, "Export dependencies to lib/ folder next to the JAR")
	skipJar    = flag.Bool("nojar", false, "Do not create JAR (runtime packaging requires a JAR)")
	inputJar   = flag.String("jar", "", "Use an existing JAR instead of compiling sources (skip javac/jar)")

	runtimeEnable     = flag.Bool("runtime", false, "Package app-image runtime using jlink+jpackage (requires JDK 25)")
	runtimeAppName    = flag.String("runtime-app-name", "", "Application name for jpackage (default: JAR file name without .jar)")
	runtimeMainClass  = flag.String("runtime-main-class", "", "Main class for jpackage (default: use built/inferred main class, then JAR manifest)")
	runtimeImageDir   = flag.String("runtime-image-dir", "", "Output directory for jlink runtime image (default: <out>/runtime)")
	runtimeDistDir    = flag.String("runtime-dist-dir", "", "Output directory for jpackage app-image (default: <out>/dist)")
	runtimeJavaOpts   = flag.String("runtime-java-opts", "-Xms256m;-Xmx1g", "Semicolon-separated java options for jpackage")
	runtimeAddModules = flag.String("runtime-add-modules", "jdk.unsupported,java.naming", "Extra modules appended to jlink add-modules")
	runtimeWinConsole = flag.Bool("runtime-win-console", true, "Enable jpackage --win-console on Windows")

	vscFlag     = flag.String("vsc", "", "Generate VS Code Java project in specified folder")
	eclipseFlag = flag.String("eclipse", "", "Generate Eclipse Java project in specified folder")
	ideaFlag    = flag.String("idea", "", "Generate IntelliJ IDEA Java project in specified folder")
	mavenFlag   = flag.String("maven", "", "Generate Maven Java project in specified folder")
	springFlag  = flag.String("spring", "", "Generate Spring Boot Java project in specified folder")
)

var (
	info    = pkg.Info
	fail    = pkg.Fail
	mustAbs = pkg.MustAbs
)

func main() {
	flag.Parse()

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
	var resolvedMainClass string

	isMavenProject := pkg.DetectMavenProject()

	if strings.TrimSpace(*inputJar) != "" {
		jarPath = mustAbs(*inputJar)
		resolvedMainClass = strings.TrimSpace(*mainFQCN)
		info("[JAR] Using existing JAR: %s", jarPath)
	} else if isMavenProject {
		var err error
		jarPath, resolvedMainClass, err = pkg.BuildMavenProject(
			*mainFQCN, *outDir, *jarName,
			*skipJar, *staticMode, *expLib,
		)
		if err != nil {
			fail(err)
		}
	} else {
		var err error
		jarPath, resolvedMainClass, err = pkg.BuildJavaProject(
			*srcDir, *outDir, *jarName, *mainFQCN,
			*skipJar, *staticMode, *expLib,
		)
		if err != nil {
			fail(err)
		}
	}

	if *runtimeEnable {
		if strings.TrimSpace(jarPath) == "" {
			fail(errors.New("-runtime requires a JAR. Please provide -jar or remove -nojar"))
		}
		mainClassForRuntime := strings.TrimSpace(*runtimeMainClass)
		if mainClassForRuntime == "" {
			mainClassForRuntime = strings.TrimSpace(*mainFQCN)
		}
		if mainClassForRuntime == "" {
			mainClassForRuntime = resolvedMainClass
		}
		if err := pkg.BuildRuntimeAppImage(
			jarPath,
			*outDir,
			*runtimeAppName,
			mainClassForRuntime,
			*runtimeImageDir,
			*runtimeDistDir,
			*runtimeJavaOpts,
			*runtimeAddModules,
			*runtimeWinConsole,
		); err != nil {
			fail(err)
		}
	}
}
