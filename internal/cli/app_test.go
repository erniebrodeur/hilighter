package cli

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/erniebrodeur/hilighter/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolveOptions", func() {
	It("uses the built-in app command when --cmd is not provided", func() {
		opts, err := resolveOptions(Options{
			App:       "syslog",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("/usr/bin/log stream --style syslog"))
	})

	It("uses the docker built-in app command when --cmd is not provided", func() {
		opts, err := resolveOptions(Options{
			App:       "docker",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("docker ps -a"))
	})

	It("prefers an explicit command over the built-in app command", func() {
		opts, err := resolveOptions(Options{
			App:       "syslog",
			Command:   "printf custom",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("printf custom"))
	})

	It("loads a default command from an explicit rules file", func() {
		rulesPath := filepath.Join(GinkgoT().TempDir(), "rules.yaml")
		Expect(os.WriteFile(rulesPath, []byte("cmd: printf from-rules\nrules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())

		opts, err := resolveOptions(Options{
			RulesPath: rulesPath,
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Command).To(Equal("printf from-rules"))
	})

	It("resolves a tail profile and defaults its file under the current working directory", func() {
		configDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("profiles:\n  rails-log:\n    rules: /tmp/rails.yaml\n    file: log/development.log\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile("/tmp/rails.yaml", []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		DeferCleanup(func() { _ = os.Remove("/tmp/rails.yaml") })

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "development.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "rails-log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.RulesPath).To(Equal("/tmp/rails.yaml"))
		Expect(opts.FilePath).To(Equal("./log/development.log"))
		Expect(opts.Command).To(Equal(`tail -f "./log/development.log"`))
	})

	It("prefers an explicit tail file argument over the profile default", func() {
		configDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("profiles:\n  rails-log:\n    file: log/development.log\n"), 0o644)).To(Succeed())

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "production.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "rails-log",
			FilePath:  "log/production.log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.FilePath).To(Equal("./log/production.log"))
		Expect(opts.Command).To(Equal(`tail -f "./log/production.log"`))
	})

	It("treats an unknown file-mode target as a direct file path", func() {
		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "development.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "log/development.log",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Profile).To(BeEmpty())
		Expect(opts.FilePath).To(Equal("./log/development.log"))
		Expect(opts.Command).To(Equal(`tail -f "./log/development.log"`))
	})

	It("auto-detects a saved profile from the file path and uses its rules", func() {
		configDir := GinkgoT().TempDir()
		rulesPath := filepath.Join(configDir, "rails.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("profiles:\n  rails-log:\n    rules: "+rulesPath+"\n    file: log/development.log\n"), 0o644)).To(Succeed())

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "development.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "log/development.log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Profile).To(BeEmpty())
		Expect(opts.RulesPath).To(Equal(rulesPath))
		Expect(opts.Command).To(Equal(`tail -f "./log/development.log"`))
	})

	It("auto-detects a rules file from the config rules directory based on the target path", func() {
		configDir := GinkgoT().TempDir()
		rulesDir := filepath.Join(configDir, "rules")
		Expect(os.MkdirAll(rulesDir, 0o755)).To(Succeed())
		dockerRules := filepath.Join(rulesDir, "docker.yaml")
		Expect(os.WriteFile(dockerRules, []byte("rules:\n  - name: container\n    pattern: 'container'\n    style: info\n"), 0o644)).To(Succeed())

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "logs"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "logs", "docker.log"), []byte("container booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "cat",
			Profile:   "logs/docker.log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Profile).To(BeEmpty())
		Expect(opts.RulesPath).To(Equal(dockerRules))
		Expect(opts.Command).To(Equal(`cat "./logs/docker.log"`))
	})

	It("auto-detects a built-in app from the target path when no profile or rules file match", func() {
		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "var", "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "var", "log", "syslog"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "head",
			Profile:   "var/log/syslog",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Profile).To(BeEmpty())
		Expect(opts.App).To(Equal("syslog"))
		Expect(opts.Command).To(Equal(`head "./var/log/syslog"`))
	})

	It("skips auto-detection when --no-detect is set", func() {
		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "var", "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "var", "log", "syslog"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:        "head",
			Profile:     "var/log/syslog",
			NoDetect:    true,
			DebugDetect: true,
			ConfigDir:   GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.App).To(BeEmpty())
		Expect(opts.RulesPath).To(BeEmpty())
		Expect(opts.DetectionNote).To(Equal("auto-detection disabled"))
	})

	It("returns a clear error when a direct file path is combined with an extra file argument", func() {
		_, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "log/development.log",
			FilePath:  "log/production.log",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).To(MatchError(`tail accepts either a file path alone or a profile name with an optional file path`))
	})

	It("resolves a cat profile through the shared file-mode helper", func() {
		configDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("profiles:\n  rails-log:\n    file: log/development.log\n"), 0o644)).To(Succeed())

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "development.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "cat",
			Profile:   "rails-log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.FilePath).To(Equal("./log/development.log"))
		Expect(opts.Command).To(Equal(`cat "./log/development.log"`))
	})

	It("resolves a head profile through the shared file-mode helper", func() {
		configDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("profiles:\n  rails-log:\n    file: log/development.log\n"), 0o644)).To(Succeed())

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "development.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "head",
			Profile:   "rails-log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.FilePath).To(Equal("./log/development.log"))
		Expect(opts.Command).To(Equal(`head "./log/development.log"`))
	})

	It("prefers profile rules over root config defaults", func() {
		configDir := GinkgoT().TempDir()
		globalRules := filepath.Join(configDir, "global-rules.yaml")
		profileRules := filepath.Join(configDir, "rails-rules.yaml")
		Expect(os.WriteFile(globalRules, []byte("rules:\n  - name: warn\n    pattern: 'WARN'\n    style: warning\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(profileRules, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("rules: "+globalRules+"\nprofiles:\n  rails-log:\n    rules: "+profileRules+"\n    file: log/development.log\n"), 0o644)).To(Succeed())

		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "log", "development.log"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "rails-log",
			ConfigDir: configDir,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(opts.RulesPath).To(Equal(profileRules))
	})

	It("returns a clear error when a requested profile does not exist outside file modes", func() {
		_, err := resolveOptions(Options{
			Profile:   "missing",
			ConfigDir: GinkgoT().TempDir(),
		})

		Expect(err).To(MatchError(`unknown profile "missing"`))
	})

	It("returns a clear error when a tail profile has no file and no file argument was provided", func() {
		configDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("profiles:\n  rails-log:\n    app: logs\n"), 0o644)).To(Succeed())

		_, err := resolveOptions(Options{
			Mode:      "tail",
			Profile:   "rails-log",
			ConfigDir: configDir,
		})

		Expect(err).To(MatchError(`profile "rails-log" does not define a default file and no file argument was provided`))
	})
})

