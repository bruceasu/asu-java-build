package main
import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"archive/zip"
    "io"
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
	skipJar  = flag.Bool("nojar", false, "Do not create JAR (will disable native build)")
// 直接使用已有 JAR 做 native（跳过 javac/jar 阶段）
	inputJar = flag.String("jar", "", "Use an existing JAR instead of compiling sources (skip javac/jar)")
// Native 构建开关与模式
	nativeEnable = flag.Bool("native", false, "Build a native image after JAR is ready")
	nativeDriver = flag.String("native-driver", "docker", "native builder: docker | local")
	nativeTarget = flag.String("native-target", "linux", "target OS for native when using docker: linux | windows (windows image rarely used)")
	nativeName   = flag.String("native-name", "", "native binary name (default: app-<os>-amd64)")
	nativeArgs   = flag.String("native-args", "", "extra args to pass to native-image (e.g. '-Ob -march=native')")
	dockerImage  = flag.String("docker-image", "ghcr.io/graalvm/native-image-community:21-ol9", "Docker image for linux native build")
)
func main() {
	flag.Parse()
	if *nativeName == "" {
		if *nativeTarget == "windows" {
			*nativeName = "app-windows-amd64"
		} else {
			*nativeName = "app-linux-amd64"
		}
	}
    var jarPath string
	var mainClass string
    if strings.TrimSpace(*inputJar) != "" {
		// 直接使用外部 JAR
		jarPath = mustAbs(*inputJar)
		mainClass = strings.TrimSpace(*mainFQCN) // native-image 可从 JAR manifest 读 Main-Class；留空也可
		info("[JAR] Using existing JAR: %s", jarPath)
		if *nativeEnable && *skipJar {
			fail(errors.New("-nojar conflicts with -jar input; remove -nojar when using -jar"))
		}
	} else {
		// 正常流程：收集源码 -> 编译 -> 打 JAR
		javaFiles, err := collectJavaFiles(*srcDir)
		if err != nil {
			fail(err)
		}
		if len(javaFiles) == 0 {
			fail(fmt.Errorf("no .java files under %q", *srcDir))
		}
        // 推断/获取主类
		mainClass = strings.TrimSpace(*mainFQCN)
		if mainClass == "" {
			if len(javaFiles) == 1 {
				mc, err := inferMainFromSingle(*srcDir, javaFiles[0])
				if err != nil {
					fail(fmt.Errorf("failed to infer main class: %w", err))
				}
				mainClass = mc
				info("[MAIN] Single source; inferred main: %s", mainClass)
			} else {
				info("[HINT] Multiple source files detected. Set -main com.example.Main if you want a runnable JAR/native.")
			}
		}
        // 输出目录
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fail(fmt.Errorf("create out dir: %w", err))
		}
		classesDir := filepath.Join(filepath.Dir(*srcDir), "classes")
        os.MkdirAll(classesDir, 0755)
        // 编译
		if err := javacCompile(javaFiles, classesDir); err != nil {
			fail(err)
		}
		ok("Compiled %d source file(s) to %s", len(javaFiles), classesDir)
        // 打 JAR（可选）
		if !*skipJar && mainClass != "" {
			jarPath = filepath.Join(*outDir, *jarName)
			if err := buildJar(jarPath, classesDir, mainClass, *staticMode); err != nil {
				fail(err)
			}
			ok("Built runnable JAR: %s", jarPath)
			info(`[RUN] java -jar "%s"`, jarPath)
		} else {
			info(`[RUN] java -cp "%s" %s`, classesDir, mainClassOrHint(mainClass))
		}
		
		if *expLib {
			deps := collectDeps()
			copyExplib(*outDir, deps)
		}
	}

    // native 构建
	if *nativeEnable {
		if jarPath == "" {
			fail(errors.New("native build requires a JAR (either compile one or provide -jar)"))
		}
		switch strings.ToLower(*nativeDriver) {
		case "docker":
			if err := buildNativeDocker(jarPath, *nativeName, *nativeArgs, *dockerImage, *nativeTarget); err != nil {
				fail(err)
			}
			ok("Native binary: %s", filepath.Join(*outDir, *nativeName))
			if *nativeTarget == "linux" {
				info("Run in WSL2/Linux: chmod +x %s && ./%s", filepath.Join(*outDir, *nativeName), *nativeName)
			}
		case "local":
			// 本地 native-image（仅在当前 OS 生成同平台二进制）
			if err := buildNativeLocal(jarPath, *nativeName, *nativeArgs); err != nil {
				fail(err)
			}
			ok("Native binary: %s", filepath.Join(*outDir, *nativeName))
		default:
			fail(fmt.Errorf("unknown -native-driver: %s (allowed: docker|local)", *nativeDriver))
		}
	}
}
// ---------- Java 编译 & JAR ----------
func collectJavaFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".java") {
			ap, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			files = append(files, ap)
		}
		return nil
	})
	return files, err
}
var pkgRe = regexp.MustCompile(`(?m)^\s*package\s+([a-zA-Z0-9_\.]+)\s*;`)
func inferMainFromSingle(srcRoot, onlyFile string) (string, error) {
	// 尝试解析 package
	if b, err := os.ReadFile(onlyFile); err == nil {
		if m := pkgRe.FindSubmatch(b); len(m) == 2 {
			pkg := string(m[1])
			clz := strings.TrimSuffix(filepath.Base(onlyFile), ".java")
			return pkg + "." + clz, nil
		}
	}
	// 按路径推断
	srcAbs, err := filepath.Abs(srcRoot)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(srcAbs, onlyFile)
	if err != nil {
		return "", err
	}
	noExt := strings.TrimSuffix(rel, ".java")
	fqcn := filepath.ToSlash(noExt)
	fqcn = strings.ReplaceAll(fqcn, "/", ".")
	return fqcn, nil
}

