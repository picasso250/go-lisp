def (add x y)
  (+ x y)
end

(print (add 10 20))

def (square x) (* x x) end
(print (square 5))

def (multi-step x)
  (print "calculating square of" x)
  (* x x)
end
(print (multi-step 4))

; EXPECT ["30", "25", "calculating square of 4", "16"]
