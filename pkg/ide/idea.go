package ide

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bruceasu/asu-java-build/pkg"
)

func GenerateIdea(folder string) error {
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
	dirs := []string{filepath.Join(baseDir, ".idea"), filepath.Join(baseDir, "src"), filepath.Join(baseDir, "lib"), filepath.Join(baseDir, "classes"), filepath.Join(baseDir, "bin")}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create .idea/misc.xml
	miscPath := filepath.Join(baseDir, ".idea", "misc.xml")
	miscContent := `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="ProjectRootManager" version="2" languageLevel="JDK_11" project-jdk-name="11" project-jdk-type="JavaSDK">
    <output url="file://$PROJECT_DIR$/classes" />
  </component>
</project>`
	if err := os.WriteFile(miscPath, []byte(miscContent), 0644); err != nil {
		return fmt.Errorf("failed to write misc.xml: %w", err)
	}

	// Create .idea/modules.xml
	modulesPath := filepath.Join(baseDir, ".idea", "modules.xml")
	modulesContent := `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="ProjectModuleManager">
    <modules>
      <module fileurl="file://$PROJECT_DIR$/.idea/JavaProject.iml" filepath="$PROJECT_DIR$/.idea/JavaProject.iml" />
    </modules>
  </component>
</project>`
	if err := os.WriteFile(modulesPath, []byte(modulesContent), 0644); err != nil {
		return fmt.Errorf("failed to write modules.xml: %w", err)
	}

	// Create .idea/JavaProject.iml
	imlPath := filepath.Join(baseDir, ".idea", "JavaProject.iml")
	imlContent := `<?xml version="1.0" encoding="UTF-8"?>
<module type="JAVA_MODULE" version="4">
  <component name="NewModuleRootManager" inherit-compiler-output="true">
    <exclude-output />
    <content url="file://$MODULE_DIR$">
      <sourceFolder url="file://$MODULE_DIR$/src" isTestSource="false" />
    </content>
    <orderEntry type="inheritedJdk" />
    <orderEntry type="sourceFolder" forTests="false" />
    <orderEntry type="library" name="lib" level="project" />
  </component>
</module>`
	if err := os.WriteFile(imlPath, []byte(imlContent), 0644); err != nil {
		return fmt.Errorf("failed to write JavaProject.iml: %w", err)
	}

	pkg.Info("Generated IntelliJ IDEA Java project settings in %s", baseDir)

	if err := pkg.EnsureGitSetup(baseDir); err != nil {
		pkg.Info("[GIT] %v", err)
	}
	return nil
}
