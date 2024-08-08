import re
import matplotlib.pyplot as plt

def parse_benchmark(lines):
    data = {}
    for line in lines:
        match = re.match(r'Benchmark([A-Za-z]+)(/keys_(\d+)-\d+)\s+(\d+)\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op', line)
        if match:
            test_name = match.group(1)
            keys = int(match.group(3))
            time = float(match.group(5)) / 1e6  # Convert nanoseconds to milliseconds
            memory = int(match.group(6))
            allocs = int(match.group(7))
            
            if test_name not in data:
                data[test_name] = {'keys': [], 'time': [], 'memory': [], 'allocs': []}
            
            data[test_name]['keys'].append(keys)
            data[test_name]['time'].append(time)
            data[test_name]['memory'].append(memory)
            data[test_name]['allocs'].append(allocs)
    
    return data

def plot_results(data, metric, filename_suffix=''):
    plt.figure(figsize=(15, 8))
    for test_name, results in data.items():
        if results['keys'] and results[metric]:
            plt.plot(results['keys'], results[metric], marker='o', label=test_name)

    plt.xscale('log')
    plt.yscale('log')
    plt.xlabel('Number of Keys', fontsize=12)
    
    if metric == 'time':
        plt.ylabel('Execution Time (milliseconds)', fontsize=12)
        plt.title('Query Execution Time vs Number of Keys', fontsize=14)
    elif metric == 'memory':
        plt.ylabel('Memory Allocation (bytes)', fontsize=12)
        plt.title('Memory Allocation vs Number of Keys', fontsize=14)
    else:  # metric == 'allocs'
        plt.ylabel('Number of Allocations', fontsize=12)
        plt.title('Number of Allocations vs Number of Keys', fontsize=14)
    
    plt.legend(bbox_to_anchor=(1.05, 1), loc='upper left', fontsize=10)
    plt.grid(True, which="both", ls="-", alpha=0.2)
    
    plt.gca().yaxis.set_major_formatter(plt.FuncFormatter(lambda x, p: f'{x:.2f}' if metric == 'time' else format(int(x), ',')))
    
    plt.xticks(rotation=45, ha='right')
    
    plt.tight_layout()
    plt.savefig(f'benchmark_{metric}{filename_suffix}.png', dpi=300, bbox_inches='tight')
    plt.close()

def group_tests(data):
    groups = {
        'OrderBy': [],
        'Where': [],
        'Limit': [],
        'Other': []
    }
    
    for test_name in data.keys():
        if 'OrderBy' in test_name:
            groups['OrderBy'].append(test_name)
        elif 'Where' in test_name:
            groups['Where'].append(test_name)
        elif 'Limit' in test_name:
            groups['Limit'].append(test_name)
        else:
            groups['Other'].append(test_name)
    
    return groups

# Read benchmark data
with open('benchmark_results.txt', 'r') as f:
    lines = f.readlines()

# Parse benchmark data
data = parse_benchmark(lines)

# Check if any data was parsed
if not data:
    print("No data was parsed from the file. Please check the file contents and format.")
else:
    # Generate overall plots
    plot_results(data, 'time', '_overall')
    plot_results(data, 'memory', '_overall')
    plot_results(data, 'allocs', '_overall')
    print("Overall plots have been generated: benchmark_time_overall.png, benchmark_memory_overall.png, and benchmark_allocs_overall.png")

    # Generate grouped plots
    groups = group_tests(data)
    for group_name, test_names in groups.items():
        group_data = {name: data[name] for name in test_names}
        plot_results(group_data, 'time', f'_{group_name}')
        plot_results(group_data, 'memory', f'_{group_name}')
        plot_results(group_data, 'allocs', f'_{group_name}')
        print(f"Plots for {group_name} have been generated: benchmark_time_{group_name}.png, benchmark_memory_{group_name}.png, and benchmark_allocs_{group_name}.png")