var _ = Describe("shouldRunCommand", func() {
	It("always runs file modes even when stdin is piped", func() {
		run := shouldRunCommand(
			Options{Mode: "head", Profile: "rails-log"},
			Options{Mode: "head", Profile: "rails-log", Command: `head "./log/development.log"`},
			0,
		)

		Expect(run).To(BeTrue())
	})

	It("runs an explicit command even when stdin is piped", func() {
		run := shouldRunCommand(
			Options{Command: "printf explicit"},
			Options{Command: "printf explicit"},
			0,
		)

		Expect(run).To(BeTrue())
	})

	It("uses stdin when a built-in app only contributes a default command and stdin is piped", func() {
		run := shouldRunCommand(
			Options{App: "docker"},
			Options{App: "docker", Command: "docker ps -a"},
			0,
		)

		Expect(run).To(BeFalse())
	})

	It("runs the built-in default command when stdin is interactive", func() {
		run := shouldRunCommand(
			Options{App: "docker"},
			Options{App: "docker", Command: "docker ps -a"},
			fs.ModeCharDevice,
		)

		Expect(run).To(BeTrue())
	})
})

var _ = Describe("shouldShowHelp", func() {
	It("shows help for an interactive session with no arguments", func() {
		show := shouldShowHelp(Options{}, fs.ModeCharDevice)
		Expect(show).To(BeTrue())
	})

	It("does not show help when stdin is piped", func() {
		show := shouldShowHelp(Options{}, 0)
		Expect(show).To(BeFalse())
	})

	It("does not show help when a version request is present", func() {
		show := shouldShowHelp(Options{ShowVersion: true}, fs.ModeCharDevice)
		Expect(show).To(BeFalse())
	})
})

