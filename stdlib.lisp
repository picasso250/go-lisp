(define not (lambda (x) (if x false true)))
(define list-empty? (lambda (l) (null? l)))
(define foldl reduce)
(define even? (lambda (x) (= 0 (% x 2))))
