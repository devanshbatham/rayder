package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Vars     map[string]string `yaml:"vars"`
	Parallel bool              `yaml:"parallel"`
	Tasks    []struct {
		Name    string   `yaml:"name"`
		Cmds    []string `yaml:"cmds"`
		Silent  bool     `yaml:"silent"`
	} `yaml:"modules"`
}

func main() {
	var (
		taskFile  string
		variables map[string]string
		quietMode  bool // Flag to indicate quiet mode
	)

	flag.StringVar(&taskFile, "w", "", "Path to the workflow YAML file")
	flag.BoolVar(&quietMode, "q", false, "Suppress banner")
	flag.Parse()
	log.SetFlags(0)

	// Color formatting functions
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	// Print banner
	if !quietMode {
		// Print banner only if quiet mode is not enabled
		fmt.Printf("\n%s\n\n", white(`
	                         __         
	   _____________  ______/ /__  _____
	  / ___/ __  / / / / __  / _ \/ ___/
	 / /  / /_/ / /_/ / /_/ /  __/ /    
	/_/   \____/\___ /\____/\___/_/     
	           /____/                   

	           		- v0.0.3 ⚡

`))

}

	var defaultVars map[string]string
	yamlFileContent, err := ioutil.ReadFile(taskFile)
	if err == nil {
		var config Config
		err = yaml.Unmarshal(yamlFileContent, &config)
		if err == nil {
			defaultVars = config.Vars
		}
	}

	variables = parseArgs(defaultVars)


	if taskFile == "" {
		fmt.Println("Usage: rayder -w workflow.yaml [variable assignments e.g. DOMAIN=example.host]")
		return
	}

	taskFileContent, err := ioutil.ReadFile(taskFile)
	if err != nil {
		log.Fatalf("Error reading workflow file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(taskFileContent, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	runAllTasks(config, variables, cyan, magenta, white, yellow, red, green)
}

func parseArgs(defaultVars map[string]string) map[string]string {
	variables := make(map[string]string)

	for _, arg := range flag.Args() {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			if variables == nil {
				variables = make(map[string]string)
			}
			variables[parts[0]] = parts[1]
		}
	}

	// Apply default values if not provided by the user
	for key, defaultValue := range defaultVars {
		if _, exists := variables[key]; !exists {
			variables[key] = defaultValue
		}
	}

	return variables
}


func runAllTasks(config Config, variables map[string]string, cyan, magenta, white, yellow, red, green func(a ...interface{}) string) {
	var wg sync.WaitGroup
	var errorOccurred bool
	var errorMutex sync.Mutex

	for _, task := range config.Tasks {
		if config.Parallel {
			wg.Add(1)
			go func(name string, cmds []string, silent bool, vars map[string]string) {
				defer wg.Done()
				err := runTask(name, cmds, silent, vars, cyan, magenta, white, yellow, red, green)
				if err != nil {
					errorMutex.Lock()
					errorOccurred = true
					fmt.Printf("[%s] [%s] Module '%s' %s ❌\n", yellow(currentTime()), red("INFO"), cyan(name), red("errored"))
					errorMutex.Unlock()
				}
			}(task.Name, task.Cmds, task.Silent, variables)
		} else {
			err := runTask(task.Name, task.Cmds, task.Silent, variables, cyan, magenta, white, yellow, red, green)
			if err != nil {
				errorOccurred = true
				fmt.Printf("[%s] [%s] Module '%s' %s ❌\n", yellow(currentTime()), red("INFO"), cyan(task.Name), red("errored"))
				return // Exit the function immediately if an error occurs
			}
		}
	} 

	if config.Parallel {
		wg.Wait()
	}

	if !config.Parallel && errorOccurred {
		fmt.Printf("[%s] [%s] Errors occurred during execution. Exiting program ❌\n", yellow(currentTime()), red("INFO"))
		os.Exit(1) // Exit with error code 1
	}

	if errorOccurred {
		fmt.Printf("[%s] [%s] Errors occurred during execution of some command(s) ❌\n", yellow(currentTime()), red("INFO"))
	} else {
		fmt.Printf("[%s] [%s] All modules completed successfully ✅\n", yellow(currentTime()), yellow("INFO"))
	}
}

func runTask(taskName string, cmds []string, silent bool, vars map[string]string, cyan, magenta, white, yellow, red, green func(a ...interface{}) string) error {
	currentTime()
	fmt.Printf("[%s] [%s] Module '%s' %s ⚡\n", yellow(currentTime()), yellow("INFO"), cyan(taskName), yellow("running"))

	var hasError bool
	for _, cmd := range cmds {
		err := executeCommand(cmd, silent, vars)
		if err != nil {
			hasError = true
			break
		}
	}

	if hasError {
		return fmt.Errorf("Module '%s' %s ❌", taskName, red("errored"))
	}

	fmt.Printf("[%s] [%s] Module '%s' %s ✅\n", yellow(currentTime()), yellow("INFO"), cyan(taskName), green("completed"))
	return nil
}

func executeCommand(cmdStr string, silent bool, vars map[string]string) error {
	cmdStr = replacePlaceholders(cmdStr, vars)
	execCmd := exec.Command("sh", "-c", cmdStr)

	if silent {
		execCmd.Stdout = nil
		execCmd.Stderr = nil
	} else {
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
	}

	err := execCmd.Run()
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}

func replacePlaceholders(input string, vars map[string]string) string {
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		input = strings.ReplaceAll(input, placeholder, value)
	}
	return input
}

func currentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
