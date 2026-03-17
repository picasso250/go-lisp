(print (concat "string-at '中' 0: " (string-at "中" 0)))
(print (concat "string->list '中a': " (str (string->list "中a"))))
(print (concat "string-length '中国': " (str (string-length "中国"))))

; EXPECT ["string-at '中' 0: 中", "string->list '中a': [中 a]", "string-length '中国': 2"]
