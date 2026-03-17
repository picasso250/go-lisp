(concat "hello" " world") (concat "sum: " (+ 1 2)) (string-split "a,b,c" ",") (string-trim "  hello  ") (string-length "hello") (string-contains? "hello" "ell") (string-replace "abcabc" "a" "x")
; EXPECT ["hello world", "sum: 3", "[a b c]", "hello", "5", "true", "xbcxbc"]
