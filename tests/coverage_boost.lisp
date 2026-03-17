(print "Invalid escape: \z")
(print 123)
(print (type undefined_symbol))
(print (if false 1))
(print (and true true))
(print (or false false))
; EXPECT ["Invalid escape: \\z", "123", "<nil>", "<nil>", "true", "false"]