// 收集依赖：lib/**/*.jar + dep.lib
func collectDeps() []string {
    var deps []string

    // 1. lib/**/*.jar
    filepath.WalkDir("lib", func(p string, d fs.DirEntry, err error) error {
        if err == nil && !d.IsDir() && strings.HasSuffix(p, ".jar") {
            abs, _ := filepath.Abs(p)
            deps = append(deps, abs)
        }
        return nil
    })

    // 2. dep.lib
    depFile := "dep.lib"
    if b, err := os.ReadFile(depFile); err == nil {
        scanner := bufio.NewScanner(bytes.NewReader(b))
        for scanner.Scan() {
            line := strings.TrimSpace(scanner.Text())
            if line == "" { continue }

            // 绝对路径
            if filepath.IsAbs(line) {
                if _, err := os.Stat(line); err == nil {
                    deps = append(deps, line)
                } else {
                    fmt.Printf("[WARN] dependency not found: %s\n", line)
                }
                continue
            }

            // 相对路径
            if _, err := os.Stat(line); err == nil {
                abs, _ := filepath.Abs(line)
                deps = append(deps, abs)
            } else {
                home, _ := os.UserHomeDir()
                alt := filepath.Join(home, ".m2", "repository", line)
                if _, err := os.Stat(alt); err == nil {
                    deps = append(deps, alt)
                } else {
                    fmt.Printf("[WARN] dependency not found: %s\n", line)
                }
            }
        }
    }
    
    fmt.Println("[INFO] Collected dependencies:")
    for _, d := range deps {
        fmt.Println("   ", d)
    }
    return deps
}

