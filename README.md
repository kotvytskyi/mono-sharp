# Mono Sharp

Mono Sharp is a lightway change-tracking tool for .NET solutions. It matches changed files to a solution projects.

* Goes up though a solution graph to resolve the whole dependency chain 
* Relies on **dotnet CLI**

## How It Works
1. Parses projects from **dotnet CLI** output
2. Finds changed files via **git diff**
3. Finds changes intersection with projects
4. Searches though projects references to find **affected** projects 

## Build

``
go build
``

## Usage

### Examples

<pre>
mono-sharp
mono-sharp --slnDir /some/path
mono-sharp --to HEAD~5
</pre>

### Help
<pre>
Usage of ./mono-sharp:
  -from string
    	'from' git commit (default "HEAD")
  -slnDir string
    	solution file directory (default "./")
  -to string
    	'to' git commit (default "HEAD~1")
</pre>

### Output
Prints all the affected projects separated by new line character