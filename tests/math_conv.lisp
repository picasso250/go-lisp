(abs -50) (pow 2 10) (str 123) (int "456") (integer 789) (bool 0) (bool 1) (bool '()) (divmod 10 3) (round 3.6) (round 3.4)
; EXPECT ["50", "1024", "123", "456", "789", "false", "true", "false", "[3 1]", "4", "3"]
