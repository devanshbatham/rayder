package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"time"
	"encoding/json"
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Name       string            `yaml:"workflow"`
	Parallel   bool              `yaml:"parallel"`
	Workers    int               `yaml:"workers"`
	Silent     bool              `yaml:"silent"`
	OutputDir  string            `yaml:"output-dir"`
	OutputFile string            `yaml:"output-file"`
	Steps      map[string]string `yaml:"steps"`
	Errors     map[string]string `yaml:"-"`
}

func main() {
	var workflowFile string
	var placeholdersJSON string
	flag.StringVar(&workflowFile, "workflow", "", "Path to the workflow YAML file")
	flag.StringVar(&workflowFile, "w", "", "Path to the workflow YAML file (alias)")

	flag.StringVar(&placeholdersJSON, "placeholders", "", "JSON representation of placeholders and values")
	flag.StringVar(&placeholdersJSON, "p", "", "JSON representation of placeholders and values (alias)")
	flag.Parse()

	log.SetFlags(0)

	log.Print(`


	                         __         
	   _____________  ______/ /__  _____
	  / ___/ __  / / / / __  / _ \/ ___/
	 / /  / /_/ / /_/ / /_/ /  __/ /    
	/_/   \____/\___ /\____/\___/_/     
	           /____/                   


 `)
	// Parse the placeholders JSON data (if provided)
	var placeholders map[string]string
	if placeholdersJSON != "" {
		if err := json.Unmarshal([]byte(placeholdersJSON), &placeholders); err != nil {
			log.Fatal("Error parsing placeholders JSON:", err)
		}
	}

	// Read the YAML file
	yamlData, err := ioutil.ReadFile(workflowFile)
	if err != nil {
		log.Fatal("Error reading YAML file:", err)
	}

	// Replace placeholders in the YAML content (if placeholders provided)
	yamlContent := string(yamlData)
	for placeholder, value := range placeholders {
		placeholderTag := fmt.Sprintf("<<%s>>", placeholder)
		yamlContent = strings.ReplaceAll(yamlContent, placeholderTag, value)
	}

	// Parse the modified YAML data into a Workflow struct
	var workflow Workflow
	err = yaml.Unmarshal([]byte(yamlContent), &workflow)
	if err != nil {
		log.Fatal("Error parsing YAML:", err)
	}

	// Ensure the output directory exists
	if workflow.OutputDir != "" {
		err = os.MkdirAll(workflow.OutputDir, 0755)
		if err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}

	fmt.Printf("[%s] [%s] Executing workflow %s\n", getTimeStamp(), getColorizedLog("INFO", "green"), workflow.Name)

	// Determine the number of workers
	numWorkers := workflow.Workers
	if numWorkers == 0 {
		numWorkers = 10
	}

	var wg sync.WaitGroup
	// Execute steps sequentially if parallel is false
	if !workflow.Parallel {
		runSequential(workflow.Steps, workflow.OutputDir, workflow.OutputFile, workflow.Silent, &workflow)
	} else {
		runParallel(workflow.Steps, numWorkers, workflow.OutputDir, workflow.OutputFile, workflow.Silent, &workflow, &wg)
	}

	// Print error messages for steps that failed
	printErrorMessages(workflow.Errors)

	fmt.Printf("[%s] [%s] Output saved in '%s' dir\n", getTimeStamp(), getColorizedLog("INFO", "green"), workflow.OutputDir)
	fmt.Printf("[%s] [%s] Job Finished", getTimeStamp(), getColorizedLog("INFO", "green"))
}



