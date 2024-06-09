import socket
import logging
from datetime import datetime

logging.basicConfig(level=logging.INFO)


def new_socket(port: int) -> socket.socket:
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
    s.bind(("", port))
    return s


port = 8080
timeFmt = "%H:%M:%S"
s = new_socket(port)

logging.info("new client created at " + datetime.now().strftime(timeFmt))

while True:
    time = s.recvfrom(1024)[0].decode()
    logging.info("received " + time)
