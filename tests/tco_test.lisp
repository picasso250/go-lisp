(define count-down (lambda (n)
  (if (= n 0)
      "done"
      (count-down (- n 1)))))

(print (count-down 10000))

; EXPECT ["done"]
