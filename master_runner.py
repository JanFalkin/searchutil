#!/usr/bin/env python

import sys
import subprocess
import json
import string


def run_script(input, output):
    try:
        process = subprocess.Popen(["mp/process-multi", input, "-1", output],
                                    stdout=subprocess.PIPE,
                                    stderr=subprocess.STDOUT,
                                    universal_newlines=True)
        for line in process.stdout:
            line = line.rstrip('\n')
            sys.stdout.write('\r' + line)
            sys.stdout.flush()
        process.wait()
    except subprocess.CalledProcessError as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    base_letter = sys.argv[1]
    base_data = sys.argv[2]
    base_output = sys.argv[3]

    for letter in string.ascii_lowercase:
        input_file = base_data+"/tok-file"+base_letter+letter+".json"
        output_file = base_output+"/output"+base_letter+letter+".txt"
        run_script(input_file, output_file)