func javacCompile(javaFiles []string, out string) error {
	argsFile, err := os.CreateTemp("", "javac-sources-*.txt")
	if err != nil {
		return fmt.Errorf("create temp args file: %w", err)
	}
	defer os.Remove(argsFile.Name())
	w := bufio.NewWriter(argsFile)
	for _, f := range javaFiles {
		fmt.Fprintln(w, f)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := argsFile.Close(); err != nil {
		return err
	}
	
	deps := collectDeps()
    cp := []string{out}
    cp = append(cp, deps...)
    
    // 打印 javac 命令
    fmt.Println("[INFO] javac command:")
    fmt.Print("javac -d ",  out, " -cp ", strings.Join(cp, string(os.PathListSeparator)), " @"+argsFile.Name())
    fmt.Println()
    
	cmd := exec.Command("javac", "-d", out, "-cp", strings.Join(cp, string(os.PathListSeparator)), "@"+argsFile.Name())
	outb, errb, err := run(cmd)
	if err != nil {
		return fmt.Errorf("javac failed: %v\n%s\n%s", err, outb, errb)
	}
	return nil
}

func buildJar(jarPath, classesDir, mainClass string, staticMode bool) error {
	deps := collectDeps()
	manifest, err := os.CreateTemp("", "MANIFEST-*.mf")
	if err != nil {
		return fmt.Errorf("create manifest: %w", err)
	}
	defer os.Remove(manifest.Name())
	
    if err := writeManifest(manifest.Name(), mainClass, deps, staticMode); err != nil {
        return fmt.Errorf("write manifest: %w", err)
    }
	if staticMode {
        return buildFatJar(jarPath, classesDir, manifest.Name(), deps)
    }

    // 普通 JAR
    cmd := exec.Command("jar", "cfm", jarPath, manifest.Name(), "-C", classesDir, ".")
	outb, errb, err := run(cmd)
	if err != nil {
		hint := ""
		if runtime.GOOS == "windows" {
			hint = "\n(Hint: ensure a JDK is in PATH so 'jar' is available)"
		}
		return fmt.Errorf("jar failed: %v%s\n%s\n%s", err, hint, outb, errb)
	}
	// 写 META-INF/dependencies.lib
    if len(deps) > 0 {
        tmpDir := filepath.Join(os.TempDir(), "deps-"+filepath.Base(jarPath))
        metaInf := filepath.Join(tmpDir, "META-INF")
        os.MkdirAll(metaInf, 0755)
        depFile := filepath.Join(metaInf, "dependencies.lib")

        f, _ := os.Create(depFile)
        for _, d := range deps {
            fmt.Fprintln(f, "lib/"+filepath.Base(d))
        }
        f.Close()

        cmd = exec.Command("jar", "uf", jarPath, "-C", tmpDir, "META-INF/dependencies.lib")
        run(cmd)

        os.RemoveAll(tmpDir)
    }
	return nil
}
func writeManifest(path, mainClass string, deps []string, staticMode bool) error {
    var buf bytes.Buffer
    fmt.Fprintf(&buf, "Manifest-Version: 1.0\r\n")
    fmt.Fprintf(&buf, "Main-Class: %s\r\n", mainClass)

    if !staticMode && len(deps) > 0 {
        cpLine := "Class-Path:"
        for _, d := range deps {
            part := " lib/" + filepath.Base(d)
            if len(cpLine)+len(part) > 70 { // 超过72要换行
                buf.WriteString(cpLine + "\r\n")
                cpLine = " " + strings.TrimSpace(part) // 换行后必须空格开头
            } else {
                cpLine += part
            }
        }
        buf.WriteString(cpLine + "\r\n")
    }

    buf.WriteString("\r\n") // manifest 结束必须有空行

    return os.WriteFile(path, buf.Bytes(), 0644)
}
func buildFatJar(jarPath, classesDir, manifest string, deps []string) error {
    tmpJar := jarPath + ".tmp"
    exec.Command("jar", "cfm", tmpJar, manifest, "-C", classesDir, ".").Run()

    out, _ := os.Create(jarPath)
    zw := zip.NewWriter(out)

    // 先写主 JAR（保留所有内容）
    addJarKeepManifest(tmpJar, zw, nil)

    // 收集所有 services 文件的内容
    services := make(map[string][]string)

    // 再写依赖 JAR
    for _, d := range deps {
        addJarSkipManifest(d, zw, services)
    }

    // 合并写 META-INF/services/*
    for name, lines := range services {
        w, err := zw.Create(name)
        if err != nil {
            continue
        }
        for _, l := range lines {
            fmt.Fprintln(w, l)
        }
    }

    zw.Close()
    out.Close()
    os.Remove(tmpJar)

    // 写 META-INF/dependencies-static.lib
    buf := new(bytes.Buffer)
    for _, d := range deps {
        fmt.Fprintln(buf, "lib/"+filepath.Base(d))
    }
    addFileToJar(jarPath, "META-INF/dependencies-static.lib", buf.Bytes())

    return nil
}
// 主 JAR：保留所有文件（包括 MANIFEST.MF）
func addJarKeepManifest(path string, zw *zip.Writer, services map[string][]string) {
    r, _ := zip.OpenReader(path)
    defer r.Close()
    for _, f := range r.File {
        if f.FileInfo().IsDir() {
            continue
        }
        name := f.Name
        // 如果是 services 文件，收集内容
        if services != nil && strings.HasPrefix(strings.ToLower(name), "meta-inf/services/") {
        	fmt.Printf("[INFO] merging service file: %s from %s\n", name, filepath.Base(path))
            rc, _ := f.Open()
            scanner := bufio.NewScanner(rc)
            for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if line != "" && !strings.HasPrefix(line, "#") {
                    services[name] = append(services[name], line)
                }
            }
            rc.Close()
            continue
        }

        // 直接写入
        w, err := zw.Create(name)
        if err != nil {
            continue
        }
        rc, _ := f.Open()
        io.Copy(w, rc)
        rc.Close()
    }
}

