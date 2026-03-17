(define (test-multi x)
  (print "Log: x is" x)
  (+ x 10))

(print (test-multi 5))
; EXPECT ["Log: x is 5", "15"]
