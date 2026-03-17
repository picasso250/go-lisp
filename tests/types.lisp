(integer? 1) (float? 1.0) (string? "hi") (integer 1.5) (float 2) (float? (float 1)) (type 1) (type 1.5) (type "hi") (float 1.5)
; EXPECT ["true", "true", "true", "1", "2", "true", "int64", "float64", "string", "1.5"]