var _ = Describe("formattedVersion", func() {
	It("formats the app-wide version marker", func() {
		previous := Version
		Version = "1.0.0"
		DeferCleanup(func() { Version = previous })

		Expect(formattedVersion()).To(Equal("hilighter-1.0.0"))
	})
})

var _ = Describe("control modes", func() {
	It("lists built-in apps", func() {
		var out bytes.Buffer

		err := runList(Options{Subject: "apps"}, config.Config{}, &out)

		Expect(err).NotTo(HaveOccurred())
		Expect(out.String()).To(ContainSubstring("docker"))
		Expect(out.String()).To(ContainSubstring("syslog"))
	})

	It("shows a profile definition", func() {
		var out bytes.Buffer
		cfg := config.Config{
			Profiles: map[string]config.Profile{
				"rails-log": {
					RulesPath: "/tmp/rails.yaml",
					ThemePath: "/tmp/default.yaml",
					FilePath:  "log/development.log",
				},
			},
		}

		err := runShow(Options{Subject: "profile", Name: "rails-log"}, cfg, &out)

		Expect(err).NotTo(HaveOccurred())
		Expect(out.String()).To(ContainSubstring("profile: rails-log"))
		Expect(out.String()).To(ContainSubstring("rules: /tmp/rails.yaml"))
		Expect(out.String()).To(ContainSubstring("file: log/development.log"))
	})

	It("validates an explicit rules file and theme file", func() {
		configDir := GinkgoT().TempDir()
		rulesPath := filepath.Join(configDir, "rules.yaml")
		themePath := filepath.Join(configDir, "theme.yaml")
		Expect(os.WriteFile(rulesPath, []byte("rules:\n  - name: error\n    pattern: 'ERROR'\n    style: error\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(themePath, []byte("styles:\n  error:\n    fg: white\n    bg: red\n"), 0o644)).To(Succeed())

		var out bytes.Buffer
		err := runValidate(Options{
			ConfigDir: configDir,
			RulesPath: rulesPath,
			ThemePath: themePath,
		}, config.Config{}, false, &out)

		Expect(err).NotTo(HaveOccurred())
		Expect(out.String()).To(ContainSubstring("ok:"))
		Expect(out.String()).To(ContainSubstring(rulesPath))
		Expect(out.String()).To(ContainSubstring(themePath))
	})

	It("reports validation failures for a broken app reference in config", func() {
		var out bytes.Buffer
		err := runValidate(Options{}, config.Config{
			Profiles: map[string]config.Profile{
				"broken": {App: "missing-app"},
			},
		}, true, &out)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`profile "broken": unknown built-in app profile "missing-app"`))
	})

	It("prints a debug detection message for a detected built-in", func() {
		cwd := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(cwd, "var", "log"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cwd, "var", "log", "syslog"), []byte("booted\n"), 0o644)).To(Succeed())

		previous, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(cwd)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(previous) })

		opts, err := resolveOptions(Options{
			Mode:        "head",
			Profile:     "var/log/syslog",
			DebugDetect: true,
			ConfigDir:   GinkgoT().TempDir(),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(strings.ToLower(opts.DetectionNote)).To(ContainSubstring("matched built-in app"))
	})
})
