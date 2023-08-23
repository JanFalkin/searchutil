#!/usr/bin/env python

import sys
import subprocess
import json
import string


def run_script(input, output):
    command = []
    try:
        output = subprocess.Popen(["python", "process_tokens.py", input, "-1", output], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)#subprocess.check_output(command, universal_newlines=True)
    except subprocess.CalledProcessError as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    base_letter = sys.argv[1]
    for letter in string.ascii_lowercase:
        input_file = "tok-file"+base_letter+letter+".json"
        output_file = "output"+base_letter+letter+".txt"
        run_script(input_file, output_file)
