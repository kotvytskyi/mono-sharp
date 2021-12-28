# Mono Sharp

Mono Sharp is a lightweight change-tracking tool for .NET solutions. It matches changed files to a solution projects.

* Goes up though a solution graph to resolve the whole dependency chain 
* No **dotnet CLI** dependency

## How It Works
1. Parses solution and projects files
2. Finds changed files via **git diff**
3. Finds changes intersection with projects
4. Searches though projects references to find **affected** projects 

### Example

File system
<pre>
-- eShop
   |-- Basket
     |-- FileA.cs
     |-- FileB.cs
     |-- Basket.csproj
   |-- Catalog
     |-- FileC.cs
     |-- FileD.cs - <ins>changed at HEAD~1</ins>
     |-- Catalog.csproj
   |-- Order
     |-- FileE.cs
     |-- FileF.cs
     |-- Order.csproj
   |-- User
     |-- FileG.cs
     |-- FileH.cs
     |-- User.csproj
</pre>

Dependencies
<pre>
Basket -> Catalog
Order -> Catalog
Order -> User
</pre>

Output
**monosharp --slnDir ./eShop**
<pre>
eShop/Catalog/Catalog.csproj
eShop/Basket/Basket.csproj
eShop/Order/Order.csproj
</pre>

## Build

``
go build
``

## Usage
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
