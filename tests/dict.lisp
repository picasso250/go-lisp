(define d (dict "a" 1 "b" 2))
(print (dict-get d "a"))
(print (dict-get d "b"))
(print (dict-set! d "c" 3))
(print (dict-get d "c"))
(print (dict-has? d "a"))
(print (dict-has? d "z"))
(print (length (dict-keys d)))
; EXPECT ["1", "2", "3", "3", "true", "false", "3"]
