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

![rayder](https://github.com/devanshbatham/rayder/blob/main/static/banner.png?raw=true)


## About

Rayder is a command-line tool designed to simplify the orchestration and execution of workflows. It allows you to define a series of modules in a YAML file, each consisting of commands to be executed. Rayder helps you automate complex processes, making it easy to streamline repetitive modules and execute them parallelly if the commands do not depend on each other.

## Installation

To install Rayder, ensure you have Go (1.16 or higher) installed on your system. Then, run the following command:

```sh
go install github.com/devanshbatham/rayder@v0.0.4
```

## Usage

Rayder offers a straightforward way to execute workflows defined in YAML files. Use the following command:

```sh
rayder -w path/to/workflow.yaml
```

## Workflow Configuration

A workflow is defined in a YAML file with the following structure:

```yaml
vars:
  VAR_NAME: value
  # Add more variables...

parallel: true|false
modules:
  - name: task-name
    cmds:
      - command-1
      - command-2
      # Add more commands...
    silent: true|false
  # Add more modules...
```

## Using Variables in Workflows

Rayder allows you to use variables in your workflow configuration, making it easy to parameterize your commands and achieve more flexibility. You can define variables in the `vars` section of your workflow YAML file. These variables can then be referenced within your command strings using double curly braces (`{{}}`).

### Defining Variables

To define variables, add them to the `vars` section of your workflow YAML file:

```yaml
vars:
  VAR_NAME: value
  ANOTHER_VAR: another_value
  # Add more variables...
```

### Referencing Variables in Commands

You can reference variables within your command strings using double curly braces (`{{}}`). For example, if you defined a variable `OUTPUT_DIR`, you can use it like this:

```yaml
modules:
  - name: example-task
    cmds:
      - echo "Output directory {{OUTPUT_DIR}}"
```

### Supplying Variables via the Command Line

You can also supply values for variables via the command line when executing your workflow. Use the format `VARIABLE_NAME=value` to provide values for specific variables. For example:

```sh
rayder -w path/to/workflow.yaml VAR_NAME=new_value ANOTHER_VAR=updated_value
```

If you don't provide values for variables via the command line, Rayder will automatically apply default values defined in the `vars` section of your workflow YAML file.

Remember that variables supplied via the command line will override the default values defined in the YAML configuration.

## Example

### Example 1: 

Here's an example of how you can define, reference, and supply variables in your workflow configuration:

```yaml
vars:
  ORG: "example.org"
  OUTPUT_DIR: "results"

modules:
  - name: example-task
    cmds:
      - echo "Organization {{ORG}}"
      - echo "Output directory {{OUTPUT_DIR}}"
```

When executing the workflow, you can provide values for `ORG` and `OUTPUT_DIR` via the command line like this:

```sh
rayder -w path/to/workflow.yaml ORG=custom_org OUTPUT_DIR=custom_results_dir
```

This will override the default values and use the provided values for these variables.




### Example 2: 

Here's an example workflow configuration tailored for reverse whois recon and processing the root domains into subdomains, resolving them and checking which ones are alive:

```yaml
vars:
  ORG: "Acme, Inc"
  OUTPUT_DIR: "results-dir"

parallel: false
modules:
  - name: reverse-whois
    silent: false
    cmds:
      - mkdir -p {{OUTPUT_DIR}}
      - revwhoix -k "{{ORG}}" > {{OUTPUT_DIR}}/root-domains.txt

  - name: finding-subdomains
    cmds:
      - xargs -I {} -a {{OUTPUT_DIR}}/root-domains.txt echo "subfinder -d {} -o {}.out" | quaithe -workers 30 
    silent: false

  - name: cleaning-subdomains
    cmds:
      -  cat *.out > {{OUTPUT_DIR}}/root-subdomains.txt
      -  rm *.out
    silent: true

  - name: resolving-subdomains
    cmds:
      - cat {{OUTPUT_DIR}}/root-subdomains.txt | dnsx -silent -threads 100 -o {{OUTPUT_DIR}}/resolved-subdomains.txt
    silent: false

  - name: checking-alive-subdomains
    cmds:
      - cat {{OUTPUT_DIR}}/resolved-subdomains.txt | httpx -silent -threads 1000 -o {{OUTPUT_DIR}}/alive-subdomains.txt
    silent: false
```


To execute the above workflow, run the following command:

```sh
rayder -w path/to/reverse-whois.yaml ORG="Yelp, Inc" OUTPUT_DIR=results
```

## Parallel Execution

The `parallel` field in the workflow configuration determines whether modules should be executed in parallel or sequentially. Setting `parallel` to `true` allows modules to run concurrently, making it suitable for modules with no dependencies. When set to `false`, modules will execute one after another.

## Workflows

Explore a collection of sample workflows and examples in the [Rayder workflows repository](https://github.com/devanshbatham/rayder-workflows). Stay tuned for more additions!

## Inspiration
Inspiration of this project comes from Awesome [taskfile](https://taskfile.dev/) project. 