// 依赖 JAR：跳过 MANIFEST.MF，合并 services 文件
func addJarSkipManifest(path string, zw *zip.Writer, services map[string][]string) {
    r, _ := zip.OpenReader(path)
    defer r.Close()
    for _, f := range r.File {
        if f.FileInfo().IsDir() {
            continue
        }
        name := f.Name
        lname := strings.ToLower(name)

        // 跳过依赖的 MANIFEST.MF
        if lname == "meta-inf/manifest.mf" {
            continue
        }

        // 合并 META-INF/services/*
        if strings.HasPrefix(lname, "meta-inf/services/") {
        	fmt.Printf("[INFO] merging service file: %s from %s\n", name, filepath.Base(path))
            rc, _ := f.Open()
            scanner := bufio.NewScanner(rc)
            for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if line != "" && !strings.HasPrefix(line, "#") {
                    services[name] = append(services[name], line)
                }
            }
            rc.Close()
            continue
        }

        // 其他文件 → 覆盖写
        w, err := zw.Create(name)
        if err != nil {
            continue
        }
        rc, _ := f.Open()
        io.Copy(w, rc)
        rc.Close()
    }
}



func addFileToJar(jarPath, name string, content []byte) {
    tmp := jarPath + ".tmp2"
    os.Rename(jarPath, tmp)
    out, _ := os.Create(jarPath)
    zw := zip.NewWriter(out)

    r, _ := zip.OpenReader(tmp)
    for _, f := range r.File {
        rc, _ := f.Open()
        w, _ := zw.Create(f.Name)
        io.Copy(w, rc)
        rc.Close()
    }
    r.Close()

    w, _ := zw.Create(name)
    w.Write(content)
    zw.Close(); 
    out.Close()
    os.Remove(tmp)
}

// copyExplib 会把依赖复制到输出 JAR 同目录下的 lib/ 目录
func copyExplib(outDir string, deps []string) {
    libDir := filepath.Join(outDir, "lib")
    os.MkdirAll(libDir, 0755)

    for _, d := range deps {
        target := filepath.Join(libDir, filepath.Base(d))
        data, err := os.ReadFile(d)
        if err != nil {
            fmt.Printf("[WARN] failed to read dependency %s: %v\n", d, err)
            continue
        }
        err = os.WriteFile(target, data, 0644)
        if err != nil {
            fmt.Printf("[WARN] failed to write dependency %s: %v\n", target, err)
        }
    }
    ok("Exported %d dependencies to %s", len(deps), libDir)
}