func runParallel(steps map[string]string, numWorkers int, outputDir, outputFile string, silent bool, workflow *Workflow, wg *sync.WaitGroup) {
	// Collect errors for steps that failed
	var mu sync.Mutex
	workflow.Errors = make(map[string]string)

	semaphore := make(chan struct{}, numWorkers)

	for stepName, command := range steps {
		wg.Add(1)
		go func(name string, cmd string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("[%s] [%s] Executing step: %s: %s\n", getTimeStamp(), getColorizedLog("INFO", "green"), name, cmd)
			execCmd := exec.Command("sh", "-c", cmd)

			if silent {
				execCmd.Stdout = ioutil.Discard
				execCmd.Stderr = ioutil.Discard
			} else {
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
			}

			if outputDir != "" && outputFile != "" {
				outputPath := filepath.Join(outputDir, outputFile)
				file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Printf("[%s] [%s] Error opening output file for step '%s': %v\n", getTimeStamp(), getColorizedLog("ERROR", "red"), name, err)
					return
				}
				defer file.Close()
				execCmd.Stdout = io.MultiWriter(file, execCmd.Stdout)
				execCmd.Stderr = io.MultiWriter(file, execCmd.Stderr)
			}

			err := execCmd.Run()
			if err != nil {
				errorMessage := fmt.Sprintf("Failed to execute: %v", err)
				mu.Lock()
				workflow.Errors[name] = errorMessage
				mu.Unlock()

				fmt.Printf("[%s] [%s] %s: %s\n", getTimeStamp(), getColorizedLog("ERROR", "red"), name, errorMessage)
			}
		}(stepName, command)
	}

	wg.Wait()
}

func runSequential(steps map[string]string, outputDir, outputFile string, silent bool, workflow *Workflow) {
	stepOrder := make([]string, 0, len(steps))
	for stepName := range steps {
		stepOrder = append(stepOrder, stepName)
	}

	// Sort the steps based on their order
	sort.Strings(stepOrder)

	// Initialize the Errors map
	workflow.Errors = make(map[string]string)

	for _, stepName := range stepOrder {
		cmd, exists := steps[stepName]
		if !exists {
			fmt.Printf("[%s] [%s] %s: %s\n", getTimeStamp(), getColorizedLog("ERROR", "red"), stepName, "Step not found")
			continue
		}

		fmt.Printf("[%s] [%s] Executing step: %s: %s\n", getTimeStamp(), getColorizedLog("INFO", "green"), stepName, cmd)
		execCmd := exec.Command("sh", "-c", cmd)


		if silent {
			execCmd.Stdout = ioutil.Discard
			execCmd.Stderr = ioutil.Discard
		} else {
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
		}

		if outputDir != "" && outputFile != "" {
			outputPath := filepath.Join(outputDir, outputFile)
			file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("[%s] [%s] Error opening output file for step '%s': %v\n", getTimeStamp(), getColorizedLog("ERROR", "red"), stepName, err)
				continue
			}
			defer file.Close()
			execCmd.Stdout = io.MultiWriter(file, execCmd.Stdout)
			execCmd.Stderr = io.MultiWriter(file, execCmd.Stderr)
		}

		err := execCmd.Run()
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to execute: %v", err)
			workflow.Errors[stepName] = errorMessage

			fmt.Printf("[%s] [%s] %s: %s\n", getTimeStamp(), getColorizedLog("ERROR", "red"), stepName, errorMessage)
			fmt.Printf("[%s] [%s] Exiting due to error in step: %s\n", getTimeStamp(), getColorizedLog("ERROR", "red"), stepName)
			os.Exit(1)
		}
	}
}




func printErrorMessages(errors map[string]string) {
	for stepName, errorMessage := range errors {
		fmt.Printf("[%s] [%s] %s: %s\n", getTimeStamp(), getColorizedLog("ERROR", "red"), stepName, errorMessage)
	}
}

func getTimeStamp() string {
	return time.Now().Format("15:04:05")
}

func getColorizedLog(text, color string) string {
	colorCode := getColorCode(color)
	return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, text)
}

func getColorCode(color string) string {
	colorMap := map[string]string{
		"black":   "30",
		"red":     "31",
		"green":   "32",
		"yellow":  "33",
		"blue":    "34",
		"magenta": "35",
		"cyan":    "36",
		"white":   "37",
	}

	return colorMap[color]
}