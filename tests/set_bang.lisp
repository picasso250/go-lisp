(define x 10)
(print x)
(set! x 20)
(print x)

(define f (lambda (y) (set! x y)))
(f 30)
(print x)

(let ((z 5))
  (begin
    (set! z 100)
    (print z)))
; EXPECT ["10", "20", "30", "100"]
