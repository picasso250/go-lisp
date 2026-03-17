import re
import sys

def extract_cov0(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Error: {file_path} not found. Run 'go test -coverprofile=coverage.out' and 'go tool cover -html=coverage.out -o coverage.html' first.")
        return
    
    matches = re.findall(r'<span class="cov0"[^>]*>(.*?)</span>', content, re.DOTALL)
    
    print(f"--- Uncovered Code (cov0) in {file_path} ---")
    if not matches:
        print("Perfect! No uncovered code found.")
        return

    for i, match in enumerate(matches, 1):
        clean_text = match.replace('&quot;', '"').replace('&lt;', '<').replace('&gt;', '>').replace('&amp;', '&')
        print(f"[{i}]:\n{clean_text.strip()}\n")

if __name__ == "__main__":
    path = sys.argv[1] if len(sys.argv) > 1 else 'coverage.html'
    extract_cov0(path)
