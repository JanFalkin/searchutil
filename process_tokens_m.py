#!/usr/bin/env python

import concurrent.futures
import sys
import subprocess
import json



def qfab_cli(line):
    command = ["qfab_cli", "tools", "decode", line]
    # print("command = {}".format(command))
    try:
        output = subprocess.check_output(command, universal_newlines=True)
        split_out = "".join(output.split("\n")[2:])
        post_split = split_out.split("TOKEN")
        test_json = json.loads(post_split[0])
        # print(test_json)
        prefix_info = post_split[1].split("PREFIX      ")
        # prefix_info.split("")
        # out_file.write(prefix_info[1] + "\n")
        if prefix_info[1].find("asc=state-channel") != -1:
            state_channel_count += 1

        return state_channel_count
    except subprocess.CalledProcessError as e:
        print(f"Error: {e}")

def process_file_alt(filename, nprocessor):
    total_lines = 0
    with open(filename, 'r') as file:
        # Create an iterator for the file
        lines_iter = iter(file)
        
        with concurrent.futures.ThreadPoolExecutor(max_workers=nprocessor) as executor:
            # Submit initial tasks to the thread pool
            futures = {executor.submit(qfab_cli, line.strip()): line for line in lines_iter}
            
            # Process the remaining lines as tasks complete
            while futures:
                # Wait for any task to complete
                completed, _ = concurrent.futures.wait(futures, return_when=concurrent.futures.FIRST_COMPLETED)
                
                # Process the completed tasks
                for future in completed:
                    result = future.result()
                    total_lines += 1
                    print("total={}".format(total_lines))
                    # Handle the result if needed
                    
                    # Get the next line from the iterator
                    try:
                        next_line = next(lines_iter)
                        # Submit the next task to the thread pool
                        futures[executor.submit(qfab_cli, next_line.strip())] = next_line
                    except StopIteration:
                        # No more lines in the file, remove the future
                        del futures[future]

def process_file(filename, nprocessor):
    total_count = 0
    processed = 0
    with open(filename, 'r') as file:
        with concurrent.futures.ThreadPoolExecutor(max_workers=nprocessor) as executor:
            # Submit tasks to the thread pool
            futures = [executor.submit(qfab_cli, line.strip()) for line in file]

            # Wait for all tasks to complete
            for future in concurrent.futures.as_completed(futures):
                total_count += 1
                result = future.result()
                print("total={}".format(total_count), end="\r")
                # Handle the result if needed


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: ./script.py <input_file> <num_tokens>")
        sys.exit(1)

    input_file = sys.argv[1]
    num_tokens = int(sys.argv[2])
    process_file_alt(input_file, num_tokens)
