#!/usr/bin/env python

import sys
import subprocess
import json

def count_lines(filename):
    line_count = 0
    with open(filename, 'r') as file:
        for _ in file:
            line_count += 1
    return line_count


def process_file(file_path, num_tokens, output_file):
    i = 0
    line_count = count_lines(file_path)
    if num_tokens < 0:
        num_tokens = line_count
    with open(file_path, "r") as file:
        with open(output_file, "w") as out_file:
            state_channel_count = 0 
            total_count = 0
            for line in file:
                if i == num_tokens:
                    break
                tokens = line.strip()
                # print("tokens = {}".format(tokens))
                print(f"{i} out of {num_tokens} stat: state channel = {state_channel_count} total = {total_count}", end="\r")
                sc, t = run_qfab_cli(tokens, out_file)
                state_channel_count += sc
                total_count += t
                i = i + 1
    with open(output_file, "a") as finals:
        finals.write(f"state channel = {state_channel_count} \n total = {total_count}")


def run_qfab_cli(formatted_line, out_file):
    command = ["qfab_cli", "tools", "decode", formatted_line]
    state_channel_count = 0 
    total_count = 0
    # print("command = {}".format(command))
    try:
        output = subprocess.check_output(command, universal_newlines=True)
        if output.find("legacy token: ") != -1:
            out_file.write(output + "\n")
            return 0,1
        split_out = "".join(output.split("\n")[2:])
        post_split = split_out.split("TOKEN")
        #test_json = json.loads(post_split[0])
        # print(test_json)
        prefix_info = post_split[1].split("PREFIX      ")
        # prefix_info.split("")
        if len(prefix_info) > 1:
            out_file.write(prefix_info[1] + "\n")
            if prefix_info[1].find("asc=state-channel") != -1:
                state_channel_count += 1
                total_count += 1
            else:
                total_count += 1
        else:
            ps = "".join(post_split)
            print(f"prefix info unexpected = {prefix_info} post_split={ps}")
        return state_channel_count, total_count
    except subprocess.CalledProcessError as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    if len(sys.argv) < 4:
        print("Usage: ./script.py <input_file> <num_tokens> <output_file>")
        sys.exit(1)

    input_file = sys.argv[1]
    num_tokens = int(sys.argv[2])
    output_file = sys.argv[3]
    process_file(input_file, num_tokens, output_file)
