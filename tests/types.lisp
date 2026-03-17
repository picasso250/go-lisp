(integer? 1) (float? 1.0) (string? "hi") (integer 1.5) (float 2) (float? (float 1)) (type 1) (type 1.5) (type "hi")
; EXPECT ["true", "true", "true", "1", "2", "true", "*big.Int", "float64", "string"]
