(integer? 1) (float? 1.0) (integer 1.5) (float 2) (float? (float 1))
; EXPECT ["true", "true", "1", "2", "true"]
