package ide

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bruceasu/asu-java-build/pkg"
)

func GenerateVSC(folder string) error {
	// If folder is "." or current directory, use current directory
	baseDir := folder
	if folder == "." {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else {
		abs, err := filepath.Abs(folder)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", folder, err)
		}
		baseDir = abs
	}

	// Create directories
	dirs := []string{filepath.Join(baseDir, ".vscode"), filepath.Join(baseDir, "src"), filepath.Join(baseDir, "lib"), filepath.Join(baseDir, "classes"), filepath.Join(baseDir, "bin")}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create .vscode/settings.json
	settingsPath := filepath.Join(baseDir, ".vscode", "settings.json")
	settingsContent := `{
"java.project.sourcePaths": ["src"],
"java.project.referencedLibraries": ["lib/**/*.jar"],
"java.project.outputPath": "classes",
"java.project.outputPathForJar": "bin"
}`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	// Create sample Main.java
	mainJavaPath := filepath.Join(baseDir, "src", "Main.java")
	mainJavaContent := `public class Main {
    public static void main(String[] args) {
        System.out.println("Hello from VS Code Java Project!");
    }
}
`
	if err := os.WriteFile(mainJavaPath, []byte(mainJavaContent), 0644); err != nil {
		return fmt.Errorf("failed to write Main.java: %w", err)
	}

	pkg.Info("Generated VS Code Java project settings in %s", baseDir)

	if err := pkg.EnsureGitSetup(baseDir); err != nil {
		pkg.Info("[GIT] %v", err)
	}
	return nil
}
