(define not (lambda (x) (if x false true)))

(define sum (lambda (l) (reduce + l 0)))

(define all (lambda (l) 
  (reduce (lambda (acc x) (and acc (bool x))) l true)))

(define any (lambda (l) 
  (reduce (lambda (acc x) (or acc (bool x))) l false)))

(define even? (lambda (x) (= 0 (% x 2))))
(define odd? (lambda (x) (not (even? x))))

(define zip (lambda (l1 l2)
  (if (or (null? l1) (null? l2))
      '()
      (cons (list (car l1) (car l2)) (zip (cdr l1) (cdr l2))))))

(define enumerate (lambda (l)
  (let ((idx-range (range (length l))))
    (zip idx-range l))))

(define foldl reduce)
