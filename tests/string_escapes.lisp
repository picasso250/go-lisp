(concat "line1\nline2" "\"quote\"") (concat "tab\ttest" "backslash\\test")
; EXPECT ["line1", "line2\"quote\"", "tab\ttestbackslash\\test"]
