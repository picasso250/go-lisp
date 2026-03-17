(define not (lambda (x) (if x false true)))

(define sum (lambda (l) (reduce + l 0)))

(define all (lambda (l)
  (reduce (lambda (acc x) (and acc (bool x))) l true)))

(define any (lambda (l)
  (reduce (lambda (acc x) (or acc (bool x))) l false)))

(define even? (lambda (x) (= 0 (% x 2))))
(define odd? (lambda (x) (not (even? x))))

(define zip (lambda (l1 l2)
  (if (or (null? l1) (null? l2))
      '()
      (cons (list (car l1) (car l2)) (zip (cdr l1) (cdr l2))))))

(define enumerate (lambda (l)
  (let ((idx-range (range (length l))))
    (zip idx-range l))))

(define foldl reduce)

(define string-join (lambda (l sep)
  (if (null? l) ""
      (if (= 1 (length l)) 
          (str (car l))
          (concat (str (car l)) sep (string-join (cdr l) sep))))))

(define json-stringify (lambda (x)
  (if (string? x) 
      (concat "\"" (string-replace (string-replace x "\\" "\\\\") "\"" "\\\"") "\"")
      (if (integer? x) 
          (str x)
          (if (float? x) 
              (str x)
              (if (bool? x) 
                  (if x "true" "false")
                  (if (nil? x) 
                      "null"
                      (if (list? x) 
                          (concat "[" (string-join (map json-stringify x) ", ") "]")
                          (if (dict? x) 
                              (concat "{" (string-join (map (lambda (k) (concat "\"" k "\": " (json-stringify (dict-get x k)))) (sorted (dict-keys x))) ", ") "}")
                              "null")))))))))

(define json--skip-ws (lambda (chars)
  (if (null? chars) 
      '()
      (let ((c (car chars)))
        (if (or (= c " ") (or (= c "\n") (or (= c "\r") (= c "\t"))))
            (json--skip-ws (cdr chars))
            chars)))))

(define json--parse-str-extract (lambda (chars acc)
  (if (= (car chars) "\"") 
      (list acc (cdr chars))
      (if (= (car chars) "\\") 
          (json--parse-str-extract (cdr (cdr chars)) (concat acc (car (cdr chars))))
          (json--parse-str-extract (cdr chars) (concat acc (car chars)))))))

(define json--parse-str (lambda (chars)
  (json--parse-str-extract chars "")))

(define json--parse-num-extract (lambda (chars acc)
  (if (or (null? chars) (not (string-contains? "0123456789.eE-" (car chars))))
      (list (float acc) chars)
      (json--parse-num-extract (cdr chars) (concat acc (car chars))))))

(define json--parse-num (lambda (chars)
  (json--parse-num-extract chars "")))

(define json--parse-val (lambda (chars)
  (let ((chars (json--skip-ws chars)))
    (if (null? chars) 
        '()
        (let ((c (car chars)))
          (if (= c "{") (json--parse-obj (cdr chars))
          (if (= c "[") (json--parse-arr (cdr chars))
          (if (= c "\"") (json--parse-str (cdr chars))
          (if (string-contains? "0123456789-" c) (json--parse-num chars)
          (if (= c "t") (list true (cdr (cdr (cdr (cdr chars)))))
          (if (= c "f") (list false (cdr (cdr (cdr (cdr (cdr chars))))))
          (if (= c "n") (list nil (cdr (cdr (cdr (cdr chars)))))
          '()))))))))))))

(define json--parse-arr-extract (lambda (chars acc)
  (let ((chars (json--skip-ws chars)))
    (if (= (car chars) "]") 
        (list acc (cdr chars))
        (if (= (car chars) ",") 
            (json--parse-arr-extract (cdr chars) acc)
            (let ((res (json--parse-val chars)))
              (json--parse-arr-extract (car (cdr res)) (append acc (list (car res))))))))))

(define json--parse-arr (lambda (chars)
  (json--parse-arr-extract chars '())))

(define json--parse-obj-extract (lambda (chars acc)
  (let ((chars (json--skip-ws chars)))
    (if (= (car chars) "}") 
        (list acc (cdr chars))
        (if (= (car chars) ",") 
            (json--parse-obj-extract (cdr chars) acc)
            (let ((key-res (json--parse-val chars)))
              (let ((val-start (json--skip-ws (cdr (json--skip-ws (car (cdr key-res)))))))
                (let ((val-res (json--parse-val val-start)))
                  (begin
                    (dict-set! acc (car key-res) (car val-res))
                    (json--parse-obj-extract (car (cdr val-res)) acc))))))))))

(define json--parse-obj (lambda (chars)
  (json--parse-obj-extract chars (dict))))

(define json-parse (lambda (s) (car (json--parse-val (string->list s)))))
