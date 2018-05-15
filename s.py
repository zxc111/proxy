# coding: utf8
import socket
import logging
import threading


def auth(conn):
    data = conn.recv(256)
    conn.sendall(b"\x05\x00")

def connect(conn):
    # import ipdb; ipdb.set_trace()
    data = conn.recv(2048)
    if data[3] == 1:
        host = "%s.%s.%s.%s" % (data[4], data[5], data[6], data[7])
    else:
        host = data[3:-2]

    port = data[-1] + data[-2]*256

    print(host,port)
    # import ipdb; ipdb.set_trace()
    conn.sendall(b"\x05\x00\x00\x01"+data[3:])
    return host, port

def build(host, port):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.connect((host, port))
    return s

def read(conn, dest):
    while True:
            data = conn.recv(1024*256)
            if not data:
                break
            print(data)
            dest.sendall(data)


if __name__ == "__main__":
    HOST = ''                 # Symbolic name meaning all available interfaces
    PORT = 10801              # Arbitrary non-privileged port
    lock = threading.Lock()
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        s.bind((HOST, PORT))
        s.listen(1)

        conn, addr = s.accept()
        auth(conn)
        host, port = connect(conn)
        dest = build(host, port)
        threading.Thread(target=read, args=(conn, dest)).start()
        threading.Thread(target=read, args=(dest, conn)).start()

    except Exception as e:
        logging.warning(e)
        conn.close()
        s.close()
