# Generic Go Benchmark and Plotting

The scripts in this directory can be used to run benchmarks on the executor and generate plots based on the benchmark results.

## Prerequisites

- Go (for running benchmarks)
- Python 3 (for generating plots)
- matplotlib (Python library for plotting)

## Running Benchmarks

1. Navigate to the directory containing the Go files you want to benchmark:

   ```bash
   cd core
   ```

2. Run the benchmarks using the `go test` command. You can specify patterns to run specific benchmarks:

   ```bash
   go test -bench=BenchmarkExecuteQuery -benchmem
   ```

## Generating Plots

1. Place your `benchmark_results.txt` file in the same directory as your `plot_benchmarks.py` script.

2. Install the required Python library:

   ```bash
   # Create a virtual environment
   python3 -m venv venv

    # Activate the virtual environment
   source venv/bin/activat

    # Install matplotlib
   pip install matplotlib
   ```

3. Run the Python script to generate plots:

   ```bash
   python plot_benchmarks.py
   ```

4. The script will read the `benchmark_results.txt` file and generate plots based on the benchmark results.
