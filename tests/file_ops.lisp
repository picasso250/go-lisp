(write-file "tmp.txt" "hello file") (read-file "tmp.txt")
; EXPECT ["true", "hello file"]
