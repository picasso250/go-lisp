import sys
import os
import re

def parse_coverage(file_path):
    if not os.path.exists(file_path):
        print(f"Error: {file_path} not found.")
        return

    with open(file_path, "r") as f:
        lines = f.readlines()

    if not lines or not lines[0].startswith("mode:"):
        print("Error: Invalid coverage file.")
        return

    # Skip the 'mode: ...' line
    data_lines = lines[1:]

    files = {}
    
    for line in data_lines:
        # Format: name.go:line.char,line.char num_stmt count
        parts = line.split()
        if len(parts) != 3:
            continue
            
        location = parts[0]
        num_stmt = int(parts[1])
        count = int(parts[2])
        
        # Example location: go-lisp/builtin.go:14.50,15.50
        file_match = re.match(r"([^:]+):(\d+)\.(\d+),(\d+)\.(\d+)", location)
        if not file_match:
            continue
            
        full_path, start_line, start_col, end_line, end_col = file_match.groups()
        start_line, start_col, end_line, end_col = map(int, [start_line, start_col, end_line, end_col])
        
        # Strip module name if present (assuming go-lisp/...)
        file_name = full_path
        if "/" in full_path:
            # Try to find the file locally
            parts = full_path.split("/")
            for i in range(len(parts)):
                candidate = "/".join(parts[i:])
                if os.path.exists(candidate):
                    file_name = candidate
                    break

        if file_name not in files:
            files[file_name] = {"total_stmts": 0, "covered_stmts": 0, "missed_blocks": []}
            
        files[file_name]["total_stmts"] += num_stmt
        if count > 0:
            files[file_name]["covered_stmts"] += num_stmt
        else:
            files[file_name]["missed_blocks"].append({
                "start_line": start_line,
                "start_col": start_col,
                "end_line": end_line,
                "end_col": end_col
            })

    # Generate Markdown
    md = "# Coverage Report\n\n"
    md += "## Summary\n\n"
    md += "| File | Coverage | Stmts | Covered | Missed |\n"
    md += "| :--- | :--- | :--- | :--- | :--- |\n"
    
    total_all_stmts = 0
    total_all_covered = 0
    
    file_list = sorted(files.items())
    for file_name, stats in file_list:
        total = stats["total_stmts"]
        covered = stats["covered_stmts"]
        missed = total - covered
        percentage = (covered / total * 100) if total > 0 else 0
        
        md += f"| {file_name} | {percentage:.1f}% | {total} | {covered} | {missed} |\n"
        
        total_all_stmts += total
        total_all_covered += covered
        
    total_percentage = (total_all_covered / total_all_stmts * 100) if total_all_stmts > 0 else 0
    total_missed = total_all_stmts - total_all_covered
    
    md += f"| **Total** | **{total_percentage:.1f}%** | **{total_all_stmts}** | **{total_all_covered}** | **{total_missed}** |\n\n"

    md += "## Missed Details\n\n"
    
    for file_name, stats in file_list:
        if not stats["missed_blocks"]:
            continue
            
        md += f"### {file_name}\n\n"
        
        try:
            with open(file_name, "r", encoding="utf-8") as f:
                source_lines = f.readlines()
        except Exception as e:
            md += f"Error reading file {file_name}: {e}\n\n"
            continue

        md += "| Lines | Uncovered Code |\n"
        md += "| :--- | :--- |\n"
        
        for block in stats["missed_blocks"]:
            s_line = block["start_line"]
            e_line = block["end_line"]
            
            # Extract code snippet
            code_snippet = ""
            if s_line == e_line:
                line_text = source_lines[s_line-1]
                # Note: cols are 1-based but slice is 0-based
                code_snippet = line_text[block["start_col"]-1 : block["end_col"]-1].strip()
                line_range = f"L{s_line}"
            else:
                first_line = source_lines[s_line-1][block["start_col"]-1 :].strip()
                last_line = source_lines[e_line-1][: block["end_col"]-1].strip()
                
                mid_lines = []
                for i in range(s_line, e_line - 1):
                    mid_lines.append(source_lines[i].strip())
                
                parts = [first_line] + mid_lines + [last_line]
                code_snippet = " ".join([p for p in parts if p])
                line_range = f"L{s_line}-L{e_line}"
            
            # Escape pipe for markdown table
            code_snippet = code_snippet.replace("|", "\\|")
            md += f"| {line_range} | `{code_snippet}` |\n"
        
        md += "\n"
    
    return md

if __name__ == "__main__":
    input_file = "coverage.out"
    output_file = "coverage_report.md"
    
    report = parse_coverage(input_file)
    if report:
        with open(output_file, "w", encoding="utf-8") as f:
            f.write(report)
        print(f"Report generated: {output_file}")
