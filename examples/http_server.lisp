(define server (tcp-listen "127.0.0.1:8080"))
(print "HTTP Server starting on http://127.0.0.1:8080")

(define handle-request (lambda (conn)
  (let ((req (tcp-read conn 1024)))
    (begin
      (print "Received request:")
      (print (car (string-split req "\n")))
      (tcp-send conn "HTTP/1.1 200 OK\r\n")
      (tcp-send conn "Content-Type: text/html\r\n")
      (tcp-send conn "Connection: close\r\n")
      (tcp-send conn "\r\n")
      (tcp-send conn "<h1>Hello from Lisp HTTP Server!</h1>")
      (tcp-send conn (concat "<p>Timestamp: " (str 12345678) "</p>"))
      (tcp-close conn)))))

(define loop (lambda ()
  (begin
    (print "Waiting for connection...")
    (let ((conn (tcp-accept server)))
      (begin
        (handle-request conn)
        (loop))))))

(loop)
