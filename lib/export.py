#!/usr/bin/python
import re
import json

export_go_path = "export.go"  # Path to your export.go file
output_json_path = "export.json"  # Path to the output JSON file

def parse_export_go(file_path):
    functions = []
    current_func = None
    in_function = False
    brace_level = 0
    func_body_lines = []
    var_declarations = {}

    with open(file_path, 'r') as file:
        lines = file.readlines()

    func_pattern = re.compile(r'//export (\w+)\s*\nfunc (\w+)\(([^)]*)\) C\.Result', re.MULTILINE)
    cinput_pattern = re.compile(r'cInput\([^,]+,\s*([^,]+),\s*&([^,]+)\)', re.MULTILINE)
    var_pattern = re.compile(r'var\s+(\w+)\s+(\S+)', re.MULTILINE)

    matches = list(func_pattern.finditer('\n'.join(lines)))
    for i, func_match in enumerate(matches):
        if current_func:
            functions.append(current_func)
        
        export_name = func_match.group(1)
        func_name = func_match.group(2)
        params_str = func_match.group(3)
        
        print(f"Parsing function: {func_name} with params: {params_str}")

        params = []
        param_parts = [p.strip() for p in params_str.split(',')]
        last_type = None
        for part in param_parts:
            parts = part.split()
            if len(parts) == 2:
                param_name, param_type = parts
                last_type = param_type
            else:
                param_name = parts[0]
                param_type = last_type
            params.append({
                "name": param_name,
                "type": param_type,
                "binding": None
            })
        current_func = {
            "name": func_name,
            "parameters": params
        }
        
        func_body_start = func_match.end()
        func_body_end = matches[i+1].start() if i+1 < len(matches) else len('\n'.join(lines))
        func_body = '\n'.join(lines)[func_body_start:func_body_end]

        brace_level = 0
        in_function = False
        func_body_lines = []
        for line in func_body.splitlines():
            brace_level += line.count('{') - line.count('}')
            if not in_function and brace_level > 0:
                in_function = True
            if in_function and brace_level == 0:
                break
            func_body_lines.append(line)

        func_body_full = '\n'.join(func_body_lines)

        for var_match in var_pattern.finditer(func_body_full):
            var_name = var_match.group(1).strip()
            var_type = var_match.group(2).strip()
            var_declarations[var_name] = var_type
            print(f"Found var declaration: {var_name} -> {var_type}")

        for cinput_match in cinput_pattern.finditer(func_body_full):
            param_name = cinput_match.group(1).strip()
            var_name = cinput_match.group(2).strip()
            go_struct = var_declarations.get(var_name)
            print(f"Found cInput call: {param_name} -> {go_struct}")
            for param in current_func["parameters"]:
                if param["name"] == param_name:
                    param["binding"] = go_struct

    if current_func:
        functions.append(current_func)
        print(f"Added function: {current_func['name']}")

    return functions

def generate_json(functions, output_file):
    output = {
        "export": functions,
        "types": {}
    }
    with open(output_file, 'w') as file:
        json.dump(output, file, indent=4)

if __name__ == "__main__":
    functions = parse_export_go(export_go_path)
    print(f"Functions extracted: {json.dumps(functions, indent=4)}")
    generate_json(functions, output_json_path)
    print(f"JSON file generated: {output_json_path}")
