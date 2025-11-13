package ide

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bruceasu/asu-java-build/pkg"
)

func GenerateMaven(folder string, findMavenCommand func() string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var targetParent string
	var desiredPath string
	var defaultArtifact string
	if folder == "." {
		targetParent = cwd
		defaultArtifact = filepath.Base(cwd)
		if defaultArtifact == "." || defaultArtifact == string(filepath.Separator) {
			defaultArtifact = "app"
		}
	} else {
		abs, err := filepath.Abs(folder)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", folder, err)
		}
		desiredPath = abs
		targetParent = filepath.Dir(abs)
		defaultArtifact = filepath.Base(abs)
		if err := os.MkdirAll(targetParent, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory %s: %w", targetParent, err)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	defaultGroup := "me.asu"
	fmt.Printf("Enter groupId (default: %s): ", defaultGroup)
	groupId := defaultGroup
	if scanner.Scan() {
		if v := strings.TrimSpace(scanner.Text()); v != "" {
			groupId = v
		}
	} else if err := scanner.Err(); err != nil {
		return fmt.Errorf("read groupId: %w", err)
	}

	fmt.Printf("Enter artifactId (default: %s): ", defaultArtifact)
	artifactId := defaultArtifact
	if scanner.Scan() {
		if v := strings.TrimSpace(scanner.Text()); v != "" {
			artifactId = v
		}
	} else if err := scanner.Err(); err != nil {
		return fmt.Errorf("read artifactId: %w", err)
	}
	if strings.TrimSpace(artifactId) == "" {
		return errors.New("artifactId is required")
	}

	defaultVersion := "1.0.0-SNAPSHOT"
	fmt.Printf("Enter version (default: %s): ", defaultVersion)
	version := defaultVersion
	if scanner.Scan() {
		if v := strings.TrimSpace(scanner.Text()); v != "" {
			version = v
		}
	} else if err := scanner.Err(); err != nil {
		return fmt.Errorf("read version: %w", err)
	}

	fmt.Print("Select Java version (8/17/21, default: 21): ")
	javaVersion := "21"
	if scanner.Scan() {
		if v := strings.TrimSpace(scanner.Text()); v != "" {
			switch v {
			case "8", "17", "21":
				javaVersion = v
			default:
				pkg.Info("[MAVEN] Unsupported Java version %s, fallback to 21", v)
			}
		}
	} else if err := scanner.Err(); err != nil {
		return fmt.Errorf("read Java version: %w", err)
	}

	generatedPath := filepath.Join(targetParent, artifactId)
	if _, err := os.Stat(generatedPath); err == nil {
		return fmt.Errorf("target directory %s already exists", generatedPath)
	}

	mvnCmd := findMavenCommand()
	pkg.Info("[MAVEN] Generating project using %s", mvnCmd)
	args := []string{
		"archetype:generate",
		"-DgroupId=" + groupId,
		"-DartifactId=" + artifactId,
		"-Dversion=" + version,
		"-DarchetypeGroupId=org.apache.maven.archetypes",
		"-DarchetypeArtifactId=maven-archetype-quickstart",
		"-DarchetypeVersion=1.5",
		"-DinteractiveMode=false",
	}
	outb, errb, err := pkg.RunCommandDir(targetParent, mvnCmd, args...)
	if err != nil {
		return fmt.Errorf("maven archetype generation failed: %v\n%s\n%s", err, outb, errb)
	}
	pkg.Info("[MAVEN] Archetype generation completed")

	baseDir := generatedPath
	if desiredPath != "" {
		desiredPath = filepath.Clean(desiredPath)
		if !pkg.SamePath(baseDir, desiredPath) {
			if _, err := os.Stat(desiredPath); err == nil {
				return fmt.Errorf("destination %s already exists", desiredPath)
			}
			if err := os.Rename(baseDir, desiredPath); err != nil {
				return fmt.Errorf("failed to move project to %s: %w", desiredPath, err)
			}
			baseDir = desiredPath
		}
	}

	// 更新 pom.xml 中的 Java 版本和构建配置，而不是完全覆盖
	pomPath := filepath.Join(baseDir, "pom.xml")
	pomBytes, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml: %w", err)
	}

	pomStr := string(pomBytes)

	// 更新 Java 版本
	pomStr = strings.ReplaceAll(pomStr, "<maven.compiler.source>1.7</maven.compiler.source>",
		fmt.Sprintf("<maven.compiler.source>%s</maven.compiler.source>", javaVersion))
	pomStr = strings.ReplaceAll(pomStr, "<maven.compiler.target>1.7</maven.compiler.target>",
		fmt.Sprintf("<maven.compiler.target>%s</maven.compiler.target>", javaVersion))

	// 在 </build> 之前添加自定义插件配置
	if !strings.Contains(pomStr, "<plugins>") || !strings.Contains(pomStr, "</plugins>") {
		// 如果没有 <plugins> 标签，添加整个 plugins 块
		pluginsBlock := fmt.Sprintf(`  <build>
    <plugins>
      <plugin>
        <groupId>org.apache.maven.plugins</groupId>
        <artifactId>maven-compiler-plugin</artifactId>
        <version>3.11.0</version>
        <configuration>
          <source>%s</source>
          <target>%s</target>
          <encoding>UTF-8</encoding>
          <parameters>true</parameters>
        </configuration>
      </plugin>
      <plugin>
        <groupId>org.apache.maven.plugins</groupId>
        <artifactId>maven-jar-plugin</artifactId>
        <version>3.3.0</version>
        <configuration>
          <archive>
            <manifest>
              <addClasspath>true</addClasspath>
              <classpathPrefix>lib/</classpathPrefix>
              <mainClass>${project.groupId}.App</mainClass>
            </manifest>
          </archive>
        </configuration>
      </plugin>
    </plugins>
  </build>`, javaVersion, javaVersion)
		pomStr = strings.ReplaceAll(pomStr, "</build>", pluginsBlock)
	} else {
		// 如果已有 <plugins>，在 </plugins> 前添加我们的插件
		compilerPlugin := fmt.Sprintf(`      <plugin>
        <groupId>org.apache.maven.plugins</groupId>
        <artifactId>maven-compiler-plugin</artifactId>
        <version>3.11.0</version>
        <configuration>
          <source>%s</source>
          <target>%s</target>
          <encoding>UTF-8</encoding>
          <parameters>true</parameters>
        </configuration>
      </plugin>
      <plugin>
        <groupId>org.apache.maven.plugins</groupId>
        <artifactId>maven-jar-plugin</artifactId>
        <version>3.3.0</version>
        <configuration>
          <archive>
            <manifest>
              <addClasspath>true</addClasspath>
              <classpathPrefix>lib/</classpathPrefix>
              <mainClass>${project.groupId}.App</mainClass>
            </manifest>
          </archive>
        </configuration>
      </plugin>
    </plugins>`, javaVersion, javaVersion)

		// 在 </pluginManagement> 后添加 <plugins>
		if strings.Contains(pomStr, "</pluginManagement>") {
			pomStr = strings.ReplaceAll(pomStr, "</pluginManagement>", "</pluginManagement>\n    <plugins>\n"+compilerPlugin)
		} else {
			pomStr = strings.ReplaceAll(pomStr, "</plugins>", compilerPlugin+"</plugins>")
		}
	}

	if err := os.WriteFile(pomPath, []byte(pomStr), 0644); err != nil {
		return fmt.Errorf("failed to update pom.xml: %w", err)
	}

	pkg.Info("Generated Maven project in %s", baseDir)

	if err := pkg.EnsureGitSetup(baseDir); err != nil {
		pkg.Info("[GIT] %v", err)
	}

	return nil
}
