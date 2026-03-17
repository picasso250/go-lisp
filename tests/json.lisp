(define data (dict "name" "Gemini" "age" 1 "features" (list "json" "tcp" "dict") "alive" true))
(define json-str (json-stringify data))
(print json-str)

(define parsed (json-parse json-str))
(print (dict-get parsed "name"))
(print (nth (dict-get parsed "features") 1))
(print (dict-get parsed "alive"))

(define complex-json "{\"a\": [1, {\"b\": 2}], \"c\": null}")
(define complex-parsed (json-parse complex-json))
(print (dict-get (nth (dict-get complex-parsed "a") 1) "b"))
(print (dict-get complex-parsed "c"))
; EXPECT ["{\"age\": 1, \"alive\": true, \"features\": [\"json\", \"tcp\", \"dict\"], \"name\": \"Gemini\"}", "Gemini", "tcp", "true", "2", "<nil>"]
