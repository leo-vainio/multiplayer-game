import argparse
import socket

parser = argparse.ArgumentParser(description='http server address')
parser.add_argument('--addr', default='127.0.0.1', help='Usage: --addr <ip>')
parser.add_argument('--port', type=int,  default=8080, help='Usage: --port <port>')
args = parser.parse_args()

s = socket.socket()
s.connect((args.addr, args.port))

# read and write
def readWrite():
    while True:
        s.send(b'HI THERE\n')
        print(s.recv(1024).decode())

readWrite()