// ---------- Native 构建 ----------
func buildNativeDocker(jarPath, outName, extraArgs, image, target string) error {
    wd := mustGetwd()
    vol := wd + ":/work"

    if err := os.MkdirAll(*outDir, 0755); err != nil {
        return fmt.Errorf("create out dir: %w", err)
    }

    jarIn := "/work/" + toSlash(mustRel(wd, jarPath))
    outDirIn := "/work/" + toSlash(*outDir)
    nameOnly := filepath.Base(outName)

    args := []string{"run", "--rm", "-v", vol, "-w", "/work", image,
        "-jar", jarIn, "-H:Path=" + outDirIn, "-H:Name=" + nameOnly}
    if strings.TrimSpace(extraArgs) != "" {
        args = append(args, strings.Fields(extraArgs)...)
    }

    info("[NATIVE][docker] docker %s", strings.Join(args, " "))
    cmd := exec.Command("docker", args...)
    outb, errb, err := run(cmd)
    if err != nil {
        return fmt.Errorf("docker/native-image failed: %v\n%s\n%s", err, outb, errb)
    }
    return nil
}

func buildNativeLocal(jarPath, outName, extraArgs string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		return fmt.Errorf("create out dir: %w", err)
	}

	// 选择本机 native-image 可执行名（Windows 可能是 .cmd 或 .exe）
	bin := "native-image"
	if _, err := exec.LookPath(bin); err != nil {
		if _, err2 := exec.LookPath("native-image.cmd"); err2 == nil {
			bin = "native-image.cmd"
		} else if _, err3 := exec.LookPath("native-image.exe"); err3 == nil {
			bin = "native-image.exe"
		} else {
			return fmt.Errorf("native-image not found in PATH; please install GraalVM and add native-image to PATH")
		}
	}

	// -H:Path 指目录；-H:Name 只放文件名
	nameOnly := filepath.Base(outName)
	args := []string{
		"-jar", jarPath,
		"-H:Path=" + *outDir,
		"-H:Name=" + nameOnly,
	}
	if s := strings.TrimSpace(extraArgs); s != "" {
		args = append(args, strings.Fields(s)...)
	}

	info("[NATIVE][local] %s %s", bin, strings.Join(args, " "))
	cmd := exec.Command(bin, args...)
	outb, errb, err := run(cmd)
	if err != nil {
		return fmt.Errorf("local native-image failed: %v\n%s\n%s", err, outb, errb)
	}
	return nil
}

// ---------- 工具函数 ----------
func run(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
func info(f string, a ...any) { fmt.Printf("[INFO] "+f+"\n", a...) }
func ok(f string, a ...any)   { fmt.Printf("[OK]   "+f+"\n", a...) }
func fail(err error) {
	fmt.Fprintf(os.Stderr, "[ERR]  %v\n", err)
	os.Exit(1)
}
func mustAbs(p string) string {
	ap, err := filepath.Abs(p)
	if err != nil {
		fail(err)
	}
	return ap
}
func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	return wd
}
func mustRel(base, p string) string {
	rel, err := filepath.Rel(base, p)
	if err != nil {
		fail(err)
	}
	return rel
}
func toSlash(p string) string { return filepath.ToSlash(p) }
// 简易 shell 引号（供 bash -lc）
func shellQuote(s string) string {
	// wrap with single quotes, escape existing single quotes by closing/opening
	return "'" + strings.ReplaceAll(s, `'`, `'\''`) + "'"
}
func mainClassOrHint(mc string) string {
	if mc != "" {
		return mc
	}
	return "YourMainClass"
}

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