import socket

s = socket.socket()
localhost = "127.0.0.1"
port = 1234
s.connect((localhost, port))

# read and write
def readWrite():
    while True:
        s.send(b'HI THERE\n')
        print(s.recv(1024).decode())

readWrite()