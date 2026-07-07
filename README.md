# Pips Solver backend

This package implements the core logic for solving the NYT Pips puzzle. 

It implements:
1. An API to get the Pips puzzle for a given date
2. An Integer Linear Programming modeler for solving the board

## Running

This has only currently been tested on MacOS (but is a work in progress)

1. Run the shell script at the package root.
2. Add your Apify API token to the ```.env``` in the backend root
3. Run ```cd board```
4. Run ```go run main.go```

