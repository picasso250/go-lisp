def (make-adder x)
  def (adder y)
    (+ x y)
  end
  adder
end

def add5 (make-adder 5) end
def add10 (make-adder 10) end

(print (add5 3))
(print (add10 3))

; EXPECT ["8", "13"]
