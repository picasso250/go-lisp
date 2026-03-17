(begin
  (define a 10)
  (define b 20)
  (define cmp-test (list (< a b) (<= a a) (> b a) (>= b b) (= a a) (!= a b)))

  (define my-list '(1 2 3))
  (define q-test (quote hello))
  (define car-test (car my-list))
  (define cdr-test (cdr my-list))
  (define cons-test (cons 0 my-list))
  (define len-test (length cons-test))
  (define null-test (null? (quote ())))

  (define let-test 
    (let ((x 100) (y 200))
      (+ x y)))

  (list cmp-test q-test car-test cdr-test cons-test len-test null-test let-test)
)
