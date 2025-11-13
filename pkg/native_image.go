package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ---------- Native 构建 ----------
func buildNativeDocker(jarPath, outDir, outName, extraArgs, image, target string) error {
	wd := MustGetwd()
	vol := wd + ":/work"

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create out dir: %w", err)
	}

	jarIn := "/work/" + ToSlash(MustRel(wd, jarPath))
	outDirIn := "/work/" + ToSlash(outDir)
	nameOnly := filepath.Base(outName)

	args := []string{"run", "--rm", "-v", vol, "-w", "/work", image,
		"-jar", jarIn, "-H:Path=" + outDirIn, "-H:Name=" + nameOnly}
	if strings.TrimSpace(extraArgs) != "" {
		args = append(args, strings.Fields(extraArgs)...)
	}

	Info("[NATIVE][docker] docker %s", strings.Join(args, " "))
	cmd := exec.Command("docker", args...)
	outb, errb, err := Run(cmd)
	if err != nil {
		return fmt.Errorf("docker/native-image failed: %v\n%s\n%s", err, outb, errb)
	}
	return nil
}

func buildNativeLocal(jarPath, outDir, outName, extraArgs string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(outDir, 0755); err != nil {
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
		"-H:Path=" + outDir,
		"-H:Name=" + nameOnly,
	}
	if s := strings.TrimSpace(extraArgs); s != "" {
		args = append(args, strings.Fields(s)...)
	}

	Info("[NATIVE][local] %s %s", bin, strings.Join(args, " "))
	cmd := exec.Command(bin, args...)
	outb, errb, err := Run(cmd)
	if err != nil {
		return fmt.Errorf("local native-image failed: %v\n%s\n%s", err, outb, errb)
	}
	return nil
}

// BuildNative 处理 native 构建流程
func BuildNative(jarPath, outDir, nativeName, nativeArgs, dockerImage, nativeDriver, nativeTarget string) error {
	if jarPath == "" {
		return fmt.Errorf("native build requires a JAR (either compile one or provide -jar)")
	}
	if nativeName == "" {
		if nativeTarget == "windows" {
			nativeName = "app-windows-amd64"
		} else {
			nativeName = "app-linux-amd64"
		}
	}
	switch strings.ToLower(nativeDriver) {
	case "docker":
		if err := buildNativeDocker(jarPath, outDir, nativeName, nativeArgs, dockerImage, nativeTarget); err != nil {
			return err
		}
		Ok("Native binary: %s", filepath.Join(outDir, nativeName))
		if nativeTarget == "linux" {
			Info("Run in WSL2/Linux: chmod +x %s && ./%s", filepath.Join(outDir, nativeName), nativeName)
		}
	case "local":
		// 本地 native-image（仅在当前 OS 生成同平台二进制）
		if err := buildNativeLocal(jarPath, outDir, nativeName, nativeArgs); err != nil {
			return err
		}
		Ok("Native binary: %s", filepath.Join(outDir, nativeName))
	default:
		return fmt.Errorf("unknown -native-driver: %s (allowed: docker|local)", nativeDriver)
	}

	return nil
}
