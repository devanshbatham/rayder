<h1 align="center">
  Rayder
</h1>

<p align="center">
  <strong>A lightweight tool for orchestrating and organizing your command-line workflows</strong>
</p>

<p align="center">
  <a href="#about">About</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#workflow-configuration">Workflow Configuration</a> •
  <a href="#parallel-execution">Parallel Execution</a> •
  <a href="#workflows">Workflows</a>
</p>

---

## About

Rayder is a command-line tool designed to simplify the orchestration and execution of workflows. It allows you to define a series of steps in a YAML file, each consisting of a command to be executed and optional output redirection. Rayder helps automate complex processes, making it easy to streamline repetitive tasks, and executing them parallelly if the commands do not depend on each other. 

## Installation

To install Rayder, ensure you have Go (1.16 or higher) installed on your system. Then, run the following command:

```sh
go install github.com/devanshbatham/rayder@latest
```

## Usage

Rayder offers a straightforward way to execute workflows defined in YAML files. Use the following command:

```sh
rayder -w path/to/workflow.yaml
```

## Workflow Configuration

A workflow is defined in a YAML file with the following structure:

```yaml
workflow: workflow-name
parallel: true|false
silent: true|false
output-dir: output-directory
output-file: workflow.log
steps:
  step-1: command-1
  step-2: command-2
  # Add more steps...
```


## Example workflow with placeholder

```yaml
workflow: reverse-whois
parallel: false
silent: true
output-dir: results
output-file: workflow.log
steps:
  step-1: revwhoix -k "<<ORG>>" > results/root-domains.txt
  step-2: xargs -I {} -a results/root-domains.txt echo "subfinder -d {} -o {}.out" | quaithe -workers 30 -silent
  step-3: cat *.out > results/root-subdomains.txt
  step-4: rm *.out
  step-5: cat results/root-subdomains.txt | dnsx -silent -threads 100 -o results/resolved-subdomains.txt
```

To be executed as: 

```sh
rayder -w workflow.yaml -p '{"ORG":"Yelp, Inc"}'
```



## Parallel Execution

The `parallel` field in the workflow configuration determines whether steps should be executed in parallel or sequentially. Setting `parallel` to `true` allows steps to run concurrently, making it suitable for tasks with no dependencies. When set to `false`, steps will execute one after another.


## Workflows

A collection of Rayder workflows to be publish soon. Stay tuned..