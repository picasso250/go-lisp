(print (filter (lambda (x) (> x 1)) '(1 2 3))) (print (reduce (lambda (acc x) (+ acc x)) '(1 2 3) 0))
; EXPECT ["[2 3]", "6"]
